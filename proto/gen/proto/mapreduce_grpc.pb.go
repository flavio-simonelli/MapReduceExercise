// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.21.12
// source: proto/mapreduce.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	MapReduceService_Mapping_FullMethodName  = "/MapReduceExercise.MapReduceService/Mapping"
	MapReduceService_Reducing_FullMethodName = "/MapReduceExercise.MapReduceService/Reducing"
)

// MapReduceServiceClient is the client API for MapReduceService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MapReduceServiceClient interface {
	Mapping(ctx context.Context, in *Chunk, opts ...grpc.CallOption) (*Response, error)
	Reducing(ctx context.Context, in *Chunk, opts ...grpc.CallOption) (*Response, error)
}

type mapReduceServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMapReduceServiceClient(cc grpc.ClientConnInterface) MapReduceServiceClient {
	return &mapReduceServiceClient{cc}
}

func (c *mapReduceServiceClient) Mapping(ctx context.Context, in *Chunk, opts ...grpc.CallOption) (*Response, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Response)
	err := c.cc.Invoke(ctx, MapReduceService_Mapping_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mapReduceServiceClient) Reducing(ctx context.Context, in *Chunk, opts ...grpc.CallOption) (*Response, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Response)
	err := c.cc.Invoke(ctx, MapReduceService_Reducing_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MapReduceServiceServer is the server API for MapReduceService service.
// All implementations must embed UnimplementedMapReduceServiceServer
// for forward compatibility.
type MapReduceServiceServer interface {
	Mapping(context.Context, *Chunk) (*Response, error)
	Reducing(context.Context, *Chunk) (*Response, error)
	mustEmbedUnimplementedMapReduceServiceServer()
}

// UnimplementedMapReduceServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedMapReduceServiceServer struct{}

func (UnimplementedMapReduceServiceServer) Mapping(context.Context, *Chunk) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Mapping not implemented")
}
func (UnimplementedMapReduceServiceServer) Reducing(context.Context, *Chunk) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Reducing not implemented")
}
func (UnimplementedMapReduceServiceServer) mustEmbedUnimplementedMapReduceServiceServer() {}
func (UnimplementedMapReduceServiceServer) testEmbeddedByValue()                          {}

// UnsafeMapReduceServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MapReduceServiceServer will
// result in compilation errors.
type UnsafeMapReduceServiceServer interface {
	mustEmbedUnimplementedMapReduceServiceServer()
}

func RegisterMapReduceServiceServer(s grpc.ServiceRegistrar, srv MapReduceServiceServer) {
	// If the following call pancis, it indicates UnimplementedMapReduceServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&MapReduceService_ServiceDesc, srv)
}

func _MapReduceService_Mapping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Chunk)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MapReduceServiceServer).Mapping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MapReduceService_Mapping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MapReduceServiceServer).Mapping(ctx, req.(*Chunk))
	}
	return interceptor(ctx, in, info, handler)
}

func _MapReduceService_Reducing_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Chunk)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MapReduceServiceServer).Reducing(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MapReduceService_Reducing_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MapReduceServiceServer).Reducing(ctx, req.(*Chunk))
	}
	return interceptor(ctx, in, info, handler)
}

// MapReduceService_ServiceDesc is the grpc.ServiceDesc for MapReduceService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MapReduceService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "MapReduceExercise.MapReduceService",
	HandlerType: (*MapReduceServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Mapping",
			Handler:    _MapReduceService_Mapping_Handler,
		},
		{
			MethodName: "Reducing",
			Handler:    _MapReduceService_Reducing_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/mapreduce.proto",
}