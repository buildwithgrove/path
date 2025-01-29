// To generate the Go code from this proto file, run: `make proto_generate`
// See `proto.mk` for more details.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.3
// source: path/gateway.proto

package observation

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// RequestType captures the origin of the request.
// As of PR #72, it is one of:
//  1. Organic: a real user sent a service request to a PATH instance
//  2. Synthetic: internal infrastructure generated the service request for simulation and data purposes.
type RequestType int32

const (
	RequestType_REQUEST_TYPE_UNSPECIFIED RequestType = 0
	RequestType_REQUEST_TYPE_ORGANIC     RequestType = 1 // Service request sent by a user.
	RequestType_REQUEST_TYPE_SYNTHETIC   RequestType = 2 // Service request sent by the endpoint hydrator: see gateway/hydrator.go.
)

// Enum value maps for RequestType.
var (
	RequestType_name = map[int32]string{
		0: "REQUEST_TYPE_UNSPECIFIED",
		1: "REQUEST_TYPE_ORGANIC",
		2: "REQUEST_TYPE_SYNTHETIC",
	}
	RequestType_value = map[string]int32{
		"REQUEST_TYPE_UNSPECIFIED": 0,
		"REQUEST_TYPE_ORGANIC":     1,
		"REQUEST_TYPE_SYNTHETIC":   2,
	}
)

func (x RequestType) Enum() *RequestType {
	p := new(RequestType)
	*p = x
	return p
}

