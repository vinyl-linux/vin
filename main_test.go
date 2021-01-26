package main

import (
	"testing"
)

func TestSetup(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			t.Errorf("unexpected panic: %+v", err)
		}
	}()

	pkgDir = "testdata/manifests/valid-manifests"
	configFile = "testdata/test-config.toml"
	cacheDir = "/tmp"
	sockAddr = "/tmp/vin-test.sock"

	Setup()
}
