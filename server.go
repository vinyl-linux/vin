package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/vinyl-linux/vin/server"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	DefaultProfile = "default"
)

type Server struct {
	server.UnimplementedVinServer

	sdb StateDB
	mdb ManifestDB

	operationLock *sync.Mutex
}

func NewServer(m ManifestDB, sdb StateDB) (s Server, err error) {
	return Server{
		UnimplementedVinServer: server.UnimplementedVinServer{},
		sdb:                    sdb,
		mdb:                    m,
		operationLock:          &sync.Mutex{},
	}, nil
}

func (s *Server) getOpsLock(oc chan string) {
	done := make(chan bool, 0)
	go func() {
		c := time.Tick(time.Second)

		for {
			select {
			case <-done:
				return
			case <-c:
				oc <- "waiting for lock"
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

	output := NewOutputter(vs)
	output.Prefix = "setup"

	defer close(output.C)
	go output.Dispatch()

	output.C <- fmt.Sprintf("installing %s", is.Pkg)

	s.getOpsLock(output.C)
	defer s.operationLock.Unlock()

	var ver version.Constraints

	if is.Version != "" {
		ver, err = version.NewConstraint(is.Version)
		if err != nil {
			return
		}
	}

	g := NewGraph(&s.mdb, &s.sdb, output.C)
	tasks, err := g.Solve(DefaultProfile, is.Pkg, ver)
	if err != nil {
		return
	}

	output.C <- installingLine(tasks)

	var (
		cmd string
	)

	// Write world db to disk on return
	defer s.sdb.Write()

	// Store finaliser commands
	finalisers := make([]*Manifest, 0)

	// for each pkg
	for _, task := range tasks {
		output.Prefix = task.ID

		if s.sdb.IsInstalled(task.ID) && !is.Force {
			output.C <- fmt.Sprintf("%s is already installed, skipping", task.ID)

			continue
		}

		if task.Meta {
			output.C <- fmt.Sprintf("%s is a meta-package, skipping", task.ID)

			continue
		}

		err = task.Prepare(output.C)
		if err != nil {
			return
		}

		err = task.Commands.Patch(task.Commands.absoluteWorkingDir, output.C)
		if err != nil {
			return
		}

		for _, raw := range task.Commands.Slice() {
			cmd, err = task.Commands.installationValues.Expand(raw)
			if err != nil {
				return
			}

			err = execute(task.Commands.absoluteWorkingDir, cmd, task.Commands.Skipenv, output.C)
			if err != nil {
				return
			}
		}

		if task.ServiceDir != "" {
			output.C <- "installing service directory"
			err = installServiceDir(filepath.Join(task.ManifestDir, task.ServiceDir))
			if err != nil {
				return
			}
		}

		if task.Commands.Finaliser != "" {
			finalisers = append(finalisers, task)
		}

		s.sdb.AddInstalled(task.ID, time.Now())
	}

	for _, task := range finalisers {
		var cmd string
		cmd, err = task.Commands.installationValues.Expand(task.Commands.Finaliser)
		if err != nil {
			return
		}

		err = execute(task.Commands.absoluteWorkingDir, cmd, task.Commands.Skipenv, output.C)
		if err != nil {
			return
		}
	}

	s.sdb.AddWorld(is.Pkg, is.Version)

	return
}

func (s Server) Reload(_ *emptypb.Empty, vs server.Vin_ReloadServer) (err error) {
	output := NewOutputter(vs)
	output.Prefix = "reload"

	defer close(output.C)
	go output.Dispatch()

	output.C <- "reloading config"

	s.getOpsLock(output.C)
	defer s.operationLock.Unlock()

	// reload mdb
	s.mdb.Reload()

	output.C <- "reloaded"

	return
}

func (s Server) Version(ctx context.Context, _ *emptypb.Empty) (v *server.VersionMessage, err error) {
	return &server.VersionMessage{
		Ref:       ref,
		BuildUser: buildUser,
		BuiltOn:   builtOn,
	}, nil
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