func (x RequestType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (RequestType) Descriptor() protoreflect.EnumDescriptor {
	return file_path_gateway_proto_enumTypes[0].Descriptor()
}

func (RequestType) Type() protoreflect.EnumType {
	return &file_path_gateway_proto_enumTypes[0]
}

func (x RequestType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use RequestType.Descriptor instead.
func (RequestType) EnumDescriptor() ([]byte, []int) {
	return file_path_gateway_proto_rawDescGZIP(), []int{0}
}

// GatewayObservations is the set of observations on a service request, made from the perspective of a gateway.
// Examples include the geographic region of the request, the request type, etc.
type GatewayObservations struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The geographic region where the gateway serving the request was located.
	// DEV_NOTE: It aligns with typical cloud regions (e.g. us-east) but kept as a string for simplicity.
	// TODO_FUTURE: this may need to move into a separate message if more details regarding the PATH instance are required.
	Region string `protobuf:"bytes,1,opt,name=region,proto3" json:"region,omitempty"`
	// Specifies the request type.
	// For example, wWhether the request was sent by a user or synthetically generated (e.g. by the endpoint hydrator).
	RequestType RequestType `protobuf:"varint,2,opt,name=request_type,json=requestType,proto3,enum=path.RequestType" json:"request_type,omitempty"`
	// service_id stores the service ID specified by the request.
	// As of PR #72, this can only be specified through a custom header on the HTTP request, extracted in `request/parser.go`.
	ServiceId string `protobuf:"bytes,3,opt,name=service_id,json=serviceId,proto3" json:"service_id,omitempty"`
	// received_time specifies the time the request was received.
	ReceivedTime *timestamppb.Timestamp `protobuf:"bytes,4,opt,name=received_time,json=receivedTime,proto3" json:"received_time,omitempty"`
	// completed_time specifies the time the request processing was completed, i.e. once a response was returned to the user.
	CompletedTime *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=completed_time,json=completedTime,proto3" json:"completed_time,omitempty"`
	// response_size captures the size of the response payload returned to the user.
	ResponseSize uint64 `protobuf:"varint,6,opt,name=response_size,json=responseSize,proto3" json:"response_size,omitempty"`
}

func (x *GatewayObservations) Reset() {
	*x = GatewayObservations{}
	if protoimpl.UnsafeEnabled {
		mi := &file_path_gateway_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GatewayObservations) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GatewayObservations) ProtoMessage() {}

func (x *GatewayObservations) ProtoReflect() protoreflect.Message {
	mi := &file_path_gateway_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GatewayObservations.ProtoReflect.Descriptor instead.
func (*GatewayObservations) Descriptor() ([]byte, []int) {
	return file_path_gateway_proto_rawDescGZIP(), []int{0}
}

func (x *GatewayObservations) GetRegion() string {
	if x != nil {
		return x.Region
	}
	return ""
}

func (x *GatewayObservations) GetRequestType() RequestType {
	if x != nil {
		return x.RequestType
	}
	return RequestType_REQUEST_TYPE_UNSPECIFIED
}

func (x *GatewayObservations) GetServiceId() string {
	if x != nil {
		return x.ServiceId
	}
	return ""
}

func (x *GatewayObservations) GetReceivedTime() *timestamppb.Timestamp {
	if x != nil {
		return x.ReceivedTime
	}
	return nil
}

func (x *GatewayObservations) GetCompletedTime() *timestamppb.Timestamp {
	if x != nil {
		return x.CompletedTime
	}
	return nil
}

func (x *GatewayObservations) GetResponseSize() uint64 {
	if x != nil {
		return x.ResponseSize
	}
	return 0
}

var File_path_gateway_proto protoreflect.FileDescriptor

var file_path_gateway_proto_rawDesc = []byte{
	0x0a, 0x12, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x70, 0x61, 0x74, 0x68, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xab, 0x02, 0x0a, 0x13,
	0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x12, 0x34, 0x0a, 0x0c, 0x72,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x11, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x54, 0x79, 0x70, 0x65, 0x52, 0x0b, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x54, 0x79, 0x70,
	0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x69, 0x64, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x49, 0x64,
	0x12, 0x3f, 0x0a, 0x0d, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x64, 0x5f, 0x74, 0x69, 0x6d,
	0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x52, 0x0c, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x64, 0x54, 0x69, 0x6d,
	0x65, 0x12, 0x41, 0x0a, 0x0e, 0x63, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x65, 0x64, 0x5f, 0x74,
	0x69, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0d, 0x63, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x65, 0x64,
	0x54, 0x69, 0x6d, 0x65, 0x12, 0x23, 0x0a, 0x0d, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0c, 0x72, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x2a, 0x61, 0x0a, 0x0b, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1c, 0x0a, 0x18, 0x52, 0x45, 0x51, 0x55,
	0x45, 0x53, 0x54, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49,
	0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x18, 0x0a, 0x14, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53,
	0x54, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x4f, 0x52, 0x47, 0x41, 0x4e, 0x49, 0x43, 0x10, 0x01,
	0x12, 0x1a, 0x0a, 0x16, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x5f, 0x54, 0x59, 0x50, 0x45,
	0x5f, 0x53, 0x59, 0x4e, 0x54, 0x48, 0x45, 0x54, 0x49, 0x43, 0x10, 0x02, 0x42, 0x2c, 0x5a, 0x2a,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x75, 0x69, 0x6c, 0x64,
	0x77, 0x69, 0x74, 0x68, 0x67, 0x72, 0x6f, 0x76, 0x65, 0x2f, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x6f,
	0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_path_gateway_proto_rawDescOnce sync.Once
	file_path_gateway_proto_rawDescData = file_path_gateway_proto_rawDesc
)

func file_path_gateway_proto_rawDescGZIP() []byte {
	file_path_gateway_proto_rawDescOnce.Do(func() {
		file_path_gateway_proto_rawDescData = protoimpl.X.CompressGZIP(file_path_gateway_proto_rawDescData)
	})
	return file_path_gateway_proto_rawDescData
}

var file_path_gateway_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_path_gateway_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_path_gateway_proto_goTypes = []any{
	(RequestType)(0),              // 0: path.RequestType
	(*GatewayObservations)(nil),   // 1: path.GatewayObservations
	(*timestamppb.Timestamp)(nil), // 2: google.protobuf.Timestamp
}
var file_path_gateway_proto_depIdxs = []int32{
	0, // 0: path.GatewayObservations.request_type:type_name -> path.RequestType
	2, // 1: path.GatewayObservations.received_time:type_name -> google.protobuf.Timestamp
	2, // 2: path.GatewayObservations.completed_time:type_name -> google.protobuf.Timestamp
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_path_gateway_proto_init() }
func file_path_gateway_proto_init() {
	if File_path_gateway_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_path_gateway_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*GatewayObservations); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_path_gateway_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_path_gateway_proto_goTypes,
		DependencyIndexes: file_path_gateway_proto_depIdxs,
		EnumInfos:         file_path_gateway_proto_enumTypes,
		MessageInfos:      file_path_gateway_proto_msgTypes,
	}.Build()
	File_path_gateway_proto = out.File
	file_path_gateway_proto_rawDesc = nil
	file_path_gateway_proto_goTypes = nil
	file_path_gateway_proto_depIdxs = nil
}
