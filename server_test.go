package main

import (
	"path/filepath"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/hashicorp/go-version"
	"github.com/vinyl-linux/vin/server"
	"google.golang.org/grpc"
)

type mockInstallServer struct {
	grpc.ServerStream
	messages []*server.Output
}

func (mis *mockInstallServer) Send(m *server.Output) error {
	mis.messages = append(mis.messages, m)

	return nil
}

func TestServer_Install(t *testing.T) {
	pkgDir = "testdata/manifests/valid-manifests"
	configFile = "testdata/test-config.toml"
	cacheDir = "/tmp"
	sockAddr = "/tmp/vin-test.sock"

	c, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	mdb, err := LoadDB()
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	s, err := NewServer(c, mdb, StateDB{})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	for _, test := range []struct {
		name        string
		pkg         string
		ver         string
		expectError bool
	}{
		{"valid package, explicit version", "standalone", "1.0.0", false},
		{"valid package, empty version", "standalone", "", false},
		{"valid package, missing version", "standalone", "> 2.0.0", true},
		{"invalid package", "foo", "", true},
		{"valid package, invalid version", "standalone", "zzzzz", true},
		{"valid package, bad checksum", "standalone", "0.1.1", true},
		{"valid package, bad command template", "standalone", "0.1.2", true},
		{"valid package, 404 archive", "standalone", "0.1.3", true},
		{"erroring commands", "standalone", "0.1.0", true},
		{"missing package", "", "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			// create an empty statedb
			stateDB = filepath.Join("tmp", uuid.Must(uuid.NewV4()).String(), "vin-test.db")
			s.sdb, _ = LoadStateDB()

			is := &server.InstallSpec{
				Pkg:     test.pkg,
				Version: test.ver,
			}

			vs := &mockInstallServer{}

			err := s.Install(is, vs)
			if err == nil && test.expectError {
				t.Errorf("expected error")
			} else if err != nil && !test.expectError {
				t.Errorf("unexpected error: %+v", err)
			}

			t.Logf("err: %+v", err)
		})
	}
}

func TestServer_Reload(t *testing.T) {
	pkgDir = "testdata/manifests/valid-manifests"
	configFile = "testdata/test-config.toml"
	cacheDir = "/tmp"
	sockAddr = "/tmp/vin-test.sock"
	stateDB = filepath.Join("/tmp", uuid.Must(uuid.NewV4()).String())

	c, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	mdb, err := LoadDB()
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	s, err := NewServer(c, mdb, StateDB{})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	emptyConstraint, _ := version.NewConstraint(">= 0.0.0")

	// Reset ManifestDB path, hit a server reload, and see whether the ManifestDB now
	// contains a package that wasn't there before, thus suggesting that the DB has been
	// updated
	for _, test := range []struct {
		name        string
		pkgDir      string
		pkg         string
		expectPkg   bool
		expectError bool
	}{
		{"happy path", "testdata/manifests/null-manifests", "app-1", true, false},
		{"no matching packages (anymore)", "testdata/no-such-manifest", "", false, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			vs := &mockInstallServer{}

			pkgDir = test.pkgDir

			err = s.Reload(nil, vs)
			if err == nil && test.expectError {
				t.Errorf("expected error")
			} else if err != nil && !test.expectError {
				t.Errorf("unexpected error: %+v", err)
			}

			m, _ := s.mdb.Satisfies(test.pkg, emptyConstraint)
			if (len(m) > 0) != test.expectPkg {
				t.Errorf("received %d manifests, expectPkg is %v", len(m), test.expectPkg)
			}
		})
	}
}
