package main

import (
	"testing"

	"github.com/hashicorp/go-version"
)

func TestManifestDB_Satisfies(t *testing.T) {
	oldPkgDir := pkgDir
	defer func() {
		pkgDir = oldPkgDir
	}()

	pkgDir = "testdata/manifests/valid-manifests"

	d, err := LoadDB()
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	gte1, _ := version.NewConstraint(">= 1.0.0")

	for _, test := range []struct {
		name       string
		pkg        string
		constraint version.Constraints
		expect     string
	}{
		{"simple, versioned package", "sample-app", gte1, "sample-app 1.0.0"},
		{"simple app, 'latest' version", "sample-app", latest, "sample-app 1.0.0"},
		{"app with only so-called pre-releases", "complex-versions-app", latest, "complex-versions-app 3.2.1-r1"},
	} {
		t.Run(test.name, func(t *testing.T) {
			s, err := d.Satisfies(test.pkg, test.constraint)
			if err != nil {
				t.Fatalf("unexpected error: %+v", err)
			}

			if len(s) != 1 {
				t.Fatalf("expected 1 satisfying manifest, received %d", len(s))
			}

			if s[0].ID != test.expect {
				t.Errorf("expected %q, received %q", test.expect, s[0].ID)
			}
		})
	}
}
