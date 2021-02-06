package cmd

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	vin "github.com/vinyl-linux/vin/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// all output streams have the same signature
type outputStream interface {
	Recv() (*vin.Output, error)
	grpc.ClientStream
}

type DummyVinClient interface {
	vin.VinClient
	outputStream

	// installSpec should return the installspec passed to this interface
	installSpec() *vin.InstallSpec

	// outut returns any output message streamed to this client
	getOutput() []string
}

type dummyInstallClient struct {
	is        *vin.InstallSpec
	output    []string
	outputPos int
	err       bool
	recvErr   bool
}

// Install(ctx context.Context, in *InstallSpec, opts ...grpc.CallOption) (Vin_InstallClient, error)
func (c *dummyInstallClient) Install(_ context.Context, is *vin.InstallSpec, _ ...grpc.CallOption) (vic vin.Vin_InstallClient, err error) {
	c.is = is
	if c.output == nil {
		c.output = make([]string, 0)
	}

	c.outputPos = 0

	if c.err {
		err = fmt.Errorf("an error")

		return
	}

	return c, nil
}

func (c *dummyInstallClient) Recv() (*vin.Output, error) {
	if c.recvErr {
		return nil, fmt.Errorf("an error")
	}

	defer func() { c.outputPos += 1 }()

	if len(c.output) <= c.outputPos {
		return nil, io.EOF
	}

	return &vin.Output{
		Line: c.output[c.outputPos],
	}, nil
}

func (c *dummyInstallClient) Header() (metadata.MD, error)  { return nil, nil }
func (c *dummyInstallClient) Trailer() metadata.MD          { return nil }
func (c *dummyInstallClient) CloseSend() error              { return nil }
func (c *dummyInstallClient) Context() context.Context      { return context.TODO() }
func (c *dummyInstallClient) SendMsg(m interface{}) error   { return nil }
func (c *dummyInstallClient) RecvMsg(m interface{}) error   { return nil }
func (c *dummyInstallClient) installSpec() *vin.InstallSpec { return c.is }
func (c *dummyInstallClient) getOutput() []string           { return c.output }

func TestClient_Install(t *testing.T) {
	for _, test := range []struct {
		name         string
		pkg          string
		ver          string
		client       DummyVinClient
		expectSpec   *vin.InstallSpec
		expectOutput []string
		expectError  bool
	}{
		{"valid package, 'latest' version", "foo", "latest", &dummyInstallClient{}, &vin.InstallSpec{Pkg: "foo"}, []string{}, false},
		{"valid package, empty version", "foo", "", &dummyInstallClient{}, &vin.InstallSpec{Pkg: "foo"}, []string{}, false},
		{"valid package, set version", "foo", "1.0.0", &dummyInstallClient{}, &vin.InstallSpec{Pkg: "foo", Version: "1.0.0"}, []string{}, false},
		{"valid package, 'latest' version, output", "foo", "latest", &dummyInstallClient{output: []string{"line-1", "line-2"}}, &vin.InstallSpec{Pkg: "foo"}, []string{"line-1", "line-2"}, false},

		{"vind throws error", "foo", "1.0.0", &dummyInstallClient{err: true}, &vin.InstallSpec{Pkg: "foo", Version: "1.0.0"}, []string{}, true},
		{"vind stream error", "foo", "1.0.0", &dummyInstallClient{recvErr: true}, &vin.InstallSpec{Pkg: "foo", Version: "1.0.0"}, []string{}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := client{c: test.client}

			err := c.install(test.pkg, test.ver, false)
			if err == nil && test.expectError {
				t.Errorf("expected error")
			} else if err != nil && !test.expectError {
				t.Errorf("unexpected error %#v", err)
			}

			output := test.client.getOutput()
			if !reflect.DeepEqual(test.expectOutput, output) {
				t.Errorf("expected %v, received %v", test.expectOutput, output)
			}

			is := test.client.installSpec()
			if !reflect.DeepEqual(test.expectSpec, is) {
				t.Errorf("expected %#v, received %#v", test.expectSpec, is)
			}
		})
	}
}

func TestNewClient_NoSock(t *testing.T) {
	_, err := newClient("/no/such/file")
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestParseAddr(t *testing.T) {
	emptySock, emptyAlsoSock, err := setupParseAddr()
	if err != nil {
		t.Fatalf("setup: %#v", err)
	}

	for _, test := range []struct {
		name        string
		addr        string
		expect      string
		expectError bool
	}{
		{"Unix scheme, file exists", emptySock, emptySock, false},
		{"Unix scheme, file is symlink", emptyAlsoSock, emptySock, false},
		{"Missing file", "unix://./testdata/non-such.sock", "", true},
		{"HTTPS scheme", "https://example.com", "https://example.com", false},
		{"Malformed URI", "/tmp/vind.sock\n", "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseAddr(test.addr)
			if err == nil && test.expectError {
				t.Errorf("expected error")
			} else if err != nil && !test.expectError {
				t.Errorf("unexpected error %#v", err)
			}

			if test.expect != got {
				t.Errorf("expected\n%q\nreceived\n%q", test.expect, got)
			}
		})
	}
}

// create empty.sock as an empty file
// create emptyAlsoSock as a symlink from empty-also.sock -> empty.sock
func setupParseAddr() (emptySock, emptyAlsoSock string, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	dir, err := ioutil.TempDir(filepath.Join(cwd, "testdata/tmp"), "")
	if err != nil {
		return
	}

	emptySockRaw := filepath.Join(dir, "empty.sock")
	emptySock = fmt.Sprintf("unix://%s", emptySockRaw)

	emptyAlsoSockRaw := filepath.Join(dir, "empty-also.sock")
	emptyAlsoSock = fmt.Sprintf("unix://%s", emptyAlsoSockRaw)

	err = ioutil.WriteFile(emptySockRaw, []byte{}, 0644)
	if err != nil {
		return
	}

	err = os.Symlink(emptySockRaw, emptyAlsoSockRaw)

	return
}
