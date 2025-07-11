// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: path/qos/jsonrpc.proto

package qos

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// JsonRpcRequest represents essential fields of a JSON-RPC request for observation purposes.
// Reference: https://www.jsonrpc.org/specification#request_object
type JsonRpcRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Client-established identifier. Must be a String, Number, or NULL if present.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Name of the JSON-RPC method being called (e.g., eth_chainId for EVM chains)
	Method        string `protobuf:"bytes,2,opt,name=method,proto3" json:"method,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *JsonRpcRequest) Reset() {
	*x = JsonRpcRequest{}
	mi := &file_path_qos_jsonrpc_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *JsonRpcRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JsonRpcRequest) ProtoMessage() {}

func (x *JsonRpcRequest) ProtoReflect() protoreflect.Message {
	mi := &file_path_qos_jsonrpc_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JsonRpcRequest.ProtoReflect.Descriptor instead.
func (*JsonRpcRequest) Descriptor() ([]byte, []int) {
	return file_path_qos_jsonrpc_proto_rawDescGZIP(), []int{0}
}

func (x *JsonRpcRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *JsonRpcRequest) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

// JsonRpcResponse represents essential fields of a JSON-RPC response for observation purposes.
// Reference: https://www.jsonrpc.org/specification#response_object
type JsonRpcResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Must match the id value from the corresponding request
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// JSON-serializable response data
	Result string `protobuf:"bytes,2,opt,name=result,proto3" json:"result,omitempty"`
	// Error details, if the request failed
	Err           *JsonRpcResponseError `protobuf:"bytes,3,opt,name=err,proto3,oneof" json:"err,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *JsonRpcResponse) Reset() {
	*x = JsonRpcResponse{}
	mi := &file_path_qos_jsonrpc_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *JsonRpcResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JsonRpcResponse) ProtoMessage() {}

func (x *JsonRpcResponse) ProtoReflect() protoreflect.Message {
	mi := &file_path_qos_jsonrpc_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JsonRpcResponse.ProtoReflect.Descriptor instead.
func (*JsonRpcResponse) Descriptor() ([]byte, []int) {
	return file_path_qos_jsonrpc_proto_rawDescGZIP(), []int{1}
}

func (x *JsonRpcResponse) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *JsonRpcResponse) GetResult() string {
	if x != nil {
		return x.Result
	}
	return ""
}

func (x *JsonRpcResponse) GetErr() *JsonRpcResponseError {
	if x != nil {
		return x.Err
	}
	return nil
}

// JsonRpcResponseError represents core error fields from a JSON-RPC response.
// Reference: https://www.jsonrpc.org/specification#error_object
//
// Only includes fields required for QoS observations.
type JsonRpcResponseError struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Error code indicating the type of failure
	Code int64 `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	// Human-readable error description
	Message       string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *JsonRpcResponseError) Reset() {
	*x = JsonRpcResponseError{}
	mi := &file_path_qos_jsonrpc_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *JsonRpcResponseError) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JsonRpcResponseError) ProtoMessage() {}

func (x *JsonRpcResponseError) ProtoReflect() protoreflect.Message {
	mi := &file_path_qos_jsonrpc_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JsonRpcResponseError.ProtoReflect.Descriptor instead.
func (*JsonRpcResponseError) Descriptor() ([]byte, []int) {
	return file_path_qos_jsonrpc_proto_rawDescGZIP(), []int{2}
}

func (x *JsonRpcResponseError) GetCode() int64 {
	if x != nil {
		return x.Code
	}
	return 0
}

func (x *JsonRpcResponseError) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_path_qos_jsonrpc_proto protoreflect.FileDescriptor

const file_path_qos_jsonrpc_proto_rawDesc = "" +
	"\n" +
	"\x16path/qos/jsonrpc.proto\x12\bpath.qos\"8\n" +
	"\x0eJsonRpcRequest\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x16\n" +
	"\x06method\x18\x02 \x01(\tR\x06method\"x\n" +
	"\x0fJsonRpcResponse\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x16\n" +
	"\x06result\x18\x02 \x01(\tR\x06result\x125\n" +
	"\x03err\x18\x03 \x01(\v2\x1e.path.qos.JsonRpcResponseErrorH\x00R\x03err\x88\x01\x01B\x06\n" +
	"\x04_err\"D\n" +
	"\x14JsonRpcResponseError\x12\x12\n" +
	"\x04code\x18\x01 \x01(\x03R\x04code\x12\x18\n" +
	"\amessage\x18\x02 \x01(\tR\amessageB0Z.github.com/buildwithgrove/path/observation/qosb\x06proto3"

var (
	file_path_qos_jsonrpc_proto_rawDescOnce sync.Once
	file_path_qos_jsonrpc_proto_rawDescData []byte
)

func file_path_qos_jsonrpc_proto_rawDescGZIP() []byte {
	file_path_qos_jsonrpc_proto_rawDescOnce.Do(func() {
		file_path_qos_jsonrpc_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_path_qos_jsonrpc_proto_rawDesc), len(file_path_qos_jsonrpc_proto_rawDesc)))
	})
	return file_path_qos_jsonrpc_proto_rawDescData
}

var file_path_qos_jsonrpc_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_path_qos_jsonrpc_proto_goTypes = []any{
	(*JsonRpcRequest)(nil),       // 0: path.qos.JsonRpcRequest
	(*JsonRpcResponse)(nil),      // 1: path.qos.JsonRpcResponse
	(*JsonRpcResponseError)(nil), // 2: path.qos.JsonRpcResponseError
}
var file_path_qos_jsonrpc_proto_depIdxs = []int32{
	2, // 0: path.qos.JsonRpcResponse.err:type_name -> path.qos.JsonRpcResponseError
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_path_qos_jsonrpc_proto_init() }
func file_path_qos_jsonrpc_proto_init() {
	if File_path_qos_jsonrpc_proto != nil {
		return
	}
	file_path_qos_jsonrpc_proto_msgTypes[1].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_path_qos_jsonrpc_proto_rawDesc), len(file_path_qos_jsonrpc_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_path_qos_jsonrpc_proto_goTypes,
		DependencyIndexes: file_path_qos_jsonrpc_proto_depIdxs,
		MessageInfos:      file_path_qos_jsonrpc_proto_msgTypes,
	}.Build()
	File_path_qos_jsonrpc_proto = out.File
	file_path_qos_jsonrpc_proto_goTypes = nil
	file_path_qos_jsonrpc_proto_depIdxs = nil
}
