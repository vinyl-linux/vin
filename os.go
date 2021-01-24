package main

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/h2non/filetype"
	"github.com/zeebo/blake3"
)

var (
	pkgDir     = getEnv("VIN_PATH", "/etc/vinyl/pkg")
	configFile = getEnv("VIN_CONFIG", "/etc/vinyl/vin.toml")
	cacheDir   = getEnv("VIN_CACHE", "/var/cache/vinyl/vin/packages")
)

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

// untar will untar a tarball
//
// this solution is, in part, taken from https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07
func untar(fn, dir string) (err error) {
	// detect filetype
	var compressorReader io.Reader

	f, err := os.Open(fn)
	if err != nil {
		return
	}

	defer f.Close()

	// Read header, determine compression method
	head := make([]byte, 261)
	f.Read(head)

	ty, err := filetype.Match(head)
	if err != nil {
		return
	}

	// Go back to the start
	f.Seek(0, os.SEEK_SET)

	switch ty.MIME.Value {
	case "application/gzip":
		compressorReader, err = gzip.NewReader(f)
		defer compressorReader.(*gzip.Reader).Close()
	case "application/x-bzip2":
		compressorReader = bzip2.NewReader(f)
	default:
		err = fmt.Errorf("untar: tarball is of (unsupported) type %q", ty.MIME.Value)
	}

	if err != nil {
		return
	}

	tr := tar.NewReader(compressorReader)

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
		target := filepath.Join(dir, header.Name)

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
		}
	}
}
