package main

import (
	"os"
	"path/filepath"
)

var (
	pkgDir = getEnv("PKGROOT", "/etc/vinyl/pkg")
)

func getEnv(key, def string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		v = def
	}

	return v
}

func manifestPath(pkg, ver string) string {
	return filepath.Join(pkgDir, pkg, ver, ManifestFilename)
}
