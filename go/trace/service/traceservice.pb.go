// Code generated by protoc-gen-go. DO NOT EDIT.
// source: traceservice.proto

/*
Package traceservice is a generated protocol buffer package.

It is generated from these files:
	traceservice.proto

It has these top-level messages:
	Empty
	CommitID
	Params
	MissingParamsRequest
	MissingParamsResponse
	ParamsPair
	AddParamsRequest
	StoredEntry
	ValuePair
	AddRequest
	RemoveRequest
	ListRequest
	ListResponse
	GetValuesRequest
	GetValuesResponse
	GetParamsRequest
	GetParamsResponse
	GetValuesRawResponse
	GetTraceIDsRequest
	TraceIDPair
	GetTraceIDsResponse
	ListMD5Request
	CommitMD5
	ListMD5Response
*/
package traceservice

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

type Empty struct {
}

func (m *Empty) Reset()                    { *m = Empty{} }
func (m *Empty) String() string            { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()               {}
func (*Empty) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// CommitID identifies one commit, or trybot try.
type CommitID struct {
	// The id of a commit, either a git hash, or a Reitveld patch id.
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	// The source of the commit, either a git branch name, or a Reitveld issue id.
	Source string `protobuf:"bytes,2,opt,name=source" json:"source,omitempty"`
	// The timestamp of the commit or trybot patch.
	Timestamp int64 `protobuf:"varint,3,opt,name=timestamp" json:"timestamp,omitempty"`
}

func (m *CommitID) Reset()                    { *m = CommitID{} }
func (m *CommitID) String() string            { return proto.CompactTextString(m) }
func (*CommitID) ProtoMessage()               {}
func (*CommitID) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *CommitID) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *CommitID) GetSource() string {
	if m != nil {
		return m.Source
	}
	return ""
}

func (m *CommitID) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

// Params are the key-value pairs for a single trace.
//
// All of the key-value parameters should be present, the ones used to
// construct the traceid, along with optional parameters.
type Params struct {
	Params map[string]string `protobuf:"bytes,1,rep,name=params" json:"params,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *Params) Reset()                    { *m = Params{} }
func (m *Params) String() string            { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()               {}
func (*Params) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Params) GetParams() map[string]string {
	if m != nil {
		return m.Params
	}
	return nil
}

type MissingParamsRequest struct {
	Traceids []string `protobuf:"bytes,1,rep,name=traceids" json:"traceids,omitempty"`
}

func (m *MissingParamsRequest) Reset()                    { *m = MissingParamsRequest{} }
func (m *MissingParamsRequest) String() string            { return proto.CompactTextString(m) }
func (*MissingParamsRequest) ProtoMessage()               {}
func (*MissingParamsRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *MissingParamsRequest) GetTraceids() []string {
	if m != nil {
		return m.Traceids
	}
	return nil
}

type MissingParamsResponse struct {
	Traceids []string `protobuf:"bytes,1,rep,name=traceids" json:"traceids,omitempty"`
}

func (m *MissingParamsResponse) Reset()                    { *m = MissingParamsResponse{} }
func (m *MissingParamsResponse) String() string            { return proto.CompactTextString(m) }
func (*MissingParamsResponse) ProtoMessage()               {}
func (*MissingParamsResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *MissingParamsResponse) GetTraceids() []string {
	if m != nil {
		return m.Traceids
	}
	return nil
}

type ParamsPair struct {
	Key    string            `protobuf:"bytes,1,opt,name=key" json:"key,omitempty"`
	Params map[string]string `protobuf:"bytes,2,rep,name=params" json:"params,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *ParamsPair) Reset()                    { *m = ParamsPair{} }
func (m *ParamsPair) String() string            { return proto.CompactTextString(m) }
func (*ParamsPair) ProtoMessage()               {}
func (*ParamsPair) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *ParamsPair) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *ParamsPair) GetParams() map[string]string {
	if m != nil {
		return m.Params
	}
	return nil
}

type AddParamsRequest struct {
	Params []*ParamsPair `protobuf:"bytes,4,rep,name=params" json:"params,omitempty"`
}

func (m *AddParamsRequest) Reset()                    { *m = AddParamsRequest{} }
func (m *AddParamsRequest) String() string            { return proto.CompactTextString(m) }
func (*AddParamsRequest) ProtoMessage()               {}
func (*AddParamsRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *AddParamsRequest) GetParams() []*ParamsPair {
	if m != nil {
		return m.Params
	}
	return nil
}

// StoredEntry is used to serialize the Params to be stored in the BoltBD.
type StoredEntry struct {
	Params *Params `protobuf:"bytes,2,opt,name=params" json:"params,omitempty"`
}

