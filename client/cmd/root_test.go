package cmd

import (
	"bytes"
	"testing"
)

var (
	defaultExpect  = "vin provides package management stuff for vinyl linux\n\nIt offers:\n  * Speed and extensibility- no mucking around with byzantine package manager configs\n  * Modern tooling- sha and md5 are slow and unwieldly. We've moved on. PGP signing package manifests. Why? Sod it, let git handle it\n  * Low barrier to entry for contributing- why are we mucking about with granting access to servers/ mailing lists to make changes? Github/ gitlab/ gitea/ all of these solve these issues better. Slap a reasonably permissive CLA onto a repo somewhere, and let people do what they do.\n\nUsage:\n  vin [command]\n\nAvailable Commands:\n  advise      generate a vin config.toml for this machine\n  help        Help about any command\n  install     install package(s)\n  reload      trigger a reload of manifests\n  version     Return server and client version information\n\nFlags:\n      --config string   config file (default is $HOME/.vin.yaml)\n  -h, --help            help for vin\n      --sock string     path to the vin socket file (default \"unix:///var/run/vin.sock\")\n\nUse \"vin [command] --help\" for more information about a command.\n"
	defaultInstall = "Usage:\n  vin install [package name] | [package name] [package name] [flags]\n\nFlags:\n  -f, --force            Force installation, even when targets are marked as installed\n  -h, --help             help for install\n  -v, --version string   Version constraint to install. This command allows strict versions, such as \"1.2.3\", or loose versions such as \">=1.20, <1.3.5\" (default \"latest\")\n\nGlobal Flags:\n      --config string   config file (default is $HOME/.vin.yaml)\n      --sock string     path to the vin socket file (default \"unix:///var/run/vin.sock\")\n"
	defaultReload  = ""
)

func TestInitConfig(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			t.Errorf("unexpected error: %#v", err)
		}
	}()

	initConfig()
}

func TestExecute(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			t.Errorf("unexpected error: %#v", err)
		}
	}()

	Execute()
}

func TestRootCmd(t *testing.T) {
	// Since this test is checking stdout/stderr it's a fair question
	// to ask why use testing.T/ associated methods, rather than example
	// tests (https://golang.org/pkg/testing/#hdr-Examples) which exist
	// for this precise purpose.
	//
	// The reason is simple- we're testing lots of output from inputs passed
	// to rootCmd, and writing table driven tests to DRY this up a little
	// seemed to make the most sense.
	//
	// Keeps our test output a little less verbose, too (though this reason
	// only occurs to me now, and not at time or architecting tests)

	for _, test := range []struct {
		name        string
		args        []string
		expect      string
		expectError bool
	}{
		{"no args", []string{}, defaultExpect, false},
		{"install, missing arg", []string{"install"}, defaultInstall, true},
		{"install, error connecting to vind", []string{"install", "foo"}, "", true},
		{"install multiple packages, error connecting to vind", []string{"install", "foo", "bar", "baz"}, "", true},
		{"reload, error connecting to vind", []string{"reload"}, defaultReload, true},
		{"reload, error connecting to vind", []string{"version"}, defaultReload, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			rootCmd.SetArgs(test.args)

			b := &bytes.Buffer{}
			rootCmd.SetOut(b)

			err := rootCmd.Execute()
			if err == nil && test.expectError {
				t.Errorf("expected error")
			} else if err != nil && !test.expectError {
				t.Errorf("unexpected error %#v", err)
			}

			got := b.String()
			if test.expect != got {
				t.Errorf("expected\n\t%q\nreceived\n\t%q", test.expect, got)
			}
		})
	}
}

func TestRootCmd_Advise(t *testing.T) {
	// Separate this from the above; we don't care so much _what_ the output
	// is, just that there is some and nothing crashes

	defer func() {
		err := recover()
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	}()

	rootCmd.SetArgs([]string{"advise"})

	b := &bytes.Buffer{}
	rootCmd.SetOut(b)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %#v", err)
	}

	got := b.String()
	if got == "" {
		t.Fatalf("expected _some_ output")
	}
}
