package config

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
