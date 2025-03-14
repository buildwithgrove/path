// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
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

// MorseEndpointErrorType enumerates possible relay errors when interacting with Morse endpoints
type MorseEndpointErrorType int32

const (
	MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_UNSPECIFIED       MorseEndpointErrorType = 0
	MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_CONNECTION_FAILED MorseEndpointErrorType = 1
	MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_TIMEOUT           MorseEndpointErrorType = 2
	MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_MAXED_OUT         MorseEndpointErrorType = 3
	MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_MISCONFIGURED     MorseEndpointErrorType = 4
	MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_INVALID_RESPONSE  MorseEndpointErrorType = 5
	MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_INTERNAL          MorseEndpointErrorType = 6 // Added for internal gateway errors
)

// Enum value maps for MorseEndpointErrorType.
var (
	MorseEndpointErrorType_name = map[int32]string{
		0: "MORSE_ENDPOINT_ERROR_UNSPECIFIED",
		1: "MORSE_ENDPOINT_ERROR_CONNECTION_FAILED",
		2: "MORSE_ENDPOINT_ERROR_TIMEOUT",
		3: "MORSE_ENDPOINT_ERROR_MAXED_OUT",
		4: "MORSE_ENDPOINT_ERROR_MISCONFIGURED",
		5: "MORSE_ENDPOINT_ERROR_INVALID_RESPONSE",
		6: "MORSE_ENDPOINT_ERROR_INTERNAL",
	}
	MorseEndpointErrorType_value = map[string]int32{
		"MORSE_ENDPOINT_ERROR_UNSPECIFIED":       0,
		"MORSE_ENDPOINT_ERROR_CONNECTION_FAILED": 1,
		"MORSE_ENDPOINT_ERROR_TIMEOUT":           2,
		"MORSE_ENDPOINT_ERROR_MAXED_OUT":         3,
		"MORSE_ENDPOINT_ERROR_MISCONFIGURED":     4,
		"MORSE_ENDPOINT_ERROR_INVALID_RESPONSE":  5,
		"MORSE_ENDPOINT_ERROR_INTERNAL":          6,
	}
)

func (x MorseEndpointErrorType) Enum() *MorseEndpointErrorType {
	p := new(MorseEndpointErrorType)
	*p = x
	return p
}

