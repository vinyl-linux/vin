package main

import (
	"context"
	"os"
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

	err := loadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	mdb, err := LoadDB()
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	s, err := NewServer(mdb, StateDB{})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	for _, test := range []struct {
		name        string
		pkg         []string
		ver         string
		expectError bool
	}{
		{"valid package, explicit version", []string{"standalone"}, "1.0.0", false},
		{"valid package, empty version", []string{"standalone"}, "", false},
		{"valid package, missing version", []string{"standalone"}, "> 2.0.0", true},
		{"invalid package", []string{"foo"}, "", true},
		{"valid package, invalid version", []string{"standalone"}, "zzzzz", true},
		{"valid package, bad checksum", []string{"standalone"}, "0.1.1", true},
		{"valid package, bad command template", []string{"standalone"}, "0.1.2", true},
		{"valid package, 404 archive", []string{"standalone"}, "0.1.3", true},
		{"erroring commands", []string{"standalone"}, "0.1.0", true},
		{"missing package", []string{}, "", true},
		{"empty package", []string{""}, "", true},
		{"multiple empty packages", []string{"", ""}, "", true},
		{"multiple valid packages", []string{"standalone", "metaz"}, "", false},
		{"meta package", []string{"metaz"}, "", false},
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

func TestServer_Install_WithService(t *testing.T) {
	pkgDir = "testdata/manifests-with-services"
	configFile = "testdata/test-config.toml"
	cacheDir = "/tmp"
	sockAddr = "/tmp/vin-test.sock"

	svcDir, _ = os.MkdirTemp("", "")
	t.Logf(svcDir)

	err := loadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	mdb, err := LoadDB()
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	s, err := NewServer(mdb, StateDB{})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	stateDB = filepath.Join("tmp", uuid.Must(uuid.NewV4()).String(), "vin-test.db")
	s.sdb, _ = LoadStateDB()

	is := &server.InstallSpec{
		Pkg:     []string{"another-sample-app"},
		Version: "",
	}

	vs := &mockInstallServer{}

	err = s.Install(is, vs)
	if err != nil {
		t.Fatal(err)
	}
}

func TestServer_Reload(t *testing.T) {
	pkgDir = "testdata/manifests/valid-manifests"
	configFile = "testdata/test-config.toml"
	cacheDir = "/tmp"
	sockAddr = "/tmp/vin-test.sock"
	stateDB = filepath.Join("/tmp", uuid.Must(uuid.NewV4()).String())

	err := loadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	mdb, err := LoadDB()
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	s, err := NewServer(mdb, StateDB{})
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

func TestServer_Version(t *testing.T) {
	s := Server{}

	out, err := s.Version(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		key    string
		value  string
		expect string
	}{
		{"Ref", out.Ref, "unknown"},
		{"BuildUser", out.BuildUser, "unknown"},
		{"BuiltOn", out.BuiltOn, "unknown"},
	} {
		t.Run(test.key, func(t *testing.T) {
			if test.expect != test.value {
				t.Errorf("expected %q, received %q", test.expect, test.value)
			}
		})
	}
}
