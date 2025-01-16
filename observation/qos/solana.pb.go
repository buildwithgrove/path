// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.2
// 	protoc        v5.29.3
// source: path/qos/solana.proto

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

// SolanaObservations stores all the observations made by the Solana QoS on a single request.
// A single request can result in multiple observations if:
//   - The originally selected endpoint returns an invalid response, AND
//   - A retry mechanism sends the request to additional endpoint(s).
type SolanaObservations struct {
	state                protoimpl.MessageState       `protogen:"open.v1"`
	JsonrpcRequest       *JsonRpcRequest              `protobuf:"bytes,1,opt,name=jsonrpc_request,json=jsonrpcRequest,proto3" json:"jsonrpc_request,omitempty"`
	EndpointObservations []*SolanaEndpointObservation `protobuf:"bytes,2,rep,name=endpoint_observations,json=endpointObservations,proto3" json:"endpoint_observations,omitempty"`
	unknownFields        protoimpl.UnknownFields
	sizeCache            protoimpl.SizeCache
}

func (x *SolanaObservations) Reset() {
	*x = SolanaObservations{}
	mi := &file_path_qos_solana_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SolanaObservations) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SolanaObservations) ProtoMessage() {}

func (x *SolanaObservations) ProtoReflect() protoreflect.Message {
	mi := &file_path_qos_solana_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SolanaObservations.ProtoReflect.Descriptor instead.
func (*SolanaObservations) Descriptor() ([]byte, []int) {
	return file_path_qos_solana_proto_rawDescGZIP(), []int{0}
}

func (x *SolanaObservations) GetJsonrpcRequest() *JsonRpcRequest {
	if x != nil {
		return x.JsonrpcRequest
	}
	return nil
}

func (x *SolanaObservations) GetEndpointObservations() []*SolanaEndpointObservation {
	if x != nil {
		return x.EndpointObservations
	}
	return nil
}

// SolanaEndpointObservation captures a single observation regarding a single endpoint.
// e.g.: an `ok` response to a `getHealth` request from an endpoint.
type SolanaEndpointObservation struct {
	state        protoimpl.MessageState `protogen:"open.v1"`
	EndpointAddr string                 `protobuf:"bytes,1,opt,name=endpoint_addr,json=endpointAddr,proto3" json:"endpoint_addr,omitempty"`
	// Types that are valid to be assigned to ResponseObservation:
	//
	//	*SolanaEndpointObservation_GetEpochInfoResponse
	//	*SolanaEndpointObservation_GetHealthResponse
	//	*SolanaEndpointObservation_UnrecognizedResponse
	ResponseObservation isSolanaEndpointObservation_ResponseObservation `protobuf_oneof:"response_observation"`
	unknownFields       protoimpl.UnknownFields
	sizeCache           protoimpl.SizeCache
}

func (x *SolanaEndpointObservation) Reset() {
	*x = SolanaEndpointObservation{}
	mi := &file_path_qos_solana_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SolanaEndpointObservation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SolanaEndpointObservation) ProtoMessage() {}

func (x *SolanaEndpointObservation) ProtoReflect() protoreflect.Message {
	mi := &file_path_qos_solana_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SolanaEndpointObservation.ProtoReflect.Descriptor instead.
func (*SolanaEndpointObservation) Descriptor() ([]byte, []int) {
	return file_path_qos_solana_proto_rawDescGZIP(), []int{1}
}

func (x *SolanaEndpointObservation) GetEndpointAddr() string {
	if x != nil {
		return x.EndpointAddr
	}
	return ""
}

func (x *SolanaEndpointObservation) GetResponseObservation() isSolanaEndpointObservation_ResponseObservation {
	if x != nil {
		return x.ResponseObservation
	}
	return nil
}

func (x *SolanaEndpointObservation) GetGetEpochInfoResponse() *SolanaGetEpochInfoResponse {
	if x != nil {
		if x, ok := x.ResponseObservation.(*SolanaEndpointObservation_GetEpochInfoResponse); ok {
			return x.GetEpochInfoResponse
		}
	}
	return nil
}

func (x *SolanaEndpointObservation) GetGetHealthResponse() *SolanaGetHealthResponse {
	if x != nil {
		if x, ok := x.ResponseObservation.(*SolanaEndpointObservation_GetHealthResponse); ok {
			return x.GetHealthResponse
		}
	}
	return nil
}

func (x *SolanaEndpointObservation) GetUnrecognizedResponse() *SolanaUnrecognizedResponse {
	if x != nil {
		if x, ok := x.ResponseObservation.(*SolanaEndpointObservation_UnrecognizedResponse); ok {
			return x.UnrecognizedResponse
		}
	}
	return nil
}

type isSolanaEndpointObservation_ResponseObservation interface {
	isSolanaEndpointObservation_ResponseObservation()
}

type SolanaEndpointObservation_GetEpochInfoResponse struct {
	GetEpochInfoResponse *SolanaGetEpochInfoResponse `protobuf:"bytes,2,opt,name=get_epoch_info_response,json=getEpochInfoResponse,proto3,oneof"`
}

type SolanaEndpointObservation_GetHealthResponse struct {
	GetHealthResponse *SolanaGetHealthResponse `protobuf:"bytes,3,opt,name=get_health_response,json=getHealthResponse,proto3,oneof"`
}

type SolanaEndpointObservation_UnrecognizedResponse struct {
	UnrecognizedResponse *SolanaUnrecognizedResponse `protobuf:"bytes,4,opt,name=unrecognized_response,json=unrecognizedResponse,proto3,oneof"`
}

func (*SolanaEndpointObservation_GetEpochInfoResponse) isSolanaEndpointObservation_ResponseObservation() {
}

func (*SolanaEndpointObservation_GetHealthResponse) isSolanaEndpointObservation_ResponseObservation() {
}

func (*SolanaEndpointObservation_UnrecognizedResponse) isSolanaEndpointObservation_ResponseObservation() {
}

// SolanaEpochInfoResponse stores an endpoint's response to a `getEpochInfo` request.
// See the following link for more details:
// https://solana.com/docs/rpc/http/getepochinfo
type SolanaGetEpochInfoResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// block_height is stored as a string to allow validation by any observing PATH instance.
	BlockHeight uint64 `protobuf:"varint,1,opt,name=block_height,json=blockHeight,proto3" json:"block_height,omitempty"`
	// epoch is stored as a string to allow validation by any observing PATH instance.
	Epoch         uint64 `protobuf:"varint,2,opt,name=epoch,proto3" json:"epoch,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SolanaGetEpochInfoResponse) Reset() {
	*x = SolanaGetEpochInfoResponse{}
	mi := &file_path_qos_solana_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SolanaGetEpochInfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SolanaGetEpochInfoResponse) ProtoMessage() {}

func (x *SolanaGetEpochInfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_path_qos_solana_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SolanaGetEpochInfoResponse.ProtoReflect.Descriptor instead.
func (*SolanaGetEpochInfoResponse) Descriptor() ([]byte, []int) {
	return file_path_qos_solana_proto_rawDescGZIP(), []int{2}
}

func (x *SolanaGetEpochInfoResponse) GetBlockHeight() uint64 {
	if x != nil {
		return x.BlockHeight
	}
	return 0
}

func (x *SolanaGetEpochInfoResponse) GetEpoch() uint64 {
	if x != nil {
		return x.Epoch
	}
	return 0
}

// SolanaGetHealthResponse stores an endpoint's response to a `getHealth` request.
// See the following link for more details:
// https://solana.com/docs/rpc/http/gethealth
type SolanaGetHealthResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Result        string                 `protobuf:"bytes,1,opt,name=result,proto3" json:"result,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SolanaGetHealthResponse) Reset() {
	*x = SolanaGetHealthResponse{}
	mi := &file_path_qos_solana_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SolanaGetHealthResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SolanaGetHealthResponse) ProtoMessage() {}

func (x *SolanaGetHealthResponse) ProtoReflect() protoreflect.Message {
	mi := &file_path_qos_solana_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SolanaGetHealthResponse.ProtoReflect.Descriptor instead.
func (*SolanaGetHealthResponse) Descriptor() ([]byte, []int) {
	return file_path_qos_solana_proto_rawDescGZIP(), []int{3}
}

func (x *SolanaGetHealthResponse) GetResult() string {
	if x != nil {
		return x.Result
	}
	return ""
}

// SolanaUnrecognizedResponse is utilized if the request's method is ignored by state update and endpoint validation methods.
// For example, as of PR #72, neither of `getTokenSupply` or `getTransaction` requests are used for endpoint validation.
// Therefore only generic fields of the JSONRPC response (like `id`) are stored.
type SolanaUnrecognizedResponse struct {
	state           protoimpl.MessageState `protogen:"open.v1"`
	JsonrpcResponse *JsonRpcResponse       `protobuf:"bytes,1,opt,name=jsonrpc_response,json=jsonrpcResponse,proto3" json:"jsonrpc_response,omitempty"`
	unknownFields   protoimpl.UnknownFields
	sizeCache       protoimpl.SizeCache
}

func (x *SolanaUnrecognizedResponse) Reset() {
	*x = SolanaUnrecognizedResponse{}
	mi := &file_path_qos_solana_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SolanaUnrecognizedResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SolanaUnrecognizedResponse) ProtoMessage() {}

func (x *SolanaUnrecognizedResponse) ProtoReflect() protoreflect.Message {
	mi := &file_path_qos_solana_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SolanaUnrecognizedResponse.ProtoReflect.Descriptor instead.
func (*SolanaUnrecognizedResponse) Descriptor() ([]byte, []int) {
	return file_path_qos_solana_proto_rawDescGZIP(), []int{4}
}

func (x *SolanaUnrecognizedResponse) GetJsonrpcResponse() *JsonRpcResponse {
	if x != nil {
		return x.JsonrpcResponse
	}
	return nil
}

var File_path_qos_solana_proto protoreflect.FileDescriptor

var file_path_qos_solana_proto_rawDesc = []byte{
	0x0a, 0x15, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x71, 0x6f, 0x73, 0x2f, 0x73, 0x6f, 0x6c, 0x61, 0x6e,
	0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x71, 0x6f,
	0x73, 0x1a, 0x16, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x71, 0x6f, 0x73, 0x2f, 0x6a, 0x73, 0x6f, 0x6e,
	0x72, 0x70, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xb1, 0x01, 0x0a, 0x12, 0x53, 0x6f,
	0x6c, 0x61, 0x6e, 0x61, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x12, 0x41, 0x0a, 0x0f, 0x6a, 0x73, 0x6f, 0x6e, 0x72, 0x70, 0x63, 0x5f, 0x72, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x70, 0x61, 0x74, 0x68,
	0x2e, 0x71, 0x6f, 0x73, 0x2e, 0x4a, 0x73, 0x6f, 0x6e, 0x52, 0x70, 0x63, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x52, 0x0e, 0x6a, 0x73, 0x6f, 0x6e, 0x72, 0x70, 0x63, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x58, 0x0a, 0x15, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x5f,
	0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x23, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x71, 0x6f, 0x73, 0x2e, 0x53, 0x6f,
	0x6c, 0x61, 0x6e, 0x61, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x4f, 0x62, 0x73, 0x65,
	0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x14, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e,
	0x74, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0xe9, 0x02,
	0x0a, 0x19, 0x53, 0x6f, 0x6c, 0x61, 0x6e, 0x61, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74,
	0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x23, 0x0a, 0x0d, 0x65,
	0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0c, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x41, 0x64, 0x64, 0x72,
	0x12, 0x5d, 0x0a, 0x17, 0x67, 0x65, 0x74, 0x5f, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x5f, 0x69, 0x6e,
	0x66, 0x6f, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x24, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x71, 0x6f, 0x73, 0x2e, 0x53, 0x6f, 0x6c,
	0x61, 0x6e, 0x61, 0x47, 0x65, 0x74, 0x45, 0x70, 0x6f, 0x63, 0x68, 0x49, 0x6e, 0x66, 0x6f, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x48, 0x00, 0x52, 0x14, 0x67, 0x65, 0x74, 0x45, 0x70,
	0x6f, 0x63, 0x68, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x53, 0x0a, 0x13, 0x67, 0x65, 0x74, 0x5f, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x5f, 0x72, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x70,
	0x61, 0x74, 0x68, 0x2e, 0x71, 0x6f, 0x73, 0x2e, 0x53, 0x6f, 0x6c, 0x61, 0x6e, 0x61, 0x47, 0x65,
	0x74, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x48,
	0x00, 0x52, 0x11, 0x67, 0x65, 0x74, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x5b, 0x0a, 0x15, 0x75, 0x6e, 0x72, 0x65, 0x63, 0x6f, 0x67, 0x6e,
	0x69, 0x7a, 0x65, 0x64, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x71, 0x6f, 0x73, 0x2e, 0x53,
	0x6f, 0x6c, 0x61, 0x6e, 0x61, 0x55, 0x6e, 0x72, 0x65, 0x63, 0x6f, 0x67, 0x6e, 0x69, 0x7a, 0x65,
	0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x48, 0x00, 0x52, 0x14, 0x75, 0x6e, 0x72,
	0x65, 0x63, 0x6f, 0x67, 0x6e, 0x69, 0x7a, 0x65, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x42, 0x16, 0x0a, 0x14, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x5f, 0x6f, 0x62,
	0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x55, 0x0a, 0x1a, 0x53, 0x6f, 0x6c,
	0x61, 0x6e, 0x61, 0x47, 0x65, 0x74, 0x45, 0x70, 0x6f, 0x63, 0x68, 0x49, 0x6e, 0x66, 0x6f, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x5f, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0b, 0x62,
	0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x70,
	0x6f, 0x63, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x65, 0x70, 0x6f, 0x63, 0x68,
	0x22, 0x31, 0x0a, 0x17, 0x53, 0x6f, 0x6c, 0x61, 0x6e, 0x61, 0x47, 0x65, 0x74, 0x48, 0x65, 0x61,
	0x6c, 0x74, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x72,
	0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x22, 0x62, 0x0a, 0x1a, 0x53, 0x6f, 0x6c, 0x61, 0x6e, 0x61, 0x55, 0x6e, 0x72,
	0x65, 0x63, 0x6f, 0x67, 0x6e, 0x69, 0x7a, 0x65, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x44, 0x0a, 0x10, 0x6a, 0x73, 0x6f, 0x6e, 0x72, 0x70, 0x63, 0x5f, 0x72, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x70, 0x61,
	0x74, 0x68, 0x2e, 0x71, 0x6f, 0x73, 0x2e, 0x4a, 0x73, 0x6f, 0x6e, 0x52, 0x70, 0x63, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x52, 0x0f, 0x6a, 0x73, 0x6f, 0x6e, 0x72, 0x70, 0x63, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x30, 0x5a, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x77, 0x69, 0x74, 0x68, 0x67,
	0x72, 0x6f, 0x76, 0x65, 0x2f, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x71, 0x6f, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_path_qos_solana_proto_rawDescOnce sync.Once
	file_path_qos_solana_proto_rawDescData = file_path_qos_solana_proto_rawDesc
)

func file_path_qos_solana_proto_rawDescGZIP() []byte {
	file_path_qos_solana_proto_rawDescOnce.Do(func() {
		file_path_qos_solana_proto_rawDescData = protoimpl.X.CompressGZIP(file_path_qos_solana_proto_rawDescData)
	})
	return file_path_qos_solana_proto_rawDescData
}

var file_path_qos_solana_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_path_qos_solana_proto_goTypes = []any{
	(*SolanaObservations)(nil),         // 0: path.qos.SolanaObservations
	(*SolanaEndpointObservation)(nil),  // 1: path.qos.SolanaEndpointObservation
	(*SolanaGetEpochInfoResponse)(nil), // 2: path.qos.SolanaGetEpochInfoResponse
	(*SolanaGetHealthResponse)(nil),    // 3: path.qos.SolanaGetHealthResponse
	(*SolanaUnrecognizedResponse)(nil), // 4: path.qos.SolanaUnrecognizedResponse
	(*JsonRpcRequest)(nil),             // 5: path.qos.JsonRpcRequest
	(*JsonRpcResponse)(nil),            // 6: path.qos.JsonRpcResponse
}
var file_path_qos_solana_proto_depIdxs = []int32{
	5, // 0: path.qos.SolanaObservations.jsonrpc_request:type_name -> path.qos.JsonRpcRequest
	1, // 1: path.qos.SolanaObservations.endpoint_observations:type_name -> path.qos.SolanaEndpointObservation
	2, // 2: path.qos.SolanaEndpointObservation.get_epoch_info_response:type_name -> path.qos.SolanaGetEpochInfoResponse
	3, // 3: path.qos.SolanaEndpointObservation.get_health_response:type_name -> path.qos.SolanaGetHealthResponse
	4, // 4: path.qos.SolanaEndpointObservation.unrecognized_response:type_name -> path.qos.SolanaUnrecognizedResponse
	6, // 5: path.qos.SolanaUnrecognizedResponse.jsonrpc_response:type_name -> path.qos.JsonRpcResponse
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_path_qos_solana_proto_init() }
func file_path_qos_solana_proto_init() {
	if File_path_qos_solana_proto != nil {
		return
	}
	file_path_qos_jsonrpc_proto_init()
	file_path_qos_solana_proto_msgTypes[1].OneofWrappers = []any{
		(*SolanaEndpointObservation_GetEpochInfoResponse)(nil),
		(*SolanaEndpointObservation_GetHealthResponse)(nil),
		(*SolanaEndpointObservation_UnrecognizedResponse)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_path_qos_solana_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_path_qos_solana_proto_goTypes,
		DependencyIndexes: file_path_qos_solana_proto_depIdxs,
		MessageInfos:      file_path_qos_solana_proto_msgTypes,
	}.Build()
	File_path_qos_solana_proto = out.File
	file_path_qos_solana_proto_rawDesc = nil
	file_path_qos_solana_proto_goTypes = nil
	file_path_qos_solana_proto_depIdxs = nil
}
