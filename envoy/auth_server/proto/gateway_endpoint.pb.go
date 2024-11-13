// To generate the Go code from this proto file, run:
// `make proto_generate`
// which runs:
// `protoc --go_out=./envoy/auth_server/proto --go-grpc_out=./envoy/auth_server/proto envoy/auth_server/proto/gateway_endpoint.proto`

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v5.28.3
// source: envoy/auth_server/proto/gateway_endpoint.proto

package proto

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

// CapacityLimitPeriod is the period over which the capacity limit is enforced.
// For example: CapacityLimit=`100,000` and CapacityLimitPeriod=`daily`
// enforces a rate limit of 100,000 requests per day.
type CapacityLimitPeriod int32

const (
	CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_UNSPECIFIED CapacityLimitPeriod = 0
	CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_DAILY       CapacityLimitPeriod = 1
	CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_WEEKLY      CapacityLimitPeriod = 2
	CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY     CapacityLimitPeriod = 3
)

// Enum value maps for CapacityLimitPeriod.
var (
	CapacityLimitPeriod_name = map[int32]string{
		0: "CAPACITY_LIMIT_PERIOD_UNSPECIFIED",
		1: "CAPACITY_LIMIT_PERIOD_DAILY",
		2: "CAPACITY_LIMIT_PERIOD_WEEKLY",
		3: "CAPACITY_LIMIT_PERIOD_MONTHLY",
	}
	CapacityLimitPeriod_value = map[string]int32{
		"CAPACITY_LIMIT_PERIOD_UNSPECIFIED": 0,
		"CAPACITY_LIMIT_PERIOD_DAILY":       1,
		"CAPACITY_LIMIT_PERIOD_WEEKLY":      2,
		"CAPACITY_LIMIT_PERIOD_MONTHLY":     3,
	}
)

func (x CapacityLimitPeriod) Enum() *CapacityLimitPeriod {
	p := new(CapacityLimitPeriod)
	*p = x
	return p
}

func (x CapacityLimitPeriod) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (CapacityLimitPeriod) Descriptor() protoreflect.EnumDescriptor {
	return file_envoy_auth_server_proto_gateway_endpoint_proto_enumTypes[0].Descriptor()
}

func (CapacityLimitPeriod) Type() protoreflect.EnumType {
	return &file_envoy_auth_server_proto_gateway_endpoint_proto_enumTypes[0]
}

func (x CapacityLimitPeriod) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use CapacityLimitPeriod.Descriptor instead.
func (CapacityLimitPeriod) EnumDescriptor() ([]byte, []int) {
	return file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescGZIP(), []int{0}
}

type AuthDataRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *AuthDataRequest) Reset() {
	*x = AuthDataRequest{}
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AuthDataRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthDataRequest) ProtoMessage() {}

