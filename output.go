package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/vinyl-linux/vin/server"
)

var (
	prefixColour = color.New(color.FgCyan, color.Bold)
)

func init() {
	// Ensure colour escape codes are generated properly, in both
	// gitlab actions _and_ when running `vind` under s6 - in both cases
	// the `color` package doesn't create escape codes because we're not
	// running as a real terminal.
	//
	// In the case of `vind` this is fine; we might not be a terminal, but
	// the `vin` client is certainly running under one
	color.NoColor = false

}

// Outputter handles writing strings to vin
// clients by reading strings from a chan and writing
// them, prefixed with an additional string, to a server.OutputSender
type Outputter struct {
	C      chan string
	Prefix string

	o server.OutputSender
}

// NewOutputter takes an OutputSender and returns an Outputter
// ready to be dispatched across
func NewOutputter(o server.OutputSender) Outputter {

	return Outputter{
		C:      make(chan string),
		Prefix: "",
		o:      o,
	}
}

// Dispatch loops over o.C, prefixes each message with o.Prefix, and writes
// it to the internal OutputSender.
//
// It is generally called as a `go func`
func (o *Outputter) Dispatch() {
	for msg := range o.C {
		// generate this for each message; the prefix can (and does) change often
		prefix := prefixColour.Sprintf(o.Prefix)

		msg = strings.TrimSpace(msg)

		for _, line := range strings.Split(msg, "\n") {
			o.o.Send(&server.Output{
				Line: fmt.Sprintf("%s\t%s", prefix, line),
			})
		}
	}
}
