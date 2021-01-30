package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/vinyl-linux/vin/server"
)

const (
	DefaultProfile = "default"
)

type OutputSender interface {
	Send(m *server.Output) error
}

type Server struct {
	server.UnimplementedVinServer

	config Config
	sdb    StateDB
	mdb    ManifestDB

	operationLock *sync.Mutex
}

func NewServer(c Config, m ManifestDB, sdb StateDB) (s Server, err error) {
	return Server{
		UnimplementedVinServer: server.UnimplementedVinServer{},
		config:                 c,
		mdb:                    m,
		sdb:                    sdb,
		operationLock:          &sync.Mutex{},
	}, nil
}

func (s *Server) getOpsLock(sender OutputSender) {
	done := make(chan bool, 0)
	go func() {
		c := time.Tick(time.Second)

		for {
			select {
			case <-done:
				return
			case <-c:
				sender.Send(&server.Output{Line: "waiting for lock"})
			}
		}
	}()

	s.operationLock.Lock()
	done <- true
}

func (s Server) Install(is *server.InstallSpec, vs server.Vin_InstallServer) (err error) {
	// find root manifest
	vs.Send(&server.Output{
		Line: fmt.Sprintf("installing %s", is.Pkg),
	})

	s.getOpsLock(vs)
	defer s.operationLock.Unlock()

	var ver version.Constraints

	if is.Version != "" {
		ver, err = version.NewConstraint(is.Version)
		if err != nil {
			return
		}
	}

	// create output chan and iterate over it, sending messages
	output := make(chan string, 0)
	defer close(output)

	go dispatchOutput(vs, output)

	g := NewGraph(&s.mdb, &s.sdb, output)
	tasks, err := g.Solve(DefaultProfile, is.Pkg, ver)
	if err != nil {
		return
	}

	vs.Send(&server.Output{
		Line: fmt.Sprintf(installingLine(tasks)),
	})

	var (
		cmd string
	)

	// for each pkg
	for _, task := range tasks {
		if s.sdb.IsInstalled(task.ID) {
			output <- fmt.Sprintf("%s is already installed, skipping", task.ID)

			continue
		}

		err = task.Prepare(output)
		if err != nil {
			return
		}

		iv := InstallationValues{s.config, task}
		workDir := filepath.Join(task.dir, task.Commands.WorkingDir)

		for _, raw := range task.Commands.Slice() {
			cmd, err = iv.Expand(raw)
			if err != nil {
				return
			}

			err = execute(workDir, cmd, output)
			if err != nil {
				return
			}
		}

		s.sdb.AddInstalled(task.ID, time.Now())
	}

	s.sdb.AddWorld(is.Pkg, is.Version)
	s.sdb.Write()

	return
}

func dispatchOutput(vs server.Vin_InstallServer, output chan string) {
	for msg := range output {
		vs.Send(&server.Output{Line: msg})
	}
}

func installingLine(tasks []*Manifest) string {
	sb := strings.Builder{}

	sb.WriteString("installing dependencies:")
	for _, t := range tasks {
		sb.WriteString("\n")
		sb.WriteString("\t")
		sb.WriteString(t.ID)
	}

	return sb.String()
}
