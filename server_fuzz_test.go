// +build fuzz

package main

import (
	"context"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/gofuzz"
	"github.com/vinyl-linux/vin/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const (
	bufSize = 1024 * 1024
	fuzzDur = time.Minute * 5
)

var (
	lis       *bufconn.Listener
	fuzzConc  int
	fuzzCount int64
	fuzzWG    sync.WaitGroup
)

func init() {
	lis = bufconn.Listen(bufSize)

	pkgDir = "testdata/manifests/valid-manifests"
	configFile = "testdata/test-config.toml"
	cacheDir = "/tmp"
	stateDB = "/tmp/vinyl-fuzz.db"

	// Disable logs for fuzz test (otherwise it's super verbose and we miss anything useful)
	logger = zap.NewNop()

	s := Setup()
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()

	// set concurrent fuzz threads
	switch n := runtime.NumCPU(); n {
	case 1, 2:
		fuzzConc = 1

	case 3, 4, 5, 6, 7, 8:
		fuzzConc = int(n / 2)

	default:
		fuzzConc = 4
	}
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestFuzzServer_Install(t *testing.T) {
	end := time.Now().Add(fuzzDur)
	fuzzWG.Add(fuzzConc)

	t.Logf("fuzzing Install until %v", end)
	t.Logf("with %d workers", fuzzConc)

	for i := 0; i < fuzzConc; i++ {
		go fuzzLoop(t, end)
	}

	fuzzWG.Wait()

	t.Logf("ran %d tests", fuzzCount)
}

func fuzzLoop(t *testing.T, end time.Time) {
	defer func() {
		err := recover()

		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	}()

	defer fuzzWG.Done()

	f := fuzz.New().Funcs(
		func(is *server.InstallSpec, c fuzz.Continue) {
			switch c.Intn(3) {
			case 0:
				is.Version = ""
			case 1:
				is.Version = "1.0.1"
			default:
				c.Fuzz(&is.Version)
			}

			c.Fuzz(&is.Pkg)
			c.Fuzz(&is.Force)
		},
	)

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := server.NewVinClient(conn)

	for {
		if time.Now().After(end) {
			return
		}

		is := &server.InstallSpec{}

		f.Fuzz(is)

		resp, err := client.Install(context.Background(), is)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
		for {
			_, err = resp.Recv()
			if err != nil {
				goto nxt
			}
		}

	nxt:
		atomic.AddInt64(&fuzzCount, 1)
	}

}
