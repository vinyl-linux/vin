package config

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	for _, test := range []struct {
		path        string
		expectError bool
	}{
		{"/this/file/does/not.exist", true},
		{"../testdata/test-config.toml", false},
	} {
		t.Run(test.path, func(t *testing.T) {
			_, err := Load(test.path)
			if err == nil && test.expectError {
				t.Error("expected error, received none")
			} else if err != nil && !test.expectError {
				t.Errorf("unexpected error: %+v", err)
			}
		})
	}
}

func TestConfig_String(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	}()

	c := Config{}.String()
	expect := "CFLAGS = \"\"\nCXXFLAGS = \"\"\nMAKEOPTS = \"\"\nconfigure_flags = \"\"\n"

	if expect != c {
		t.Errorf("expected:\n%sreceived:\n%s", expect, c)
	}
}
