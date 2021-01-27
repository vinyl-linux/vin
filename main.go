package main

import (
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/vinyl-linux/vin/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("unix", sockAddr)
	if err != nil {
		panic(err)
	}

	Setup().Serve(lis)
}

// Setup configures grpc servers, handles startup conditions, and returns
// a grpc server.
//
// This function cynically exists as a way to avoid dropping code coverage
func Setup() *grpc.Server {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	sugar := logger.Sugar()

	sugar.Info("starting")
	sugar.Info("loading config")

	c, err := LoadConfig()
	if err != nil {
		sugar.Panic(err)
	}

	sugar.Info("loaded")
	sugar.Info("loading manifests")

	mdb, err := LoadDB()
	if err != nil {
		sugar.Panic(err)
	}

	sugar.Info("loaded")
	sugar.Info("loading state")

	sdb, err := LoadStateDB()
	if err != nil {
		sugar.Panic(err)
	}

	sugar.Info("loaded")
	sugar.Info("starting server")

	s, err := NewServer(c, mdb, sdb)
	if err != nil {
		sugar.Panic(err)
	}

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_zap.StreamServerInterceptor(logger),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(logger),
		)),
	)

	server.RegisterVinServer(grpcServer, s)
	reflection.Register(grpcServer)

	sugar.Info("serving vin server :-)")

	return grpcServer
}
