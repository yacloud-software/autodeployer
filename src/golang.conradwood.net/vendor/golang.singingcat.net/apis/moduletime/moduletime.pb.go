// Code generated by protoc-gen-go.
// source: golang.singingcat.net/apis/moduletime/moduletime.proto
// DO NOT EDIT!

/*
Package moduletime is a generated protocol buffer package.

It is generated from these files:
	golang.singingcat.net/apis/moduletime/moduletime.proto

It has these top-level messages:
	DebugTimeDate
	ModuleList
	ClockSyncDetail
	ClockSyncReq
	ModParasReq
*/
package moduletime

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import common "golang.conradwood.net/apis/common"
import scmodcomms "golang.singingcat.net/apis/scmodcomms"

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

type DebugTimeDate struct {
	ID             uint64 `protobuf:"varint,1,opt,name=ID" json:"ID,omitempty"`
	Received       uint32 `protobuf:"varint,2,opt,name=Received" json:"Received,omitempty"`
	ModuleID       uint64 `protobuf:"varint,3,opt,name=ModuleID" json:"ModuleID,omitempty"`
	Now            uint64 `protobuf:"varint,4,opt,name=Now" json:"Now,omitempty"`
	RTC            uint64 `protobuf:"varint,5,opt,name=RTC" json:"RTC,omitempty"`
	SyncAge        uint64 `protobuf:"varint,6,opt,name=SyncAge" json:"SyncAge,omitempty"`
	RTD            uint64 `protobuf:"varint,7,opt,name=RTD" json:"RTD,omitempty"`
	Drift          int64  `protobuf:"varint,8,opt,name=Drift" json:"Drift,omitempty"`
	LTD            int64  `protobuf:"varint,9,opt,name=LTD" json:"LTD,omitempty"`
	TimeAsReceived string `protobuf:"bytes,10,opt,name=TimeAsReceived" json:"TimeAsReceived,omitempty"`
	DateAsReceived string `protobuf:"bytes,11,opt,name=DateAsReceived" json:"DateAsReceived,omitempty"`
	Mul            uint32 `protobuf:"varint,12,opt,name=Mul" json:"Mul,omitempty"`
	Div            uint32 `protobuf:"varint,13,opt,name=Div" json:"Div,omitempty"`
	OldLastSync    uint32 `protobuf:"varint,14,opt,name=OldLastSync" json:"OldLastSync,omitempty"`
	OldDrift       int64  `protobuf:"varint,15,opt,name=OldDrift" json:"OldDrift,omitempty"`
}

func (m *DebugTimeDate) Reset()                    { *m = DebugTimeDate{} }
func (m *DebugTimeDate) String() string            { return proto.CompactTextString(m) }
func (*DebugTimeDate) ProtoMessage()               {}
func (*DebugTimeDate) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *DebugTimeDate) GetID() uint64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *DebugTimeDate) GetReceived() uint32 {
	if m != nil {
		return m.Received
	}
	return 0
}

func (m *DebugTimeDate) GetModuleID() uint64 {
	if m != nil {
		return m.ModuleID
	}
	return 0
}

func (m *DebugTimeDate) GetNow() uint64 {
	if m != nil {
		return m.Now
	}
	return 0
}

func (m *DebugTimeDate) GetRTC() uint64 {
	if m != nil {
		return m.RTC
	}
	return 0
}

func (m *DebugTimeDate) GetSyncAge() uint64 {
	if m != nil {
		return m.SyncAge
	}
	return 0
}

func (m *DebugTimeDate) GetRTD() uint64 {
	if m != nil {
		return m.RTD
	}
	return 0
}

func (m *DebugTimeDate) GetDrift() int64 {
	if m != nil {
		return m.Drift
	}
	return 0
}

func (m *DebugTimeDate) GetLTD() int64 {
	if m != nil {
		return m.LTD
	}
	return 0
}

func (m *DebugTimeDate) GetTimeAsReceived() string {
	if m != nil {
		return m.TimeAsReceived
	}
	return ""
}

func (m *DebugTimeDate) GetDateAsReceived() string {
	if m != nil {
		return m.DateAsReceived
	}
	return ""
}

func (m *DebugTimeDate) GetMul() uint32 {
	if m != nil {
		return m.Mul
	}
	return 0
}

func (m *DebugTimeDate) GetDiv() uint32 {
	if m != nil {
		return m.Div
	}
	return 0
}

func (m *DebugTimeDate) GetOldLastSync() uint32 {
	if m != nil {
		return m.OldLastSync
	}
	return 0
}

