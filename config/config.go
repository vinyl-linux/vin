package config

import (
	"io/ioutil"

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
	ConfigureFlags string `toml:"configure_flags"`

	// MakeOpts are passed to make
	MakeOpts string `toml:"MAKEOPTS"`

	// CFlags and CXXFlags are inserted into the environment when
	// running build commands
	CFlags   string `toml:"CFLAGS"`
	CXXFlags string `toml:CXXFLAGS"`
}

func Load(configFile string) (c Config, err error) {
	d, err := ioutil.ReadFile(configFile)
	if err != nil {
		return
	}

	err = toml.Unmarshal(d, &c)

	return
}

func (c Config) String() string {
	// swallow errors, they're both massively unlikely, and
	// not used anywhere where error handling is necessary
	b, _ := toml.Marshal(c)

	return string(b)
}
