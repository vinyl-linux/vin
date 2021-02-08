package cmd

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	vin "github.com/vinyl-linux/vin/server"
	"google.golang.org/grpc"
)

type client struct {
	c vin.VinClient
}

func newClient(addr string) (c client, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resolvedAddr, err := parseAddr(addr)
	if err != nil {
		return
	}

	conn, err := grpc.DialContext(ctx,
		resolvedAddr,
		grpc.WithBlock(),
		grpc.WithInsecure(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}))

	if err != nil {
		return
	}

	c.c = vin.NewVinClient(conn)

	return
}

func parseAddr(addr string) (s string, err error) {
	u, err := url.Parse(addr)
	if err != nil {
		return
	}

	if u.Scheme == "unix" {
		p := u.Path

		u.Path, err = filepath.EvalSymlinks(p)
		if err != nil {
			return
		}
	}

	return u.String(), nil
}

func (c client) install(pkg, version string, force bool) (err error) {
	is := &vin.InstallSpec{
		Pkg:   pkg,
		Force: force,
	}

	if version != "" && version != "latest" {
		is.Version = version
	}

	vic, err := c.c.Install(context.Background(), is)
	if err != nil {
		return
	}

	var output *vin.Output
	for {
		output, err = vic.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return
		}

		fmt.Println(strings.TrimSpace(output.Line))
	}

	return nil
}
