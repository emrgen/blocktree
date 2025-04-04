// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: apis/v1/blocktree.proto

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

const (
	Blocktree_Apply_FullMethodName          = "/apis.v1.Blocktree/Apply"
	Blocktree_CreateSpace_FullMethodName    = "/apis.v1.Blocktree/CreateSpace"
	Blocktree_GetBlock_FullMethodName       = "/apis.v1.Blocktree/GetBlock"
	Blocktree_GetChildren_FullMethodName    = "/apis.v1.Blocktree/GetChildren"
	Blocktree_GetDescendants_FullMethodName = "/apis.v1.Blocktree/GetDescendants"
	Blocktree_GetPage_FullMethodName        = "/apis.v1.Blocktree/GetPage"
	Blocktree_GetBackLinks_FullMethodName   = "/apis.v1.Blocktree/GetBackLinks"
	Blocktree_GetUpdates_FullMethodName     = "/apis.v1.Blocktree/GetUpdates"
)

// BlocktreeClient is the client API for Blocktree service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BlocktreeClient interface {
	Apply(ctx context.Context, in *TransactionsRequest, opts ...grpc.CallOption) (*TransactionsResponse, error)
	CreateSpace(ctx context.Context, in *CreateSpaceRequest, opts ...grpc.CallOption) (*CreateSpaceResponse, error)
	GetBlock(ctx context.Context, in *GetBlockRequest, opts ...grpc.CallOption) (*GetBlockResponse, error)
	GetChildren(ctx context.Context, in *GetBlockChildrenRequest, opts ...grpc.CallOption) (*GetBlockChildrenResponse, error)
	GetDescendants(ctx context.Context, in *GetBlockDescendantsRequest, opts ...grpc.CallOption) (*GetBlockDescendantsResponse, error)
	GetPage(ctx context.Context, in *GetBlockPageRequest, opts ...grpc.CallOption) (*GetBlockPageResponse, error)
	GetBackLinks(ctx context.Context, in *GetBackLinksRequest, opts ...grpc.CallOption) (*GetBackLinksResponse, error)
	GetUpdates(ctx context.Context, in *GetUpdatesRequest, opts ...grpc.CallOption) (*GetUpdatesResponse, error)
}

type blocktreeClient struct {
	cc grpc.ClientConnInterface
}

func NewBlocktreeClient(cc grpc.ClientConnInterface) BlocktreeClient {
	return &blocktreeClient{cc}
}

