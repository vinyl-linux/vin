package main

import (
	"reflect"
	"testing"
)

func TestReadManifest(t *testing.T) {
	oldPkgDir := pkgDir
	defer func() {
		pkgDir = oldPkgDir
	}()

	pkgDir = "testdata/manifests"

	for _, test := range []struct {
		name        string
		pkg         string
		ver         string
		expect      Manifest
		expectError bool
	}{
		{"valid vin manifest", "vin", "0.0.0-rc0", Manifest{ID: "vin 0.0.0-rc0", Provides: "vin", VersionStr: "0.0.0-rc0", Licence: "BSD3", Tarball: "https://github.com/vinyl-linux/vin/archive/0.0.0-rc0.tar.gz", Profiles: map[string]Profile{"default": Profile{Deps: []Dep{Dep{"go", ">= 1.12"}}}}, Commands: Commands{Configure: "true", Compile: "make", Install: "make install"}}, false},
		{"missing manifest", "unknown", "0", Manifest{}, true},
		{"invalid manifest", "invalid", "0.1.0", Manifest{}, true},
		{"invalid dep", "invalid", "0.1.1", Manifest{ID: "invalid 0.1.1", Provides: "invalid", VersionStr: "0.1.1", Profiles: map[string]Profile{"default": Profile{Deps: []Dep{Dep{"bash", "xxx"}}}}}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			received, err := ReadManifest(test.pkg, test.ver)

			if err == nil && test.expectError {
				t.Error("expected error, received none")
			} else if err != nil && !test.expectError {
				t.Errorf("unexpected error: %+v", err)
			}

			// Remove Version pointer to make testing easier
			received.Version = nil

			if !reflect.DeepEqual(test.expect, received) {
				t.Errorf("expected:\t\n%#v\nreceived:\t\n%#v", test.expect, received)
			}
		})
	}
}