func (m *DebugTimeDate) GetOldDrift() int64 {
	if m != nil {
		return m.OldDrift
	}
	return 0
}

type ModuleList struct {
	ModuleIDs []uint64 `protobuf:"varint,1,rep,packed,name=ModuleIDs" json:"ModuleIDs,omitempty"`
}

func (m *ModuleList) Reset()                    { *m = ModuleList{} }
func (m *ModuleList) String() string            { return proto.CompactTextString(m) }
func (*ModuleList) ProtoMessage()               {}
func (*ModuleList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *ModuleList) GetModuleIDs() []uint64 {
	if m != nil {
		return m.ModuleIDs
	}
	return nil
}

type ClockSyncDetail struct {
	DTD *DebugTimeDate `protobuf:"bytes,1,opt,name=DTD" json:"DTD,omitempty"`
}

func (m *ClockSyncDetail) Reset()                    { *m = ClockSyncDetail{} }
func (m *ClockSyncDetail) String() string            { return proto.CompactTextString(m) }
func (*ClockSyncDetail) ProtoMessage()               {}
func (*ClockSyncDetail) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *ClockSyncDetail) GetDTD() *DebugTimeDate {
	if m != nil {
		return m.DTD
	}
	return nil
}

type ClockSyncReq struct {
	ModuleID uint64 `protobuf:"varint,1,opt,name=ModuleID" json:"ModuleID,omitempty"`
}

func (m *ClockSyncReq) Reset()                    { *m = ClockSyncReq{} }
func (m *ClockSyncReq) String() string            { return proto.CompactTextString(m) }
func (*ClockSyncReq) ProtoMessage()               {}
func (*ClockSyncReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *ClockSyncReq) GetModuleID() uint64 {
	if m != nil {
		return m.ModuleID
	}
	return 0
}

type ModParasReq struct {
	ModuleID            uint64 `protobuf:"varint,1,opt,name=ModuleID" json:"ModuleID,omitempty"`
	SetMul              bool   `protobuf:"varint,2,opt,name=SetMul" json:"SetMul,omitempty"`
	Mul                 int32  `protobuf:"varint,3,opt,name=Mul" json:"Mul,omitempty"`
	SetDiv              bool   `protobuf:"varint,4,opt,name=SetDiv" json:"SetDiv,omitempty"`
	Div                 int32  `protobuf:"varint,5,opt,name=Div" json:"Div,omitempty"`
	CalculatedTimestamp uint32 `protobuf:"varint,6,opt,name=CalculatedTimestamp" json:"CalculatedTimestamp,omitempty"`
	AppliedTimestamp    uint32 `protobuf:"varint,7,opt,name=AppliedTimestamp" json:"AppliedTimestamp,omitempty"`
}

func (m *ModParasReq) Reset()                    { *m = ModParasReq{} }
func (m *ModParasReq) String() string            { return proto.CompactTextString(m) }
func (*ModParasReq) ProtoMessage()               {}
func (*ModParasReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *ModParasReq) GetModuleID() uint64 {
	if m != nil {
		return m.ModuleID
	}
	return 0
}

func (m *ModParasReq) GetSetMul() bool {
	if m != nil {
		return m.SetMul
	}
	return false
}

func (m *ModParasReq) GetMul() int32 {
	if m != nil {
		return m.Mul
	}
	return 0
}

func (m *ModParasReq) GetSetDiv() bool {
	if m != nil {
		return m.SetDiv
	}
	return false
}

func (m *ModParasReq) GetDiv() int32 {
	if m != nil {
		return m.Div
	}
	return 0
}

func (m *ModParasReq) GetCalculatedTimestamp() uint32 {
	if m != nil {
		return m.CalculatedTimestamp
	}
	return 0
}

func (m *ModParasReq) GetAppliedTimestamp() uint32 {
	if m != nil {
		return m.AppliedTimestamp
	}
	return 0
}

