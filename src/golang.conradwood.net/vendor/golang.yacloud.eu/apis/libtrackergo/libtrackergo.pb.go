// Code generated by protoc-gen-go.
// source: golang.yacloud.eu/apis/libtrackergo/libtrackergo.proto
// DO NOT EDIT!

/*
Package libtrackergo is a generated protocol buffer package.

It is generated from these files:
	golang.yacloud.eu/apis/libtrackergo/libtrackergo.proto

It has these top-level messages:
	ModuleCheckRequest
	ModuleCheckResponse
	RequestTracker
*/
package libtrackergo

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

type ModuleCheckRequest struct {
	FullName string `protobuf:"bytes,1,opt,name=FullName" json:"FullName,omitempty"`
	Version  string `protobuf:"bytes,2,opt,name=Version" json:"Version,omitempty"`
}

func (m *ModuleCheckRequest) Reset()                    { *m = ModuleCheckRequest{} }
func (m *ModuleCheckRequest) String() string            { return proto.CompactTextString(m) }
func (*ModuleCheckRequest) ProtoMessage()               {}
func (*ModuleCheckRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *ModuleCheckRequest) GetFullName() string {
	if m != nil {
		return m.FullName
	}
	return ""
}

func (m *ModuleCheckRequest) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

type ModuleCheckResponse struct {
	Granted bool `protobuf:"varint,1,opt,name=Granted" json:"Granted,omitempty"`
}

func (m *ModuleCheckResponse) Reset()                    { *m = ModuleCheckResponse{} }
func (m *ModuleCheckResponse) String() string            { return proto.CompactTextString(m) }
func (*ModuleCheckResponse) ProtoMessage()               {}
func (*ModuleCheckResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *ModuleCheckResponse) GetGranted() bool {
	if m != nil {
		return m.Granted
	}
	return false
}

type RequestTracker struct {
	ID       uint64 `protobuf:"varint,1,opt,name=ID" json:"ID,omitempty"`
	FullName string `protobuf:"bytes,2,opt,name=FullName" json:"FullName,omitempty"`
	Version  string `protobuf:"bytes,3,opt,name=Version" json:"Version,omitempty"`
	Created  uint32 `protobuf:"varint,4,opt,name=Created" json:"Created,omitempty"`
	UserID   string `protobuf:"bytes,5,opt,name=UserID" json:"UserID,omitempty"`
}

func (m *RequestTracker) Reset()                    { *m = RequestTracker{} }
func (m *RequestTracker) String() string            { return proto.CompactTextString(m) }
func (*RequestTracker) ProtoMessage()               {}
func (*RequestTracker) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *RequestTracker) GetID() uint64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *RequestTracker) GetFullName() string {
	if m != nil {
		return m.FullName
	}
	return ""
}

func (m *RequestTracker) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *RequestTracker) GetCreated() uint32 {
	if m != nil {
		return m.Created
	}
	return 0
}

func (m *RequestTracker) GetUserID() string {
	if m != nil {
		return m.UserID
	}
	return ""
}

