// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.3
// source: path/qos/evm.proto

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

type EVMObservations struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	JsonrpcRequest *JsonRpcRequest `protobuf:"bytes,1,opt,name=jsonrpc_request,json=jsonrpcRequest,proto3" json:"jsonrpc_request,omitempty"`
	// A single request may create multiple observations.
	// This can happen if:
	//  1. The originally selected endpoint fails, AND
	//  2. The request is sent to additional endpoints.
	EndpointObservations []*EVMEndpointObservation `protobuf:"bytes,2,rep,name=endpoint_observations,json=endpointObservations,proto3" json:"endpoint_observations,omitempty"`
}

func (x *EVMObservations) Reset() {
	*x = EVMObservations{}
	if protoimpl.UnsafeEnabled {
		mi := &file_path_qos_evm_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EVMObservations) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EVMObservations) ProtoMessage() {}

func (x *EVMObservations) ProtoReflect() protoreflect.Message {
	mi := &file_path_qos_evm_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EVMObservations.ProtoReflect.Descriptor instead.
func (*EVMObservations) Descriptor() ([]byte, []int) {
	return file_path_qos_evm_proto_rawDescGZIP(), []int{0}
}

func (x *EVMObservations) GetJsonrpcRequest() *JsonRpcRequest {
	if x != nil {
		return x.JsonrpcRequest
	}
	return nil
}

func (x *EVMObservations) GetEndpointObservations() []*EVMEndpointObservation {
	if x != nil {
		return x.EndpointObservations
	}
	return nil
}

// EVMEndpointObservation stores a single observation regarding an endpoint.
// e.g. a specific endpoint's response to an `eth_getBlockNumber` request.
type EVMEndpointObservation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EndpointAddr string `protobuf:"bytes,1,opt,name=endpoint_addr,json=endpointAddr,proto3" json:"endpoint_addr,omitempty"`
	// Types that are assignable to ResponseObservation:
	//
	//	*EVMEndpointObservation_ChainIdResponse
	//	*EVMEndpointObservation_BlockHeightResponse
	//	*EVMEndpointObservation_UnrecognizedResponse
	ResponseObservation isEVMEndpointObservation_ResponseObservation `protobuf_oneof:"response_observation"`
}

func (x *EVMEndpointObservation) Reset() {
	*x = EVMEndpointObservation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_path_qos_evm_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EVMEndpointObservation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EVMEndpointObservation) ProtoMessage() {}

func (x *EVMEndpointObservation) ProtoReflect() protoreflect.Message {
	mi := &file_path_qos_evm_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EVMEndpointObservation.ProtoReflect.Descriptor instead.
func (*EVMEndpointObservation) Descriptor() ([]byte, []int) {
	return file_path_qos_evm_proto_rawDescGZIP(), []int{1}
}

func (x *EVMEndpointObservation) GetEndpointAddr() string {
	if x != nil {
		return x.EndpointAddr
	}
	return ""
}

func (m *EVMEndpointObservation) GetResponseObservation() isEVMEndpointObservation_ResponseObservation {
	if m != nil {
		return m.ResponseObservation
	}
	return nil
}

func (x *EVMEndpointObservation) GetChainIdResponse() *EVMChainIDResponse {
	if x, ok := x.GetResponseObservation().(*EVMEndpointObservation_ChainIdResponse); ok {
		return x.ChainIdResponse
	}
	return nil
}

func (x *EVMEndpointObservation) GetBlockHeightResponse() *EVMBlockHeightResponse {
	if x, ok := x.GetResponseObservation().(*EVMEndpointObservation_BlockHeightResponse); ok {
		return x.BlockHeightResponse
	}
	return nil
}

func (x *EVMEndpointObservation) GetUnrecognizedResponse() *EVMUnrecognizedResponse {
	if x, ok := x.GetResponseObservation().(*EVMEndpointObservation_UnrecognizedResponse); ok {
		return x.UnrecognizedResponse
	}
	return nil
}

type isEVMEndpointObservation_ResponseObservation interface {
	isEVMEndpointObservation_ResponseObservation()
}

type EVMEndpointObservation_ChainIdResponse struct {
	ChainIdResponse *EVMChainIDResponse `protobuf:"bytes,2,opt,name=chain_id_response,json=chainIdResponse,proto3,oneof"`
}

type EVMEndpointObservation_BlockHeightResponse struct {
	BlockHeightResponse *EVMBlockHeightResponse `protobuf:"bytes,3,opt,name=block_height_response,json=blockHeightResponse,proto3,oneof"`
}

type EVMEndpointObservation_UnrecognizedResponse struct {
	UnrecognizedResponse *EVMUnrecognizedResponse `protobuf:"bytes,4,opt,name=unrecognized_response,json=unrecognizedResponse,proto3,oneof"`
}

func (*EVMEndpointObservation_ChainIdResponse) isEVMEndpointObservation_ResponseObservation() {}

func (*EVMEndpointObservation_BlockHeightResponse) isEVMEndpointObservation_ResponseObservation() {}

func (*EVMEndpointObservation_UnrecognizedResponse) isEVMEndpointObservation_ResponseObservation() {}

// EVMChainIDResponse stores the response to an `eth_chainId` request.
type EVMChainIDResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ChainIdResponse string `protobuf:"bytes,1,opt,name=chain_id_response,json=chainIdResponse,proto3" json:"chain_id_response,omitempty"`
}