func (m *StoredEntry) Reset()                    { *m = StoredEntry{} }
func (m *StoredEntry) String() string            { return proto.CompactTextString(m) }
func (*StoredEntry) ProtoMessage()               {}
func (*StoredEntry) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *StoredEntry) GetParams() *Params {
	if m != nil {
		return m.Params
	}
	return nil
}

type ValuePair struct {
	Key   string `protobuf:"bytes,1,opt,name=key" json:"key,omitempty"`
	Value []byte `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *ValuePair) Reset()                    { *m = ValuePair{} }
func (m *ValuePair) String() string            { return proto.CompactTextString(m) }
func (*ValuePair) ProtoMessage()               {}
func (*ValuePair) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *ValuePair) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *ValuePair) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

type AddRequest struct {
	// The id of the commit/trybot we are adding data about.
	Commitid *CommitID    `protobuf:"bytes,1,opt,name=commitid" json:"commitid,omitempty"`
	Values   []*ValuePair `protobuf:"bytes,3,rep,name=values" json:"values,omitempty"`
}

func (m *AddRequest) Reset()                    { *m = AddRequest{} }
func (m *AddRequest) String() string            { return proto.CompactTextString(m) }
func (*AddRequest) ProtoMessage()               {}
func (*AddRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *AddRequest) GetCommitid() *CommitID {
	if m != nil {
		return m.Commitid
	}
	return nil
}

func (m *AddRequest) GetValues() []*ValuePair {
	if m != nil {
		return m.Values
	}
	return nil
}

type RemoveRequest struct {
	// The id of the commit/trybot we are removing.
	Commitid *CommitID `protobuf:"bytes,1,opt,name=commitid" json:"commitid,omitempty"`
}

func (m *RemoveRequest) Reset()                    { *m = RemoveRequest{} }
func (m *RemoveRequest) String() string            { return proto.CompactTextString(m) }
func (*RemoveRequest) ProtoMessage()               {}
func (*RemoveRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *RemoveRequest) GetCommitid() *CommitID {
	if m != nil {
		return m.Commitid
	}
	return nil
}

type ListRequest struct {
	// begin is the unix timestamp to start searching from.
	Begin int64 `protobuf:"varint,1,opt,name=begin" json:"begin,omitempty"`
	// end is the unix timestamp to search to (inclusive).
	End int64 `protobuf:"varint,2,opt,name=end" json:"end,omitempty"`
}

func (m *ListRequest) Reset()                    { *m = ListRequest{} }
func (m *ListRequest) String() string            { return proto.CompactTextString(m) }
func (*ListRequest) ProtoMessage()               {}
func (*ListRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

func (m *ListRequest) GetBegin() int64 {
	if m != nil {
		return m.Begin
	}
	return 0
}

func (m *ListRequest) GetEnd() int64 {
	if m != nil {
		return m.End
	}
	return 0
}

type ListResponse struct {
	// A list of CommitIDs that fall between the given timestamps in
	// ListRequest.
	Commitids []*CommitID `protobuf:"bytes,3,rep,name=commitids" json:"commitids,omitempty"`
}

func (m *ListResponse) Reset()                    { *m = ListResponse{} }
func (m *ListResponse) String() string            { return proto.CompactTextString(m) }
func (*ListResponse) ProtoMessage()               {}
func (*ListResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

func (m *ListResponse) GetCommitids() []*CommitID {
	if m != nil {
		return m.Commitids
	}
	return nil
}

type GetValuesRequest struct {
	Commitid *CommitID `protobuf:"bytes,1,opt,name=commitid" json:"commitid,omitempty"`
}

func (m *GetValuesRequest) Reset()                    { *m = GetValuesRequest{} }
func (m *GetValuesRequest) String() string            { return proto.CompactTextString(m) }
func (*GetValuesRequest) ProtoMessage()               {}
func (*GetValuesRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

func (m *GetValuesRequest) GetCommitid() *CommitID {
	if m != nil {
		return m.Commitid
	}
	return nil
}

type GetValuesResponse struct {
	Values []*ValuePair `protobuf:"bytes,4,rep,name=values" json:"values,omitempty"`
	Md5    string       `protobuf:"bytes,5,opt,name=md5" json:"md5,omitempty"`
}

func (m *GetValuesResponse) Reset()                    { *m = GetValuesResponse{} }
func (m *GetValuesResponse) String() string            { return proto.CompactTextString(m) }
func (*GetValuesResponse) ProtoMessage()               {}
func (*GetValuesResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{14} }

func (m *GetValuesResponse) GetValues() []*ValuePair {
	if m != nil {
		return m.Values
	}
	return nil
}

func (m *GetValuesResponse) GetMd5() string {
	if m != nil {
		return m.Md5
	}
	return ""
}

type GetParamsRequest struct {
	// A list of traceids.
	Traceids []string `protobuf:"bytes,1,rep,name=traceids" json:"traceids,omitempty"`
}

func (m *GetParamsRequest) Reset()                    { *m = GetParamsRequest{} }
func (m *GetParamsRequest) String() string            { return proto.CompactTextString(m) }
func (*GetParamsRequest) ProtoMessage()               {}
func (*GetParamsRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{15} }

func (m *GetParamsRequest) GetTraceids() []string {
	if m != nil {
		return m.Traceids
	}
	return nil
}

type GetParamsResponse struct {
	Params []*ParamsPair `protobuf:"bytes,4,rep,name=params" json:"params,omitempty"`
}

func (m *GetParamsResponse) Reset()                    { *m = GetParamsResponse{} }
func (m *GetParamsResponse) String() string            { return proto.CompactTextString(m) }
func (*GetParamsResponse) ProtoMessage()               {}
func (*GetParamsResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{16} }

func (m *GetParamsResponse) GetParams() []*ParamsPair {
	if m != nil {
		return m.Params
	}
	return nil
}

type GetValuesRawResponse struct {
	// Raw byte slice that can be decoded with NewCommitInfo.
	Value []byte `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	Md5   string `protobuf:"bytes,2,opt,name=md5" json:"md5,omitempty"`
}