func (x *AuthDataRequest) ProtoReflect() protoreflect.Message {
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthDataRequest.ProtoReflect.Descriptor instead.
func (*AuthDataRequest) Descriptor() ([]byte, []int) {
	return file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescGZIP(), []int{0}
}

type AuthDataResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Endpoints map[string]*GatewayEndpoint `protobuf:"bytes,1,rep,name=endpoints,proto3" json:"endpoints,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *AuthDataResponse) Reset() {
	*x = AuthDataResponse{}
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AuthDataResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthDataResponse) ProtoMessage() {}

func (x *AuthDataResponse) ProtoReflect() protoreflect.Message {
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthDataResponse.ProtoReflect.Descriptor instead.
func (*AuthDataResponse) Descriptor() ([]byte, []int) {
	return file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescGZIP(), []int{1}
}

func (x *AuthDataResponse) GetEndpoints() map[string]*GatewayEndpoint {
	if x != nil {
		return x.Endpoints
	}
	return nil
}

type AuthDataUpdatesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *AuthDataUpdatesRequest) Reset() {
	*x = AuthDataUpdatesRequest{}
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AuthDataUpdatesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthDataUpdatesRequest) ProtoMessage() {}

func (x *AuthDataUpdatesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthDataUpdatesRequest.ProtoReflect.Descriptor instead.
func (*AuthDataUpdatesRequest) Descriptor() ([]byte, []int) {
	return file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescGZIP(), []int{2}
}

// An AuthDataUpdate message is sent from the remote gRPC server when a GatewayEndpoint is created, updated, or deleted.
type AuthDataUpdate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A unique identifier for a user account's GatewayEndpoint. It must be passed as the last path segment of
	// the service request URL. For example: POST https://api.path.xyz/v1/{gateway_endpoint_id}
	// This is used to identify the GatewayEndpoint when making a service request.
	EndpointId string `protobuf:"bytes,1,opt,name=endpoint_id,json=endpointId,proto3" json:"endpoint_id,omitempty"`
	// The GatewayEndpoint to upsert to the database. If delete is true, the GatewayEndpoint will be deleted and this field will be empty in the gRPC data.
	GatewayEndpoint *GatewayEndpoint `protobuf:"bytes,2,opt,name=gateway_endpoint,json=gatewayEndpoint,proto3" json:"gateway_endpoint,omitempty"`
	// Indicates whether the GatewayEndpoint should be deleted
	Delete bool `protobuf:"varint,3,opt,name=delete,proto3" json:"delete,omitempty"`
}

func (x *AuthDataUpdate) Reset() {
	*x = AuthDataUpdate{}
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AuthDataUpdate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthDataUpdate) ProtoMessage() {}

func (x *AuthDataUpdate) ProtoReflect() protoreflect.Message {
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthDataUpdate.ProtoReflect.Descriptor instead.
func (*AuthDataUpdate) Descriptor() ([]byte, []int) {
	return file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescGZIP(), []int{3}
}

func (x *AuthDataUpdate) GetEndpointId() string {
	if x != nil {
		return x.EndpointId
	}
	return ""
}

func (x *AuthDataUpdate) GetGatewayEndpoint() *GatewayEndpoint {
	if x != nil {
		return x.GatewayEndpoint
	}
	return nil
}

func (x *AuthDataUpdate) GetDelete() bool {
	if x != nil {
		return x.Delete
	}
	return false
}

// A GatewayEndpoint represents a user account's endpoint, which has two primary functions:
// 1. Identifying which endpoint is being used to make a service request.
// 2. Allowing configuration of endpoint-specific settings, such as API key authorization, etc.
//
// A GatewayEndpoint is associated to a single UserAccount. A UserAccount can have multiple GatewayEndpoints.
// Settings related to service requests, such as enforcing API key authorization, are configured per GatewayEndpoint.
type GatewayEndpoint struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The unique identifier for a user account's endpoint, which must be passed
	// as the last path segment of the service request URL. eg `/v1/{endpoint_id}`
	// This is used to identify the GatewayEndpoint when making a service request.
	EndpointId string `protobuf:"bytes,1,opt,name=endpoint_id,json=endpointId,proto3" json:"endpoint_id,omitempty"`
	// The authorization settings for the GatewayEndpoint.
	Auth *Auth `protobuf:"bytes,2,opt,name=auth,proto3" json:"auth,omitempty"`
	// The UserAccount that the GatewayEndpoint belongs to, including the PlanType.
	UserAccount *UserAccount `protobuf:"bytes,3,opt,name=user_account,json=userAccount,proto3" json:"user_account,omitempty"`
	// The rate limiting settings for the GatewayEndpoint, which includes both
	// the throughput (TPS) limit and the capacity (longer period) limit.
	RateLimiting *RateLimiting `protobuf:"bytes,4,opt,name=rate_limiting,json=rateLimiting,proto3" json:"rate_limiting,omitempty"`
}

func (x *GatewayEndpoint) Reset() {
	*x = GatewayEndpoint{}
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GatewayEndpoint) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GatewayEndpoint) ProtoMessage() {}

func (x *GatewayEndpoint) ProtoReflect() protoreflect.Message {
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GatewayEndpoint.ProtoReflect.Descriptor instead.
func (*GatewayEndpoint) Descriptor() ([]byte, []int) {
	return file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescGZIP(), []int{4}
}

func (x *GatewayEndpoint) GetEndpointId() string {
	if x != nil {
		return x.EndpointId
	}
	return ""
}

func (x *GatewayEndpoint) GetAuth() *Auth {
	if x != nil {
		return x.Auth
	}
	return nil
}

func (x *GatewayEndpoint) GetUserAccount() *UserAccount {
	if x != nil {
		return x.UserAccount
	}
	return nil
}

func (x *GatewayEndpoint) GetRateLimiting() *RateLimiting {
	if x != nil {
		return x.RateLimiting
	}
	return nil
}

// The authorization settings for a GatewayEndpoint.
type Auth struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A boolean indicating whether the GatewayEndpoint requires authentication.
	RequireAuth bool `protobuf:"varint,1,opt,name=require_auth,json=requireAuth,proto3" json:"require_auth,omitempty"`
	// A map of ProviderUserIDs authorized to access this UserAccount's GatewayEndpoints.
	AuthorizedUsers map[string]*Empty `protobuf:"bytes,2,rep,name=authorized_users,json=authorizedUsers,proto3" json:"authorized_users,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Auth) Reset() {
	*x = Auth{}
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Auth) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Auth) ProtoMessage() {}

