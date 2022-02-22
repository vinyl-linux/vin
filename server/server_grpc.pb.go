// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.1
// source: server.proto

package server

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// VinClient is the client API for Vin service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type VinClient interface {
	Install(ctx context.Context, in *InstallSpec, opts ...grpc.CallOption) (Vin_InstallClient, error)
	Reload(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (Vin_ReloadClient, error)
	Version(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*VersionMessage, error)
}

type vinClient struct {
	cc grpc.ClientConnInterface
}

func NewVinClient(cc grpc.ClientConnInterface) VinClient {
	return &vinClient{cc}
}

func (c *vinClient) Install(ctx context.Context, in *InstallSpec, opts ...grpc.CallOption) (Vin_InstallClient, error) {
	stream, err := c.cc.NewStream(ctx, &Vin_ServiceDesc.Streams[0], "/server.Vin/Install", opts...)
	if err != nil {
		return nil, err
	}
	x := &vinInstallClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Vin_InstallClient interface {
	Recv() (*Output, error)
	grpc.ClientStream
}

type vinInstallClient struct {
	grpc.ClientStream
}

func (x *vinInstallClient) Recv() (*Output, error) {
	m := new(Output)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *vinClient) Reload(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (Vin_ReloadClient, error) {
	stream, err := c.cc.NewStream(ctx, &Vin_ServiceDesc.Streams[1], "/server.Vin/Reload", opts...)
	if err != nil {
		return nil, err
	}
	x := &vinReloadClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Vin_ReloadClient interface {
	Recv() (*Output, error)
	grpc.ClientStream
}

type vinReloadClient struct {
	grpc.ClientStream
}

func (x *vinReloadClient) Recv() (*Output, error) {
	m := new(Output)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *vinClient) Version(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*VersionMessage, error) {
	out := new(VersionMessage)
	err := c.cc.Invoke(ctx, "/server.Vin/Version", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// VinServer is the server API for Vin service.
// All implementations must embed UnimplementedVinServer
// for forward compatibility
type VinServer interface {
	Install(*InstallSpec, Vin_InstallServer) error
	Reload(*emptypb.Empty, Vin_ReloadServer) error
	Version(context.Context, *emptypb.Empty) (*VersionMessage, error)
	mustEmbedUnimplementedVinServer()
}

// UnimplementedVinServer must be embedded to have forward compatible implementations.
type UnimplementedVinServer struct {
}

func (UnimplementedVinServer) Install(*InstallSpec, Vin_InstallServer) error {
	return status.Errorf(codes.Unimplemented, "method Install not implemented")
}
func (UnimplementedVinServer) Reload(*emptypb.Empty, Vin_ReloadServer) error {
	return status.Errorf(codes.Unimplemented, "method Reload not implemented")
}
func (UnimplementedVinServer) Version(context.Context, *emptypb.Empty) (*VersionMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Version not implemented")
}
func (UnimplementedVinServer) mustEmbedUnimplementedVinServer() {}

// UnsafeVinServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to VinServer will
// result in compilation errors.
type UnsafeVinServer interface {
	mustEmbedUnimplementedVinServer()
}

func RegisterVinServer(s grpc.ServiceRegistrar, srv VinServer) {
	s.RegisterService(&Vin_ServiceDesc, srv)
}

func _Vin_Install_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(InstallSpec)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(VinServer).Install(m, &vinInstallServer{stream})
}

type Vin_InstallServer interface {
	Send(*Output) error
	grpc.ServerStream
}

type vinInstallServer struct {
	grpc.ServerStream
}

func (x *vinInstallServer) Send(m *Output) error {
	return x.ServerStream.SendMsg(m)
}

func _Vin_Reload_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(emptypb.Empty)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(VinServer).Reload(m, &vinReloadServer{stream})
}

type Vin_ReloadServer interface {
	Send(*Output) error
	grpc.ServerStream
}

type vinReloadServer struct {
	grpc.ServerStream
}

func (x *vinReloadServer) Send(m *Output) error {
	return x.ServerStream.SendMsg(m)
}

func _Vin_Version_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VinServer).Version(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/server.Vin/Version",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VinServer).Version(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// Vin_ServiceDesc is the grpc.ServiceDesc for Vin service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Vin_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "server.Vin",
	HandlerType: (*VinServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Version",
			Handler:    _Vin_Version_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Install",
			Handler:       _Vin_Install_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Reload",
			Handler:       _Vin_Reload_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "server.proto",
}
