// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v5.28.3
// source: path/qos/jsonrpc.proto

package qos

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type JsonRpcRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id     string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Method string `protobuf:"bytes,2,opt,name=method,proto3" json:"method,omitempty"`
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

type JsonRpcResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id     string                `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Result string                `protobuf:"bytes,2,opt,name=result,proto3" json:"result,omitempty"`
	Err    *JsonRpcResponseError `protobuf:"bytes,3,opt,name=err,proto3,oneof" json:"err,omitempty"`
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

type JsonRpcResponseError struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code    int64  `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
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

var file_path_qos_jsonrpc_proto_rawDesc = []byte{
	0x0a, 0x16, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x71, 0x6f, 0x73, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x72,
	0x70, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x71,
	0x6f, 0x73, 0x22, 0x38, 0x0a, 0x0e, 0x4a, 0x73, 0x6f, 0x6e, 0x52, 0x70, 0x63, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x22, 0x78, 0x0a, 0x0f,
	0x4a, 0x73, 0x6f, 0x6e, 0x52, 0x70, 0x63, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12,
	0x16, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x35, 0x0a, 0x03, 0x65, 0x72, 0x72, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x71, 0x6f, 0x73, 0x2e,
	0x4a, 0x73, 0x6f, 0x6e, 0x52, 0x70, 0x63, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x45,
	0x72, 0x72, 0x6f, 0x72, 0x48, 0x00, 0x52, 0x03, 0x65, 0x72, 0x72, 0x88, 0x01, 0x01, 0x42, 0x06,
	0x0a, 0x04, 0x5f, 0x65, 0x72, 0x72, 0x22, 0x44, 0x0a, 0x14, 0x4a, 0x73, 0x6f, 0x6e, 0x52, 0x70,
	0x63, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x12,
	0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x63, 0x6f,
	0x64, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x42, 0x30, 0x5a, 0x2e,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x75, 0x69, 0x6c, 0x64,
	0x77, 0x69, 0x74, 0x68, 0x67, 0x72, 0x6f, 0x76, 0x65, 0x2f, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x6f,
	0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x71, 0x6f, 0x73, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_path_qos_jsonrpc_proto_rawDescOnce sync.Once
	file_path_qos_jsonrpc_proto_rawDescData = file_path_qos_jsonrpc_proto_rawDesc
)

func file_path_qos_jsonrpc_proto_rawDescGZIP() []byte {
	file_path_qos_jsonrpc_proto_rawDescOnce.Do(func() {
		file_path_qos_jsonrpc_proto_rawDescData = protoimpl.X.CompressGZIP(file_path_qos_jsonrpc_proto_rawDescData)
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
			RawDescriptor: file_path_qos_jsonrpc_proto_rawDesc,
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
	file_path_qos_jsonrpc_proto_rawDesc = nil
	file_path_qos_jsonrpc_proto_goTypes = nil
	file_path_qos_jsonrpc_proto_depIdxs = nil
}