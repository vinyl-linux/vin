package main

import (
	"testing"
)

func TestInstallationValues_Expand(t *testing.T) {
	configFile = "testdata/test-config.toml"

	err := loadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	m := &Manifest{
		ManifestDir: "/tmp/app/1.0.0/",
		Commands: Commands{
			installationValues: InstallationValues{Manifest: &Manifest{ManifestDir: "/tmp/app/1.0.0/"}},
		},
	}

	for _, test := range []struct {
		name        string
		in          string
		expect      string
		expectError bool
	}{
		{"default configure", m.Commands.GetConfigure(), "./configure --enable-hello-world", false},
		{"default compile", m.Commands.GetCompile(), "make -j100", false},
		{"default install", m.Commands.GetInstall(), "make install -j100", false},
		{"values from manifest", "rm -rf {{ .ManifestDir }}", "rm -rf /tmp/app/1.0.0/", false},
		{"dodgy template", " {{ .HelloWorld }}", "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := m.Commands.installationValues.Expand(test.in)

			if err == nil && test.expectError {
				t.Error("expected error, received none")
			} else if err != nil && !test.expectError {
				t.Errorf("unexpected error: %+v", err)
			}

			if test.expect != got {
				t.Errorf("expected %q, received %q", test.expect, got)
			}
		})
	}
}