func (m *GetValuesRawResponse) Reset()                    { *m = GetValuesRawResponse{} }
func (m *GetValuesRawResponse) String() string            { return proto.CompactTextString(m) }
func (*GetValuesRawResponse) ProtoMessage()               {}
func (*GetValuesRawResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{17} }

func (m *GetValuesRawResponse) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *GetValuesRawResponse) GetMd5() string {
	if m != nil {
		return m.Md5
	}
	return ""
}

type GetTraceIDsRequest struct {
	Id []uint64 `protobuf:"varint,1,rep,packed,name=id" json:"id,omitempty"`
}

func (m *GetTraceIDsRequest) Reset()                    { *m = GetTraceIDsRequest{} }
func (m *GetTraceIDsRequest) String() string            { return proto.CompactTextString(m) }
func (*GetTraceIDsRequest) ProtoMessage()               {}
func (*GetTraceIDsRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{18} }

func (m *GetTraceIDsRequest) GetId() []uint64 {
	if m != nil {
		return m.Id
	}
	return nil
}

type TraceIDPair struct {
	Id64 uint64 `protobuf:"varint,1,opt,name=id64" json:"id64,omitempty"`
	Id   string `protobuf:"bytes,2,opt,name=id" json:"id,omitempty"`
}

func (m *TraceIDPair) Reset()                    { *m = TraceIDPair{} }
func (m *TraceIDPair) String() string            { return proto.CompactTextString(m) }
func (*TraceIDPair) ProtoMessage()               {}
func (*TraceIDPair) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{19} }

func (m *TraceIDPair) GetId64() uint64 {
	if m != nil {
		return m.Id64
	}
	return 0
}

func (m *TraceIDPair) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type GetTraceIDsResponse struct {
	Ids []*TraceIDPair `protobuf:"bytes,1,rep,name=ids" json:"ids,omitempty"`
}

func (m *GetTraceIDsResponse) Reset()                    { *m = GetTraceIDsResponse{} }
func (m *GetTraceIDsResponse) String() string            { return proto.CompactTextString(m) }
func (*GetTraceIDsResponse) ProtoMessage()               {}
func (*GetTraceIDsResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{20} }

func (m *GetTraceIDsResponse) GetIds() []*TraceIDPair {
	if m != nil {
		return m.Ids
	}
	return nil
}

type ListMD5Request struct {
	Commitid []*CommitID `protobuf:"bytes,1,rep,name=commitid" json:"commitid,omitempty"`
}

func (m *ListMD5Request) Reset()                    { *m = ListMD5Request{} }
func (m *ListMD5Request) String() string            { return proto.CompactTextString(m) }
func (*ListMD5Request) ProtoMessage()               {}
func (*ListMD5Request) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{21} }

func (m *ListMD5Request) GetCommitid() []*CommitID {
	if m != nil {
		return m.Commitid
	}
	return nil
}

type CommitMD5 struct {
	Commitid *CommitID `protobuf:"bytes,1,opt,name=commitid" json:"commitid,omitempty"`
	Md5      string    `protobuf:"bytes,2,opt,name=md5" json:"md5,omitempty"`
}

func (m *CommitMD5) Reset()                    { *m = CommitMD5{} }
func (m *CommitMD5) String() string            { return proto.CompactTextString(m) }
func (*CommitMD5) ProtoMessage()               {}
func (*CommitMD5) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{22} }

func (m *CommitMD5) GetCommitid() *CommitID {
	if m != nil {
		return m.Commitid
	}
	return nil
}

