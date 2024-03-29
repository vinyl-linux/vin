package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

func TestUntar_WithLinks(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	err = untar("testdata/complex.tar.bz2", dir)
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	t.Run("checksum", func(t *testing.T) {
		expect := "9fbefb949709fc7086ab4be43544d08406f0ededa0733f6683424e774a6cb799"

		sum, err := checksum(filepath.Join(dir, "raw"))
		if err != nil {
			t.Fatalf("unexpected error: %+v", err)
		}

		if expect != sum {
			t.Errorf("expected %q, received %q", expect, sum)
		}
	})

	t.Run("hard links", func(t *testing.T) {
		_, err := os.Stat(filepath.Join(dir, "hardlink"))
		if err != nil {
			t.Fatalf("unexpected error: %+v", err)
		}
	})

	t.Run("symlinks", func(t *testing.T) {
		_, err := os.Stat(filepath.Join(dir, "symlink"))
		if err != nil {
			t.Fatalf("unexpected error: %+v", err)
		}
	})
}

func TestExecute(t *testing.T) {
	for _, test := range []struct {
		name        string
		dir         string
		command     string
		expect      string
		expectError bool
	}{
		{"happy path, no output", ".", "true", "", false},
		{"read file", "testdata", "cat hello-world", "hello, world!\n", false},
		{"no such command", ".", "ahkjsdhkajshkqrzokhz", "", true},
		{"good command, bad file", ".", "cat nonsuch.txt", "cat: nonsuch.txt: No such file or directory\n", true},
		{"empty command", "", "", "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			output := make(chan string)
			outputB := strings.Builder{}
			go func() {
				for s := range output {
					outputB.WriteString(s)
				}
			}()

			err := execute(test.dir, test.command, false, output)
			close(output)
			if err == nil && test.expectError {
				t.Errorf("expected error")
			}

			if err != nil && !test.expectError {
				t.Errorf("unexpected error: %+v", err)
			}

			got := outputB.String()
			if test.expect != got {
				t.Errorf("expected %q, received %q", test.expect, got)
			}
		})
	}
}
