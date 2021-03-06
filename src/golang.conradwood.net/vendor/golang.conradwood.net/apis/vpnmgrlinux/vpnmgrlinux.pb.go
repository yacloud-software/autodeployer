// Code generated by protoc-gen-go.
// source: golang.conradwood.net/apis/vpnmgrlinux/vpnmgrlinux.proto
// DO NOT EDIT!

/*
Package vpnmgrlinux is a generated protocol buffer package.

It is generated from these files:
	golang.conradwood.net/apis/vpnmgrlinux/vpnmgrlinux.proto

It has these top-level messages:
	VPNUpDownRequest
	VPNStatusResponse
	Route
	VPNRoutes
	EmptyResponse
*/
package vpnmgrlinux

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type VPNUpDownRequest struct {
	VPNID  uint64 `protobuf:"varint,1,opt,name=VPNID" json:"VPNID,omitempty"`
	CameUp bool   `protobuf:"varint,2,opt,name=CameUp" json:"CameUp,omitempty"`
}

func (m *VPNUpDownRequest) Reset()                    { *m = VPNUpDownRequest{} }
func (m *VPNUpDownRequest) String() string            { return proto.CompactTextString(m) }
func (*VPNUpDownRequest) ProtoMessage()               {}
func (*VPNUpDownRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *VPNUpDownRequest) GetVPNID() uint64 {
	if m != nil {
		return m.VPNID
	}
	return 0
}

func (m *VPNUpDownRequest) GetCameUp() bool {
	if m != nil {
		return m.CameUp
	}
	return false
}

type VPNStatusResponse struct {
	VPN []*VPNUpDownRequest `protobuf:"bytes,1,rep,name=VPN" json:"VPN,omitempty"`
}

func (m *VPNStatusResponse) Reset()                    { *m = VPNStatusResponse{} }
func (m *VPNStatusResponse) String() string            { return proto.CompactTextString(m) }
func (*VPNStatusResponse) ProtoMessage()               {}
func (*VPNStatusResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *VPNStatusResponse) GetVPN() []*VPNUpDownRequest {
	if m != nil {
		return m.VPN
	}
	return nil
}

type Route struct {
	Ip   string `protobuf:"bytes,1,opt,name=Ip" json:"Ip,omitempty"`
	Size uint32 `protobuf:"varint,2,opt,name=Size" json:"Size,omitempty"`
}

func (m *Route) Reset()                    { *m = Route{} }
func (m *Route) String() string            { return proto.CompactTextString(m) }
func (*Route) ProtoMessage()               {}
func (*Route) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Route) GetIp() string {
	if m != nil {
		return m.Ip
	}
	return ""
}

func (m *Route) GetSize() uint32 {
	if m != nil {
		return m.Size
	}
	return 0
}

type VPNRoutes struct {
	Routes []*Route `protobuf:"bytes,1,rep,name=Routes" json:"Routes,omitempty"`
}

func (m *VPNRoutes) Reset()                    { *m = VPNRoutes{} }
func (m *VPNRoutes) String() string            { return proto.CompactTextString(m) }
func (*VPNRoutes) ProtoMessage()               {}
func (*VPNRoutes) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *VPNRoutes) GetRoutes() []*Route {
	if m != nil {
		return m.Routes
	}
	return nil
}

type EmptyResponse struct {
}

