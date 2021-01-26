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

	// DefaultConfigure is the command used to configure packages
	// where a configure command is not provided
	DefaultConfigure = "./configure {{ .ConfigureFlags }}"

	// DefaultCompile is the command used to configure packages
	// where a configure command is not provided
	DefaultCompile = "make {{ .MakeOpts }}"

	// DefaultInstall is the command used to configure packages
	// where a configure command is not provided
	DefaultInstall = "make install {{ .MakeOpts }}"
)

// Dep represents a dependency tuple.
//
// A dependency tuple is characterised as (package, version constraint)
type Dep [2]string

// Valid returns true if a Dep has both fields set
// and if both fields are 'correct'
func (d Dep) Valid() bool           { return d[0] != "" && d[1] != "" && d.validConstraint() }
func (d Dep) validConstraint() bool { _, err := d.Constraint(); return err == nil }

// Package returns the package name for this dependency
func (d Dep) Package() string { return d[0] }

// Constraint returns a version.Constraints type built from
// the constraint set in the tuple
func (d Dep) Constraint() (version.Constraints, error) { return version.NewConstraint(d[1]) }

// InvalidDepError is a wrapper around a Dep purely because adding an Error() func to dep
// seems confusing, and to allow us to check types later
type InvalidDepError Dep

// Error implements the `error` interface and is returned when a Dep
// is invalid in some way
func (d InvalidDepError) Error() string {
	return fmt.Sprintf(`dependency "%s %s" is invalid`, d[0], d[1])
}

// ManifestByVersion implements the sort.Interface interface, in
// order to sort manifests by version
type ManifestByVersion []*Manifest

// Len returns the length of the Manifest set
func (m ManifestByVersion) Len() int { return len(m) }

// Swap will swap the position of two Manifests
func (m ManifestByVersion) Swap(i, j int) { m[i], m[j] = m[j], m[i] }

// Less returns true if m[i] is a lower version that m[j]
func (m ManifestByVersion) Less(i, j int) bool { return m[i].Version.LessThan(m[j].Version) }

// Manifest represents the config needed to install a package, including
// build commands, dependencies, sources, versions, and associated funtions
// that do stuff with those things
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

// String represents the canonical name of a package as provided by a Manifest
func (m Manifest) String() string { return fmt.Sprintf("%s %s", m.Provides, m.VersionStr) }

// Prepare accepts a chan to send messages to (for buffering messages when multithreaded) and
// returns an error.
//
// It handles things like downloading and verifying tarballs, and subsequently untarring
func (m *Manifest) Prepare(output chan string) (err error) {
	// This function will download the Manifest Tarball, checksum it, un-tar it, and so on. At some
	// point we could even think about things like applying optional patches

	m.dir = filepath.Join(cacheDir, m.Provides, m.VersionStr)
	err = os.MkdirAll(m.dir, 0755)
	if err != nil {
		return
	}

	// download m.Tarball to tempdir/.tarball
	fn := filepath.Join(m.dir, ".tarball")
	output <- fmt.Sprintf("%s: downloading %q to %s", m.ID, m.Tarball, fn)

	err = download(fn, m.Tarball)
	if err != nil {
		return
	}

	// generate checksum for tarball
	output <- fmt.Sprintf("%s: comparing blake3 checksums", m.ID)

	sum, err := checksum(fn)
	if err != nil {
		return
	}

	// compare checksum with m.Checksum
	if m.Checksum != sum {
		return fmt.Errorf("checksum error: expected %q, downloaded file was %q", m.Checksum, sum)
	}

	// un-tar tarball
	output <- fmt.Sprintf("%s: extracting sources", m.ID)

	return untar(fn, m.dir)
}

// Profile holds a set of dependencies associated with a 'profile'.
//
// A 'profile' is a way of splitting dependencies into groups, such
// as only including GUI dependencies when building X11 apps, or
// bundling extra packages for larger/ less disk constrained systems
type Profile struct {
	Deps []Dep
}

// Commands provides a set of 'commands' which are used in our three stages:
//
//   1. Configuring packages/ apps/ whosits
//   2. Compiling packages/ binaries
//   3. Installing the resulting compiled stuff into the filesystem
//
// Empty commands receive the default for each item, so use something like `true`
// where a stage command is not necessary
type Commands struct {
	Configure *string
	Compile   *string
	Install   *string
}

// Slice returns each command in an ordered slice
func (c Commands) Slice() []string {
	return []string{
		c.GetConfigure(),
		c.GetCompile(),
		c.GetInstall(),
	}
}

// GetConfigure returns either c.Configure, or the default
// configure command
func (c Commands) GetConfigure() string {
	if c.Configure == nil {
		return DefaultConfigure
	}

	return *c.Configure
}

// GetCompile returns either c.Compile, or the default
// configure command
func (c Commands) GetCompile() string {
	if c.Compile == nil {
		return DefaultCompile
	}

	return *c.Compile
}

// GetInstall returns either c.Install, or the default
// configure command
func (c Commands) GetInstall() string {
	if c.Install == nil {
		return DefaultInstall
	}

	return *c.Install
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
