// To generate the Go code from this proto file, run: `make proto_generate`
// See `proto.mk` for more details.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: path/auth.proto

package observation

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

// RequestAuth captures all fields related to the authentication of the request.
// These are all external to PATH, i.e. reported to PATH by the authentication layer.
// Used in generating observations for the data pipeline.
type RequestAuth struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Unique identifier for tracing the service request.
	// Used by the data pipeline.
	TraceId string `protobuf:"bytes,1,opt,name=trace_id,json=traceId,proto3" json:"trace_id,omitempty"`
	// The geographic region where the gateway serving the request was located.
	// DEV_NOTE: It aligns with typical cloud regions (e.g. us-east) but kept as a string for simplicity.
	// TODO_FUTURE: this may need to move into a separate message if more details regarding the PATH instance are required.
	Region string `protobuf:"bytes,2,opt,name=region,proto3" json:"region,omitempty"`
	// Tracks Grove portal credentials.
	PortalCredentials *PortalCredentials `protobuf:"bytes,3,opt,name=portal_credentials,json=portalCredentials,proto3,oneof" json:"portal_credentials,omitempty"`
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
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

func (x *RequestAuth) GetTraceId() string {
	if x != nil {
		return x.TraceId
	}
	return ""
}

func (x *RequestAuth) GetRegion() string {
	if x != nil {
		return x.Region
	}
	return ""
}

func (x *RequestAuth) GetPortalCredentials() *PortalCredentials {
	if x != nil {
		return x.PortalCredentials
	}
	return nil
}

// PortalCredentials captures fields related to the Grove portal's request authentication.
type PortalCredentials struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The ID of the Grove portal account behind the service request
	PortalAccountId string `protobuf:"bytes,3,opt,name=portal_account_id,json=portalAccountId,proto3" json:"portal_account_id,omitempty"`
	// The ID of the Grove portal application authenticating the service request.
	PortalApplicationId string `protobuf:"bytes,4,opt,name=portal_application_id,json=portalApplicationId,proto3" json:"portal_application_id,omitempty"`
	unknownFields       protoimpl.UnknownFields
	sizeCache           protoimpl.SizeCache
}

func (x *PortalCredentials) Reset() {
	*x = PortalCredentials{}
	mi := &file_path_auth_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PortalCredentials) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PortalCredentials) ProtoMessage() {}

func (x *PortalCredentials) ProtoReflect() protoreflect.Message {
	mi := &file_path_auth_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PortalCredentials.ProtoReflect.Descriptor instead.
func (*PortalCredentials) Descriptor() ([]byte, []int) {
	return file_path_auth_proto_rawDescGZIP(), []int{1}
}

func (x *PortalCredentials) GetPortalAccountId() string {
	if x != nil {
		return x.PortalAccountId
	}
	return ""
}

func (x *PortalCredentials) GetPortalApplicationId() string {
	if x != nil {
		return x.PortalApplicationId
	}
	return ""
}

var File_path_auth_proto protoreflect.FileDescriptor

const file_path_auth_proto_rawDesc = "" +
	"\n" +
	"\x0fpath/auth.proto\x12\x04path\"\xa4\x01\n" +
	"\vRequestAuth\x12\x19\n" +
	"\btrace_id\x18\x01 \x01(\tR\atraceId\x12\x16\n" +
	"\x06region\x18\x02 \x01(\tR\x06region\x12K\n" +
	"\x12portal_credentials\x18\x03 \x01(\v2\x17.path.PortalCredentialsH\x00R\x11portalCredentials\x88\x01\x01B\x15\n" +
	"\x13_portal_credentials\"s\n" +
	"\x11PortalCredentials\x12*\n" +
	"\x11portal_account_id\x18\x03 \x01(\tR\x0fportalAccountId\x122\n" +
	"\x15portal_application_id\x18\x04 \x01(\tR\x13portalApplicationIdB,Z*github.com/buildwithgrove/path/observationb\x06proto3"

var (
	file_path_auth_proto_rawDescOnce sync.Once
	file_path_auth_proto_rawDescData []byte
)

func file_path_auth_proto_rawDescGZIP() []byte {
	file_path_auth_proto_rawDescOnce.Do(func() {
		file_path_auth_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_path_auth_proto_rawDesc), len(file_path_auth_proto_rawDesc)))
	})
	return file_path_auth_proto_rawDescData
}

var file_path_auth_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_path_auth_proto_goTypes = []any{
	(*RequestAuth)(nil),       // 0: path.RequestAuth
	(*PortalCredentials)(nil), // 1: path.PortalCredentials
}
var file_path_auth_proto_depIdxs = []int32{
	1, // 0: path.RequestAuth.portal_credentials:type_name -> path.PortalCredentials
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_path_auth_proto_init() }
func file_path_auth_proto_init() {
	if File_path_auth_proto != nil {
		return
	}
	file_path_auth_proto_msgTypes[0].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_path_auth_proto_rawDesc), len(file_path_auth_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_path_auth_proto_goTypes,
		DependencyIndexes: file_path_auth_proto_depIdxs,
		MessageInfos:      file_path_auth_proto_msgTypes,
	}.Build()
	File_path_auth_proto = out.File
	file_path_auth_proto_goTypes = nil
	file_path_auth_proto_depIdxs = nil
}