func (x MorseEndpointErrorType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (MorseEndpointErrorType) Descriptor() protoreflect.EnumDescriptor {
	return file_path_protocol_morse_proto_enumTypes[0].Descriptor()
}

func (MorseEndpointErrorType) Type() protoreflect.EnumType {
	return &file_path_protocol_morse_proto_enumTypes[0]
}

func (x MorseEndpointErrorType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use MorseEndpointErrorType.Descriptor instead.
func (MorseEndpointErrorType) EnumDescriptor() ([]byte, []int) {
	return file_path_protocol_morse_proto_rawDescGZIP(), []int{0}
}

// TODO_DOCUMENT(@adshmh): Document the sanction system in the Morse protocol implementation:
// - Enumerate all sanction types with their specific triggers
// - Detail error conditions that activate each sanction category
// - Explain the rationale behind each sanction's severity level
// - Specify sanction durations and how they're calculated
// - Document potential escalation path for repeated violations
// - Include examples of boundary cases where sanctions apply/don't apply
//
// MorseSanctionType specifies the duration type for endpoint sanctions
type MorseSanctionType int32

const (
	MorseSanctionType_MORSE_SANCTION_UNSPECIFIED MorseSanctionType = 0
	MorseSanctionType_MORSE_SANCTION_SESSION     MorseSanctionType = 1 // Valid only for current session
	MorseSanctionType_MORSE_SANCTION_PERMANENT   MorseSanctionType = 2 // Sanction persists indefinitely; can only be cleared by Gateway restart (e.g., redeploying the K8s pod or restarting the binary)
)

// Enum value maps for MorseSanctionType.
var (
	MorseSanctionType_name = map[int32]string{
		0: "MORSE_SANCTION_UNSPECIFIED",
		1: "MORSE_SANCTION_SESSION",
		2: "MORSE_SANCTION_PERMANENT",
	}
	MorseSanctionType_value = map[string]int32{
		"MORSE_SANCTION_UNSPECIFIED": 0,
		"MORSE_SANCTION_SESSION":     1,
		"MORSE_SANCTION_PERMANENT":   2,
	}
)

func (x MorseSanctionType) Enum() *MorseSanctionType {
	p := new(MorseSanctionType)
	*p = x
	return p
}

func (x MorseSanctionType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (MorseSanctionType) Descriptor() protoreflect.EnumDescriptor {
	return file_path_protocol_morse_proto_enumTypes[1].Descriptor()
}

func (MorseSanctionType) Type() protoreflect.EnumType {
	return &file_path_protocol_morse_proto_enumTypes[1]
}

func (x MorseSanctionType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use MorseSanctionType.Descriptor instead.
func (MorseSanctionType) EnumDescriptor() ([]byte, []int) {
	return file_path_protocol_morse_proto_rawDescGZIP(), []int{1}
}

// MorseRequestObservations contains Morse-specific observations collected from relays
// handling a single service request.
type MorseRequestObservations struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Service ID (i.e. chain ID) for which the observation was made
	ServiceId string `protobuf:"bytes,1,opt,name=service_id,json=serviceId,proto3" json:"service_id,omitempty"`
	// Multiple observations possible if:
	// - Original endpoint returns invalid response
	// - Retry mechanism activates
	EndpointObservations []*MorseEndpointObservation `protobuf:"bytes,2,rep,name=endpoint_observations,json=endpointObservations,proto3" json:"endpoint_observations,omitempty"`
}

func (x *MorseRequestObservations) Reset() {
	*x = MorseRequestObservations{}
	if protoimpl.UnsafeEnabled {
		mi := &file_path_protocol_morse_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MorseRequestObservations) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MorseRequestObservations) ProtoMessage() {}

func (x *MorseRequestObservations) ProtoReflect() protoreflect.Message {
	mi := &file_path_protocol_morse_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MorseRequestObservations.ProtoReflect.Descriptor instead.
func (*MorseRequestObservations) Descriptor() ([]byte, []int) {
	return file_path_protocol_morse_proto_rawDescGZIP(), []int{0}
}

func (x *MorseRequestObservations) GetServiceId() string {
	if x != nil {
		return x.ServiceId
	}
	return ""
}

func (x *MorseRequestObservations) GetEndpointObservations() []*MorseEndpointObservation {
	if x != nil {
		return x.EndpointObservations
	}
	return nil
}

// MorseEndpointObservation stores a single observation from an endpoint
type MorseEndpointObservation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Address of the endpoint handling the request
	EndpointAddr string `protobuf:"bytes,1,opt,name=endpoint_addr,json=endpointAddr,proto3" json:"endpoint_addr,omitempty"`
	// Application address that signed the associated relay
	AppAddress string `protobuf:"bytes,2,opt,name=app_address,json=appAddress,proto3" json:"app_address,omitempty"`
	// Session information when available
	SessionKey       string `protobuf:"bytes,3,opt,name=session_key,json=sessionKey,proto3" json:"session_key,omitempty"`
	SessionServiceId string `protobuf:"bytes,4,opt,name=session_service_id,json=sessionServiceId,proto3" json:"session_service_id,omitempty"`
	SessionHeight    int32  `protobuf:"varint,5,opt,name=session_height,json=sessionHeight,proto3" json:"session_height,omitempty"`
	// Error type if relay to this endpoint failed
	ErrorType *MorseEndpointErrorType `protobuf:"varint,6,opt,name=error_type,json=errorType,proto3,enum=path.protocol.MorseEndpointErrorType,oneof" json:"error_type,omitempty"`
	// Additional error details when available
	ErrorDetails *string `protobuf:"bytes,7,opt,name=error_details,json=errorDetails,proto3,oneof" json:"error_details,omitempty"`
	// Recommended sanction type based on the error
	RecommendedSanction *MorseSanctionType `protobuf:"varint,8,opt,name=recommended_sanction,json=recommendedSanction,proto3,enum=path.protocol.MorseSanctionType,oneof" json:"recommended_sanction,omitempty"`
}

func (x *MorseEndpointObservation) Reset() {
	*x = MorseEndpointObservation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_path_protocol_morse_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MorseEndpointObservation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MorseEndpointObservation) ProtoMessage() {}

func (x *MorseEndpointObservation) ProtoReflect() protoreflect.Message {
	mi := &file_path_protocol_morse_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MorseEndpointObservation.ProtoReflect.Descriptor instead.
func (*MorseEndpointObservation) Descriptor() ([]byte, []int) {
	return file_path_protocol_morse_proto_rawDescGZIP(), []int{1}
}

func (x *MorseEndpointObservation) GetEndpointAddr() string {
	if x != nil {
		return x.EndpointAddr
	}
	return ""
}

func (x *MorseEndpointObservation) GetAppAddress() string {
	if x != nil {
		return x.AppAddress
	}
	return ""
}

func (x *MorseEndpointObservation) GetSessionKey() string {
	if x != nil {
		return x.SessionKey
	}
	return ""
}

func (x *MorseEndpointObservation) GetSessionServiceId() string {
	if x != nil {
		return x.SessionServiceId
	}
	return ""
}

func (x *MorseEndpointObservation) GetSessionHeight() int32 {
	if x != nil {
		return x.SessionHeight
	}
	return 0
}

func (x *MorseEndpointObservation) GetErrorType() MorseEndpointErrorType {
	if x != nil && x.ErrorType != nil {
		return *x.ErrorType
	}
	return MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_UNSPECIFIED
}

func (x *MorseEndpointObservation) GetErrorDetails() string {
	if x != nil && x.ErrorDetails != nil {
		return *x.ErrorDetails
	}
	return ""
}

func (x *MorseEndpointObservation) GetRecommendedSanction() MorseSanctionType {
	if x != nil && x.RecommendedSanction != nil {
		return *x.RecommendedSanction
	}
	return MorseSanctionType_MORSE_SANCTION_UNSPECIFIED
}

// MorseObservationsList is a wrapper message that enables embedding lists of
// Morse observations in other protocol buffers.
type MorseObservationsList struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Observations []*MorseRequestObservations `protobuf:"bytes,1,rep,name=observations,proto3" json:"observations,omitempty"`
}

func (x *MorseObservationsList) Reset() {
	*x = MorseObservationsList{}
	if protoimpl.UnsafeEnabled {
		mi := &file_path_protocol_morse_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MorseObservationsList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MorseObservationsList) ProtoMessage() {}

func (x *MorseObservationsList) ProtoReflect() protoreflect.Message {
	mi := &file_path_protocol_morse_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
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
	return file_path_protocol_morse_proto_rawDescGZIP(), []int{2}
}

func (x *MorseObservationsList) GetObservations() []*MorseRequestObservations {
	if x != nil {
		return x.Observations
	}
	return nil
}

var File_path_protocol_morse_proto protoreflect.FileDescriptor

var file_path_protocol_morse_proto_rawDesc = []byte{
	0x0a, 0x19, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2f,
	0x6d, 0x6f, 0x72, 0x73, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x70, 0x61, 0x74,
	0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x22, 0x97, 0x01, 0x0a, 0x18, 0x4d,
	0x6f, 0x72, 0x73, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x4f, 0x62, 0x73, 0x65, 0x72,
	0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x49, 0x64, 0x12, 0x5c, 0x0a, 0x15, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69,
	0x6e, 0x74, 0x5f, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x4d, 0x6f, 0x72, 0x73, 0x65, 0x45, 0x6e, 0x64, 0x70, 0x6f,
	0x69, 0x6e, 0x74, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x14,
	0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x22, 0xdf, 0x03, 0x0a, 0x18, 0x4d, 0x6f, 0x72, 0x73, 0x65, 0x45, 0x6e,
	0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x23, 0x0a, 0x0d, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x5f, 0x61, 0x64,
	0x64, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69,
	0x6e, 0x74, 0x41, 0x64, 0x64, 0x72, 0x12, 0x1f, 0x0a, 0x0b, 0x61, 0x70, 0x70, 0x5f, 0x61, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x61, 0x70, 0x70,
	0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x1f, 0x0a, 0x0b, 0x73, 0x65, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x73, 0x65,
	0x73, 0x73, 0x69, 0x6f, 0x6e, 0x4b, 0x65, 0x79, 0x12, 0x2c, 0x0a, 0x12, 0x73, 0x65, 0x73, 0x73,
	0x69, 0x6f, 0x6e, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x10, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x49, 0x64, 0x12, 0x25, 0x0a, 0x0e, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f,
	0x6e, 0x5f, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0d,
	0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x49, 0x0a,
	0x0a, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x25, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f,
	0x6c, 0x2e, 0x4d, 0x6f, 0x72, 0x73, 0x65, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x45,
	0x72, 0x72, 0x6f, 0x72, 0x54, 0x79, 0x70, 0x65, 0x48, 0x00, 0x52, 0x09, 0x65, 0x72, 0x72, 0x6f,
	0x72, 0x54, 0x79, 0x70, 0x65, 0x88, 0x01, 0x01, 0x12, 0x28, 0x0a, 0x0d, 0x65, 0x72, 0x72, 0x6f,
	0x72, 0x5f, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x48,
	0x01, 0x52, 0x0c, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x88,
	0x01, 0x01, 0x12, 0x58, 0x0a, 0x14, 0x72, 0x65, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x64, 0x65,
	0x64, 0x5f, 0x73, 0x61, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x20, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c,
	0x2e, 0x4d, 0x6f, 0x72, 0x73, 0x65, 0x53, 0x61, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x79,
	0x70, 0x65, 0x48, 0x02, 0x52, 0x13, 0x72, 0x65, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x64, 0x65,
	0x64, 0x53, 0x61, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x88, 0x01, 0x01, 0x42, 0x0d, 0x0a, 0x0b,
	0x5f, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x42, 0x10, 0x0a, 0x0e, 0x5f,
	0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x42, 0x17, 0x0a,
	0x15, 0x5f, 0x72, 0x65, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x64, 0x65, 0x64, 0x5f, 0x73, 0x61,
	0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x64, 0x0a, 0x15, 0x4d, 0x6f, 0x72, 0x73, 0x65, 0x4f,
	0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x4c, 0x69, 0x73, 0x74, 0x12,
	0x4b, 0x0a, 0x0c, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x70, 0x61, 0x74, 0x68, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x4d, 0x6f, 0x72, 0x73, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x0c,
	0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2a, 0xa6, 0x02, 0x0a,
	0x16, 0x4d, 0x6f, 0x72, 0x73, 0x65, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x45, 0x72,
	0x72, 0x6f, 0x72, 0x54, 0x79, 0x70, 0x65, 0x12, 0x24, 0x0a, 0x20, 0x4d, 0x4f, 0x52, 0x53, 0x45,
	0x5f, 0x45, 0x4e, 0x44, 0x50, 0x4f, 0x49, 0x4e, 0x54, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x5f,
	0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x2a, 0x0a,
	0x26, 0x4d, 0x4f, 0x52, 0x53, 0x45, 0x5f, 0x45, 0x4e, 0x44, 0x50, 0x4f, 0x49, 0x4e, 0x54, 0x5f,
	0x45, 0x52, 0x52, 0x4f, 0x52, 0x5f, 0x43, 0x4f, 0x4e, 0x4e, 0x45, 0x43, 0x54, 0x49, 0x4f, 0x4e,
	0x5f, 0x46, 0x41, 0x49, 0x4c, 0x45, 0x44, 0x10, 0x01, 0x12, 0x20, 0x0a, 0x1c, 0x4d, 0x4f, 0x52,
	0x53, 0x45, 0x5f, 0x45, 0x4e, 0x44, 0x50, 0x4f, 0x49, 0x4e, 0x54, 0x5f, 0x45, 0x52, 0x52, 0x4f,
	0x52, 0x5f, 0x54, 0x49, 0x4d, 0x45, 0x4f, 0x55, 0x54, 0x10, 0x02, 0x12, 0x22, 0x0a, 0x1e, 0x4d,
	0x4f, 0x52, 0x53, 0x45, 0x5f, 0x45, 0x4e, 0x44, 0x50, 0x4f, 0x49, 0x4e, 0x54, 0x5f, 0x45, 0x52,
	0x52, 0x4f, 0x52, 0x5f, 0x4d, 0x41, 0x58, 0x45, 0x44, 0x5f, 0x4f, 0x55, 0x54, 0x10, 0x03, 0x12,
	0x26, 0x0a, 0x22, 0x4d, 0x4f, 0x52, 0x53, 0x45, 0x5f, 0x45, 0x4e, 0x44, 0x50, 0x4f, 0x49, 0x4e,
	0x54, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x5f, 0x4d, 0x49, 0x53, 0x43, 0x4f, 0x4e, 0x46, 0x49,
	0x47, 0x55, 0x52, 0x45, 0x44, 0x10, 0x04, 0x12, 0x29, 0x0a, 0x25, 0x4d, 0x4f, 0x52, 0x53, 0x45,
	0x5f, 0x45, 0x4e, 0x44, 0x50, 0x4f, 0x49, 0x4e, 0x54, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x5f,
	0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x5f, 0x52, 0x45, 0x53, 0x50, 0x4f, 0x4e, 0x53, 0x45,
	0x10, 0x05, 0x12, 0x21, 0x0a, 0x1d, 0x4d, 0x4f, 0x52, 0x53, 0x45, 0x5f, 0x45, 0x4e, 0x44, 0x50,
	0x4f, 0x49, 0x4e, 0x54, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x5f, 0x49, 0x4e, 0x54, 0x45, 0x52,
	0x4e, 0x41, 0x4c, 0x10, 0x06, 0x2a, 0x6d, 0x0a, 0x11, 0x4d, 0x6f, 0x72, 0x73, 0x65, 0x53, 0x61,
	0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1e, 0x0a, 0x1a, 0x4d, 0x4f,
	0x52, 0x53, 0x45, 0x5f, 0x53, 0x41, 0x4e, 0x43, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x55, 0x4e, 0x53,
	0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x1a, 0x0a, 0x16, 0x4d, 0x4f,
	0x52, 0x53, 0x45, 0x5f, 0x53, 0x41, 0x4e, 0x43, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x53, 0x45, 0x53,
	0x53, 0x49, 0x4f, 0x4e, 0x10, 0x01, 0x12, 0x1c, 0x0a, 0x18, 0x4d, 0x4f, 0x52, 0x53, 0x45, 0x5f,
	0x53, 0x41, 0x4e, 0x43, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x50, 0x45, 0x52, 0x4d, 0x41, 0x4e, 0x45,
	0x4e, 0x54, 0x10, 0x02, 0x42, 0x35, 0x5a, 0x33, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x77, 0x69, 0x74, 0x68, 0x67, 0x72, 0x6f, 0x76,
	0x65, 0x2f, 0x70, 0x61, 0x74, 0x68, 0x2f, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
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

var file_path_protocol_morse_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_path_protocol_morse_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_path_protocol_morse_proto_goTypes = []any{
	(MorseEndpointErrorType)(0),      // 0: path.protocol.MorseEndpointErrorType
	(MorseSanctionType)(0),           // 1: path.protocol.MorseSanctionType
	(*MorseRequestObservations)(nil), // 2: path.protocol.MorseRequestObservations
	(*MorseEndpointObservation)(nil), // 3: path.protocol.MorseEndpointObservation
	(*MorseObservationsList)(nil),    // 4: path.protocol.MorseObservationsList
}
var file_path_protocol_morse_proto_depIdxs = []int32{
	3, // 0: path.protocol.MorseRequestObservations.endpoint_observations:type_name -> path.protocol.MorseEndpointObservation
	0, // 1: path.protocol.MorseEndpointObservation.error_type:type_name -> path.protocol.MorseEndpointErrorType
	1, // 2: path.protocol.MorseEndpointObservation.recommended_sanction:type_name -> path.protocol.MorseSanctionType
	2, // 3: path.protocol.MorseObservationsList.observations:type_name -> path.protocol.MorseRequestObservations
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_path_protocol_morse_proto_init() }
func file_path_protocol_morse_proto_init() {
	if File_path_protocol_morse_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_path_protocol_morse_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*MorseRequestObservations); i {
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
		file_path_protocol_morse_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*MorseEndpointObservation); i {
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
		file_path_protocol_morse_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*MorseObservationsList); i {
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
	file_path_protocol_morse_proto_msgTypes[1].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_path_protocol_morse_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_path_protocol_morse_proto_goTypes,
		DependencyIndexes: file_path_protocol_morse_proto_depIdxs,
		EnumInfos:         file_path_protocol_morse_proto_enumTypes,
		MessageInfos:      file_path_protocol_morse_proto_msgTypes,
	}.Build()
	File_path_protocol_morse_proto = out.File
	file_path_protocol_morse_proto_rawDesc = nil
	file_path_protocol_morse_proto_goTypes = nil
	file_path_protocol_morse_proto_depIdxs = nil
}
