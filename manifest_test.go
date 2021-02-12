package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func strP(s string) *string { return &s }

func TestReadManifest(t *testing.T) {
	pkgDir = "testdata/manifests"

	for _, test := range []struct {
		name        string
		pkg         string
		ver         string
		expect      Manifest
		expectError bool
	}{
		{"valid vin manifest", "vin", "0.0.0-rc0", Manifest{ID: "vin 0.0.0-rc0", Provides: "vin", VersionStr: "0.0.0-rc0", Licence: "BSD3", Tarball: "https://github.com/vinyl-linux/vin/archive/0.0.0-rc0.tar.gz", ManifestDir: "testdata/manifests/vin/0.0.0-rc0", Profiles: map[string]Profile{"default": {Deps: []Dep{{"go", ">= 1.12"}}}}, Commands: Commands{Configure: strP("true"), Compile: strP("make"), Install: strP("make install"), Patches: []string(nil)}}, false},
		{"missing manifest", "unknown", "0", Manifest{}, true},
		{"invalid manifest", "invalid", "0.1.0", Manifest{}, true},
		{"invalid dep", "invalid", "0.1.1", Manifest{ID: "invalid 0.1.1", Provides: "invalid", VersionStr: "0.1.1", ManifestDir: "testdata/manifests/invalid/0.1.1", Profiles: map[string]Profile{"default": {Deps: []Dep{{"bash", "xxx"}}}}, Commands: Commands{Patches: []string(nil)}}, true},
		{"manifest with patches", "patched", "0.0.1", Manifest{ID: "patched 0.0.1", Provides: "patched", VersionStr: "0.0.1", ManifestDir: "testdata/manifests/patched/0.0.1", Commands: Commands{Patches: []string{"testdata/manifests/patched/0.0.1/0.patch"}}}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			received, err := readManifest(filepath.Join(pkgDir, test.pkg, test.ver, ManifestFilename))

			if err == nil && test.expectError {
				t.Error("expected error, received none")
			} else if err != nil {
				if !test.expectError {
					t.Errorf("unexpected error: %+v", err)
				} else {
					t.Logf("received error: %+v", err)
				}
			}

			// Remove Version pointer to make testing easier
			received.Version = nil

			if !reflect.DeepEqual(test.expect, received) {
				t.Errorf("expected:\t\n%#v\nreceived:\t\n%#v", test.expect, received)
			}
		})
	}
}

func TestManifest_Prepare(t *testing.T) {
	cacheDir, _ = ioutil.TempDir("", "")

	man := Manifest{
		Provides:   "test-package",
		Tarball:    "https://github.com/vinyl-linux/vin/archive/0.0.0-rc0.tar.gz",
		Checksum:   "a4569f58346973cdbaaa69cda6c8d712bf824bbee824153ae0692aff54120ab7",
		VersionStr: "0.0.0-rc0",
	}

	output := make(chan string, 0)
	go func() {
		for _ = range output {
		}
	}()

	err := man.Prepare(output)
	if err != nil {
		t.Errorf("unexpected error: %+v", err)
	}

	t.Logf("extract dir: %q", man.dir)
}

func TestCommands_Slice(t *testing.T) {
	// Ensure that slice contains user specified commands, where user
	// specified commands are set
	configure := "conf.sh"
	compile := "comp.sh"
	install := "install.sh"

	c := Commands{
		Configure: &configure,
		Compile:   &compile,
		Install:   &install,
	}

	cSlice := c.Slice()

	for idx, expect := range []string{configure, compile, install} {
		t.Run("", func(t *testing.T) {
			if expect != cSlice[idx] {
				t.Errorf("expected %q, received %q", expect, cSlice[idx])
			}
		})
	}
}

func TestCommands_Patch(t *testing.T) {
	sources, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected error: %#v", err)
	}

	err = ioutil.WriteFile(filepath.Join(sources, "main.go"), []byte("package main\n"), 0600)
	if err != nil {
		t.Fatalf("unexpected error: %#v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("unexpected error: %#v", err)
	}

	testdata := filepath.Join(wd, "testdata")

	for _, test := range []struct {
		name        string
		commands    Commands
		expectError bool
	}{
		{"Simple commands, no patches", Commands{}, false},
		{"Happy path, valid patch", Commands{Patches: []string{filepath.Join(testdata, "test.patch")}}, false},
		{"Broken patch", Commands{Patches: []string{filepath.Join(testdata, "broken.patch")}}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			output := make(chan string, 0)
			go func() {
				for m := range output {
					t.Log(m)
				}
			}()

			err = test.commands.Patch(sources, output)
			if !test.expectError && err != nil {
				t.Errorf("unexpected error: %+v", err)
			} else if test.expectError && err == nil {
				t.Errorf("expected error, received none")
			}

		})
	}
}