func (m *CommitMD5) GetMd5() string {
	if m != nil {
		return m.Md5
	}
	return ""
}

type ListMD5Response struct {
	Commitmd5 []*CommitMD5 `protobuf:"bytes,1,rep,name=commitmd5" json:"commitmd5,omitempty"`
}

func (m *ListMD5Response) Reset()                    { *m = ListMD5Response{} }
func (m *ListMD5Response) String() string            { return proto.CompactTextString(m) }
func (*ListMD5Response) ProtoMessage()               {}
func (*ListMD5Response) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{23} }

func (m *ListMD5Response) GetCommitmd5() []*CommitMD5 {
	if m != nil {
		return m.Commitmd5
	}
	return nil
}

func init() {
	proto.RegisterType((*Empty)(nil), "traceservice.Empty")
	proto.RegisterType((*CommitID)(nil), "traceservice.CommitID")
	proto.RegisterType((*Params)(nil), "traceservice.Params")
	proto.RegisterType((*MissingParamsRequest)(nil), "traceservice.MissingParamsRequest")
	proto.RegisterType((*MissingParamsResponse)(nil), "traceservice.MissingParamsResponse")
	proto.RegisterType((*ParamsPair)(nil), "traceservice.ParamsPair")
	proto.RegisterType((*AddParamsRequest)(nil), "traceservice.AddParamsRequest")
	proto.RegisterType((*StoredEntry)(nil), "traceservice.StoredEntry")
	proto.RegisterType((*ValuePair)(nil), "traceservice.ValuePair")
	proto.RegisterType((*AddRequest)(nil), "traceservice.AddRequest")
	proto.RegisterType((*RemoveRequest)(nil), "traceservice.RemoveRequest")
	proto.RegisterType((*ListRequest)(nil), "traceservice.ListRequest")
	proto.RegisterType((*ListResponse)(nil), "traceservice.ListResponse")
	proto.RegisterType((*GetValuesRequest)(nil), "traceservice.GetValuesRequest")
	proto.RegisterType((*GetValuesResponse)(nil), "traceservice.GetValuesResponse")
	proto.RegisterType((*GetParamsRequest)(nil), "traceservice.GetParamsRequest")
	proto.RegisterType((*GetParamsResponse)(nil), "traceservice.GetParamsResponse")
	proto.RegisterType((*GetValuesRawResponse)(nil), "traceservice.GetValuesRawResponse")
	proto.RegisterType((*GetTraceIDsRequest)(nil), "traceservice.GetTraceIDsRequest")
	proto.RegisterType((*TraceIDPair)(nil), "traceservice.TraceIDPair")
	proto.RegisterType((*GetTraceIDsResponse)(nil), "traceservice.GetTraceIDsResponse")
	proto.RegisterType((*ListMD5Request)(nil), "traceservice.ListMD5Request")
	proto.RegisterType((*CommitMD5)(nil), "traceservice.CommitMD5")
	proto.RegisterType((*ListMD5Response)(nil), "traceservice.ListMD5Response")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for TraceService service

type TraceServiceClient interface {
	// Returns a list of traceids that don't have Params stored in the datastore.
	MissingParams(ctx context.Context, in *MissingParamsRequest, opts ...grpc.CallOption) (*MissingParamsResponse, error)
	// Adds Params for a set of traceids.
	AddParams(ctx context.Context, in *AddParamsRequest, opts ...grpc.CallOption) (*Empty, error)
	// Adds data for a set of traces for a particular commitid.
	Add(ctx context.Context, in *AddRequest, opts ...grpc.CallOption) (*Empty, error)
	// Removes data for a particular commitid.
	Remove(ctx context.Context, in *RemoveRequest, opts ...grpc.CallOption) (*Empty, error)
	// List returns all the CommitIDs that exist in the given time range.
	List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error)
	// GetValues returns all the trace values stored for the given CommitID.
	GetValues(ctx context.Context, in *GetValuesRequest, opts ...grpc.CallOption) (*GetValuesResponse, error)
	// GetParams returns the Params for all of the given traces.
	GetParams(ctx context.Context, in *GetParamsRequest, opts ...grpc.CallOption) (*GetParamsResponse, error)
	// GetValuesRaw returns all the trace values stored for the given CommitID in
	// the raw format stored in BoltDB. The decoding can be done by calling
	// NewCommitInfo() on the returned byte slice.
	GetValuesRaw(ctx context.Context, in *GetValuesRequest, opts ...grpc.CallOption) (*GetValuesRawResponse, error)
	// GetTraceIDs returns the traceids for the given trace64ids. These are used
	// in decoding the bytes returned from GetValuesRaw.
	GetTraceIDs(ctx context.Context, in *GetTraceIDsRequest, opts ...grpc.CallOption) (*GetTraceIDsResponse, error)
	// ListMD5 returns the MD5 hashes for the given CommitIDs.
	ListMD5(ctx context.Context, in *ListMD5Request, opts ...grpc.CallOption) (*ListMD5Response, error)
	// Ping should always succeed. Used to test if the service is up and
	// running.
	Ping(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error)
}

