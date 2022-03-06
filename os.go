package main

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/h2non/filetype"
	"github.com/vinyl-linux/vin/config"
	"github.com/zeebo/blake3"
)

var (
	pkgDir     = getEnv("VIN_PATH", "/etc/vinyl/pkg")
	configFile = getEnv("VIN_CONFIG", "/etc/vinyl/vin.toml")
	cacheDir   = getEnv("VIN_CACHE", "/var/cache/vinyl/vin/packages")
	sockAddr   = getEnv("VIN_SOCKET_ADDR", "/var/run/vin.sock")
	stateDB    = getEnv("VIN_STATE_DB", "/etc/vinyl/vin.db")
	svcDir     = getEnv("VINIT_SVC_DIR", "/etc/vinit/services")
)

// ChanWriter wraps a string channel, and implements the io.Writer
// interface in order to attach it to the Stdout and Stderr of a process
type ChanWriter chan string

// Write implements the io.Writer interface; it writes p to the underlying
// channel as a string
func (cw ChanWriter) Write(p []byte) (n int, err error) {
	cw <- string(p)

	return len(p), nil
}

func getEnv(key, def string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		v = def
	}

	return v
}

// checksum takes a filename, opens it, and generates a Blake3 sum from it
//
// Why blake3? Because it's fast, (supposedly) collision free, and is used
// little enough that package publishers will be forced to generate checksums
// solely for vin, thus ensuring the checksums are correct.
//
// This is as opposed to copy/pasting any old checksum from any old source.
func checksum(fn string) (sum string, err error) {
	f, err := os.Open(fn)
	if err != nil {
		return
	}

	defer f.Close()

	h := blake3.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return
	}

	outB := make([]byte, 0)
	sumB := h.Sum(outB)

	sum = hex.EncodeToString(sumB)

	return
}

// download a file, saving to fn
//
// this solution is, in part, taken from https://stackoverflow.com/a/33853856
func download(fn, url string) (err error) {
	out, err := os.Create(fn)
	if err != nil {
		return
	}

	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading %s: expected status 200, received %s", url, resp.Status)
	}

	_, err = io.Copy(out, resp.Body)

	return
}

// decompress opens a tarball and decompresses it according to
// which ever method was used to compress it
func decompress(data *os.File) (decompressor io.Reader, err error) {
	// Read header, determine compression method
	head := make([]byte, 261)
	data.Read(head)

	ty, err := filetype.Match(head)
	if err != nil {
		return
	}

	// Go back to the start
	data.Seek(0, os.SEEK_SET)

	switch ty.MIME.Value {
	case "application/gzip":
		decompressor, err = gzip.NewReader(data)
	case "application/x-bzip2":
		decompressor = bzip2.NewReader(data)
	default:
		err = fmt.Errorf("untar: tarball is of (unsupported) type %q", ty.MIME.Value)
	}

	return
}

// untar will untar a tarball
//
// this solution is, in part, taken from https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07
func untar(fn, dir string) (err error) {
	// detect filetype
	f, err := os.Open(fn)
	if err != nil {
		return
	}

	defer f.Close()

	decompressor, err := decompress(f)
	if err != nil {
		return
	}

	switch decompressor.(type) {
	case *gzip.Reader:
		defer decompressor.(*gzip.Reader).Close()
	}

	tr := tar.NewReader(decompressor)

	return decompressLoop(tr, dir)
}

// decompressLoop iterates over a tar reader until either the tarball is
// untarred, or an error occurs
func decompressLoop(tr *tar.Reader, dest string) (err error) {
	var header *tar.Header
	for {
		header, err = tr.Next()

		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dest, header.Name)

		// ensure the parent directory exists in situations where a single file,
		// with missing toplevel dir, is compressed.
		//
		// This seemingly weird case exists in some tarballs in our test cases, so
		// it can definitely happen. I'd love to understand more /shrug
		err = os.MkdirAll(filepath.Dir(target), 0755)
		if err != nil {
			return
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			f.Close()

		case tar.TypeLink:
			// remove target if it exists.
			//
			// ignoring the error is fine here; if there's an error
			// we'll see it when we try to link anyway /shrug
			os.Remove(target)

			err = os.Link(filepath.Join(dest, header.Linkname), target)
			if err != nil {
				return
			}

		case tar.TypeSymlink:
			// see comment for the tar.TypeLink case above;
			os.Remove(target)

			// we don't need to worry about prefixing synlinks. Infact,
			// we probably don't want that at all. symlinks can point to
			// files that don't exist, and probably want to be more flexible
			// for things like relative links anyway
			err = os.Symlink(header.Linkname, target)
			if err != nil {
				return
			}

		}
	}
}

func execute(dir, command string, skipEnv bool, output chan string, c config.Config) (err error) {
	cmdSlice := strings.Fields(command)

	var args []string

	switch len(cmdSlice) {
	case 0:
		return fmt.Errorf("execute: %q is empty, or cannot be split", command)
	case 1:
		// NOP; in this case leave args as empty
	default:
		args = cmdSlice[1:]
	}

	cmd := exec.CommandContext(context.Background(), cmdSlice[0], args...)
	cmd.Dir = dir

	if !skipEnv {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, fmt.Sprintf("CFLAGS=%s", c.CFlags), fmt.Sprintf("CXXFLAGS=%s", c.CXXFlags))
	}

	outputWriter := ChanWriter(output)
	cmd.Stdout = outputWriter
	cmd.Stderr = outputWriter

	return cmd.Run()
}

func installServiceDir(src string) (err error) {
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		name := strings.TrimPrefix(path, src)
		dst := filepath.Join(svcDir, filepath.Base(src), name)

		if info.IsDir() {
			return os.MkdirAll(dst, 0700)
		}

		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			l, err := filepath.EvalSymlinks(path)
			if err != nil {
				return err
			}

			// delete destination first; otherwise os.Symlink
			// will error.
			//
			// Ignore the error; the only pertinent error is
			// that the destination does not exist (which we don't
			// care about)- anything else will be caught elsewhere
			os.Remove(dst)

			return os.Symlink(l, dst)
		}

		source, err := os.Open(path)
		if err != nil {
			return err
		}
		defer source.Close()

		destination, err := os.Create(dst)
		if err != nil {
			return err
		}

		defer destination.Close()
		_, err = io.Copy(destination, source)

		return err
	})
}