func init() {
	proto.RegisterType((*ModuleCheckRequest)(nil), "libtrackergo.ModuleCheckRequest")
	proto.RegisterType((*ModuleCheckResponse)(nil), "libtrackergo.ModuleCheckResponse")
	proto.RegisterType((*RequestTracker)(nil), "libtrackergo.RequestTracker")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for LibTrackerGo service

type LibTrackerGoClient interface {
	Check(ctx context.Context, in *ModuleCheckRequest, opts ...grpc.CallOption) (*ModuleCheckResponse, error)
}

type libTrackerGoClient struct {
	cc *grpc.ClientConn
}

func NewLibTrackerGoClient(cc *grpc.ClientConn) LibTrackerGoClient {
	return &libTrackerGoClient{cc}
}

func (c *libTrackerGoClient) Check(ctx context.Context, in *ModuleCheckRequest, opts ...grpc.CallOption) (*ModuleCheckResponse, error) {
	out := new(ModuleCheckResponse)
	err := grpc.Invoke(ctx, "/libtrackergo.LibTrackerGo/Check", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for LibTrackerGo service

type LibTrackerGoServer interface {
	Check(context.Context, *ModuleCheckRequest) (*ModuleCheckResponse, error)
}

func RegisterLibTrackerGoServer(s *grpc.Server, srv LibTrackerGoServer) {
	s.RegisterService(&_LibTrackerGo_serviceDesc, srv)
}

func _LibTrackerGo_Check_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ModuleCheckRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LibTrackerGoServer).Check(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/libtrackergo.LibTrackerGo/Check",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LibTrackerGoServer).Check(ctx, req.(*ModuleCheckRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _LibTrackerGo_serviceDesc = grpc.ServiceDesc{
	ServiceName: "libtrackergo.LibTrackerGo",
	HandlerType: (*LibTrackerGoServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Check",
			Handler:    _LibTrackerGo_Check_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "golang.yacloud.eu/apis/libtrackergo/libtrackergo.proto",
}

func init() {
	proto.RegisterFile("golang.yacloud.eu/apis/libtrackergo/libtrackergo.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 270 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x7c, 0x91, 0xcd, 0x4a, 0xc3, 0x40,
	0x14, 0x85, 0x49, 0x6c, 0x6b, 0xbd, 0xd4, 0x2e, 0x46, 0x90, 0x50, 0x44, 0x62, 0x56, 0x5d, 0x25,
	0xa0, 0xe0, 0x03, 0xd8, 0x60, 0x89, 0x54, 0x17, 0x41, 0x5d, 0xb9, 0x99, 0x24, 0x97, 0x18, 0x3a,
	0xe6, 0xc6, 0x99, 0xcc, 0xc2, 0x37, 0xf0, 0xb1, 0x25, 0xd3, 0xa4, 0x38, 0x88, 0x5d, 0x7e, 0x73,
	0xff, 0xce, 0x9c, 0x03, 0xb7, 0x25, 0x09, 0x5e, 0x97, 0xe1, 0x17, 0xcf, 0x05, 0xe9, 0x22, 0x44,
	0x1d, 0xf1, 0xa6, 0x52, 0x91, 0xa8, 0xb2, 0x56, 0xf2, 0x7c, 0x8b, 0xb2, 0x24, 0x0b, 0xc2, 0x46,
	0x52, 0x4b, 0x6c, 0xf6, 0xfb, 0x2d, 0x78, 0x00, 0xf6, 0x48, 0x85, 0x16, 0xb8, 0x7a, 0xc7, 0x7c,
	0x9b, 0xe2, 0xa7, 0x46, 0xd5, 0xb2, 0x05, 0x4c, 0xef, 0xb5, 0x10, 0x4f, 0xfc, 0x03, 0x3d, 0xc7,
	0x77, 0x96, 0x27, 0xe9, 0x9e, 0x99, 0x07, 0xc7, 0xaf, 0x28, 0x55, 0x45, 0xb5, 0xe7, 0x9a, 0xd2,
	0x80, 0x41, 0x04, 0x67, 0xd6, 0x2e, 0xd5, 0x50, 0xad, 0xcc, 0xc0, 0x5a, 0xf2, 0xba, 0xc5, 0xc2,
	0xec, 0x9a, 0xa6, 0x03, 0x06, 0xdf, 0x0e, 0xcc, 0xfb, 0x93, 0xcf, 0x3b, 0x45, 0x6c, 0x0e, 0x6e,
	0x12, 0x9b, 0xbe, 0x51, 0xea, 0x26, 0xb1, 0xa5, 0xc4, 0xfd, 0x5f, 0xc9, 0x91, 0xa5, 0xa4, 0xab,
	0xac, 0x24, 0xf2, 0xee, 0xe4, 0xc8, 0x77, 0x96, 0xa7, 0xe9, 0x80, 0xec, 0x1c, 0x26, 0x2f, 0x0a,
	0x65, 0x12, 0x7b, 0x63, 0x33, 0xd2, 0xd3, 0xf5, 0x1b, 0xcc, 0x36, 0x55, 0xd6, 0xab, 0x58, 0x13,
	0xdb, 0xc0, 0xd8, 0xfc, 0x82, 0xf9, 0xa1, 0xe5, 0xe1, 0x5f, 0xb3, 0x16, 0x57, 0x07, 0x3a, 0x76,
	0x16, 0xdc, 0x5d, 0xc2, 0x05, 0xea, 0x7d, 0x52, 0x5d, 0x4c, 0xd6, 0x4c, 0x36, 0x31, 0xd1, 0xdc,
	0xfc, 0x04, 0x00, 0x00, 0xff, 0xff, 0x16, 0x18, 0x4d, 0x51, 0xd4, 0x01, 0x00, 0x00,
}
