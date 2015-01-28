// Code generated by protoc-gen-go.
// source: pack.proto
// DO NOT EDIT!

/*
Package proto_base is a generated protocol buffer package.

It is generated from these files:
	pack.proto

It has these top-level messages:
	Pack
*/
package proto_base

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type Pack struct {
	Session          *int32 `protobuf:"varint,1,opt,name=session" json:"session,omitempty"`
	Timestamp        *int32 `protobuf:"varint,2,opt,name=timestamp" json:"timestamp,omitempty"`
	Type             *int32 `protobuf:"varint,3,opt,name=type" json:"type,omitempty"`
	Data             []byte `protobuf:"bytes,4,opt,name=data" json:"data,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *Pack) Reset()         { *m = Pack{} }
func (m *Pack) String() string { return proto.CompactTextString(m) }
func (*Pack) ProtoMessage()    {}

func (m *Pack) GetSession() int32 {
	if m != nil && m.Session != nil {
		return *m.Session
	}
	return 0
}

func (m *Pack) GetTimestamp() int32 {
	if m != nil && m.Timestamp != nil {
		return *m.Timestamp
	}
	return 0
}

func (m *Pack) GetType() int32 {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return 0
}

func (m *Pack) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
}