func init() {
	proto.RegisterType((*DebugTimeDate)(nil), "moduletime.DebugTimeDate")
	proto.RegisterType((*ModuleList)(nil), "moduletime.ModuleList")
	proto.RegisterType((*ClockSyncDetail)(nil), "moduletime.ClockSyncDetail")
	proto.RegisterType((*ClockSyncReq)(nil), "moduletime.ClockSyncReq")
	proto.RegisterType((*ModParasReq)(nil), "moduletime.ModParasReq")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for ModuleTime service

type ModuleTimeClient interface {
	CommandReceived(ctx context.Context, in *scmodcomms.Response, opts ...grpc.CallOption) (*scmodcomms.ComResponse, error)
	// stream the clock sync details
	StreamClockSync(ctx context.Context, in *ModuleList, opts ...grpc.CallOption) (ModuleTime_StreamClockSyncClient, error)
	// actively send out a sync to a module
	SendClockSync(ctx context.Context, in *ClockSyncReq, opts ...grpc.CallOption) (*common.Void, error)
	// set parameters for module
	SetModuleParameters(ctx context.Context, in *ModParasReq, opts ...grpc.CallOption) (*common.Void, error)
}

type moduleTimeClient struct {
	cc *grpc.ClientConn
}

func NewModuleTimeClient(cc *grpc.ClientConn) ModuleTimeClient {
	return &moduleTimeClient{cc}
}

func (c *moduleTimeClient) CommandReceived(ctx context.Context, in *scmodcomms.Response, opts ...grpc.CallOption) (*scmodcomms.ComResponse, error) {
	out := new(scmodcomms.ComResponse)
	err := grpc.Invoke(ctx, "/moduletime.ModuleTime/CommandReceived", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *moduleTimeClient) StreamClockSync(ctx context.Context, in *ModuleList, opts ...grpc.CallOption) (ModuleTime_StreamClockSyncClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_ModuleTime_serviceDesc.Streams[0], c.cc, "/moduletime.ModuleTime/StreamClockSync", opts...)
	if err != nil {
		return nil, err
	}
	x := &moduleTimeStreamClockSyncClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ModuleTime_StreamClockSyncClient interface {
	Recv() (*ClockSyncDetail, error)
	grpc.ClientStream
}

type moduleTimeStreamClockSyncClient struct {
	grpc.ClientStream
}

func (x *moduleTimeStreamClockSyncClient) Recv() (*ClockSyncDetail, error) {
	m := new(ClockSyncDetail)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *moduleTimeClient) SendClockSync(ctx context.Context, in *ClockSyncReq, opts ...grpc.CallOption) (*common.Void, error) {
	out := new(common.Void)
	err := grpc.Invoke(ctx, "/moduletime.ModuleTime/SendClockSync", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *moduleTimeClient) SetModuleParameters(ctx context.Context, in *ModParasReq, opts ...grpc.CallOption) (*common.Void, error) {
	out := new(common.Void)
	err := grpc.Invoke(ctx, "/moduletime.ModuleTime/SetModuleParameters", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for ModuleTime service

type ModuleTimeServer interface {
	CommandReceived(context.Context, *scmodcomms.Response) (*scmodcomms.ComResponse, error)
	// stream the clock sync details
	StreamClockSync(*ModuleList, ModuleTime_StreamClockSyncServer) error
	// actively send out a sync to a module
	SendClockSync(context.Context, *ClockSyncReq) (*common.Void, error)
	// set parameters for module
	SetModuleParameters(context.Context, *ModParasReq) (*common.Void, error)
}

func RegisterModuleTimeServer(s *grpc.Server, srv ModuleTimeServer) {
	s.RegisterService(&_ModuleTime_serviceDesc, srv)
}

func _ModuleTime_CommandReceived_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(scmodcomms.Response)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModuleTimeServer).CommandReceived(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/moduletime.ModuleTime/CommandReceived",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModuleTimeServer).CommandReceived(ctx, req.(*scmodcomms.Response))
	}
	return interceptor(ctx, in, info, handler)
}

func _ModuleTime_StreamClockSync_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ModuleList)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ModuleTimeServer).StreamClockSync(m, &moduleTimeStreamClockSyncServer{stream})
}

type ModuleTime_StreamClockSyncServer interface {
	Send(*ClockSyncDetail) error
	grpc.ServerStream
}

type moduleTimeStreamClockSyncServer struct {
	grpc.ServerStream
}

func (x *moduleTimeStreamClockSyncServer) Send(m *ClockSyncDetail) error {
	return x.ServerStream.SendMsg(m)
}