func (x *Auth) ProtoReflect() protoreflect.Message {
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Auth.ProtoReflect.Descriptor instead.
func (*Auth) Descriptor() ([]byte, []int) {
	return file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescGZIP(), []int{5}
}

func (x *Auth) GetRequireAuth() bool {
	if x != nil {
		return x.RequireAuth
	}
	return false
}

func (x *Auth) GetAuthorizedUsers() map[string]*Empty {
	if x != nil {
		return x.AuthorizedUsers
	}
	return nil
}

// A UserAccount contains the PlanType and may have multiple GatewayEndpoints.
type UserAccount struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The unique identifier for a UserAccount.
	AccountId string `protobuf:"bytes,1,opt,name=account_id,json=accountId,proto3" json:"account_id,omitempty"`
	// The plan type for a UserAccount, which identifies the pricing plan for the Account.
	PlanType string `protobuf:"bytes,2,opt,name=plan_type,json=planType,proto3" json:"plan_type,omitempty"`
}

func (x *UserAccount) Reset() {
	*x = UserAccount{}
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UserAccount) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserAccount) ProtoMessage() {}

func (x *UserAccount) ProtoReflect() protoreflect.Message {
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserAccount.ProtoReflect.Descriptor instead.
func (*UserAccount) Descriptor() ([]byte, []int) {
	return file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescGZIP(), []int{6}
}

func (x *UserAccount) GetAccountId() string {
	if x != nil {
		return x.AccountId
	}
	return ""
}

func (x *UserAccount) GetPlanType() string {
	if x != nil {
		return x.PlanType
	}
	return ""
}

// The rate limiting settings for a GatewayEndpoint.
type RateLimiting struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ThroughputLimit refers to rate limiting per-second (TPS).
	// This is used to prevent DoS or DDoS attacks, as well as enforce pricing plan limits.
	ThroughputLimit int32 `protobuf:"varint,1,opt,name=throughput_limit,json=throughputLimit,proto3" json:"throughput_limit,omitempty"`
	// CapacityLimit refers to rate limiting over longer periods, such as a day, week or month.
	// This is to prevent abuse of the services provided, as well enforce pricing plan limits.
	CapacityLimit int32 `protobuf:"varint,2,opt,name=capacity_limit,json=capacityLimit,proto3" json:"capacity_limit,omitempty"`
	// The period over which the CapacityLimit is enforced. One of `daily`, `weekly` or `monthly`.
	CapacityLimitPeriod CapacityLimitPeriod `protobuf:"varint,3,opt,name=capacity_limit_period,json=capacityLimitPeriod,proto3,enum=proto.CapacityLimitPeriod" json:"capacity_limit_period,omitempty"`
}

func (x *RateLimiting) Reset() {
	*x = RateLimiting{}
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RateLimiting) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RateLimiting) ProtoMessage() {}

func (x *RateLimiting) ProtoReflect() protoreflect.Message {
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RateLimiting.ProtoReflect.Descriptor instead.
func (*RateLimiting) Descriptor() ([]byte, []int) {
	return file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescGZIP(), []int{7}
}

func (x *RateLimiting) GetThroughputLimit() int32 {
	if x != nil {
		return x.ThroughputLimit
	}
	return 0
}

func (x *RateLimiting) GetCapacityLimit() int32 {
	if x != nil {
		return x.CapacityLimit
	}
	return 0
}

func (x *RateLimiting) GetCapacityLimitPeriod() CapacityLimitPeriod {
	if x != nil {
		return x.CapacityLimitPeriod
	}
	return CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_UNSPECIFIED
}

// An Empty message is used to indicate that a field is not set.
type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Empty) Reset() {
	*x = Empty{}
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescGZIP(), []int{8}
}

