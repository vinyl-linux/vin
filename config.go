package main

import (
	"io/ioutil"
	"strings"
	"text/template"

	"github.com/pelletier/go-toml"
)

// Config holds all of the configuration that vin needs,
// including things like make/ confdigure opts, and so on
//
// This struct can be passed into builds to allow for templating
// build opts
type Config struct {
	// Flags passed to './configure'. These options are available to
	// commands, in order to set flags during compilation.
	//
	// Note that --prefix is always set by vin
	ConfigureFlags string

	// MakeOpts are passed to make
	MakeOpts string
}

func LoadConfig() (c Config, err error) {
	d, err := ioutil.ReadFile(configFile)
	if err != nil {
		return
	}

	err = toml.Unmarshal(d, &c)

	return
}

type InstallationValues struct {
	Config
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
