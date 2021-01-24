package main

import (
	_ "log"

	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/pelletier/go-toml"
)

const (
	// ManifestFilename is the file at which manifests live
	ManifestFilename = "manifest.toml"
)

type Dep [2]string

func (d Dep) Valid() bool                              { return d[0] != "" && d[1] != "" && d.validConstraint() }
func (d Dep) validConstraint() bool                    { _, err := d.Constraint(); return err == nil }
func (d Dep) Package() string                          { return d[0] }
func (d Dep) Constraint() (version.Constraints, error) { return version.NewConstraint(d[1]) }

// InvalidDepError is a wrapper around a Dep purely because adding an Error() func to dep
// seems confusing, and to allow us to check types later
type InvalidDepError Dep

func (d InvalidDepError) Error() string {
	return fmt.Sprintf(`dependency "%s %s" is invalid`, d[0], d[1])
}

type ManifestByVersion []*Manifest

func (m ManifestByVersion) Len() int           { return len(m) }
func (m ManifestByVersion) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m ManifestByVersion) Less(i, j int) bool { return m[i].Version.LessThan(m[j].Version) }

type Manifest struct {
	ID string `toml:"-"`

	Provides   string
	VersionStr string           `toml:"version"`
	Version    *version.Version `toml:"-"`
	Checksum   string
	Licence    string
	Tarball    string

	Profiles map[string]Profile
	Commands Commands

	// dir comes after download, and signifies the location a package
	// is extracted to
	dir string
}

func (m Manifest) String() string { return fmt.Sprintf("%s %s", m.Provides, m.VersionStr) }

// Prepare accepts a chan to send messages to (for buffering messages when multithreaded) and
// returns an error.
//
// It handles things like downloading and verifying tarballs, and subsequently untarring
func (m *Manifest) Prepare(messages chan string) (err error) {
	// This function will download the Manifest Tarball, checksum it, un-tar it, and so on. At some
	// point we could even think about things like applying optional patches

	m.dir = filepath.Join(cacheDir, m.Provides, m.VersionStr)
	err = os.MkdirAll(m.dir, 0755)
	if err != nil {
		return
	}

	// download m.Tarball to tempdir/.tarball
	fn := filepath.Join(m.dir, ".tarball")
	err = download(fn, m.Tarball)
	if err != nil {
		return
	}

	// generate checksum for tarball
	sum, err := checksum(fn)
	if err != nil {
		return
	}

	// compare checksum with m.Checksum
	if m.Checksum != sum {
		return fmt.Errorf("checksum error: expected %q, downloaded file was %q", m.Checksum, sum)
	}

	// un-tar tarball
	return untar(fn, m.dir)
}

type Profile struct {
	Deps []Dep
}

type Commands struct {
	Configure string
	Compile   string
	Install   string
}

// Manifests returns a slice of all manifests in the pkgDir
func Manifests() (m []*Manifest, err error) {
	m = make([]*Manifest, 0)

	for _, dir := range filepath.SplitList(pkgDir) {
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.Name() == ManifestFilename {
				man, err := readManifest(path)
				if err != nil {
					return err
				}

				m = append(m, &man)
			}

			return nil
		})
	}

	return
}

func readManifest(filename string) (m Manifest, err error) {
	d, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	err = toml.Unmarshal(d, &m)
	if err != nil {
		return
	}

	m.Version, err = version.NewVersion(m.VersionStr)
	if err != nil {
		return
	}

	m.ID = m.String()

	for _, profile := range m.Profiles {
		for _, d := range profile.Deps {
			if !d.Valid() {
				err = InvalidDepError(d)

				return
			}
		}
	}

	return
}
