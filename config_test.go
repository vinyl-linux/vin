package main

import (
	"testing"
)

func TestConfig_Expand(t *testing.T) {
	configFile = "testdata/test-config.toml"

	c, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	m := Manifest{}

	for _, test := range []struct {
		name   string
		in     string
		expect string
	}{
		{"default configure", m.Commands.GetConfigure(), "./configure --enable-hello-world"},
		{"default compile", m.Commands.GetCompile(), "make -j100"},
		{"default install", m.Commands.GetInstall(), "make install -j100"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := c.Expand(test.in)
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}

			if test.expect != got {
				t.Errorf("expected %q, received %q", test.expect, err)
			}
		})
	}
}