func (x *EVMChainIDResponse) Reset() {
	*x = EVMChainIDResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_path_qos_evm_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EVMChainIDResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EVMChainIDResponse) ProtoMessage() {}

func (x *EVMChainIDResponse) ProtoReflect() protoreflect.Message {
	mi := &file_path_qos_evm_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EVMChainIDResponse.ProtoReflect.Descriptor instead.
func (*EVMChainIDResponse) Descriptor() ([]byte, []int) {
	return file_path_qos_evm_proto_rawDescGZIP(), []int{2}
}

func (x *EVMChainIDResponse) GetChainIdResponse() string {
	if x != nil {
		return x.ChainIdResponse
	}
	return ""
}

// EVMBlockHeightResponse stores the response to an `eth_getBlockNumber` request.
type EVMBlockHeightResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BlockHeightResponse string `protobuf:"bytes,1,opt,name=block_height_response,json=blockHeightResponse,proto3" json:"block_height_response,omitempty"`
}

func (x *EVMBlockHeightResponse) Reset() {
	*x = EVMBlockHeightResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_path_qos_evm_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EVMBlockHeightResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EVMBlockHeightResponse) ProtoMessage() {}

func (x *EVMBlockHeightResponse) ProtoReflect() protoreflect.Message {
	mi := &file_path_qos_evm_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EVMBlockHeightResponse.ProtoReflect.Descriptor instead.
func (*EVMBlockHeightResponse) Descriptor() ([]byte, []int) {
	return file_path_qos_evm_proto_rawDescGZIP(), []int{3}
}

func (x *EVMBlockHeightResponse) GetBlockHeightResponse() string {
	if x != nil {
		return x.BlockHeightResponse
	}
	return ""
}

// EVMUnrecognizedResponse is utilized if the request's method is ignored by state update and endpoint validation methods.
// For example, as of PR #72, an `eth_call` request is not used for endpoint validation.
// Therefore only generic fields of the JSONRPC response (like `id`) are stored.
type EVMUnrecognizedResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	JsonrpcResponse *JsonRpcResponse `protobuf:"bytes,1,opt,name=jsonrpc_response,json=jsonrpcResponse,proto3" json:"jsonrpc_response,omitempty"`
}

func (x *EVMUnrecognizedResponse) Reset() {
	*x = EVMUnrecognizedResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_path_qos_evm_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EVMUnrecognizedResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EVMUnrecognizedResponse) ProtoMessage() {}

