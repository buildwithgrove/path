// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.3
// source: path/qos/cometbft.proto

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

// CometBFTRequestObservations captures all observations made while serving a single CometBFT blockchain service request.
type CometBFTRequestObservations struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The CometBFT blockchain service's route request, including all params
	RouteRequest string `protobuf:"bytes,1,opt,name=route_request,json=routeRequest,proto3" json:"route_request,omitempty"`
	// CometBFT-specific observations from endpoint(s) that responded to the service request.
	// Multiple observations may occur when:
	// * Original endpoint fails
	// * Request is sent to additional endpoints for data collection
	EndpointObservations []*CometBFTEndpointObservation `protobuf:"bytes,2,rep,name=endpoint_observations,json=endpointObservations,proto3" json:"endpoint_observations,omitempty"`
}

func (x *CometBFTRequestObservations) Reset() {
	*x = CometBFTRequestObservations{}
	if protoimpl.UnsafeEnabled {
		mi := &file_path_qos_cometbft_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CometBFTRequestObservations) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CometBFTRequestObservations) ProtoMessage() {}

func (x *CometBFTRequestObservations) ProtoReflect() protoreflect.Message {
	mi := &file_path_qos_cometbft_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CometBFTRequestObservations.ProtoReflect.Descriptor instead.
func (*CometBFTRequestObservations) Descriptor() ([]byte, []int) {
	return file_path_qos_cometbft_proto_rawDescGZIP(), []int{0}
}

func (x *CometBFTRequestObservations) GetRouteRequest() string {
	if x != nil {
		return x.RouteRequest
	}
	return ""
}

func (x *CometBFTRequestObservations) GetEndpointObservations() []*CometBFTEndpointObservation {
	if x != nil {
		return x.EndpointObservations
	}
	return nil
}

// CometBFTEndpointObservation stores a single observation from an endpoint servicing the protocol response.
// Example: A Pocket node on Shannon backed by an Ethereum data node servicing an `eth_getBlockNumber` request.
type CometBFTEndpointObservation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Address of the endpoint handling the request (e.g., onchain address of a Pocket Morse/Shannon node)
	EndpointAddr string `protobuf:"bytes,1,opt,name=endpoint_addr,json=endpointAddr,proto3" json:"endpoint_addr,omitempty"`
	// Details of the response received from the endpoint
	//
	// Types that are assignable to ResponseObservation:
	//
	//	*CometBFTEndpointObservation_HealthResponse
	//	*CometBFTEndpointObservation_StatusResponse
	//	*CometBFTEndpointObservation_UnrecognizedResponse
	ResponseObservation isCometBFTEndpointObservation_ResponseObservation `protobuf_oneof:"response_observation"`
}

func (x *CometBFTEndpointObservation) Reset() {
	*x = CometBFTEndpointObservation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_path_qos_cometbft_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CometBFTEndpointObservation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CometBFTEndpointObservation) ProtoMessage() {}

