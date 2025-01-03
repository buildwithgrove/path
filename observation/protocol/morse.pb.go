// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v5.28.3
// source: path/protocol/morse.proto

package protocol

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

// MorseObservations holds Morse-specific observations gathered through sending
// relay(s) to handle a service request.
type MorseObservations struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// gigastake_app_address is the address of the application on behalf of which the
	// relay(s) associated with the service request was signed.
	GigastakeAppAddress string `protobuf:"bytes,1,opt,name=gigastake_app_address,json=gigastakeAppAddress,proto3" json:"gigastake_app_address,omitempty"`
}

func (x *MorseObservations) Reset() {
	*x = MorseObservations{}
	mi := &file_path_protocol_morse_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MorseObservations) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MorseObservations) ProtoMessage() {}

func (x *MorseObservations) ProtoReflect() protoreflect.Message {
	mi := &file_path_protocol_morse_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MorseObservations.ProtoReflect.Descriptor instead.
func (*MorseObservations) Descriptor() ([]byte, []int) {
	return file_path_protocol_morse_proto_rawDescGZIP(), []int{0}
}

func (x *MorseObservations) GetGigastakeAppAddress() string {
	if x != nil {
		return x.GigastakeAppAddress
	}
	return ""
}

type MorseObservationsList struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Observations []*MorseObservations `protobuf:"bytes,1,rep,name=observations,proto3" json:"observations,omitempty"`
}

func (x *MorseObservationsList) Reset() {
	*x = MorseObservationsList{}
	mi := &file_path_protocol_morse_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MorseObservationsList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MorseObservationsList) ProtoMessage() {}

func (x *MorseObservationsList) ProtoReflect() protoreflect.Message {
	mi := &file_path_protocol_morse_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MorseObservationsList.ProtoReflect.Descriptor instead.
func (*MorseObservationsList) Descriptor() ([]byte, []int) {
	return file_path_protocol_morse_proto_rawDescGZIP(), []int{1}
}

func (x *MorseObservationsList) GetObservations() []*MorseObservations {
	if x != nil {
		return x.Observations
	}
	return nil
}

var File_path_protocol_morse_proto protoreflect.FileDescriptor

var file_path_protocol_morse_proto_rawDesc = []byte{
	0x0a, 0x19, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2f,
	0x6d, 0x6f, 0x72, 0x73, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x70, 0x61, 0x74,
	0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x22, 0x47, 0x0a, 0x11, 0x4d, 0x6f,
	0x72, 0x73, 0x65, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12,
	0x32, 0x0a, 0x15, 0x67, 0x69, 0x67, 0x61, 0x73, 0x74, 0x61, 0x6b, 0x65, 0x5f, 0x61, 0x70, 0x70,
	0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x13,
	0x67, 0x69, 0x67, 0x61, 0x73, 0x74, 0x61, 0x6b, 0x65, 0x41, 0x70, 0x70, 0x41, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x22, 0x5d, 0x0a, 0x15, 0x4d, 0x6f, 0x72, 0x73, 0x65, 0x4f, 0x62, 0x73, 0x65,
	0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x44, 0x0a, 0x0c,
	0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x20, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63,
	0x6f, 0x6c, 0x2e, 0x4d, 0x6f, 0x72, 0x73, 0x65, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x52, 0x0c, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x42, 0x35, 0x5a, 0x33, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x77, 0x69, 0x74, 0x68, 0x67, 0x72, 0x6f, 0x76, 0x65, 0x2f,
	0x70, 0x61, 0x74, 0x68, 0x2f, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_path_protocol_morse_proto_rawDescOnce sync.Once
	file_path_protocol_morse_proto_rawDescData = file_path_protocol_morse_proto_rawDesc
)

func file_path_protocol_morse_proto_rawDescGZIP() []byte {
	file_path_protocol_morse_proto_rawDescOnce.Do(func() {
		file_path_protocol_morse_proto_rawDescData = protoimpl.X.CompressGZIP(file_path_protocol_morse_proto_rawDescData)
	})
	return file_path_protocol_morse_proto_rawDescData
}

var file_path_protocol_morse_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_path_protocol_morse_proto_goTypes = []any{
	(*MorseObservations)(nil),     // 0: path.protocol.MorseObservations
	(*MorseObservationsList)(nil), // 1: path.protocol.MorseObservationsList
}
var file_path_protocol_morse_proto_depIdxs = []int32{
	0, // 0: path.protocol.MorseObservationsList.observations:type_name -> path.protocol.MorseObservations
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_path_protocol_morse_proto_init() }
func file_path_protocol_morse_proto_init() {
	if File_path_protocol_morse_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_path_protocol_morse_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_path_protocol_morse_proto_goTypes,
		DependencyIndexes: file_path_protocol_morse_proto_depIdxs,
		MessageInfos:      file_path_protocol_morse_proto_msgTypes,
	}.Build()
	File_path_protocol_morse_proto = out.File
	file_path_protocol_morse_proto_rawDesc = nil
	file_path_protocol_morse_proto_goTypes = nil
	file_path_protocol_morse_proto_depIdxs = nil
}