func (x *EVMUnrecognizedResponse) ProtoReflect() protoreflect.Message {
	mi := &file_path_qos_evm_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EVMUnrecognizedResponse.ProtoReflect.Descriptor instead.
func (*EVMUnrecognizedResponse) Descriptor() ([]byte, []int) {
	return file_path_qos_evm_proto_rawDescGZIP(), []int{4}
}

func (x *EVMUnrecognizedResponse) GetJsonrpcResponse() *JsonRpcResponse {
	if x != nil {
		return x.JsonrpcResponse
	}
	return nil
}

var File_path_qos_evm_proto protoreflect.FileDescriptor

var file_path_qos_evm_proto_rawDesc = []byte{
	0x0a, 0x12, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x71, 0x6f, 0x73, 0x2f, 0x65, 0x76, 0x6d, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x71, 0x6f, 0x73, 0x1a, 0x16,
	0x70, 0x61, 0x74, 0x68, 0x2f, 0x71, 0x6f, 0x73, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x72, 0x70, 0x63,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xab, 0x01, 0x0a, 0x0f, 0x45, 0x56, 0x4d, 0x4f, 0x62,
	0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x41, 0x0a, 0x0f, 0x6a, 0x73,
	0x6f, 0x6e, 0x72, 0x70, 0x63, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x71, 0x6f, 0x73, 0x2e, 0x4a,
	0x73, 0x6f, 0x6e, 0x52, 0x70, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x52, 0x0e, 0x6a,
	0x73, 0x6f, 0x6e, 0x72, 0x70, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x55, 0x0a,
	0x15, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x5f, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x70,
	0x61, 0x74, 0x68, 0x2e, 0x71, 0x6f, 0x73, 0x2e, 0x45, 0x56, 0x4d, 0x45, 0x6e, 0x64, 0x70, 0x6f,
	0x69, 0x6e, 0x74, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x14,
	0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x22, 0xd3, 0x02, 0x0a, 0x16, 0x45, 0x56, 0x4d, 0x45, 0x6e, 0x64, 0x70,
	0x6f, 0x69, 0x6e, 0x74, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x23, 0x0a, 0x0d, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x5f, 0x61, 0x64, 0x64, 0x72,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74,
	0x41, 0x64, 0x64, 0x72, 0x12, 0x4a, 0x0a, 0x11, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x69, 0x64,
	0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1c, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x71, 0x6f, 0x73, 0x2e, 0x45, 0x56, 0x4d, 0x43, 0x68,
	0x61, 0x69, 0x6e, 0x49, 0x44, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x48, 0x00, 0x52,
	0x0f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x49, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x56, 0x0a, 0x15, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74,
	0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x20, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x71, 0x6f, 0x73, 0x2e, 0x45, 0x56, 0x4d, 0x42, 0x6c,
	0x6f, 0x63, 0x6b, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x48, 0x00, 0x52, 0x13, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x58, 0x0a, 0x15, 0x75, 0x6e, 0x72, 0x65,
	0x63, 0x6f, 0x67, 0x6e, 0x69, 0x7a, 0x65, 0x64, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x71,
	0x6f, 0x73, 0x2e, 0x45, 0x56, 0x4d, 0x55, 0x6e, 0x72, 0x65, 0x63, 0x6f, 0x67, 0x6e, 0x69, 0x7a,
	0x65, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x48, 0x00, 0x52, 0x14, 0x75, 0x6e,
	0x72, 0x65, 0x63, 0x6f, 0x67, 0x6e, 0x69, 0x7a, 0x65, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x42, 0x16, 0x0a, 0x14, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x5f, 0x6f,
	0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x40, 0x0a, 0x12, 0x45, 0x56,
	0x4d, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x49, 0x44, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x2a, 0x0a, 0x11, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x69, 0x64, 0x5f, 0x72, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x63, 0x68, 0x61,
	0x69, 0x6e, 0x49, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x4c, 0x0a, 0x16,
	0x45, 0x56, 0x4d, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x32, 0x0a, 0x15, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f,
	0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x13, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x65, 0x69, 0x67,
	0x68, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x5f, 0x0a, 0x17, 0x45, 0x56,
	0x4d, 0x55, 0x6e, 0x72, 0x65, 0x63, 0x6f, 0x67, 0x6e, 0x69, 0x7a, 0x65, 0x64, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x44, 0x0a, 0x10, 0x6a, 0x73, 0x6f, 0x6e, 0x72, 0x70, 0x63,
	0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x19, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x71, 0x6f, 0x73, 0x2e, 0x4a, 0x73, 0x6f, 0x6e, 0x52,
	0x70, 0x63, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x52, 0x0f, 0x6a, 0x73, 0x6f, 0x6e,
	0x72, 0x70, 0x63, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x30, 0x5a, 0x2e, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x77,
	0x69, 0x74, 0x68, 0x67, 0x72, 0x6f, 0x76, 0x65, 0x2f, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x6f, 0x62,
	0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x71, 0x6f, 0x73, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_path_qos_evm_proto_rawDescOnce sync.Once
	file_path_qos_evm_proto_rawDescData = file_path_qos_evm_proto_rawDesc
)

func file_path_qos_evm_proto_rawDescGZIP() []byte {
	file_path_qos_evm_proto_rawDescOnce.Do(func() {
		file_path_qos_evm_proto_rawDescData = protoimpl.X.CompressGZIP(file_path_qos_evm_proto_rawDescData)
	})
	return file_path_qos_evm_proto_rawDescData
}

var file_path_qos_evm_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_path_qos_evm_proto_goTypes = []any{
	(*EVMObservations)(nil),         // 0: path.qos.EVMObservations
	(*EVMEndpointObservation)(nil),  // 1: path.qos.EVMEndpointObservation
	(*EVMChainIDResponse)(nil),      // 2: path.qos.EVMChainIDResponse
	(*EVMBlockHeightResponse)(nil),  // 3: path.qos.EVMBlockHeightResponse
	(*EVMUnrecognizedResponse)(nil), // 4: path.qos.EVMUnrecognizedResponse
	(*JsonRpcRequest)(nil),          // 5: path.qos.JsonRpcRequest
	(*JsonRpcResponse)(nil),         // 6: path.qos.JsonRpcResponse
}
var file_path_qos_evm_proto_depIdxs = []int32{
	5, // 0: path.qos.EVMObservations.jsonrpc_request:type_name -> path.qos.JsonRpcRequest
	1, // 1: path.qos.EVMObservations.endpoint_observations:type_name -> path.qos.EVMEndpointObservation
	2, // 2: path.qos.EVMEndpointObservation.chain_id_response:type_name -> path.qos.EVMChainIDResponse
	3, // 3: path.qos.EVMEndpointObservation.block_height_response:type_name -> path.qos.EVMBlockHeightResponse
	4, // 4: path.qos.EVMEndpointObservation.unrecognized_response:type_name -> path.qos.EVMUnrecognizedResponse
	6, // 5: path.qos.EVMUnrecognizedResponse.jsonrpc_response:type_name -> path.qos.JsonRpcResponse
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_path_qos_evm_proto_init() }
func file_path_qos_evm_proto_init() {
	if File_path_qos_evm_proto != nil {
		return
	}
	file_path_qos_jsonrpc_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_path_qos_evm_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*EVMObservations); i {
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
		file_path_qos_evm_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*EVMEndpointObservation); i {
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
		file_path_qos_evm_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*EVMChainIDResponse); i {
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
		file_path_qos_evm_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*EVMBlockHeightResponse); i {
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
		file_path_qos_evm_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*EVMUnrecognizedResponse); i {
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
	file_path_qos_evm_proto_msgTypes[1].OneofWrappers = []any{
		(*EVMEndpointObservation_ChainIdResponse)(nil),
		(*EVMEndpointObservation_BlockHeightResponse)(nil),
		(*EVMEndpointObservation_UnrecognizedResponse)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_path_qos_evm_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_path_qos_evm_proto_goTypes,
		DependencyIndexes: file_path_qos_evm_proto_depIdxs,
		MessageInfos:      file_path_qos_evm_proto_msgTypes,
	}.Build()
	File_path_qos_evm_proto = out.File
	file_path_qos_evm_proto_rawDesc = nil
	file_path_qos_evm_proto_goTypes = nil
	file_path_qos_evm_proto_depIdxs = nil
}