func (c *blocktreeClient) Apply(ctx context.Context, in *TransactionsRequest, opts ...grpc.CallOption) (*TransactionsResponse, error) {
	out := new(TransactionsResponse)
	err := c.cc.Invoke(ctx, Blocktree_Apply_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocktreeClient) CreateSpace(ctx context.Context, in *CreateSpaceRequest, opts ...grpc.CallOption) (*CreateSpaceResponse, error) {
	out := new(CreateSpaceResponse)
	err := c.cc.Invoke(ctx, Blocktree_CreateSpace_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocktreeClient) GetBlock(ctx context.Context, in *GetBlockRequest, opts ...grpc.CallOption) (*GetBlockResponse, error) {
	out := new(GetBlockResponse)
	err := c.cc.Invoke(ctx, Blocktree_GetBlock_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocktreeClient) GetChildren(ctx context.Context, in *GetBlockChildrenRequest, opts ...grpc.CallOption) (*GetBlockChildrenResponse, error) {
	out := new(GetBlockChildrenResponse)
	err := c.cc.Invoke(ctx, Blocktree_GetChildren_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocktreeClient) GetDescendants(ctx context.Context, in *GetBlockDescendantsRequest, opts ...grpc.CallOption) (*GetBlockDescendantsResponse, error) {
	out := new(GetBlockDescendantsResponse)
	err := c.cc.Invoke(ctx, Blocktree_GetDescendants_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocktreeClient) GetPage(ctx context.Context, in *GetBlockPageRequest, opts ...grpc.CallOption) (*GetBlockPageResponse, error) {
	out := new(GetBlockPageResponse)
	err := c.cc.Invoke(ctx, Blocktree_GetPage_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocktreeClient) GetBackLinks(ctx context.Context, in *GetBackLinksRequest, opts ...grpc.CallOption) (*GetBackLinksResponse, error) {
	out := new(GetBackLinksResponse)
	err := c.cc.Invoke(ctx, Blocktree_GetBackLinks_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocktreeClient) GetUpdates(ctx context.Context, in *GetUpdatesRequest, opts ...grpc.CallOption) (*GetUpdatesResponse, error) {
	out := new(GetUpdatesResponse)
	err := c.cc.Invoke(ctx, Blocktree_GetUpdates_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BlocktreeServer is the server API for Blocktree service.
// All implementations must embed UnimplementedBlocktreeServer
// for forward compatibility
type BlocktreeServer interface {
	Apply(context.Context, *TransactionsRequest) (*TransactionsResponse, error)
	CreateSpace(context.Context, *CreateSpaceRequest) (*CreateSpaceResponse, error)
	GetBlock(context.Context, *GetBlockRequest) (*GetBlockResponse, error)
	GetChildren(context.Context, *GetBlockChildrenRequest) (*GetBlockChildrenResponse, error)
	GetDescendants(context.Context, *GetBlockDescendantsRequest) (*GetBlockDescendantsResponse, error)
	GetPage(context.Context, *GetBlockPageRequest) (*GetBlockPageResponse, error)
	GetBackLinks(context.Context, *GetBackLinksRequest) (*GetBackLinksResponse, error)
	GetUpdates(context.Context, *GetUpdatesRequest) (*GetUpdatesResponse, error)
	mustEmbedUnimplementedBlocktreeServer()
}

// UnimplementedBlocktreeServer must be embedded to have forward compatible implementations.
type UnimplementedBlocktreeServer struct {
}

func (UnimplementedBlocktreeServer) Apply(context.Context, *TransactionsRequest) (*TransactionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Apply not implemented")
}
func (UnimplementedBlocktreeServer) CreateSpace(context.Context, *CreateSpaceRequest) (*CreateSpaceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateSpace not implemented")
}
func (UnimplementedBlocktreeServer) GetBlock(context.Context, *GetBlockRequest) (*GetBlockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlock not implemented")
}
func (UnimplementedBlocktreeServer) GetChildren(context.Context, *GetBlockChildrenRequest) (*GetBlockChildrenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetChildren not implemented")
}
func (UnimplementedBlocktreeServer) GetDescendants(context.Context, *GetBlockDescendantsRequest) (*GetBlockDescendantsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDescendants not implemented")
}
func (UnimplementedBlocktreeServer) GetPage(context.Context, *GetBlockPageRequest) (*GetBlockPageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPage not implemented")
}
func (UnimplementedBlocktreeServer) GetBackLinks(context.Context, *GetBackLinksRequest) (*GetBackLinksResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBackLinks not implemented")
}
func (UnimplementedBlocktreeServer) GetUpdates(context.Context, *GetUpdatesRequest) (*GetUpdatesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUpdates not implemented")
}
func (UnimplementedBlocktreeServer) mustEmbedUnimplementedBlocktreeServer() {}

// UnsafeBlocktreeServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BlocktreeServer will
// result in compilation errors.
type UnsafeBlocktreeServer interface {
	mustEmbedUnimplementedBlocktreeServer()
}

func RegisterBlocktreeServer(s grpc.ServiceRegistrar, srv BlocktreeServer) {
	s.RegisterService(&Blocktree_ServiceDesc, srv)
}

func _Blocktree_Apply_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TransactionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocktreeServer).Apply(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Blocktree_Apply_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocktreeServer).Apply(ctx, req.(*TransactionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Blocktree_CreateSpace_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateSpaceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocktreeServer).CreateSpace(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Blocktree_CreateSpace_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocktreeServer).CreateSpace(ctx, req.(*CreateSpaceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Blocktree_GetBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocktreeServer).GetBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Blocktree_GetBlock_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocktreeServer).GetBlock(ctx, req.(*GetBlockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Blocktree_GetChildren_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlockChildrenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocktreeServer).GetChildren(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Blocktree_GetChildren_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocktreeServer).GetChildren(ctx, req.(*GetBlockChildrenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Blocktree_GetDescendants_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlockDescendantsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocktreeServer).GetDescendants(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Blocktree_GetDescendants_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocktreeServer).GetDescendants(ctx, req.(*GetBlockDescendantsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Blocktree_GetPage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlockPageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocktreeServer).GetPage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Blocktree_GetPage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocktreeServer).GetPage(ctx, req.(*GetBlockPageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Blocktree_GetBackLinks_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBackLinksRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocktreeServer).GetBackLinks(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Blocktree_GetBackLinks_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocktreeServer).GetBackLinks(ctx, req.(*GetBackLinksRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Blocktree_GetUpdates_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUpdatesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocktreeServer).GetUpdates(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Blocktree_GetUpdates_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocktreeServer).GetUpdates(ctx, req.(*GetUpdatesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Blocktree_ServiceDesc is the grpc.ServiceDesc for Blocktree service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Blocktree_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "apis.v1.Blocktree",
	HandlerType: (*BlocktreeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Apply",
			Handler:    _Blocktree_Apply_Handler,
		},
		{
			MethodName: "CreateSpace",
			Handler:    _Blocktree_CreateSpace_Handler,
		},
		{
			MethodName: "GetBlock",
			Handler:    _Blocktree_GetBlock_Handler,
		},
		{
			MethodName: "GetChildren",
			Handler:    _Blocktree_GetChildren_Handler,
		},
		{
			MethodName: "GetDescendants",
			Handler:    _Blocktree_GetDescendants_Handler,
		},
		{
			MethodName: "GetPage",
			Handler:    _Blocktree_GetPage_Handler,
		},
		{
			MethodName: "GetBackLinks",
			Handler:    _Blocktree_GetBackLinks_Handler,
		},
		{
			MethodName: "GetUpdates",
			Handler:    _Blocktree_GetUpdates_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "apis/v1/blocktree.proto",
}
