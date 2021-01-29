package cmd

import (
	"context"
	"fmt"
	"io"
	"net"
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

	conn, err := grpc.DialContext(ctx,
		addr,
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

func (c client) install(pkg, version string) (err error) {
	is := &vin.InstallSpec{
		Pkg: pkg,
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
			return nil
		}

		if err != nil {
			return
		}

		fmt.Println(strings.TrimSpace(output.Line))
	}

	return nil
}
