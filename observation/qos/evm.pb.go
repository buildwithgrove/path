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

// EVMRequestObservations captures all observations made while serving a single EVM blockchain service request.
type EVMRequestObservations struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The EVM blockchain service's JSON-RPC request.
	// TODO_TECHDEBT: Assumes EVM chains only support JSON-RPC. May need refactoring to support other protocols.
	JsonrpcRequest *JsonRpcRequest `protobuf:"bytes,1,opt,name=jsonrpc_request,json=jsonrpcRequest,proto3" json:"jsonrpc_request,omitempty"`
	// EVM-specific observations from endpoint(s) that responded to the service request.
	// Multiple observations may occur when:
	// * Original endpoint fails
	// * Request is sent to additional endpoints for data collection
	EndpointObservations []*EVMEndpointObservation `protobuf:"bytes,2,rep,name=endpoint_observations,json=endpointObservations,proto3" json:"endpoint_observations,omitempty"`
}

func (x *EVMRequestObservations) Reset() {
	*x = EVMRequestObservations{}
	if protoimpl.UnsafeEnabled {
		mi := &file_path_qos_evm_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EVMRequestObservations) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EVMRequestObservations) ProtoMessage() {}

func (x *EVMRequestObservations) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use EVMRequestObservations.ProtoReflect.Descriptor instead.
func (*EVMRequestObservations) Descriptor() ([]byte, []int) {
	return file_path_qos_evm_proto_rawDescGZIP(), []int{0}
}

func (x *EVMRequestObservations) GetJsonrpcRequest() *JsonRpcRequest {
	if x != nil {
		return x.JsonrpcRequest
	}
	return nil
}

func (x *EVMRequestObservations) GetEndpointObservations() []*EVMEndpointObservation {
	if x != nil {
		return x.EndpointObservations
	}
	return nil
}

// EVMEndpointObservation stores a single observation from an endpoint servicing the protocol response.
// Example: A Pocket node on Shannon backed by an Ethereum data node servicing an `eth_getBlockNumber` request.
type EVMEndpointObservation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Address of the endpoint handling the request (e.g., onchain address of a Pocket Morse/Shannon node)
	EndpointAddr string `protobuf:"bytes,1,opt,name=endpoint_addr,json=endpointAddr,proto3" json:"endpoint_addr,omitempty"`
	// Details of the response received from the endpoint
	//
	// Types that are assignable to ResponseObservation:
	//
	//	*EVMEndpointObservation_ChainIdResponse
	//	*EVMEndpointObservation_BlockNumberResponse
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

func (x *EVMEndpointObservation) GetBlockNumberResponse() *EVMBlockNumberResponse {
	if x, ok := x.GetResponseObservation().(*EVMEndpointObservation_BlockNumberResponse); ok {
		return x.BlockNumberResponse
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
	// Response to `eth_chainId` request
	// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_chainid
	ChainIdResponse *EVMChainIDResponse `protobuf:"bytes,2,opt,name=chain_id_response,json=chainIdResponse,proto3,oneof"`
}

type EVMEndpointObservation_BlockNumberResponse struct {
	// Response to `eth_blockNumber` request
	// References:
	// * https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_blocknumber
	// * Chain IDs: https://chainlist.org
	BlockNumberResponse *EVMBlockNumberResponse `protobuf:"bytes,3,opt,name=block_number_response,json=blockNumberResponse,proto3,oneof"`
}

type EVMEndpointObservation_UnrecognizedResponse struct {
	// Responses not used in endpoint validation (e.g., JSONRPC ID field from `eth_call`)
	UnrecognizedResponse *EVMUnrecognizedResponse `protobuf:"bytes,4,opt,name=unrecognized_response,json=unrecognizedResponse,proto3,oneof"`
}

func (*EVMEndpointObservation_ChainIdResponse) isEVMEndpointObservation_ResponseObservation() {}

func (*EVMEndpointObservation_BlockNumberResponse) isEVMEndpointObservation_ResponseObservation() {}

func (*EVMEndpointObservation_UnrecognizedResponse) isEVMEndpointObservation_ResponseObservation() {}

// EVMChainIDResponse stores the response to an `eth_chainId` request
// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_chainid
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

// EVMBlockNumberResponse stores the response to an `eth_getBlockNumber` request
// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_blocknumber
type EVMBlockNumberResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BlockNumberResponse string `protobuf:"bytes,1,opt,name=block_number_response,json=blockNumberResponse,proto3" json:"block_number_response,omitempty"`
}

func (x *EVMBlockNumberResponse) Reset() {
	*x = EVMBlockNumberResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_path_qos_evm_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EVMBlockNumberResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EVMBlockNumberResponse) ProtoMessage() {}

func (x *EVMBlockNumberResponse) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use EVMBlockNumberResponse.ProtoReflect.Descriptor instead.
func (*EVMBlockNumberResponse) Descriptor() ([]byte, []int) {
	return file_path_qos_evm_proto_rawDescGZIP(), []int{3}
}

func (x *EVMBlockNumberResponse) GetBlockNumberResponse() string {
	if x != nil {
		return x.BlockNumberResponse
	}
	return ""
}

