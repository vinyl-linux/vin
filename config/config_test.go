package config

import (
	"fmt"
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

func ExampleConfig_String() {
	fmt.Println(Config{}.String())
	// Output:
	// configure_flags = ''
	// MAKEOPTS = ''
	// CFLAGS = ''
	// CXXFLAGS = ''
}
