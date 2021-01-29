package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
	"time"
)

// StatDB holds state between installations/runs
type StateDB struct {
	// World holds the topmost packages. These packages are only
	// those directly requested by the user- this allows us to
	// handle both upgrades and migrations; upgrades by re-installing
	// world (hello gentoo, I love you gentoo), and migrations by lifting
	// and shifting the database.
	//
	// This data is stored as: {pkg: constraint}
	World map[string]string

	// LastUpdate represents the last time the StateDB was modified
	LastUpdate time.Time

	// Installed is a list of *all* installed packages, including when they
	// were last installed
	Installed map[string]time.Time
}

// LoadStateDB loads state from the filesystem
func LoadStateDB() (s StateDB, err error) {
	f, err := os.Open(stateDB)

	// If state file has not been created yet, create
	if perr, ok := err.(*os.PathError); ok && perr.Err == syscall.ENOENT {
		s = StateDB{
			World:      make(map[string]string),
			LastUpdate: time.Now(),
			Installed:  make(map[string]time.Time),
		}

		err = s.Write()

		return
	}

	// Any other error, bomb out
	if err != nil {
		return
	}

	dec := gob.NewDecoder(f)
	err = dec.Decode(&s)

	return
}

// Write persists the state database to the filesystem
func (s StateDB) Write() (err error) {
	// We write to a buffer, and then write the buffer to disk,
	// purely because I have no idea what risk there is of writing
	// broken data directly to the file. Will gob return an error
	// before the file is truncated? Or will we lose previous state?
	// If we crash in/around a bad write, can we break state?
	//
	// These are things I have no answer to, so we minimise risk
	// (imagined or not) by not touching the underlying file until
	// we have some serialsied data
	var b bytes.Buffer

	enc := gob.NewEncoder(&b)
	err = enc.Encode(s)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(stateDB, b.Bytes(), 0640)
}

// Meta returns a 'meta manifest'- a manifest which contains everything
// in s.World as a dependency with either a version constraint, where specified,
// or 'latest'
func (s StateDB) Meta() (m Manifest, err error) {
	m = Manifest{
		Provides:   "world",
		VersionStr: fmt.Sprintf("%d", s.LastUpdate.Unix()),
		Profiles:   map[string]Profile{"default": Profile{Deps: []Dep{}}},
	}

	return processManifest(m)
}

// IsInstalled returns true if this package/version combination
// is already installed
func (s StateDB) IsInstalled(pkg string) (installed bool) {
	_, installed = s.Installed[pkg]

	return
}

// AddInstalled adds a package, and the time at which it was installed
// to the Installed list
func (s *StateDB) AddInstalled(pkg string, t time.Time) {
	s.Installed[pkg] = t
	s.LastUpdate = time.Now()
}

// AddWorld adds a package to the World list
func (s *StateDB) AddWorld(pkg string, constraint string) {
	s.World[pkg] = constraint
	s.LastUpdate = time.Now()
}
