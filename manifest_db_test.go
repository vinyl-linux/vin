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

	constraint, _ := version.NewConstraint(">= 1.0.8")
	pkg := "app-utils"

	s, err := d.Satisfies(pkg, constraint)
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	if len(s) != 1 {
		t.Fatalf("expected 1 satisfying manifest, received %d", len(s))
	}

	t.Logf("%#v", s)
}