func _ModuleTime_SendClockSync_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClockSyncReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModuleTimeServer).SendClockSync(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/moduletime.ModuleTime/SendClockSync",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModuleTimeServer).SendClockSync(ctx, req.(*ClockSyncReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ModuleTime_SetModuleParameters_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ModParasReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModuleTimeServer).SetModuleParameters(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/moduletime.ModuleTime/SetModuleParameters",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModuleTimeServer).SetModuleParameters(ctx, req.(*ModParasReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _ModuleTime_serviceDesc = grpc.ServiceDesc{
	ServiceName: "moduletime.ModuleTime",
	HandlerType: (*ModuleTimeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CommandReceived",
			Handler:    _ModuleTime_CommandReceived_Handler,
		},
		{
			MethodName: "SendClockSync",
			Handler:    _ModuleTime_SendClockSync_Handler,
		},
		{
			MethodName: "SetModuleParameters",
			Handler:    _ModuleTime_SetModuleParameters_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamClockSync",
			Handler:       _ModuleTime_StreamClockSync_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "golang.singingcat.net/apis/moduletime/moduletime.proto",
}

func init() {
	proto.RegisterFile("golang.singingcat.net/apis/moduletime/moduletime.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 601 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x84, 0x54, 0xdd, 0x6e, 0x13, 0x3d,
	0x10, 0xd5, 0xe6, 0xa7, 0x3f, 0x93, 0xa6, 0xa9, 0xdc, 0xaa, 0xf5, 0x97, 0x0f, 0xa1, 0x55, 0x2e,
	0x50, 0x54, 0xa4, 0x6d, 0x55, 0x24, 0xb8, 0x41, 0x88, 0x12, 0x5f, 0x50, 0xa9, 0xa5, 0xc8, 0x89,
	0xb8, 0x77, 0x77, 0x4d, 0x64, 0xb1, 0xbb, 0x4e, 0x63, 0xa7, 0x15, 0x97, 0x3c, 0x01, 0x0f, 0xc8,
	0xcb, 0xa0, 0xb1, 0xb3, 0x1b, 0x27, 0x54, 0x70, 0x95, 0x99, 0xe3, 0x33, 0xde, 0x99, 0x73, 0x9c,
	0x81, 0xd7, 0x53, 0x9d, 0x8b, 0x72, 0x9a, 0x18, 0x55, 0x4e, 0x55, 0x39, 0x4d, 0x85, 0x4d, 0x4a,
	0x69, 0xcf, 0xc4, 0x4c, 0x99, 0xb3, 0x42, 0x67, 0x8b, 0x5c, 0x5a, 0x55, 0xc8, 0x20, 0x4c, 0x66,
	0x73, 0x6d, 0x35, 0x81, 0x15, 0xd2, 0x4f, 0x96, 0x77, 0xa4, 0xba, 0x9c, 0x8b, 0xec, 0x51, 0xeb,
	0x6c, 0x75, 0x47, 0xaa, 0x8b, 0x42, 0x97, 0xcb, 0x1f, 0x5f, 0xdb, 0xff, 0xdb, 0x37, 0x4d, 0x5a,
	0xe8, 0x0c, 0xd9, 0x61, 0xe8, 0xeb, 0x06, 0x3f, 0x9a, 0xd0, 0x65, 0xf2, 0x6e, 0x31, 0x9d, 0xa8,
	0x42, 0x32, 0x61, 0x25, 0xd9, 0x87, 0xc6, 0x15, 0xa3, 0x51, 0x1c, 0x0d, 0x5b, 0xbc, 0x71, 0xc5,
	0x48, 0x1f, 0x76, 0xb8, 0x4c, 0xa5, 0x7a, 0x90, 0x19, 0x6d, 0xc4, 0xd1, 0xb0, 0xcb, 0xeb, 0x1c,
	0xcf, 0x6e, 0x5c, 0xcf, 0x57, 0x8c, 0x36, 0x5d, 0x45, 0x9d, 0x93, 0x03, 0x68, 0x7e, 0xd2, 0x8f,
	0xb4, 0xe5, 0x60, 0x0c, 0x11, 0xe1, 0x93, 0x11, 0x6d, 0x7b, 0x84, 0x4f, 0x46, 0x84, 0xc2, 0xf6,
	0xf8, 0x7b, 0x99, 0x5e, 0x4e, 0x25, 0xdd, 0x72, 0x68, 0x95, 0x7a, 0x2e, 0xa3, 0xdb, 0x15, 0x97,
	0x91, 0x23, 0x68, 0xb3, 0xb9, 0xfa, 0x6a, 0xe9, 0x4e, 0x1c, 0x0d, 0x9b, 0xdc, 0x27, 0xc8, 0xbb,
	0x9e, 0x30, 0xba, 0xeb, 0x30, 0x0c, 0xc9, 0x0b, 0xd8, 0xc7, 0x59, 0x2e, 0x4d, 0xdd, 0x35, 0xc4,
	0xd1, 0x70, 0x97, 0x6f, 0xa0, 0xc8, 0xc3, 0x79, 0x03, 0x5e, 0xc7, 0xf3, 0xd6, 0x51, 0xfc, 0xc2,
	0xcd, 0x22, 0xa7, 0x7b, 0x6e, 0x74, 0x0c, 0x11, 0x61, 0xea, 0x81, 0x76, 0x3d, 0xc2, 0xd4, 0x03,
	0x89, 0xa1, 0x73, 0x9b, 0x67, 0xd7, 0xc2, 0x58, 0xec, 0x9f, 0xee, 0xbb, 0x93, 0x10, 0x42, 0xa5,
	0x6e, 0xf3, 0xcc, 0x0f, 0xd0, 0x73, 0xcd, 0xd6, 0xf9, 0xe0, 0x14, 0xc0, 0xab, 0x76, 0xad, 0x8c,
	0x25, 0xcf, 0x60, 0xb7, 0xd2, 0xd0, 0xd0, 0x28, 0x6e, 0x0e, 0x5b, 0x7c, 0x05, 0x0c, 0xde, 0x41,
	0x6f, 0x94, 0xeb, 0xf4, 0x1b, 0x5e, 0xca, 0xa4, 0x15, 0x2a, 0x27, 0x2f, 0xa1, 0xc9, 0x26, 0xde,
	0xb1, 0xce, 0xc5, 0x7f, 0x49, 0xf0, 0xac, 0xd6, 0x8c, 0xe5, 0xc8, 0x1a, 0x9c, 0xc2, 0x5e, 0x5d,
	0xcf, 0xe5, 0xfd, 0x9a, 0x83, 0xd1, 0xba, 0x83, 0x83, 0x5f, 0x11, 0x74, 0x6e, 0x74, 0xf6, 0x59,
	0xcc, 0x85, 0xf9, 0x07, 0x97, 0x1c, 0xc3, 0xd6, 0x58, 0x5a, 0x14, 0x0a, 0xdf, 0xc8, 0x0e, 0x5f,
	0x66, 0x95, 0x7a, 0xf8, 0x38, 0xda, 0x5e, 0x3d, 0xcf, 0x44, 0x01, 0x5b, 0x35, 0x13, 0x35, 0x5c,
	0xaa, 0xda, 0xf6, 0x4c, 0x44, 0xce, 0xe1, 0x70, 0x24, 0xf2, 0x74, 0x91, 0x0b, 0x2b, 0x33, 0x1c,
	0xc3, 0x58, 0x51, 0xcc, 0xdc, 0x4b, 0xe9, 0xf2, 0xa7, 0x8e, 0xc8, 0x29, 0x1c, 0x5c, 0xce, 0x66,
	0xb9, 0x0a, 0xe9, 0xdb, 0x8e, 0xfe, 0x07, 0x7e, 0xf1, 0xb3, 0x51, 0xc9, 0x8e, 0x18, 0x79, 0x0f,
	0xbd, 0x91, 0x2e, 0x0a, 0x51, 0x66, 0xb5, 0xf3, 0x47, 0x49, 0xf0, 0x77, 0xe1, 0xd2, 0xcc, 0x74,
	0x69, 0x64, 0xff, 0x24, 0x44, 0x47, 0xba, 0xa8, 0x0e, 0xc8, 0x47, 0xe8, 0x8d, 0xed, 0x5c, 0x8a,
	0xa2, 0x16, 0x98, 0x1c, 0x87, 0x6e, 0xac, 0x3c, 0xee, 0xff, 0x1f, 0xe2, 0x1b, 0x7e, 0x9e, 0x47,
	0xe4, 0x0d, 0x74, 0xc7, 0xb2, 0xcc, 0x56, 0xf7, 0xd0, 0x27, 0xf9, 0x5c, 0xde, 0xf7, 0xf7, 0x92,
	0xe5, 0x1a, 0xf8, 0xa2, 0x55, 0x46, 0xde, 0xc2, 0x21, 0xea, 0xee, 0xb8, 0x68, 0x5b, 0x21, 0xad,
	0x9c, 0x1b, 0x72, 0xb2, 0xd1, 0x46, 0xe5, 0xe8, 0x7a, 0xf5, 0x87, 0x18, 0x9e, 0x97, 0xd2, 0x86,
	0x2b, 0x04, 0xd7, 0x47, 0x50, 0x7b, 0xb7, 0xe5, 0x96, 0xc6, 0xab, 0xdf, 0x01, 0x00, 0x00, 0xff,
	0xff, 0x00, 0x8c, 0x07, 0xc5, 0xe2, 0x04, 0x00, 0x00,
}
