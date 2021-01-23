package main

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestUntar(t *testing.T) {
	for _, test := range []struct {
		fn          string
		expectSum   string
		expectError bool
	}{
		{"testdata/raw.tar.bz2", "9fbefb949709fc7086ab4be43544d08406f0ededa0733f6683424e774a6cb799", false},
		{"testdata/raw.tar.gz", "9fbefb949709fc7086ab4be43544d08406f0ededa0733f6683424e774a6cb799", false},
		{"testdata/no-such-file", "", true},
		{"testdata/raw", "", true},
	} {
		t.Run(test.fn, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatalf("unexpected error: %+v", err)
			}

			err = untar(test.fn, dir)
			if err == nil && test.expectError {
				t.Errorf("expected error")
			}

			if err != nil && !test.expectError {
				t.Errorf("unexpected error: %+v", err)
			}

			if !test.expectError {
				sum, err := checksum(filepath.Join(dir, "testdata", "raw"))
				if err != nil {
					t.Fatalf("unexpected error: %+v", err)
				}

				if test.expectSum != sum {
					t.Errorf("expected %q, received %q", test.expectSum, sum)
				}
			}
		})
	}
}
