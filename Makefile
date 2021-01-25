server/:
	mkdir -p $@

server/install.pb.go server/server.pb.go server/server_grpc.pb.go: server/
	protoc --proto_path=proto --go_out=server --go-grpc_out=server --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative install.proto server.proto
