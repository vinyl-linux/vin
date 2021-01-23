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

	constraint, _ := version.NewConstraint(">= 1.0.0")
	pkg := "sample-app"

	s, err := d.Satisfies(pkg, constraint)
	t.Logf("%#v", s[0])

	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	if len(s) != 1 {
		t.Errorf("expected 1 satisfying manifest, received %d", len(s))
	}

	expectID := "sample-app 1.0.0"
	if s[0].ID != expectID {
		t.Errorf("expected %q, received %q", expectID, s[0].ID)
	}
}
