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
	"google.golang.org/protobuf/types/known/emptypb"
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

func (c client) install(pkgs []string, version string, force bool) (err error) {
	is := &vin.InstallSpec{
		Pkg:   pkgs,
		Force: force,
	}

	if version != "" && version != "latest" {
		is.Version = version
	}

	vic, err := c.c.Install(context.Background(), is)
	if err != nil {
		return
	}

	return outputLoop(vic)
}

func (c client) reload() (err error) {
	vrc, err := c.c.Reload(context.Background(), &emptypb.Empty{})
	if err != nil {
		return
	}

	return outputLoop(vrc)
}

func (c client) version() (out string, err error) {
	v, err := c.c.Version(context.Background(), new(emptypb.Empty))
	if err != nil {
		return
	}

	return formatVersion(true, v.Ref, v.BuildUser, v.BuiltOn), nil
}

func outputLoop(o vin.OutputReceiver) (err error) {
	var output *vin.Output

	for {
		output, err = o.Recv()
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

func formatVersion(isServer bool, ref, user, built string) string {
	return fmt.Sprintf("%s version\n---\nVersion: %s\nBuild User: %s\nBuilt On: %s\n",
		isServerString(isServer), ref, user, built,
	)
}

func isServerString(b bool) string {
	if b {
		return "Server"
	}

	return "Client"
}