func (x *CometBFTEndpointObservation) ProtoReflect() protoreflect.Message {
	mi := &file_path_qos_cometbft_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CometBFTEndpointObservation.ProtoReflect.Descriptor instead.
func (*CometBFTEndpointObservation) Descriptor() ([]byte, []int) {
	return file_path_qos_cometbft_proto_rawDescGZIP(), []int{1}
}

func (x *CometBFTEndpointObservation) GetEndpointAddr() string {
	if x != nil {
		return x.EndpointAddr
	}
	return ""
}

func (m *CometBFTEndpointObservation) GetResponseObservation() isCometBFTEndpointObservation_ResponseObservation {
	if m != nil {
		return m.ResponseObservation
	}
	return nil
}

func (x *CometBFTEndpointObservation) GetHealthResponse() *CometBFTHealthResponse {
	if x, ok := x.GetResponseObservation().(*CometBFTEndpointObservation_HealthResponse); ok {
		return x.HealthResponse
	}
	return nil
}

func (x *CometBFTEndpointObservation) GetStatusResponse() *CometBFTStatusResponse {
	if x, ok := x.GetResponseObservation().(*CometBFTEndpointObservation_StatusResponse); ok {
		return x.StatusResponse
	}
	return nil
}

func (x *CometBFTEndpointObservation) GetUnrecognizedResponse() *CometBFTUnrecognizedResponse {
	if x, ok := x.GetResponseObservation().(*CometBFTEndpointObservation_UnrecognizedResponse); ok {
		return x.UnrecognizedResponse
	}
	return nil
}

type isCometBFTEndpointObservation_ResponseObservation interface {
	isCometBFTEndpointObservation_ResponseObservation()
}

type CometBFTEndpointObservation_HealthResponse struct {
	// Response to `health` request
	HealthResponse *CometBFTHealthResponse `protobuf:"bytes,2,opt,name=health_response,json=healthResponse,proto3,oneof"`
}

type CometBFTEndpointObservation_StatusResponse struct {
	// Response to `status` request
	StatusResponse *CometBFTStatusResponse `protobuf:"bytes,3,opt,name=status_response,json=statusResponse,proto3,oneof"`
}

type CometBFTEndpointObservation_UnrecognizedResponse struct {
	// Responses not used in endpoint validation
	UnrecognizedResponse *CometBFTUnrecognizedResponse `protobuf:"bytes,4,opt,name=unrecognized_response,json=unrecognizedResponse,proto3,oneof"`
}

func (*CometBFTEndpointObservation_HealthResponse) isCometBFTEndpointObservation_ResponseObservation() {
}

func (*CometBFTEndpointObservation_StatusResponse) isCometBFTEndpointObservation_ResponseObservation() {
}

func (*CometBFTEndpointObservation_UnrecognizedResponse) isCometBFTEndpointObservation_ResponseObservation() {
}

// CometBFTHealthResponse stores the response to a `health` request
// Reference: https://docs.cometbft.com/v1.0/spec/rpc/#health
type CometBFTHealthResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	HealthStatusResponse bool `protobuf:"varint,1,opt,name=health_status_response,json=healthStatusResponse,proto3" json:"health_status_response,omitempty"`
}

func (x *CometBFTHealthResponse) Reset() {
	*x = CometBFTHealthResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_path_qos_cometbft_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CometBFTHealthResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CometBFTHealthResponse) ProtoMessage() {}

func (x *CometBFTHealthResponse) ProtoReflect() protoreflect.Message {
	mi := &file_path_qos_cometbft_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CometBFTHealthResponse.ProtoReflect.Descriptor instead.
func (*CometBFTHealthResponse) Descriptor() ([]byte, []int) {
	return file_path_qos_cometbft_proto_rawDescGZIP(), []int{2}
}

func (x *CometBFTHealthResponse) GetHealthStatusResponse() bool {
	if x != nil {
		return x.HealthStatusResponse
	}
	return false
}

// CometBFTBlockNumberResponse stores the latest block number from a `status` request
// Reference: https://docs.cometbft.com/v1.0/spec/rpc/#status
type CometBFTStatusResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ChainId                   string `protobuf:"bytes,1,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`
	Synced                    bool   `protobuf:"varint,2,opt,name=synced,proto3" json:"synced,omitempty"`
	LatestBlockHeightResponse string `protobuf:"bytes,3,opt,name=latest_block_height_response,json=latestBlockHeightResponse,proto3" json:"latest_block_height_response,omitempty"`
}

func (x *CometBFTStatusResponse) Reset() {
	*x = CometBFTStatusResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_path_qos_cometbft_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CometBFTStatusResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CometBFTStatusResponse) ProtoMessage() {}

func (x *CometBFTStatusResponse) ProtoReflect() protoreflect.Message {
	mi := &file_path_qos_cometbft_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CometBFTStatusResponse.ProtoReflect.Descriptor instead.
func (*CometBFTStatusResponse) Descriptor() ([]byte, []int) {
	return file_path_qos_cometbft_proto_rawDescGZIP(), []int{3}
}

func (x *CometBFTStatusResponse) GetChainId() string {
	if x != nil {
		return x.ChainId
	}
	return ""
}

func (x *CometBFTStatusResponse) GetSynced() bool {
	if x != nil {
		return x.Synced
	}
	return false
}

func (x *CometBFTStatusResponse) GetLatestBlockHeightResponse() string {
	if x != nil {
		return x.LatestBlockHeightResponse
	}
	return ""
}

// CometBFTUnrecognizedResponse handles requests with methods ignored by state update
// and endpoint validation
type CometBFTUnrecognizedResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	JsonrpcResponse *JsonRpcResponse `protobuf:"bytes,1,opt,name=jsonrpc_response,json=jsonrpcResponse,proto3" json:"jsonrpc_response,omitempty"`
}

