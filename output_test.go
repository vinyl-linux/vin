package main

import (
	"strings"
	"testing"
	"time"

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

	vs := &mockInstallServer{}
	o := NewOutputter(vs)
	go o.Dispatch()

	for _, test := range []struct {
		prefix      string
		msg         string
		expectCount int
		expect      string
	}{
		{"", "Hello, world!", 1, "\x1b[36;1m\x1b[0;22m\tHello, world!"},
		{"test", "Doing Tests!", 1, "\x1b[36;1mtest\x1b[0;22m\tDoing Tests!"},
		{"multiline", "Line 1\nAnd another!", 2, "\x1b[36;1mmultiline\x1b[0;22m\tLine 1\n\x1b[36;1mmultiline\x1b[0;22m\tAnd another!"},
		{"trailing", "Trailing whitespace\n", 1, "\x1b[36;1mtrailing\x1b[0;22m\tTrailing whitespace"},
	} {
		t.Run("", func(t *testing.T) {
			vs.messages = []*server.Output{}

			o.Prefix = test.prefix
			o.C <- test.msg

			time.Sleep(time.Millisecond * 200)

			if len(vs.messages) != test.expectCount {
				t.Fatalf("vs.messages: expected %d, received %d", test.expectCount, len(vs.messages))
			}

			lines := []string{}
			for _, m := range vs.messages {
				lines = append(lines, m.Line)
			}

			got := strings.Join(lines, "\n")

			if test.expect != got {
				t.Errorf("expected %q, recveived %q", test.expect, got)
			}
		})
	}
}
