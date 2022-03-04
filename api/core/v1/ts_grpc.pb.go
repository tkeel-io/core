// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// TSClient is the client API for TS service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TSClient interface {
	GetTSData(ctx context.Context, in *GetTSDataRequest, opts ...grpc.CallOption) (*GetTSDataResponse, error)
	DownloadTSData(ctx context.Context, in *DownloadTSDataRequest, opts ...grpc.CallOption) (*DownloadTSDataResponse, error)
	GetLatestEntities(ctx context.Context, in *GetLatestEntitiesRequest, opts ...grpc.CallOption) (*GetLatestEntitiesResponse, error)
}

type tSClient struct {
	cc grpc.ClientConnInterface
}

func NewTSClient(cc grpc.ClientConnInterface) TSClient {
	return &tSClient{cc}
}

func (c *tSClient) GetTSData(ctx context.Context, in *GetTSDataRequest, opts ...grpc.CallOption) (*GetTSDataResponse, error) {
	out := new(GetTSDataResponse)
	err := c.cc.Invoke(ctx, "/api.core.v1.TS/GetTSData", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tSClient) DownloadTSData(ctx context.Context, in *DownloadTSDataRequest, opts ...grpc.CallOption) (*DownloadTSDataResponse, error) {
	out := new(DownloadTSDataResponse)
	err := c.cc.Invoke(ctx, "/api.core.v1.TS/DownloadTSData", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tSClient) GetLatestEntities(ctx context.Context, in *GetLatestEntitiesRequest, opts ...grpc.CallOption) (*GetLatestEntitiesResponse, error) {
	out := new(GetLatestEntitiesResponse)
	err := c.cc.Invoke(ctx, "/api.core.v1.TS/GetLatestEntities", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TSServer is the server API for TS service.
// All implementations must embed UnimplementedTSServer
// for forward compatibility
type TSServer interface {
	GetTSData(context.Context, *GetTSDataRequest) (*GetTSDataResponse, error)
	DownloadTSData(context.Context, *DownloadTSDataRequest) (*DownloadTSDataResponse, error)
	GetLatestEntities(context.Context, *GetLatestEntitiesRequest) (*GetLatestEntitiesResponse, error)
	mustEmbedUnimplementedTSServer()
}

// UnimplementedTSServer must be embedded to have forward compatible implementations.
type UnimplementedTSServer struct {
}

func (UnimplementedTSServer) GetTSData(context.Context, *GetTSDataRequest) (*GetTSDataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTSData not implemented")
}
func (UnimplementedTSServer) DownloadTSData(context.Context, *DownloadTSDataRequest) (*DownloadTSDataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DownloadTSData not implemented")
}
func (UnimplementedTSServer) GetLatestEntities(context.Context, *GetLatestEntitiesRequest) (*GetLatestEntitiesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLatestEntities not implemented")
}
func (UnimplementedTSServer) mustEmbedUnimplementedTSServer() {}

// UnsafeTSServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TSServer will
// result in compilation errors.
type UnsafeTSServer interface {
	mustEmbedUnimplementedTSServer()
}

func RegisterTSServer(s grpc.ServiceRegistrar, srv TSServer) {
	s.RegisterService(&TS_ServiceDesc, srv)
}

func _TS_GetTSData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTSDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TSServer).GetTSData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.core.v1.TS/GetTSData",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TSServer).GetTSData(ctx, req.(*GetTSDataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TS_DownloadTSData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DownloadTSDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TSServer).DownloadTSData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.core.v1.TS/DownloadTSData",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TSServer).DownloadTSData(ctx, req.(*DownloadTSDataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TS_GetLatestEntities_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetLatestEntitiesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TSServer).GetLatestEntities(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.core.v1.TS/GetLatestEntities",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TSServer).GetLatestEntities(ctx, req.(*GetLatestEntitiesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// TS_ServiceDesc is the grpc.ServiceDesc for TS service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TS_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.core.v1.TS",
	HandlerType: (*TSServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetTSData",
			Handler:    _TS_GetTSData_Handler,
		},
		{
			MethodName: "DownloadTSData",
			Handler:    _TS_DownloadTSData_Handler,
		},
		{
			MethodName: "GetLatestEntities",
			Handler:    _TS_GetLatestEntities_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/core/v1/ts.proto",
}