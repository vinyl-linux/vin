package main

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/vinyl-linux/vin/server"
)

const (
	DefaultProfile = "default"
)

type Server struct {
	server.UnimplementedVinServer

	config Config
	mdb    ManifestDB
}

func NewServer(c Config, m ManifestDB) (s Server, err error) {
	return Server{
		UnimplementedVinServer: server.UnimplementedVinServer{},
		config:                 c,
		mdb:                    m,
	}, nil
}

func (s Server) Install(is *server.InstallSpec, vs server.Vin_InstallServer) (err error) {
	// find root manifest
	vs.Send(&server.Output{
		Line: fmt.Sprintf("installing %s", is.Pkg),
	})

	var ver version.Constraints

	if is.Version != "" {
		ver, err = version.NewConstraint(is.Version)
		if err != nil {
			return
		}
	}

	g := NewGraph(s.mdb)
	tasks, err := g.Solve(DefaultProfile, is.Pkg, ver)
	if err != nil {
		return
	}

	vs.Send(&server.Output{
		Line: fmt.Sprintf(installingLine(tasks)),
	})

	// create output chan and iterate over it, sending messages
	output := make(chan string, 0)
	defer close(output)

	go dispatchOutput(vs, output)

	var (
		cmd string
	)

	// for each pkg
	for _, task := range tasks {
		err = task.Prepare(output)
		if err != nil {
			return
		}

		for _, raw := range task.Commands.Slice() {
			cmd, err = s.config.Expand(raw)
			if err != nil {
				return
			}

			err = execute(task.dir, cmd, output)
			if err != nil {
				return
			}
		}
	}

	return
}

func dispatchOutput(vs server.Vin_InstallServer, output chan string) {
	for msg := range output {
		vs.Send(&server.Output{Line: msg})
	}
}

func installingLine(tasks []*Manifest) string {
	sb := strings.Builder{}

	sb.WriteString("installing dependencies: \n")
	for _, t := range tasks {
		sb.WriteString("\t")
		sb.WriteString(t.ID)
		sb.WriteString("\n")
	}

	return sb.String()
}
