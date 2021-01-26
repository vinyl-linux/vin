package main

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	configFile = "this-file-should-not-exist"
	_, err := LoadConfig()
	if err == nil {
		t.Errorf("expected error, received none")
	}
}

func TestConfig_Expand(t *testing.T) {
	configFile = "testdata/test-config.toml"

	c, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	m := Manifest{}

	for _, test := range []struct {
		name        string
		in          string
		expect      string
		expectError bool
	}{
		{"default configure", m.Commands.GetConfigure(), "./configure --enable-hello-world", false},
		{"default compile", m.Commands.GetCompile(), "make -j100", false},
		{"default install", m.Commands.GetInstall(), "make install -j100", false},
		{"dodgy template", " {{ .HelloWorld }}", "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := c.Expand(test.in)

			if err == nil && test.expectError {
				t.Error("expected error, received none")
			} else if err != nil && !test.expectError {
				t.Errorf("unexpected error: %+v", err)
			}

			if test.expect != got {
				t.Errorf("expected %q, received %q", test.expect, err)
			}
		})
	}
}
