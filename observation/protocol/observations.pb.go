// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.3
// source: path/protocol/observations.proto

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

// Observations aggregates protocol-level observations collected during service request processing.
type Observations struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Height of the blockchain block when processing the service request through a relay
	BlockHeight uint64 `protobuf:"varint,1,opt,name=block_height,json=blockHeight,proto3" json:"block_height,omitempty"`
	// Protocol-specific observations.
	// Only one protocol can be associated with a single observation.
	//
	// Types that are assignable to Protocol:
	//
	//	*Observations_Morse
	//	*Observations_Shannon
	Protocol isObservations_Protocol `protobuf_oneof:"protocol"`
}

func (x *Observations) Reset() {
	*x = Observations{}
	if protoimpl.UnsafeEnabled {
		mi := &file_path_protocol_observations_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Observations) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Observations) ProtoMessage() {}

func (x *Observations) ProtoReflect() protoreflect.Message {
	mi := &file_path_protocol_observations_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Observations.ProtoReflect.Descriptor instead.
func (*Observations) Descriptor() ([]byte, []int) {
	return file_path_protocol_observations_proto_rawDescGZIP(), []int{0}
}

func (x *Observations) GetBlockHeight() uint64 {
	if x != nil {
		return x.BlockHeight
	}
	return 0
}

func (m *Observations) GetProtocol() isObservations_Protocol {
	if m != nil {
		return m.Protocol
	}
	return nil
}

func (x *Observations) GetMorse() *MorseObservationsList {
	if x, ok := x.GetProtocol().(*Observations_Morse); ok {
		return x.Morse
	}
	return nil
}

func (x *Observations) GetShannon() *ShannonObservationsList {
	if x, ok := x.GetProtocol().(*Observations_Shannon); ok {
		return x.Shannon
	}
	return nil
}

type isObservations_Protocol interface {
	isObservations_Protocol()
}

type Observations_Morse struct {
	// Morse protocol-specific observations
	Morse *MorseObservationsList `protobuf:"bytes,2,opt,name=morse,proto3,oneof"`
}

type Observations_Shannon struct {
	// Shannon protocol-specific observations
	Shannon *ShannonObservationsList `protobuf:"bytes,3,opt,name=shannon,proto3,oneof"`
}

func (*Observations_Morse) isObservations_Protocol() {}

func (*Observations_Shannon) isObservations_Protocol() {}

var File_path_protocol_observations_proto protoreflect.FileDescriptor

var file_path_protocol_observations_proto_rawDesc = []byte{
	0x0a, 0x20, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2f,
	0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x0d, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f,
	0x6c, 0x1a, 0x1b, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c,
	0x2f, 0x73, 0x68, 0x61, 0x6e, 0x6e, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19,
	0x70, 0x61, 0x74, 0x68, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2f, 0x6d, 0x6f,
	0x72, 0x73, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xbf, 0x01, 0x0a, 0x0c, 0x4f, 0x62,
	0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x21, 0x0a, 0x0c, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x5f, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x0b, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x3c, 0x0a,
	0x05, 0x6d, 0x6f, 0x72, 0x73, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x70,
	0x61, 0x74, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x4d, 0x6f, 0x72,
	0x73, 0x65, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x4c, 0x69,
	0x73, 0x74, 0x48, 0x00, 0x52, 0x05, 0x6d, 0x6f, 0x72, 0x73, 0x65, 0x12, 0x42, 0x0a, 0x07, 0x73,
	0x68, 0x61, 0x6e, 0x6e, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x70,
	0x61, 0x74, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x53, 0x68, 0x61,
	0x6e, 0x6e, 0x6f, 0x6e, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x4c, 0x69, 0x73, 0x74, 0x48, 0x00, 0x52, 0x07, 0x73, 0x68, 0x61, 0x6e, 0x6e, 0x6f, 0x6e, 0x42,
	0x0a, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x42, 0x35, 0x5a, 0x33, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x77,
	0x69, 0x74, 0x68, 0x67, 0x72, 0x6f, 0x76, 0x65, 0x2f, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x6f, 0x62,
	0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63,
	0x6f, 0x6c, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_path_protocol_observations_proto_rawDescOnce sync.Once
	file_path_protocol_observations_proto_rawDescData = file_path_protocol_observations_proto_rawDesc
)

func file_path_protocol_observations_proto_rawDescGZIP() []byte {
	file_path_protocol_observations_proto_rawDescOnce.Do(func() {
		file_path_protocol_observations_proto_rawDescData = protoimpl.X.CompressGZIP(file_path_protocol_observations_proto_rawDescData)
	})
	return file_path_protocol_observations_proto_rawDescData
}

var file_path_protocol_observations_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_path_protocol_observations_proto_goTypes = []any{
	(*Observations)(nil),            // 0: path.protocol.Observations
	(*MorseObservationsList)(nil),   // 1: path.protocol.MorseObservationsList
	(*ShannonObservationsList)(nil), // 2: path.protocol.ShannonObservationsList
}
var file_path_protocol_observations_proto_depIdxs = []int32{
	1, // 0: path.protocol.Observations.morse:type_name -> path.protocol.MorseObservationsList
	2, // 1: path.protocol.Observations.shannon:type_name -> path.protocol.ShannonObservationsList
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_path_protocol_observations_proto_init() }
func file_path_protocol_observations_proto_init() {
	if File_path_protocol_observations_proto != nil {
		return
	}
	file_path_protocol_shannon_proto_init()
	file_path_protocol_morse_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_path_protocol_observations_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*Observations); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_path_protocol_observations_proto_msgTypes[0].OneofWrappers = []any{
		(*Observations_Morse)(nil),
		(*Observations_Shannon)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_path_protocol_observations_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_path_protocol_observations_proto_goTypes,
		DependencyIndexes: file_path_protocol_observations_proto_depIdxs,
		MessageInfos:      file_path_protocol_observations_proto_msgTypes,
	}.Build()
	File_path_protocol_observations_proto = out.File
	file_path_protocol_observations_proto_rawDesc = nil
	file_path_protocol_observations_proto_goTypes = nil
	file_path_protocol_observations_proto_depIdxs = nil
}
