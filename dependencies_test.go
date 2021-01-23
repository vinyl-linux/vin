package main

import (
	"reflect"
	"testing"

	"github.com/hashicorp/go-version"
)

func TestGraph_Solve(t *testing.T) {
	oldPkgDir := pkgDir
	defer func() {
		pkgDir = oldPkgDir
	}()

	pkgDir = "testdata/manifests/valid-manifests"

	d, err := LoadDB()
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	g := NewGraph(d)

	for _, test := range []struct {
		name          string
		pkg           string
		version       string
		expectTaskIDs []string
		expectError   bool
	}{
		{"installing non-existent package", "non-such", "= 1.0.0", []string{}, true},
		{"installing non-existent version", "app-utils", ">3", []string{}, true},

		{"happy path, with re-calculate logic", "sample-app", "1.0.0", []string{"app-utils 1.0.3", "some-security-library 1.8.9", "user-lib 1.5.0", "sample-app 1.0.0"}, false},
		{"happy path, with re-calculate logic, 'latest'", "sample-app", "", []string{"app-utils 1.0.3", "some-security-library 1.8.9", "user-lib 1.5.0", "sample-app 1.0.0"}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var (
				constraint version.Constraints
				err        error
			)

			if test.version == "" {
				constraint = nil
			} else {
				constraint, err = version.NewConstraint(test.version)
				if err != nil {
					t.Fatalf("unexpected error: %+v", err)
				}
			}

			_, err = g.Solve("default", test.pkg, constraint)
			if !test.expectError && err != nil {
				t.Errorf("unexpected error: %+v", err)
			}

			if test.expectError && err == nil {
				t.Errorf("expected error, received none")
			}

			tasks := make([]string, len(g.tasks))
			for idx, m := range g.tasks {
				tasks[idx] = m.ID
			}

			if !reflect.DeepEqual(test.expectTaskIDs, tasks) {
				t.Errorf("expected %#v, received %#v", test.expectTaskIDs, tasks)
			}
		})
	}
}

func TestGraph_Solve_Circular(t *testing.T) {
	oldPkgDir := pkgDir
	defer func() {
		pkgDir = oldPkgDir
	}()

	pkgDir = "testdata/manifests/circular"

	d, err := LoadDB()
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	g := NewGraph(d)

	_, err = g.Solve("default", "app-1", nil)
	if err == nil {
		t.Fatalf("expected error")
	}

	errStr := err.Error()
	expect := `circular dependency: "app-3" -> "app-1"`

	if expect != errStr {
		t.Fatalf("expected %q, received %q", expect, errStr)
	}
}