type traceServiceClient struct {
	cc *grpc.ClientConn
}

func NewTraceServiceClient(cc *grpc.ClientConn) TraceServiceClient {
	return &traceServiceClient{cc}
}

func (c *traceServiceClient) MissingParams(ctx context.Context, in *MissingParamsRequest, opts ...grpc.CallOption) (*MissingParamsResponse, error) {
	out := new(MissingParamsResponse)
	err := grpc.Invoke(ctx, "/traceservice.TraceService/MissingParams", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *traceServiceClient) AddParams(ctx context.Context, in *AddParamsRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := grpc.Invoke(ctx, "/traceservice.TraceService/AddParams", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *traceServiceClient) Add(ctx context.Context, in *AddRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := grpc.Invoke(ctx, "/traceservice.TraceService/Add", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *traceServiceClient) Remove(ctx context.Context, in *RemoveRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := grpc.Invoke(ctx, "/traceservice.TraceService/Remove", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *traceServiceClient) List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error) {
	out := new(ListResponse)
	err := grpc.Invoke(ctx, "/traceservice.TraceService/List", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *traceServiceClient) GetValues(ctx context.Context, in *GetValuesRequest, opts ...grpc.CallOption) (*GetValuesResponse, error) {
	out := new(GetValuesResponse)
	err := grpc.Invoke(ctx, "/traceservice.TraceService/GetValues", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *traceServiceClient) GetParams(ctx context.Context, in *GetParamsRequest, opts ...grpc.CallOption) (*GetParamsResponse, error) {
	out := new(GetParamsResponse)
	err := grpc.Invoke(ctx, "/traceservice.TraceService/GetParams", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *traceServiceClient) GetValuesRaw(ctx context.Context, in *GetValuesRequest, opts ...grpc.CallOption) (*GetValuesRawResponse, error) {
	out := new(GetValuesRawResponse)
	err := grpc.Invoke(ctx, "/traceservice.TraceService/GetValuesRaw", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *traceServiceClient) GetTraceIDs(ctx context.Context, in *GetTraceIDsRequest, opts ...grpc.CallOption) (*GetTraceIDsResponse, error) {
	out := new(GetTraceIDsResponse)
	err := grpc.Invoke(ctx, "/traceservice.TraceService/GetTraceIDs", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *traceServiceClient) ListMD5(ctx context.Context, in *ListMD5Request, opts ...grpc.CallOption) (*ListMD5Response, error) {
	out := new(ListMD5Response)
	err := grpc.Invoke(ctx, "/traceservice.TraceService/ListMD5", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *traceServiceClient) Ping(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := grpc.Invoke(ctx, "/traceservice.TraceService/Ping", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for TraceService service

type TraceServiceServer interface {
	// Returns a list of traceids that don't have Params stored in the datastore.
	MissingParams(context.Context, *MissingParamsRequest) (*MissingParamsResponse, error)
	// Adds Params for a set of traceids.
	AddParams(context.Context, *AddParamsRequest) (*Empty, error)
	// Adds data for a set of traces for a particular commitid.
	Add(context.Context, *AddRequest) (*Empty, error)
	// Removes data for a particular commitid.
	Remove(context.Context, *RemoveRequest) (*Empty, error)
	// List returns all the CommitIDs that exist in the given time range.
	List(context.Context, *ListRequest) (*ListResponse, error)
	// GetValues returns all the trace values stored for the given CommitID.
	GetValues(context.Context, *GetValuesRequest) (*GetValuesResponse, error)
	// GetParams returns the Params for all of the given traces.
	GetParams(context.Context, *GetParamsRequest) (*GetParamsResponse, error)
	// GetValuesRaw returns all the trace values stored for the given CommitID in
	// the raw format stored in BoltDB. The decoding can be done by calling
	// NewCommitInfo() on the returned byte slice.
	GetValuesRaw(context.Context, *GetValuesRequest) (*GetValuesRawResponse, error)
	// GetTraceIDs returns the traceids for the given trace64ids. These are used
	// in decoding the bytes returned from GetValuesRaw.
	GetTraceIDs(context.Context, *GetTraceIDsRequest) (*GetTraceIDsResponse, error)
	// ListMD5 returns the MD5 hashes for the given CommitIDs.
	ListMD5(context.Context, *ListMD5Request) (*ListMD5Response, error)
	// Ping should always succeed. Used to test if the service is up and
	// running.
	Ping(context.Context, *Empty) (*Empty, error)
}

func RegisterTraceServiceServer(s *grpc.Server, srv TraceServiceServer) {
	s.RegisterService(&_TraceService_serviceDesc, srv)
}

func _TraceService_MissingParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MissingParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TraceServiceServer).MissingParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/traceservice.TraceService/MissingParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TraceServiceServer).MissingParams(ctx, req.(*MissingParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TraceService_AddParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TraceServiceServer).AddParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/traceservice.TraceService/AddParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TraceServiceServer).AddParams(ctx, req.(*AddParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TraceService_Add_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TraceServiceServer).Add(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/traceservice.TraceService/Add",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TraceServiceServer).Add(ctx, req.(*AddRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TraceService_Remove_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TraceServiceServer).Remove(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/traceservice.TraceService/Remove",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TraceServiceServer).Remove(ctx, req.(*RemoveRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TraceService_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TraceServiceServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/traceservice.TraceService/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TraceServiceServer).List(ctx, req.(*ListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TraceService_GetValues_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetValuesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TraceServiceServer).GetValues(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/traceservice.TraceService/GetValues",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TraceServiceServer).GetValues(ctx, req.(*GetValuesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TraceService_GetParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TraceServiceServer).GetParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/traceservice.TraceService/GetParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TraceServiceServer).GetParams(ctx, req.(*GetParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TraceService_GetValuesRaw_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetValuesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TraceServiceServer).GetValuesRaw(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/traceservice.TraceService/GetValuesRaw",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TraceServiceServer).GetValuesRaw(ctx, req.(*GetValuesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TraceService_GetTraceIDs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTraceIDsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TraceServiceServer).GetTraceIDs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/traceservice.TraceService/GetTraceIDs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TraceServiceServer).GetTraceIDs(ctx, req.(*GetTraceIDsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TraceService_ListMD5_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListMD5Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TraceServiceServer).ListMD5(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/traceservice.TraceService/ListMD5",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TraceServiceServer).ListMD5(ctx, req.(*ListMD5Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _TraceService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TraceServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/traceservice.TraceService/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TraceServiceServer).Ping(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _TraceService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "traceservice.TraceService",
	HandlerType: (*TraceServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "MissingParams",
			Handler:    _TraceService_MissingParams_Handler,
		},
		{
			MethodName: "AddParams",
			Handler:    _TraceService_AddParams_Handler,
		},
		{
			MethodName: "Add",
			Handler:    _TraceService_Add_Handler,
		},
		{
			MethodName: "Remove",
			Handler:    _TraceService_Remove_Handler,
		},
		{
			MethodName: "List",
			Handler:    _TraceService_List_Handler,
		},
		{
			MethodName: "GetValues",
			Handler:    _TraceService_GetValues_Handler,
		},
		{
			MethodName: "GetParams",
			Handler:    _TraceService_GetParams_Handler,
		},
		{
			MethodName: "GetValuesRaw",
			Handler:    _TraceService_GetValuesRaw_Handler,
		},
		{
			MethodName: "GetTraceIDs",
			Handler:    _TraceService_GetTraceIDs_Handler,
		},
		{
			MethodName: "ListMD5",
			Handler:    _TraceService_ListMD5_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _TraceService_Ping_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "traceservice.proto",
}

func init() { proto.RegisterFile("traceservice.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 780 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x56, 0xdf, 0x53, 0xd3, 0x4a,
	0x14, 0x26, 0x4d, 0x5b, 0xc8, 0x49, 0xe1, 0x72, 0x97, 0x5e, 0x6e, 0x8c, 0xa8, 0x75, 0xe5, 0xa1,
	0x33, 0x3a, 0xa8, 0x81, 0x32, 0xa8, 0x8c, 0x0e, 0x50, 0x04, 0x66, 0xc4, 0xa9, 0x81, 0xe1, 0xc1,
	0xb7, 0xd0, 0xdd, 0x61, 0x32, 0x9a, 0xa6, 0x24, 0x01, 0x87, 0x07, 0xff, 0x0e, 0xff, 0x0c, 0xff,
	0x45, 0x27, 0xbb, 0x9b, 0xcd, 0x8f, 0x26, 0x80, 0xe2, 0x13, 0xd9, 0xdd, 0x73, 0xbe, 0xfd, 0xbe,
	0xb3, 0xdf, 0x39, 0x14, 0x50, 0x14, 0x38, 0x43, 0x1a, 0xd2, 0xe0, 0xd2, 0x1d, 0xd2, 0x95, 0x71,
	0xe0, 0x47, 0x3e, 0x6a, 0x65, 0xf7, 0xf0, 0x34, 0x34, 0x76, 0xbd, 0x71, 0x74, 0x85, 0x07, 0x30,
	0xb3, 0xe3, 0x7b, 0x9e, 0x1b, 0x1d, 0xf4, 0xd1, 0x1c, 0xd4, 0x5c, 0x62, 0x28, 0x1d, 0xa5, 0xab,
	0xd9, 0x35, 0x97, 0xa0, 0x45, 0x68, 0x86, 0xfe, 0x45, 0x30, 0xa4, 0x46, 0x8d, 0xed, 0x89, 0x15,
	0x5a, 0x02, 0x2d, 0x72, 0x3d, 0x1a, 0x46, 0x8e, 0x37, 0x36, 0xd4, 0x8e, 0xd2, 0x55, 0xed, 0x74,
	0x03, 0x7f, 0x87, 0xe6, 0xc0, 0x09, 0x1c, 0x2f, 0x44, 0x1b, 0xd0, 0x1c, 0xb3, 0x2f, 0x43, 0xe9,
	0xa8, 0x5d, 0xdd, 0xea, 0xac, 0xe4, 0x78, 0xf1, 0x28, 0xf1, 0x67, 0x77, 0x14, 0x05, 0x57, 0xb6,
	0x88, 0x37, 0x5f, 0x81, 0x9e, 0xd9, 0x46, 0xf3, 0xa0, 0x7e, 0xa1, 0x57, 0x82, 0x59, 0xfc, 0x89,
	0xda, 0xd0, 0xb8, 0x74, 0xbe, 0x5e, 0x24, 0xcc, 0xf8, 0xe2, 0x75, 0x6d, 0x43, 0xc1, 0x16, 0xb4,
	0x0f, 0xdd, 0x30, 0x74, 0x47, 0x67, 0x1c, 0xc1, 0xa6, 0xe7, 0x17, 0x34, 0x8c, 0x90, 0x09, 0x33,
	0xec, 0x76, 0x97, 0x70, 0x3a, 0x9a, 0x2d, 0xd7, 0x78, 0x15, 0xfe, 0x2b, 0xe4, 0x84, 0x63, 0x7f,
	0x14, 0xd2, 0x6b, 0x93, 0x7e, 0x28, 0x00, 0x3c, 0x7c, 0xe0, 0xb8, 0x41, 0x09, 0xc7, 0x4d, 0x29,
	0xbf, 0xc6, 0xe4, 0x2f, 0x97, 0xc9, 0x8f, 0x73, 0xff, 0x76, 0x09, 0xfa, 0x30, 0xbf, 0x45, 0x48,
	0x5e, 0xfe, 0x0b, 0x49, 0xa6, 0xce, 0xc8, 0x18, 0x55, 0x64, 0x12, 0x02, 0xf8, 0x0d, 0xe8, 0x47,
	0x91, 0x1f, 0x50, 0xc2, 0x09, 0x3c, 0xcb, 0xa8, 0x51, 0xba, 0xba, 0xd5, 0x2e, 0x03, 0x90, 0xc9,
	0xab, 0xa0, 0x9d, 0xc4, 0x7c, 0x2a, 0x4a, 0x93, 0xe3, 0xde, 0x12, 0xdc, 0xf1, 0x39, 0xc0, 0x16,
	0x21, 0x09, 0x63, 0x0b, 0x66, 0x86, 0xcc, 0x99, 0xc2, 0x93, 0xba, 0xb5, 0x98, 0xbf, 0x32, 0xf1,
	0xad, 0x2d, 0xe3, 0xd0, 0x73, 0x68, 0x32, 0xa8, 0xd0, 0x50, 0x99, 0xca, 0xff, 0xf3, 0x19, 0x92,
	0x92, 0x2d, 0xc2, 0xf0, 0x0e, 0xcc, 0xda, 0xd4, 0xf3, 0x2f, 0xe9, 0x1d, 0x6e, 0xc5, 0x3d, 0xd0,
	0x3f, 0xb8, 0x61, 0x94, 0x40, 0xb4, 0xa1, 0x71, 0x4a, 0xcf, 0xdc, 0x11, 0xcb, 0x57, 0x6d, 0xbe,
	0x88, 0x8b, 0x40, 0x47, 0x84, 0x09, 0x56, 0xed, 0xf8, 0x13, 0xf7, 0xa1, 0xc5, 0xd3, 0x84, 0xd9,
	0xd6, 0x40, 0x4b, 0x20, 0x13, 0xfe, 0x55, 0x77, 0xa7, 0x81, 0xf8, 0x3d, 0xcc, 0xef, 0xd1, 0x88,
	0x29, 0x0b, 0xef, 0x22, 0xe2, 0x04, 0xfe, 0xcd, 0xe0, 0x08, 0x4a, 0x69, 0x3d, 0xeb, 0xb7, 0xaa,
	0x67, 0xac, 0xd2, 0x23, 0x3d, 0xa3, 0xc1, 0x9f, 0xda, 0x23, 0x3d, 0xbc, 0xc2, 0xf8, 0xdd, 0xbe,
	0x17, 0x77, 0x19, 0x8f, 0x42, 0x1f, 0xfe, 0xbe, 0x7b, 0xdf, 0x42, 0x3b, 0x95, 0xe3, 0x7c, 0x93,
	0x48, 0xd2, 0x79, 0x4a, 0xc6, 0x79, 0x09, 0xed, 0x5a, 0x4a, 0x7b, 0x19, 0xd0, 0x1e, 0x8d, 0x8e,
	0xe3, 0x5b, 0x0e, 0xfa, 0x92, 0x78, 0x32, 0x21, 0xd5, 0x6e, 0x3d, 0x9e, 0x90, 0xf8, 0x25, 0xe8,
	0x22, 0x84, 0x19, 0x1d, 0x41, 0xdd, 0x25, 0xeb, 0x6b, 0x0c, 0xbb, 0x6e, 0xb3, 0x6f, 0x91, 0x52,
	0x4b, 0x86, 0x2a, 0xde, 0x86, 0x85, 0x1c, 0xb0, 0xe0, 0xf5, 0x14, 0xd4, 0xa4, 0x1a, 0xba, 0x75,
	0x2f, 0x2f, 0x2f, 0x73, 0x85, 0x1d, 0x47, 0xe1, 0x3e, 0xcc, 0xc5, 0xce, 0x39, 0xec, 0xf7, 0xca,
	0x5f, 0x5c, 0xbd, 0xd5, 0x8b, 0x7f, 0x02, 0x8d, 0xef, 0x1e, 0xf6, 0x7b, 0x7f, 0xd4, 0x6d, 0x93,
	0x55, 0xdb, 0x87, 0x7f, 0x24, 0x31, 0x21, 0xac, 0x97, 0xb8, 0x3a, 0x0e, 0x55, 0xca, 0x5c, 0x24,
	0x49, 0xd8, 0x69, 0xa4, 0xf5, 0xb3, 0x09, 0x2d, 0xa6, 0xfb, 0x88, 0x47, 0xa1, 0xcf, 0x30, 0x9b,
	0x9b, 0xd1, 0x08, 0xe7, 0x51, 0xca, 0x86, 0xbe, 0xf9, 0xe4, 0xda, 0x18, 0xce, 0x10, 0x4f, 0xa1,
	0x6d, 0xd0, 0xe4, 0xc0, 0x44, 0x0f, 0xf3, 0x39, 0xc5, 0x49, 0x6a, 0x2e, 0xe4, 0xcf, 0xf9, 0xbf,
	0xd1, 0x29, 0xb4, 0x0e, 0xea, 0x16, 0x21, 0xc8, 0x98, 0xc8, 0xbe, 0x21, 0x6f, 0x13, 0x9a, 0x7c,
	0x02, 0xa1, 0xfb, 0xf9, 0x80, 0xdc, 0x5c, 0xaa, 0xca, 0x7e, 0x07, 0xf5, 0xb8, 0xe0, 0xa8, 0xe0,
	0x98, 0xcc, 0x38, 0x32, 0xcd, 0xb2, 0x23, 0x29, 0xfd, 0x23, 0x68, 0xb2, 0x4f, 0x8a, 0xd2, 0x8b,
	0x73, 0xc5, 0x7c, 0x54, 0x79, 0x5e, 0xc0, 0x2b, 0x2f, 0x65, 0x71, 0x0e, 0x94, 0xe0, 0x4d, 0x3c,
	0xcd, 0x31, 0xb4, 0xb2, 0x7d, 0x7c, 0x23, 0x45, 0x5c, 0x75, 0x9e, 0xce, 0x00, 0x86, 0xaa, 0x67,
	0x9a, 0x10, 0x75, 0x26, 0x92, 0x0a, 0x8d, 0x6f, 0x3e, 0xbe, 0x26, 0x42, 0xa2, 0xee, 0xc3, 0xb4,
	0x70, 0x3f, 0x5a, 0x9a, 0x2c, 0x7a, 0xda, 0xad, 0xe6, 0x83, 0x8a, 0x53, 0x89, 0x64, 0x41, 0x7d,
	0xe0, 0x8e, 0xce, 0x50, 0xd9, 0xab, 0x57, 0x58, 0xe1, 0xb4, 0xc9, 0x7e, 0xe7, 0xad, 0xfe, 0x0a,
	0x00, 0x00, 0xff, 0xff, 0xcf, 0x97, 0x2e, 0x23, 0xfd, 0x09, 0x00, 0x00,
}
