package main

import (
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/vinyl-linux/vin/server"
)

func TestNewOutputter(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	}()

	NewOutputter(&mockInstallServer{})
}

func TestOutputter_Dispatch(t *testing.T) {
	// This is a 'Test', rather than an 'Example' in order to test
	// escape codes

	// Ensure colour escape codes are still generated in github actions
	// see: https://github.com/fatih/color/issues/132
	color.NoColor = false

	vs := &mockInstallServer{}
	o := NewOutputter(vs)
	go o.Dispatch()

	for _, test := range []struct {
		prefix string
		msg    string
		expect string
	}{
		{"", "Hello, world!", "\x1b[36;1m\x1b[0m\tHello, world!"},
		{"test", "Doing Tests!", "\x1b[36;1mtest\x1b[0m\tDoing Tests!"},
	} {
		t.Run("", func(t *testing.T) {
			vs.messages = []*server.Output{}

			o.Prefix = test.prefix
			o.C <- test.msg

			time.Sleep(time.Millisecond * 200)

			if len(vs.messages) != 1 {
				t.Fatalf("vs.messages: expected %d, received %d", 1, len(vs.messages))
			}

			if test.expect != vs.messages[0].Line {
				t.Errorf("expected %q, recveived %q", test.expect, vs.messages[0].Line)
			}

		})
	}
}
