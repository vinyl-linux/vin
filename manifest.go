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

type Manifest struct {
	Provides   string
	VersionStr string           `toml:"version"`
	Version    *version.Version `toml:"-"`
	Licence    string
	Tarball    string

	Profiles map[string]Profile
	Commands Commands
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
func Manifests() (m []Manifest, err error) {
	m = make([]Manifest, 0)

	err = filepath.Walk(pkgDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == "manifest.toml" {
			man, err := readManifest(path)
			if err != nil {
				return err
			}

			m = append(m, man)
		}

		return nil
	})

	return
}

// ReadManifest takes a package and version, loads the necessary manifest file,
// Parses, and returns a Manifest for processing
func ReadManifest(pkg, ver string) (m Manifest, err error) {
	filename := manifestPath(pkg, ver)

	return readManifest(filename)
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