// EVMUnrecognizedResponse handles requests with methods ignored by state update and endpoint validation
// Example: As of PR #72, `eth_call` requests are not used for endpoint validation
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
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xb2, 0x01, 0x0a, 0x16, 0x45, 0x56, 0x4d, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x12, 0x41, 0x0a, 0x0f, 0x6a, 0x73, 0x6f, 0x6e, 0x72, 0x70, 0x63, 0x5f, 0x72, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x70, 0x61, 0x74,
	0x68, 0x2e, 0x71, 0x6f, 0x73, 0x2e, 0x4a, 0x73, 0x6f, 0x6e, 0x52, 0x70, 0x63, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x52, 0x0e, 0x6a, 0x73, 0x6f, 0x6e, 0x72, 0x70, 0x63, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x55, 0x0a, 0x15, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74,
	0x5f, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x71, 0x6f, 0x73, 0x2e, 0x45,
	0x56, 0x4d, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x14, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x4f,
	0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0xd3, 0x02, 0x0a, 0x16,
	0x45, 0x56, 0x4d, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x4f, 0x62, 0x73, 0x65, 0x72,
	0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x23, 0x0a, 0x0d, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69,
	0x6e, 0x74, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x65,
	0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x41, 0x64, 0x64, 0x72, 0x12, 0x4a, 0x0a, 0x11, 0x63,
	0x68, 0x61, 0x69, 0x6e, 0x5f, 0x69, 0x64, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x71, 0x6f,
	0x73, 0x2e, 0x45, 0x56, 0x4d, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x49, 0x44, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x48, 0x00, 0x52, 0x0f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x49, 0x64, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x56, 0x0a, 0x15, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x71, 0x6f,
	0x73, 0x2e, 0x45, 0x56, 0x4d, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x48, 0x00, 0x52, 0x13, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x58, 0x0a, 0x15, 0x75, 0x6e, 0x72, 0x65, 0x63, 0x6f, 0x67, 0x6e, 0x69, 0x7a, 0x65, 0x64, 0x5f,
	0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21,
	0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x71, 0x6f, 0x73, 0x2e, 0x45, 0x56, 0x4d, 0x55, 0x6e, 0x72,
	0x65, 0x63, 0x6f, 0x67, 0x6e, 0x69, 0x7a, 0x65, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x48, 0x00, 0x52, 0x14, 0x75, 0x6e, 0x72, 0x65, 0x63, 0x6f, 0x67, 0x6e, 0x69, 0x7a, 0x65,
	0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x16, 0x0a, 0x14, 0x72, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x5f, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x22, 0x40, 0x0a, 0x12, 0x45, 0x56, 0x4d, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x49, 0x44, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2a, 0x0a, 0x11, 0x63, 0x68, 0x61, 0x69, 0x6e,
	0x5f, 0x69, 0x64, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x49, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x4c, 0x0a, 0x16, 0x45, 0x56, 0x4d, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4e,
	0x75, 0x6d, 0x62, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x32, 0x0a,
	0x15, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x5f, 0x72, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x13, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x5f, 0x0a, 0x17, 0x45, 0x56, 0x4d, 0x55, 0x6e, 0x72, 0x65, 0x63, 0x6f, 0x67, 0x6e,
	0x69, 0x7a, 0x65, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x44, 0x0a, 0x10,
	0x6a, 0x73, 0x6f, 0x6e, 0x72, 0x70, 0x63, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x71, 0x6f,
	0x73, 0x2e, 0x4a, 0x73, 0x6f, 0x6e, 0x52, 0x70, 0x63, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x52, 0x0f, 0x6a, 0x73, 0x6f, 0x6e, 0x72, 0x70, 0x63, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x42, 0x30, 0x5a, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x77, 0x69, 0x74, 0x68, 0x67, 0x72, 0x6f, 0x76, 0x65, 0x2f,
	0x70, 0x61, 0x74, 0x68, 0x2f, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x2f, 0x71, 0x6f, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
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
	(*EVMRequestObservations)(nil),  // 0: path.qos.EVMRequestObservations
	(*EVMEndpointObservation)(nil),  // 1: path.qos.EVMEndpointObservation
	(*EVMChainIDResponse)(nil),      // 2: path.qos.EVMChainIDResponse
	(*EVMBlockNumberResponse)(nil),  // 3: path.qos.EVMBlockNumberResponse
	(*EVMUnrecognizedResponse)(nil), // 4: path.qos.EVMUnrecognizedResponse
	(*JsonRpcRequest)(nil),          // 5: path.qos.JsonRpcRequest
	(*JsonRpcResponse)(nil),         // 6: path.qos.JsonRpcResponse
}
var file_path_qos_evm_proto_depIdxs = []int32{
	5, // 0: path.qos.EVMRequestObservations.jsonrpc_request:type_name -> path.qos.JsonRpcRequest
	1, // 1: path.qos.EVMRequestObservations.endpoint_observations:type_name -> path.qos.EVMEndpointObservation
	2, // 2: path.qos.EVMEndpointObservation.chain_id_response:type_name -> path.qos.EVMChainIDResponse
	3, // 3: path.qos.EVMEndpointObservation.block_number_response:type_name -> path.qos.EVMBlockNumberResponse
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
			switch v := v.(*EVMRequestObservations); i {
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
			switch v := v.(*EVMBlockNumberResponse); i {
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
		(*EVMEndpointObservation_BlockNumberResponse)(nil),
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
