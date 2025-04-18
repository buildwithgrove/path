// To generate the Go code from this proto file, run: `make proto_generate`
// See `proto.mk` for more details.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v5.28.3
// source: path/auth.proto

package observation

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

// RequestAuth captures all fields related to the authentication of the request.
// These are all external to PATH, i.e. reported to PATH by the authentication layer.
// Used in generating observations for the data pipeline.
type RequestAuth struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Unique identifier for the request.
	// Used by the data pipeline.
	RequestId string `protobuf:"bytes,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	// The geographic region where the gateway serving the request was located.
	// DEV_NOTE: It aligns with typical cloud regions (e.g. us-east) but kept as a string for simplicity.
	// TODO_FUTURE: this may need to move into a separate message if more details regarding the PATH instance are required.
	Region string `protobuf:"bytes,2,opt,name=region,proto3" json:"region,omitempty"`
	// The ID of the Grove portal account behind the service request
	PortalAccountId string `protobuf:"bytes,3,opt,name=portal_account_id,json=portalAccountId,proto3" json:"portal_account_id,omitempty"`
	// The ID of the Grove portal application authenticating the service request.
	PortalApplicationId string `protobuf:"bytes,4,opt,name=portal_application_id,json=portalApplicationId,proto3" json:"portal_application_id,omitempty"`
}

func (x *RequestAuth) Reset() {
	*x = RequestAuth{}
	mi := &file_path_auth_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RequestAuth) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequestAuth) ProtoMessage() {}

func (x *RequestAuth) ProtoReflect() protoreflect.Message {
	mi := &file_path_auth_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequestAuth.ProtoReflect.Descriptor instead.
func (*RequestAuth) Descriptor() ([]byte, []int) {
	return file_path_auth_proto_rawDescGZIP(), []int{0}
}

func (x *RequestAuth) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

func (x *RequestAuth) GetRegion() string {
	if x != nil {
		return x.Region
	}
	return ""
}

func (x *RequestAuth) GetPortalAccountId() string {
	if x != nil {
		return x.PortalAccountId
	}
	return ""
}

func (x *RequestAuth) GetPortalApplicationId() string {
	if x != nil {
		return x.PortalApplicationId
	}
	return ""
}

var File_path_auth_proto protoreflect.FileDescriptor

var file_path_auth_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x04, 0x70, 0x61, 0x74, 0x68, 0x22, 0xa4, 0x01, 0x0a, 0x0b, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x41, 0x75, 0x74, 0x68, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x12, 0x2a,
	0x0a, 0x11, 0x70, 0x6f, 0x72, 0x74, 0x61, 0x6c, 0x5f, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x70, 0x6f, 0x72, 0x74, 0x61,
	0x6c, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x32, 0x0a, 0x15, 0x70, 0x6f,
	0x72, 0x74, 0x61, 0x6c, 0x5f, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x5f, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x13, 0x70, 0x6f, 0x72, 0x74, 0x61,
	0x6c, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x42, 0x2c,
	0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x75, 0x69,
	0x6c, 0x64, 0x77, 0x69, 0x74, 0x68, 0x67, 0x72, 0x6f, 0x76, 0x65, 0x2f, 0x70, 0x61, 0x74, 0x68,
	0x2f, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_path_auth_proto_rawDescOnce sync.Once
	file_path_auth_proto_rawDescData = file_path_auth_proto_rawDesc
)

func file_path_auth_proto_rawDescGZIP() []byte {
	file_path_auth_proto_rawDescOnce.Do(func() {
		file_path_auth_proto_rawDescData = protoimpl.X.CompressGZIP(file_path_auth_proto_rawDescData)
	})
	return file_path_auth_proto_rawDescData
}

var file_path_auth_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_path_auth_proto_goTypes = []any{
	(*RequestAuth)(nil), // 0: path.RequestAuth
}
var file_path_auth_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_path_auth_proto_init() }
func file_path_auth_proto_init() {
	if File_path_auth_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_path_auth_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_path_auth_proto_goTypes,
		DependencyIndexes: file_path_auth_proto_depIdxs,
		MessageInfos:      file_path_auth_proto_msgTypes,
	}.Build()
	File_path_auth_proto = out.File
	file_path_auth_proto_rawDesc = nil
	file_path_auth_proto_goTypes = nil
	file_path_auth_proto_depIdxs = nil
}