func (x *CometBFTUnrecognizedResponse) Reset() {
	*x = CometBFTUnrecognizedResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_path_qos_cometbft_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CometBFTUnrecognizedResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CometBFTUnrecognizedResponse) ProtoMessage() {}

func (x *CometBFTUnrecognizedResponse) ProtoReflect() protoreflect.Message {
	mi := &file_path_qos_cometbft_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CometBFTUnrecognizedResponse.ProtoReflect.Descriptor instead.
func (*CometBFTUnrecognizedResponse) Descriptor() ([]byte, []int) {
	return file_path_qos_cometbft_proto_rawDescGZIP(), []int{4}
}

func (x *CometBFTUnrecognizedResponse) GetJsonrpcResponse() *JsonRpcResponse {
	if x != nil {
		return x.JsonrpcResponse
	}
	return nil
}

var File_path_qos_cometbft_proto protoreflect.FileDescriptor

var file_path_qos_cometbft_proto_rawDesc = []byte{
	0x0a, 0x17, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x71, 0x6f, 0x73, 0x2f, 0x63, 0x6f, 0x6d, 0x65, 0x74,
	0x62, 0x66, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x70, 0x61, 0x74, 0x68, 0x2e,
	0x71, 0x6f, 0x73, 0x1a, 0x16, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x71, 0x6f, 0x73, 0x2f, 0x6a, 0x73,
	0x6f, 0x6e, 0x72, 0x70, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x9e, 0x01, 0x0a, 0x1b,
	0x43, 0x6f, 0x6d, 0x65, 0x74, 0x42, 0x46, 0x54, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x4f,
	0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x23, 0x0a, 0x0d, 0x72,
	0x6f, 0x75, 0x74, 0x65, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0c, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x5a, 0x0a, 0x15, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x5f, 0x6f, 0x62, 0x73,
	0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x25, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x71, 0x6f, 0x73, 0x2e, 0x43, 0x6f, 0x6d, 0x65, 0x74,
	0x42, 0x46, 0x54, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x4f, 0x62, 0x73, 0x65, 0x72,
	0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x14, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74,
	0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0xd3, 0x02, 0x0a,
	0x1b, 0x43, 0x6f, 0x6d, 0x65, 0x74, 0x42, 0x46, 0x54, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e,
	0x74, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x23, 0x0a, 0x0d,
	0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0c, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x41, 0x64, 0x64,
	0x72, 0x12, 0x4b, 0x0a, 0x0f, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x5f, 0x72, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x70, 0x61, 0x74,
	0x68, 0x2e, 0x71, 0x6f, 0x73, 0x2e, 0x43, 0x6f, 0x6d, 0x65, 0x74, 0x42, 0x46, 0x54, 0x48, 0x65,
	0x61, 0x6c, 0x74, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x48, 0x00, 0x52, 0x0e,
	0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x4b,
	0x0a, 0x0f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x71,
	0x6f, 0x73, 0x2e, 0x43, 0x6f, 0x6d, 0x65, 0x74, 0x42, 0x46, 0x54, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x48, 0x00, 0x52, 0x0e, 0x73, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x5d, 0x0a, 0x15, 0x75,
	0x6e, 0x72, 0x65, 0x63, 0x6f, 0x67, 0x6e, 0x69, 0x7a, 0x65, 0x64, 0x5f, 0x72, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x70, 0x61, 0x74,
	0x68, 0x2e, 0x71, 0x6f, 0x73, 0x2e, 0x43, 0x6f, 0x6d, 0x65, 0x74, 0x42, 0x46, 0x54, 0x55, 0x6e,
	0x72, 0x65, 0x63, 0x6f, 0x67, 0x6e, 0x69, 0x7a, 0x65, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x48, 0x00, 0x52, 0x14, 0x75, 0x6e, 0x72, 0x65, 0x63, 0x6f, 0x67, 0x6e, 0x69, 0x7a,
	0x65, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x16, 0x0a, 0x14, 0x72, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x5f, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x22, 0x4e, 0x0a, 0x16, 0x43, 0x6f, 0x6d, 0x65, 0x74, 0x42, 0x46, 0x54, 0x48, 0x65,
	0x61, 0x6c, 0x74, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x34, 0x0a, 0x16,
	0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x5f, 0x72, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x14, 0x68, 0x65,
	0x61, 0x6c, 0x74, 0x68, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x8c, 0x01, 0x0a, 0x16, 0x43, 0x6f, 0x6d, 0x65, 0x74, 0x42, 0x46, 0x54, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x19, 0x0a,
	0x08, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x07, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x79, 0x6e, 0x63,
	0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x73, 0x79, 0x6e, 0x63, 0x65, 0x64,
	0x12, 0x3f, 0x0a, 0x1c, 0x6c, 0x61, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x5f, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x19, 0x6c, 0x61, 0x74, 0x65, 0x73, 0x74, 0x42, 0x6c,
	0x6f, 0x63, 0x6b, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x64, 0x0a, 0x1c, 0x43, 0x6f, 0x6d, 0x65, 0x74, 0x42, 0x46, 0x54, 0x55, 0x6e, 0x72,
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
	file_path_qos_cometbft_proto_rawDescOnce sync.Once
	file_path_qos_cometbft_proto_rawDescData = file_path_qos_cometbft_proto_rawDesc
)

func file_path_qos_cometbft_proto_rawDescGZIP() []byte {
	file_path_qos_cometbft_proto_rawDescOnce.Do(func() {
		file_path_qos_cometbft_proto_rawDescData = protoimpl.X.CompressGZIP(file_path_qos_cometbft_proto_rawDescData)
	})
	return file_path_qos_cometbft_proto_rawDescData
}

var file_path_qos_cometbft_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_path_qos_cometbft_proto_goTypes = []any{
	(*CometBFTRequestObservations)(nil),  // 0: path.qos.CometBFTRequestObservations
	(*CometBFTEndpointObservation)(nil),  // 1: path.qos.CometBFTEndpointObservation
	(*CometBFTHealthResponse)(nil),       // 2: path.qos.CometBFTHealthResponse
	(*CometBFTStatusResponse)(nil),       // 3: path.qos.CometBFTStatusResponse
	(*CometBFTUnrecognizedResponse)(nil), // 4: path.qos.CometBFTUnrecognizedResponse
	(*JsonRpcResponse)(nil),              // 5: path.qos.JsonRpcResponse
}
var file_path_qos_cometbft_proto_depIdxs = []int32{
	1, // 0: path.qos.CometBFTRequestObservations.endpoint_observations:type_name -> path.qos.CometBFTEndpointObservation
	2, // 1: path.qos.CometBFTEndpointObservation.health_response:type_name -> path.qos.CometBFTHealthResponse
	3, // 2: path.qos.CometBFTEndpointObservation.status_response:type_name -> path.qos.CometBFTStatusResponse
	4, // 3: path.qos.CometBFTEndpointObservation.unrecognized_response:type_name -> path.qos.CometBFTUnrecognizedResponse
	5, // 4: path.qos.CometBFTUnrecognizedResponse.jsonrpc_response:type_name -> path.qos.JsonRpcResponse
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_path_qos_cometbft_proto_init() }
func file_path_qos_cometbft_proto_init() {
	if File_path_qos_cometbft_proto != nil {
		return
	}
	file_path_qos_jsonrpc_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_path_qos_cometbft_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*CometBFTRequestObservations); i {
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
		file_path_qos_cometbft_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*CometBFTEndpointObservation); i {
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
		file_path_qos_cometbft_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*CometBFTHealthResponse); i {
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
		file_path_qos_cometbft_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*CometBFTStatusResponse); i {
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
		file_path_qos_cometbft_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*CometBFTUnrecognizedResponse); i {
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
	file_path_qos_cometbft_proto_msgTypes[1].OneofWrappers = []any{
		(*CometBFTEndpointObservation_HealthResponse)(nil),
		(*CometBFTEndpointObservation_StatusResponse)(nil),
		(*CometBFTEndpointObservation_UnrecognizedResponse)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_path_qos_cometbft_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_path_qos_cometbft_proto_goTypes,
		DependencyIndexes: file_path_qos_cometbft_proto_depIdxs,
		MessageInfos:      file_path_qos_cometbft_proto_msgTypes,
	}.Build()
	File_path_qos_cometbft_proto = out.File
	file_path_qos_cometbft_proto_rawDesc = nil
	file_path_qos_cometbft_proto_goTypes = nil
	file_path_qos_cometbft_proto_depIdxs = nil
}
