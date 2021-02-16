package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/vinyl-linux/vin/config"
	"github.com/vinyl-linux/vin/server"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	DefaultProfile = "default"
)

type Server struct {
	server.UnimplementedVinServer

	sdb    StateDB
	config config.Config
	mdb    ManifestDB

	operationLock *sync.Mutex
}

func NewServer(c config.Config, m ManifestDB, sdb StateDB) (s Server, err error) {
	return Server{
		UnimplementedVinServer: server.UnimplementedVinServer{},
		sdb:                    sdb,
		config:                 c,
		mdb:                    m,
		operationLock:          &sync.Mutex{},
	}, nil
}

func (s *Server) getOpsLock(sender server.OutputSender) {
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
	if is.Pkg == "" {
		return fmt.Errorf("package must not be empty")
	}

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

	// Write world db to disk on return
	defer s.sdb.Write()

	// for each pkg
	for _, task := range tasks {
		if s.sdb.IsInstalled(task.ID) && !is.Force {
			output <- fmt.Sprintf("%s is already installed, skipping", task.ID)

			continue
		}

		if task.Meta {
			output <- fmt.Sprintf("%s is a meta-package, skipping", task.ID)

			continue
		}

		err = task.Prepare(output)
		if err != nil {
			return
		}

		iv := InstallationValues{s.config, task}
		workDir := filepath.Join(task.dir, task.Commands.WorkingDir)

		err = task.Commands.Patch(workDir, output)
		if err != nil {
			return
		}

		for _, raw := range task.Commands.Slice() {
			cmd, err = iv.Expand(raw)
			if err != nil {
				return
			}

			err = execute(workDir, cmd, output, s.config)
			if err != nil {
				return
			}
		}

		s.sdb.AddInstalled(task.ID, time.Now())
	}

	s.sdb.AddWorld(is.Pkg, is.Version)

	return
}

func (s Server) Reload(_ *emptypb.Empty, vs server.Vin_ReloadServer) (err error) {
	vs.Send(&server.Output{
		Line: "reloading config",
	})

	s.getOpsLock(vs)
	defer s.operationLock.Unlock()

	// reload mdb
	s.mdb.Reload()

	vs.Send(&server.Output{
		Line: "reloaded",
	})

	return
}

func dispatchOutput(o server.OutputSender, output chan string) {
	for msg := range output {
		o.Send(&server.Output{Line: msg})
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