var File_envoy_auth_server_proto_gateway_endpoint_proto protoreflect.FileDescriptor

var file_envoy_auth_server_proto_gateway_endpoint_proto_rawDesc = []byte{
	0x0a, 0x2e, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x5f, 0x73, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61,
	0x79, 0x5f, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x11, 0x0a, 0x0f, 0x41, 0x75, 0x74, 0x68, 0x44,
	0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0xae, 0x01, 0x0a, 0x10, 0x41,
	0x75, 0x74, 0x68, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x44, 0x0a, 0x09, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x26, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x44,
	0x61, 0x74, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x45, 0x6e, 0x64, 0x70,
	0x6f, 0x69, 0x6e, 0x74, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x09, 0x65, 0x6e, 0x64, 0x70,
	0x6f, 0x69, 0x6e, 0x74, 0x73, 0x1a, 0x54, 0x0a, 0x0e, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e,
	0x74, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x18, 0x0a, 0x16, 0x41,
	0x75, 0x74, 0x68, 0x44, 0x61, 0x74, 0x61, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x8c, 0x01, 0x0a, 0x0e, 0x41, 0x75, 0x74, 0x68, 0x44, 0x61,
	0x74, 0x61, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x65, 0x6e, 0x64, 0x70,
	0x6f, 0x69, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x65,
	0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x41, 0x0a, 0x10, 0x67, 0x61, 0x74,
	0x65, 0x77, 0x61, 0x79, 0x5f, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x47, 0x61, 0x74, 0x65,
	0x77, 0x61, 0x79, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x52, 0x0f, 0x67, 0x61, 0x74,
	0x65, 0x77, 0x61, 0x79, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x16, 0x0a, 0x06,
	0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x64, 0x65,
	0x6c, 0x65, 0x74, 0x65, 0x22, 0xc4, 0x01, 0x0a, 0x0f, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79,
	0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x65, 0x6e, 0x64, 0x70,
	0x6f, 0x69, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x65,
	0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x1f, 0x0a, 0x04, 0x61, 0x75, 0x74,
	0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x41, 0x75, 0x74, 0x68, 0x52, 0x04, 0x61, 0x75, 0x74, 0x68, 0x12, 0x35, 0x0a, 0x0c, 0x75, 0x73,
	0x65, 0x72, 0x5f, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x12, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x41, 0x63, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x52, 0x0b, 0x75, 0x73, 0x65, 0x72, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x12, 0x38, 0x0a, 0x0d, 0x72, 0x61, 0x74, 0x65, 0x5f, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x69,
	0x6e, 0x67, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x52, 0x61, 0x74, 0x65, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x69, 0x6e, 0x67, 0x52, 0x0c, 0x72,
	0x61, 0x74, 0x65, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x69, 0x6e, 0x67, 0x22, 0xc8, 0x01, 0x0a, 0x04,
	0x41, 0x75, 0x74, 0x68, 0x12, 0x21, 0x0a, 0x0c, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x5f,
	0x61, 0x75, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x72, 0x65, 0x71, 0x75,
	0x69, 0x72, 0x65, 0x41, 0x75, 0x74, 0x68, 0x12, 0x4b, 0x0a, 0x10, 0x61, 0x75, 0x74, 0x68, 0x6f,
	0x72, 0x69, 0x7a, 0x65, 0x64, 0x5f, 0x75, 0x73, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x20, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x2e, 0x41,
	0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x7a, 0x65, 0x64, 0x55, 0x73, 0x65, 0x72, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x52, 0x0f, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x7a, 0x65, 0x64, 0x55,
	0x73, 0x65, 0x72, 0x73, 0x1a, 0x50, 0x0a, 0x14, 0x41, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x7a,
	0x65, 0x64, 0x55, 0x73, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x22,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x52, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x49, 0x0a, 0x0b, 0x55, 0x73, 0x65, 0x72, 0x41, 0x63,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x61, 0x63, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x49, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x70, 0x6c, 0x61, 0x6e, 0x5f, 0x74, 0x79, 0x70,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x6c, 0x61, 0x6e, 0x54, 0x79, 0x70,
	0x65, 0x22, 0xb0, 0x01, 0x0a, 0x0c, 0x52, 0x61, 0x74, 0x65, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x69,
	0x6e, 0x67, 0x12, 0x29, 0x0a, 0x10, 0x74, 0x68, 0x72, 0x6f, 0x75, 0x67, 0x68, 0x70, 0x75, 0x74,
	0x5f, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0f, 0x74, 0x68,
	0x72, 0x6f, 0x75, 0x67, 0x68, 0x70, 0x75, 0x74, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x12, 0x25, 0x0a,
	0x0e, 0x63, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x5f, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0d, 0x63, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x4c,
	0x69, 0x6d, 0x69, 0x74, 0x12, 0x4e, 0x0a, 0x15, 0x63, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79,
	0x5f, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x5f, 0x70, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x61, 0x70, 0x61,
	0x63, 0x69, 0x74, 0x79, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x50, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x52,
	0x13, 0x63, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x50, 0x65,
	0x72, 0x69, 0x6f, 0x64, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x2a, 0xa2, 0x01,
	0x0a, 0x13, 0x43, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x50,
	0x65, 0x72, 0x69, 0x6f, 0x64, 0x12, 0x25, 0x0a, 0x21, 0x43, 0x41, 0x50, 0x41, 0x43, 0x49, 0x54,
	0x59, 0x5f, 0x4c, 0x49, 0x4d, 0x49, 0x54, 0x5f, 0x50, 0x45, 0x52, 0x49, 0x4f, 0x44, 0x5f, 0x55,
	0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x1f, 0x0a, 0x1b,
	0x43, 0x41, 0x50, 0x41, 0x43, 0x49, 0x54, 0x59, 0x5f, 0x4c, 0x49, 0x4d, 0x49, 0x54, 0x5f, 0x50,
	0x45, 0x52, 0x49, 0x4f, 0x44, 0x5f, 0x44, 0x41, 0x49, 0x4c, 0x59, 0x10, 0x01, 0x12, 0x20, 0x0a,
	0x1c, 0x43, 0x41, 0x50, 0x41, 0x43, 0x49, 0x54, 0x59, 0x5f, 0x4c, 0x49, 0x4d, 0x49, 0x54, 0x5f,
	0x50, 0x45, 0x52, 0x49, 0x4f, 0x44, 0x5f, 0x57, 0x45, 0x45, 0x4b, 0x4c, 0x59, 0x10, 0x02, 0x12,
	0x21, 0x0a, 0x1d, 0x43, 0x41, 0x50, 0x41, 0x43, 0x49, 0x54, 0x59, 0x5f, 0x4c, 0x49, 0x4d, 0x49,
	0x54, 0x5f, 0x50, 0x45, 0x52, 0x49, 0x4f, 0x44, 0x5f, 0x4d, 0x4f, 0x4e, 0x54, 0x48, 0x4c, 0x59,
	0x10, 0x03, 0x32, 0xa9, 0x01, 0x0a, 0x10, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x45, 0x6e,
	0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x73, 0x12, 0x44, 0x0a, 0x11, 0x46, 0x65, 0x74, 0x63, 0x68,
	0x41, 0x75, 0x74, 0x68, 0x44, 0x61, 0x74, 0x61, 0x53, 0x79, 0x6e, 0x63, 0x12, 0x16, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41, 0x75, 0x74,
	0x68, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x4f, 0x0a,
	0x15, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x41, 0x75, 0x74, 0x68, 0x44, 0x61, 0x74, 0x61, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x73, 0x12, 0x1d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41,
	0x75, 0x74, 0x68, 0x44, 0x61, 0x74, 0x61, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41, 0x75,
	0x74, 0x68, 0x44, 0x61, 0x74, 0x61, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x30, 0x01, 0x42, 0x09,
	0x5a, 0x07, 0x2e, 0x3b, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescOnce sync.Once
	file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescData = file_envoy_auth_server_proto_gateway_endpoint_proto_rawDesc
)

