// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v5.28.3
// source: path/observations.proto

package observation

import (
	protocol "github.com/buildwithgrove/path/observation/protocol"
	qos "github.com/buildwithgrove/path/observation/qos"
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

// RequestResponseObservations captures the observations made on every aspect of a service
// request and its response.
// e.g. service's QoS observations, protocol instance's observations, etc.
type RequestResponseObservations struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// service_id is the target service ID of the request.
	ServiceId string `protobuf:"bytes,1,opt,name=service_id,json=serviceId,proto3" json:"service_id,omitempty"`
	// HTTPRequestObservations captures all the observations made on the incoming HTTP request.
	// e.g. the request's payload size.
	HttpRequest *HTTPRequestObservations `protobuf:"bytes,2,opt,name=http_request,json=httpRequest,proto3" json:"http_request,omitempty"`
	// GatewayObservations is the set of all gateway-level observations related to the request.
	// e.g. whether the request was from a user or generated by the endpoint hydrator.
	Gateway *GatewayObservations `protobuf:"bytes,3,opt,name=gateway,proto3" json:"gateway,omitempty"`
	// ProtocolObservations is the set of protocol-level observations made on the request.
	// e.g. the block_height at which the request was served.
	Protocol *protocol.Observations `protobuf:"bytes,4,opt,name=protocol,proto3" json:"protocol,omitempty"`
	// QoSObservations is the set of QoS-level observations made on the request.
	// e.g. the serving endpoint's response to a `eth_chainId` request.
	Qos *qos.Observations `protobuf:"bytes,5,opt,name=qos,proto3" json:"qos,omitempty"`
}

func (x *RequestResponseObservations) Reset() {
	*x = RequestResponseObservations{}
	mi := &file_path_observations_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RequestResponseObservations) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequestResponseObservations) ProtoMessage() {}

func (x *RequestResponseObservations) ProtoReflect() protoreflect.Message {
	mi := &file_path_observations_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequestResponseObservations.ProtoReflect.Descriptor instead.
func (*RequestResponseObservations) Descriptor() ([]byte, []int) {
	return file_path_observations_proto_rawDescGZIP(), []int{0}
}

func (x *RequestResponseObservations) GetServiceId() string {
	if x != nil {
		return x.ServiceId
	}
	return ""
}

func (x *RequestResponseObservations) GetHttpRequest() *HTTPRequestObservations {
	if x != nil {
		return x.HttpRequest
	}
	return nil
}

func (x *RequestResponseObservations) GetGateway() *GatewayObservations {
	if x != nil {
		return x.Gateway
	}
	return nil
}

func (x *RequestResponseObservations) GetProtocol() *protocol.Observations {
	if x != nil {
		return x.Protocol
	}
	return nil
}

func (x *RequestResponseObservations) GetQos() *qos.Observations {
	if x != nil {
		return x.Qos
	}
	return nil
}

var File_path_observations_proto protoreflect.FileDescriptor

var file_path_observations_proto_rawDesc = []byte{
	0x0a, 0x17, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x70, 0x61, 0x74, 0x68, 0x1a,
	0x0f, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x68, 0x74, 0x74, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x12, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x6f, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x12, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x71, 0x6f, 0x73, 0x2f, 0x71, 0x6f, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x96, 0x02, 0x0a, 0x1b, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x49, 0x64, 0x12, 0x40, 0x0a, 0x0c, 0x68, 0x74, 0x74, 0x70, 0x5f, 0x72, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x70, 0x61,
	0x74, 0x68, 0x2e, 0x48, 0x54, 0x54, 0x50, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x4f, 0x62,
	0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x0b, 0x68, 0x74, 0x74, 0x70,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x33, 0x0a, 0x07, 0x67, 0x61, 0x74, 0x65, 0x77,
	0x61, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e,
	0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x52, 0x07, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x12, 0x37, 0x0a, 0x08,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b,
	0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x4f,
	0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x08, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12, 0x28, 0x0a, 0x03, 0x71, 0x6f, 0x73, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x16, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x71, 0x6f, 0x73, 0x2e, 0x4f, 0x62,
	0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x03, 0x71, 0x6f, 0x73, 0x42,
	0x2c, 0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x75,
	0x69, 0x6c, 0x64, 0x77, 0x69, 0x74, 0x68, 0x67, 0x72, 0x6f, 0x76, 0x65, 0x2f, 0x70, 0x61, 0x74,
	0x68, 0x2f, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_path_observations_proto_rawDescOnce sync.Once
	file_path_observations_proto_rawDescData = file_path_observations_proto_rawDesc
)

func file_path_observations_proto_rawDescGZIP() []byte {
	file_path_observations_proto_rawDescOnce.Do(func() {
		file_path_observations_proto_rawDescData = protoimpl.X.CompressGZIP(file_path_observations_proto_rawDescData)
	})
	return file_path_observations_proto_rawDescData
}

var file_path_observations_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_path_observations_proto_goTypes = []any{
	(*RequestResponseObservations)(nil), // 0: path.RequestResponseObservations
	(*HTTPRequestObservations)(nil),     // 1: path.HTTPRequestObservations
	(*GatewayObservations)(nil),         // 2: path.GatewayObservations
	(*protocol.Observations)(nil),       // 3: path.protocol.Observations
	(*qos.Observations)(nil),            // 4: path.qos.Observations
}
var file_path_observations_proto_depIdxs = []int32{
	1, // 0: path.RequestResponseObservations.http_request:type_name -> path.HTTPRequestObservations
	2, // 1: path.RequestResponseObservations.gateway:type_name -> path.GatewayObservations
	3, // 2: path.RequestResponseObservations.protocol:type_name -> path.protocol.Observations
	4, // 3: path.RequestResponseObservations.qos:type_name -> path.qos.Observations
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_path_observations_proto_init() }
func file_path_observations_proto_init() {
	if File_path_observations_proto != nil {
		return
	}
	file_path_http_proto_init()
	file_path_gateway_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_path_observations_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_path_observations_proto_goTypes,
		DependencyIndexes: file_path_observations_proto_depIdxs,
		MessageInfos:      file_path_observations_proto_msgTypes,
	}.Build()
	File_path_observations_proto = out.File
	file_path_observations_proto_rawDesc = nil
	file_path_observations_proto_goTypes = nil
	file_path_observations_proto_depIdxs = nil
}