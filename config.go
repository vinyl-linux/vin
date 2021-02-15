package main

import (
	"strings"
	"text/template"

	"github.com/vinyl-linux/vin/config"
)

type InstallationValues struct {
	config.Config
	*Manifest
}

// Expand takes a string (containing an ostensible template) and
// expands it using Config and Manifest.
//
// What an incredibly comment that was... Okay... we allow for
// the commands section of a package manifest to be a template, and
// we grant it access to certain configuration items. This function
// is where we take that template and turn it into an actual command
// which can be used
func (c InstallationValues) Expand(s string) (cmd string, err error) {
	b := &strings.Builder{}

	tmpl, err := template.New("cmd").Parse(s)
	if err != nil {
		return
	}

	err = tmpl.Execute(b, c)
	if err != nil {
		return
	}

	cmd = b.String()

	return
}