func file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescGZIP() []byte {
	file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescOnce.Do(func() {
		file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescData = protoimpl.X.CompressGZIP(file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescData)
	})
	return file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescData
}

var file_envoy_auth_server_proto_gateway_endpoint_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_envoy_auth_server_proto_gateway_endpoint_proto_goTypes = []any{
	(CapacityLimitPeriod)(0),       // 0: proto.CapacityLimitPeriod
	(*AuthDataRequest)(nil),        // 1: proto.AuthDataRequest
	(*AuthDataResponse)(nil),       // 2: proto.AuthDataResponse
	(*AuthDataUpdatesRequest)(nil), // 3: proto.AuthDataUpdatesRequest
	(*AuthDataUpdate)(nil),         // 4: proto.AuthDataUpdate
	(*GatewayEndpoint)(nil),        // 5: proto.GatewayEndpoint
	(*Auth)(nil),                   // 6: proto.Auth
	(*UserAccount)(nil),            // 7: proto.UserAccount
	(*RateLimiting)(nil),           // 8: proto.RateLimiting
	(*Empty)(nil),                  // 9: proto.Empty
	nil,                            // 10: proto.AuthDataResponse.EndpointsEntry
	nil,                            // 11: proto.Auth.AuthorizedUsersEntry
}
var file_envoy_auth_server_proto_gateway_endpoint_proto_depIdxs = []int32{
	10, // 0: proto.AuthDataResponse.endpoints:type_name -> proto.AuthDataResponse.EndpointsEntry
	5,  // 1: proto.AuthDataUpdate.gateway_endpoint:type_name -> proto.GatewayEndpoint
	6,  // 2: proto.GatewayEndpoint.auth:type_name -> proto.Auth
	7,  // 3: proto.GatewayEndpoint.user_account:type_name -> proto.UserAccount
	8,  // 4: proto.GatewayEndpoint.rate_limiting:type_name -> proto.RateLimiting
	11, // 5: proto.Auth.authorized_users:type_name -> proto.Auth.AuthorizedUsersEntry
	0,  // 6: proto.RateLimiting.capacity_limit_period:type_name -> proto.CapacityLimitPeriod
	5,  // 7: proto.AuthDataResponse.EndpointsEntry.value:type_name -> proto.GatewayEndpoint
	9,  // 8: proto.Auth.AuthorizedUsersEntry.value:type_name -> proto.Empty
	1,  // 9: proto.GatewayEndpoints.FetchAuthDataSync:input_type -> proto.AuthDataRequest
	3,  // 10: proto.GatewayEndpoints.StreamAuthDataUpdates:input_type -> proto.AuthDataUpdatesRequest
	2,  // 11: proto.GatewayEndpoints.FetchAuthDataSync:output_type -> proto.AuthDataResponse
	4,  // 12: proto.GatewayEndpoints.StreamAuthDataUpdates:output_type -> proto.AuthDataUpdate
	11, // [11:13] is the sub-list for method output_type
	9,  // [9:11] is the sub-list for method input_type
	9,  // [9:9] is the sub-list for extension type_name
	9,  // [9:9] is the sub-list for extension extendee
	0,  // [0:9] is the sub-list for field type_name
}

func init() { file_envoy_auth_server_proto_gateway_endpoint_proto_init() }
func file_envoy_auth_server_proto_gateway_endpoint_proto_init() {
	if File_envoy_auth_server_proto_gateway_endpoint_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_envoy_auth_server_proto_gateway_endpoint_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_envoy_auth_server_proto_gateway_endpoint_proto_goTypes,
		DependencyIndexes: file_envoy_auth_server_proto_gateway_endpoint_proto_depIdxs,
		EnumInfos:         file_envoy_auth_server_proto_gateway_endpoint_proto_enumTypes,
		MessageInfos:      file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes,
	}.Build()
	File_envoy_auth_server_proto_gateway_endpoint_proto = out.File
	file_envoy_auth_server_proto_gateway_endpoint_proto_rawDesc = nil
	file_envoy_auth_server_proto_gateway_endpoint_proto_goTypes = nil
	file_envoy_auth_server_proto_gateway_endpoint_proto_depIdxs = nil
}