func (m *EmptyResponse) Reset()                    { *m = EmptyResponse{} }
func (m *EmptyResponse) String() string            { return proto.CompactTextString(m) }
func (*EmptyResponse) ProtoMessage()               {}
func (*EmptyResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func init() {
	proto.RegisterType((*VPNUpDownRequest)(nil), "vpnmgrlinux.VPNUpDownRequest")
	proto.RegisterType((*VPNStatusResponse)(nil), "vpnmgrlinux.VPNStatusResponse")
	proto.RegisterType((*Route)(nil), "vpnmgrlinux.Route")
	proto.RegisterType((*VPNRoutes)(nil), "vpnmgrlinux.VPNRoutes")
	proto.RegisterType((*EmptyResponse)(nil), "vpnmgrlinux.EmptyResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for VpnMgrLinux service

type VpnMgrLinuxClient interface {
	VPNUpDown(ctx context.Context, in *VPNUpDownRequest, opts ...grpc.CallOption) (*EmptyResponse, error)
	GetVPNStatus(ctx context.Context, in *EmptyResponse, opts ...grpc.CallOption) (*VPNStatusResponse, error)
	// return the routes to all the VPNs that are currently "up"
	GetVPNRoutes(ctx context.Context, in *EmptyResponse, opts ...grpc.CallOption) (*VPNRoutes, error)
}

type vpnMgrLinuxClient struct {
	cc *grpc.ClientConn
}

func NewVpnMgrLinuxClient(cc *grpc.ClientConn) VpnMgrLinuxClient {
	return &vpnMgrLinuxClient{cc}
}

func (c *vpnMgrLinuxClient) VPNUpDown(ctx context.Context, in *VPNUpDownRequest, opts ...grpc.CallOption) (*EmptyResponse, error) {
	out := new(EmptyResponse)
	err := grpc.Invoke(ctx, "/vpnmgrlinux.VpnMgrLinux/VPNUpDown", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *vpnMgrLinuxClient) GetVPNStatus(ctx context.Context, in *EmptyResponse, opts ...grpc.CallOption) (*VPNStatusResponse, error) {
	out := new(VPNStatusResponse)
	err := grpc.Invoke(ctx, "/vpnmgrlinux.VpnMgrLinux/GetVPNStatus", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *vpnMgrLinuxClient) GetVPNRoutes(ctx context.Context, in *EmptyResponse, opts ...grpc.CallOption) (*VPNRoutes, error) {
	out := new(VPNRoutes)
	err := grpc.Invoke(ctx, "/vpnmgrlinux.VpnMgrLinux/GetVPNRoutes", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for VpnMgrLinux service

type VpnMgrLinuxServer interface {
	VPNUpDown(context.Context, *VPNUpDownRequest) (*EmptyResponse, error)
	GetVPNStatus(context.Context, *EmptyResponse) (*VPNStatusResponse, error)
	// return the routes to all the VPNs that are currently "up"
	GetVPNRoutes(context.Context, *EmptyResponse) (*VPNRoutes, error)
}

func RegisterVpnMgrLinuxServer(s *grpc.Server, srv VpnMgrLinuxServer) {
	s.RegisterService(&_VpnMgrLinux_serviceDesc, srv)
}

func _VpnMgrLinux_VPNUpDown_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VPNUpDownRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VpnMgrLinuxServer).VPNUpDown(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/vpnmgrlinux.VpnMgrLinux/VPNUpDown",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VpnMgrLinuxServer).VPNUpDown(ctx, req.(*VPNUpDownRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _VpnMgrLinux_GetVPNStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyResponse)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VpnMgrLinuxServer).GetVPNStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/vpnmgrlinux.VpnMgrLinux/GetVPNStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VpnMgrLinuxServer).GetVPNStatus(ctx, req.(*EmptyResponse))
	}
	return interceptor(ctx, in, info, handler)
}

func _VpnMgrLinux_GetVPNRoutes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyResponse)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VpnMgrLinuxServer).GetVPNRoutes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/vpnmgrlinux.VpnMgrLinux/GetVPNRoutes",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VpnMgrLinuxServer).GetVPNRoutes(ctx, req.(*EmptyResponse))
	}
	return interceptor(ctx, in, info, handler)
}

var _VpnMgrLinux_serviceDesc = grpc.ServiceDesc{
	ServiceName: "vpnmgrlinux.VpnMgrLinux",
	HandlerType: (*VpnMgrLinuxServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "VPNUpDown",
			Handler:    _VpnMgrLinux_VPNUpDown_Handler,
		},
		{
			MethodName: "GetVPNStatus",
			Handler:    _VpnMgrLinux_GetVPNStatus_Handler,
		},
		{
			MethodName: "GetVPNRoutes",
			Handler:    _VpnMgrLinux_GetVPNRoutes_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "golang.conradwood.net/apis/vpnmgrlinux/vpnmgrlinux.proto",
}

func init() {
	proto.RegisterFile("golang.conradwood.net/apis/vpnmgrlinux/vpnmgrlinux.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 325 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x7c, 0x52, 0x4f, 0x4b, 0xfb, 0x40,
	0x10, 0x25, 0xfd, 0xc7, 0xaf, 0xd3, 0x5f, 0xfd, 0x33, 0x48, 0x29, 0x05, 0xa5, 0xe4, 0x20, 0x41,
	0x21, 0x85, 0x7a, 0xd0, 0xa3, 0xd4, 0xaa, 0x54, 0x34, 0x84, 0x2d, 0xdd, 0x7b, 0xb4, 0x4b, 0x28,
	0x34, 0xbb, 0x6b, 0x76, 0x63, 0xd5, 0xef, 0xeb, 0xf7, 0x90, 0x6c, 0x52, 0x49, 0x72, 0xc8, 0x6d,
	0x66, 0xe7, 0xcd, 0x7b, 0x6f, 0x87, 0x07, 0x37, 0xa1, 0xd8, 0x06, 0x3c, 0x74, 0xdf, 0x04, 0x8f,
	0x83, 0xf5, 0x4e, 0x88, 0xb5, 0xcb, 0x99, 0x9e, 0x04, 0x72, 0xa3, 0x26, 0x1f, 0x92, 0x47, 0x61,
	0xbc, 0xdd, 0xf0, 0xe4, 0xb3, 0x58, 0xbb, 0x32, 0x16, 0x5a, 0x60, 0xaf, 0xf0, 0x64, 0xdf, 0xc2,
	0x11, 0xf5, 0xbd, 0x95, 0x9c, 0x8b, 0x1d, 0x27, 0xec, 0x3d, 0x61, 0x4a, 0xe3, 0x09, 0xb4, 0xa9,
	0xef, 0x2d, 0xe6, 0x43, 0x6b, 0x6c, 0x39, 0x2d, 0x92, 0x35, 0x38, 0x80, 0xce, 0x5d, 0x10, 0xb1,
	0x95, 0x1c, 0x36, 0xc6, 0x96, 0xf3, 0x8f, 0xe4, 0x9d, 0x3d, 0x87, 0x63, 0xea, 0x7b, 0x4b, 0x1d,
	0xe8, 0x44, 0x11, 0xa6, 0xa4, 0xe0, 0x8a, 0xe1, 0x04, 0x9a, 0xd4, 0xf7, 0x86, 0xd6, 0xb8, 0xe9,
	0xf4, 0xa6, 0xa7, 0x6e, 0xd1, 0x44, 0x55, 0x8e, 0xa4, 0x48, 0xfb, 0x12, 0xda, 0x44, 0x24, 0x9a,
	0xe1, 0x01, 0x34, 0x16, 0xd2, 0x28, 0x77, 0x49, 0x63, 0x21, 0x11, 0xa1, 0xb5, 0xdc, 0x7c, 0x33,
	0x23, 0xda, 0x27, 0xa6, 0xb6, 0xaf, 0xa1, 0x4b, 0x7d, 0xcf, 0xe0, 0x15, 0x5e, 0x40, 0x27, 0xab,
	0x72, 0x35, 0x2c, 0xa9, 0x99, 0x11, 0xc9, 0x11, 0xf6, 0x21, 0xf4, 0xef, 0x23, 0xa9, 0xbf, 0xf6,
	0x3e, 0xa7, 0x3f, 0x16, 0xf4, 0xa8, 0xe4, 0x2f, 0x61, 0xfc, 0x9c, 0xc2, 0xf1, 0xc1, 0x30, 0x67,
	0xfe, 0xb0, 0xde, 0xf7, 0x68, 0x54, 0x1a, 0x97, 0x78, 0xf1, 0x09, 0xfe, 0x3f, 0x32, 0xfd, 0x77,
	0x17, 0xac, 0xc1, 0x8e, 0xce, 0xaa, 0x32, 0x95, 0x5b, 0xce, 0xf6, 0x5c, 0xf9, 0x87, 0xeb, 0xb8,
	0x06, 0x55, 0xae, 0x6c, 0x67, 0xe6, 0xc0, 0x39, 0x67, 0xba, 0x18, 0x96, 0x3c, 0x3e, 0x69, 0x5e,
	0x8a, 0x3b, 0xaf, 0x1d, 0x13, 0x92, 0xab, 0xdf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x3a, 0x74, 0xf2,
	0x83, 0x60, 0x02, 0x00, 0x00,
}
