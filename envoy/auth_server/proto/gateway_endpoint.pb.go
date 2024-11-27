// To generate the Go code from this proto file, run:
//  `make proto_generate`
// which runs:
//  `protoc --go_out=./envoy/auth_server/proto --go-grpc_out=./envoy/auth_server/proto envoy/auth_server/proto/gateway_endpoint.proto`

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v5.28.2
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

type Auth_AuthType int32

const (
	Auth_AUTH_TYPE_UNSPECIFIED Auth_AuthType = 0
	Auth_AUTH_TYPE_API_KEY     Auth_AuthType = 1
	Auth_AUTH_TYPE_JWT         Auth_AuthType = 2
)

// Enum value maps for Auth_AuthType.
var (
	Auth_AuthType_name = map[int32]string{
		0: "AUTH_TYPE_UNSPECIFIED",
		1: "AUTH_TYPE_API_KEY",
		2: "AUTH_TYPE_JWT",
	}
	Auth_AuthType_value = map[string]int32{
		"AUTH_TYPE_UNSPECIFIED": 0,
		"AUTH_TYPE_API_KEY":     1,
		"AUTH_TYPE_JWT":         2,
	}
)

func (x Auth_AuthType) Enum() *Auth_AuthType {
	p := new(Auth_AuthType)
	*p = x
	return p
}

func (x Auth_AuthType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Auth_AuthType) Descriptor() protoreflect.EnumDescriptor {
	return file_envoy_auth_server_proto_gateway_endpoint_proto_enumTypes[1].Descriptor()
}

func (Auth_AuthType) Type() protoreflect.EnumType {
	return &file_envoy_auth_server_proto_gateway_endpoint_proto_enumTypes[1]
}

func (x Auth_AuthType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Auth_AuthType.Descriptor instead.
func (Auth_AuthType) EnumDescriptor() ([]byte, []int) {
	return file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescGZIP(), []int{5, 0}
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

	// A unique identifier for a user account's GatewayEndpoint.
	// This is used to identify the GatewayEndpoint when making a service request.
	// It must be passed as the last path segment of the service request URL.
	// For example: POST https://api.path.xyz/v1/{gateway_endpoint_id}
	EndpointId string `protobuf:"bytes,1,opt,name=endpoint_id,json=endpointId,proto3" json:"endpoint_id,omitempty"`
	// The GatewayEndpoint to upsert to the database.
	// If delete is true, this field should be empty and the associated endpoint_id will be deleted.
	GatewayEndpoint *GatewayEndpoint `protobuf:"bytes,2,opt,name=gateway_endpoint,json=gatewayEndpoint,proto3" json:"gateway_endpoint,omitempty"`
	// Indicates whether the GatewayEndpoint should be deleted.
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
//  1. Identifying which endpoint is being used to make a service request.
//  2. Allowing configuration of endpoint-specific settings, such as API key authorization, etc.
//
// A GatewayEndpoint is associated to a single UserAccount.
// A single UserAccount can have multiple GatewayEndpoints.
// Settings related to service requests, such as enforcing API key authorization, are configured per GatewayEndpoint.
type GatewayEndpoint struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The unique identifier for a UserAccount's endpoint.
	// It must be passed as the last path segment of the service request URL; eg `/v1/{endpoint_id}`.
	// This is used to identify the GatewayEndpoint when making a service request.
	EndpointId string `protobuf:"bytes,1,opt,name=endpoint_id,json=endpointId,proto3" json:"endpoint_id,omitempty"`
	// The authorization settings for the GatewayEndpoint.
	Auth *Auth `protobuf:"bytes,2,opt,name=auth,proto3" json:"auth,omitempty"`
	// The rate limiting settings for the GatewayEndpoint.
	// This includes both throughput (TPS) limit and the capacity (longer period) limit.
	RateLimiting *RateLimiting `protobuf:"bytes,3,opt,name=rate_limiting,json=rateLimiting,proto3" json:"rate_limiting,omitempty"`
	// Optional metadata for the GatewayEndpoint, which can be set to any additional information.
	Metadata *Metadata `protobuf:"bytes,4,opt,name=metadata,proto3" json:"metadata,omitempty"`
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

func (x *GatewayEndpoint) GetRateLimiting() *RateLimiting {
	if x != nil {
		return x.RateLimiting
	}
	return nil
}

func (x *GatewayEndpoint) GetMetadata() *Metadata {
	if x != nil {
		return x.Metadata
	}
	return nil
}

// The authorization settings for a GatewayEndpoint.
type Auth struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The type of authentication being used.
	AuthType Auth_AuthType `protobuf:"varint,1,opt,name=auth_type,json=authType,proto3,enum=proto.Auth_AuthType" json:"auth_type,omitempty"`
	// Types that are assignable to AuthTypeDetails:
	//
	//	*Auth_NoAuth
	//	*Auth_StaticApiKey
	//	*Auth_Jwt
	AuthTypeDetails isAuth_AuthTypeDetails `protobuf_oneof:"auth_type_details"`
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

func (x *Auth) GetAuthType() Auth_AuthType {
	if x != nil {
		return x.AuthType
	}
	return Auth_AUTH_TYPE_UNSPECIFIED
}

func (m *Auth) GetAuthTypeDetails() isAuth_AuthTypeDetails {
	if m != nil {
		return m.AuthTypeDetails
	}
	return nil
}

func (x *Auth) GetNoAuth() *Empty {
	if x, ok := x.GetAuthTypeDetails().(*Auth_NoAuth); ok {
		return x.NoAuth
	}
	return nil
}

func (x *Auth) GetStaticApiKey() *StaticAPIKey {
	if x, ok := x.GetAuthTypeDetails().(*Auth_StaticApiKey); ok {
		return x.StaticApiKey
	}
	return nil
}

func (x *Auth) GetJwt() *JWT {
	if x, ok := x.GetAuthTypeDetails().(*Auth_Jwt); ok {
		return x.Jwt
	}
	return nil
}

type isAuth_AuthTypeDetails interface {
	isAuth_AuthTypeDetails()
}

type Auth_NoAuth struct {
	NoAuth *Empty `protobuf:"bytes,2,opt,name=no_auth,json=noAuth,proto3,oneof"`
}

type Auth_StaticApiKey struct {
	// The API key authorization settings for the GatewayEndpoint.
	StaticApiKey *StaticAPIKey `protobuf:"bytes,3,opt,name=static_api_key,json=staticApiKey,proto3,oneof"`
}

type Auth_Jwt struct {
	// The JWT authorization settings for the GatewayEndpoint.
	Jwt *JWT `protobuf:"bytes,4,opt,name=jwt,proto3,oneof"`
}

func (*Auth_NoAuth) isAuth_AuthTypeDetails() {}

func (*Auth_StaticApiKey) isAuth_AuthTypeDetails() {}

func (*Auth_Jwt) isAuth_AuthTypeDetails() {}

type StaticAPIKey struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The API key to use for authentication.
	ApiKey string `protobuf:"bytes,1,opt,name=api_key,json=apiKey,proto3" json:"api_key,omitempty"`
}

func (x *StaticAPIKey) Reset() {
	*x = StaticAPIKey{}
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StaticAPIKey) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StaticAPIKey) ProtoMessage() {}

func (x *StaticAPIKey) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use StaticAPIKey.ProtoReflect.Descriptor instead.
func (*StaticAPIKey) Descriptor() ([]byte, []int) {
	return file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescGZIP(), []int{6}
}

func (x *StaticAPIKey) GetApiKey() string {
	if x != nil {
		return x.ApiKey
	}
	return ""
}

// JWT is the JSON Web Token authorization settings for a GatewayEndpoint.
type JWT struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A map of ProviderUserIDs authorized to access this UserAccount's GatewayEndpoints.
	AuthorizedUsers map[string]*Empty `protobuf:"bytes,2,rep,name=authorized_users,json=authorizedUsers,proto3" json:"authorized_users,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *JWT) Reset() {
	*x = JWT{}
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *JWT) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JWT) ProtoMessage() {}

func (x *JWT) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use JWT.ProtoReflect.Descriptor instead.
func (*JWT) Descriptor() ([]byte, []int) {
	return file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescGZIP(), []int{7}
}

func (x *JWT) GetAuthorizedUsers() map[string]*Empty {
	if x != nil {
		return x.AuthorizedUsers
	}
	return nil
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
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RateLimiting) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RateLimiting) ProtoMessage() {}

func (x *RateLimiting) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use RateLimiting.ProtoReflect.Descriptor instead.
func (*RateLimiting) Descriptor() ([]byte, []int) {
	return file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescGZIP(), []int{8}
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

// Metadata is the metadata for a GatewayEndpoint, defined as fields that are not
// required for perform technical tasks related to the Envoy Auth implementation.
//
// These fields are intended to be used for billing, metrics, and other purposes.
// All fields are optional and may be left blank if not applicable.
type Metadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AccountId   string `protobuf:"bytes,1,opt,name=account_id,json=accountId,proto3" json:"account_id,omitempty"` // Unique identifier for the user's account
	UserId      string `protobuf:"bytes,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`          // Identifier for a specific user within the system
	PlanType    string `protobuf:"bytes,3,opt,name=plan_type,json=planType,proto3" json:"plan_type,omitempty"`    // Subscription or account plan type (e.g., "Free", "Pro", "Enterprise")
	Email       string `protobuf:"bytes,4,opt,name=email,proto3" json:"email,omitempty"`                          // The user's email address
	Environment string `protobuf:"bytes,5,opt,name=environment,proto3" json:"environment,omitempty"`              // The environment the GatewayEndpoint is in (e.g., "development", "staging", "production")
}

func (x *Metadata) Reset() {
	*x = Metadata{}
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Metadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Metadata) ProtoMessage() {}

func (x *Metadata) ProtoReflect() protoreflect.Message {
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Metadata.ProtoReflect.Descriptor instead.
func (*Metadata) Descriptor() ([]byte, []int) {
	return file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescGZIP(), []int{9}
}

func (x *Metadata) GetAccountId() string {
	if x != nil {
		return x.AccountId
	}
	return ""
}

func (x *Metadata) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *Metadata) GetPlanType() string {
	if x != nil {
		return x.PlanType
	}
	return ""
}

func (x *Metadata) GetEmail() string {
	if x != nil {
		return x.Email
	}
	return ""
}

func (x *Metadata) GetEnvironment() string {
	if x != nil {
		return x.Environment
	}
	return ""
}

// An Empty message is used to indicate that a field is not set.
type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Empty) Reset() {
	*x = Empty{}
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[10]
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
	return file_envoy_auth_server_proto_gateway_endpoint_proto_rawDescGZIP(), []int{10}
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
	0x6c, 0x65, 0x74, 0x65, 0x22, 0xba, 0x01, 0x0a, 0x0f, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79,
	0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x65, 0x6e, 0x64, 0x70,
	0x6f, 0x69, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x65,
	0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x1f, 0x0a, 0x04, 0x61, 0x75, 0x74,
	0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x41, 0x75, 0x74, 0x68, 0x52, 0x04, 0x61, 0x75, 0x74, 0x68, 0x12, 0x38, 0x0a, 0x0d, 0x72, 0x61,
	0x74, 0x65, 0x5f, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x69, 0x6e, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x13, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x61, 0x74, 0x65, 0x4c, 0x69,
	0x6d, 0x69, 0x74, 0x69, 0x6e, 0x67, 0x52, 0x0c, 0x72, 0x61, 0x74, 0x65, 0x4c, 0x69, 0x6d, 0x69,
	0x74, 0x69, 0x6e, 0x67, 0x12, 0x2b, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0x22, 0xa5, 0x02, 0x0a, 0x04, 0x41, 0x75, 0x74, 0x68, 0x12, 0x31, 0x0a, 0x09, 0x61, 0x75,
	0x74, 0x68, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x54,
	0x79, 0x70, 0x65, 0x52, 0x08, 0x61, 0x75, 0x74, 0x68, 0x54, 0x79, 0x70, 0x65, 0x12, 0x27, 0x0a,
	0x07, 0x6e, 0x6f, 0x5f, 0x61, 0x75, 0x74, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x48, 0x00, 0x52, 0x06,
	0x6e, 0x6f, 0x41, 0x75, 0x74, 0x68, 0x12, 0x3b, 0x0a, 0x0e, 0x73, 0x74, 0x61, 0x74, 0x69, 0x63,
	0x5f, 0x61, 0x70, 0x69, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x69, 0x63, 0x41, 0x50, 0x49,
	0x4b, 0x65, 0x79, 0x48, 0x00, 0x52, 0x0c, 0x73, 0x74, 0x61, 0x74, 0x69, 0x63, 0x41, 0x70, 0x69,
	0x4b, 0x65, 0x79, 0x12, 0x1e, 0x0a, 0x03, 0x6a, 0x77, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4a, 0x57, 0x54, 0x48, 0x00, 0x52, 0x03,
	0x6a, 0x77, 0x74, 0x22, 0x4f, 0x0a, 0x08, 0x41, 0x75, 0x74, 0x68, 0x54, 0x79, 0x70, 0x65, 0x12,
	0x19, 0x0a, 0x15, 0x41, 0x55, 0x54, 0x48, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x55, 0x4e, 0x53,
	0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x15, 0x0a, 0x11, 0x41, 0x55,
	0x54, 0x48, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x41, 0x50, 0x49, 0x5f, 0x4b, 0x45, 0x59, 0x10,
	0x01, 0x12, 0x11, 0x0a, 0x0d, 0x41, 0x55, 0x54, 0x48, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x4a,
	0x57, 0x54, 0x10, 0x02, 0x42, 0x13, 0x0a, 0x11, 0x61, 0x75, 0x74, 0x68, 0x5f, 0x74, 0x79, 0x70,
	0x65, 0x5f, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x22, 0x27, 0x0a, 0x0c, 0x53, 0x74, 0x61,
	0x74, 0x69, 0x63, 0x41, 0x50, 0x49, 0x4b, 0x65, 0x79, 0x12, 0x17, 0x0a, 0x07, 0x61, 0x70, 0x69,
	0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x70, 0x69, 0x4b,
	0x65, 0x79, 0x22, 0xa3, 0x01, 0x0a, 0x03, 0x4a, 0x57, 0x54, 0x12, 0x4a, 0x0a, 0x10, 0x61, 0x75,
	0x74, 0x68, 0x6f, 0x72, 0x69, 0x7a, 0x65, 0x64, 0x5f, 0x75, 0x73, 0x65, 0x72, 0x73, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4a, 0x57, 0x54,
	0x2e, 0x41, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x7a, 0x65, 0x64, 0x55, 0x73, 0x65, 0x72, 0x73,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0f, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x7a, 0x65,
	0x64, 0x55, 0x73, 0x65, 0x72, 0x73, 0x1a, 0x50, 0x0a, 0x14, 0x41, 0x75, 0x74, 0x68, 0x6f, 0x72,
	0x69, 0x7a, 0x65, 0x64, 0x55, 0x73, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10,
	0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79,
	0x12, 0x22, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x0c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xb0, 0x01, 0x0a, 0x0c, 0x52, 0x61, 0x74,
	0x65, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x69, 0x6e, 0x67, 0x12, 0x29, 0x0a, 0x10, 0x74, 0x68, 0x72,
	0x6f, 0x75, 0x67, 0x68, 0x70, 0x75, 0x74, 0x5f, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x0f, 0x74, 0x68, 0x72, 0x6f, 0x75, 0x67, 0x68, 0x70, 0x75, 0x74, 0x4c,
	0x69, 0x6d, 0x69, 0x74, 0x12, 0x25, 0x0a, 0x0e, 0x63, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79,
	0x5f, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0d, 0x63, 0x61,
	0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x12, 0x4e, 0x0a, 0x15, 0x63,
	0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x5f, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x5f, 0x70, 0x65,
	0x72, 0x69, 0x6f, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1a, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2e, 0x43, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x4c, 0x69, 0x6d, 0x69, 0x74,
	0x50, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x52, 0x13, 0x63, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79,
	0x4c, 0x69, 0x6d, 0x69, 0x74, 0x50, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x22, 0x97, 0x01, 0x0a, 0x08,
	0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x63, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x61, 0x63,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x75, 0x73, 0x65, 0x72, 0x5f,
	0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64,
	0x12, 0x1b, 0x0a, 0x09, 0x70, 0x6c, 0x61, 0x6e, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x6c, 0x61, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x12, 0x14, 0x0a,
	0x05, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x6d,
	0x61, 0x69, 0x6c, 0x12, 0x20, 0x0a, 0x0b, 0x65, 0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d, 0x65,
	0x6e, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x65, 0x6e, 0x76, 0x69, 0x72, 0x6f,
	0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x2a, 0xa2,
	0x01, 0x0a, 0x13, 0x43, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x4c, 0x69, 0x6d, 0x69, 0x74,
	0x50, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x12, 0x25, 0x0a, 0x21, 0x43, 0x41, 0x50, 0x41, 0x43, 0x49,
	0x54, 0x59, 0x5f, 0x4c, 0x49, 0x4d, 0x49, 0x54, 0x5f, 0x50, 0x45, 0x52, 0x49, 0x4f, 0x44, 0x5f,
	0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x1f, 0x0a,
	0x1b, 0x43, 0x41, 0x50, 0x41, 0x43, 0x49, 0x54, 0x59, 0x5f, 0x4c, 0x49, 0x4d, 0x49, 0x54, 0x5f,
	0x50, 0x45, 0x52, 0x49, 0x4f, 0x44, 0x5f, 0x44, 0x41, 0x49, 0x4c, 0x59, 0x10, 0x01, 0x12, 0x20,
	0x0a, 0x1c, 0x43, 0x41, 0x50, 0x41, 0x43, 0x49, 0x54, 0x59, 0x5f, 0x4c, 0x49, 0x4d, 0x49, 0x54,
	0x5f, 0x50, 0x45, 0x52, 0x49, 0x4f, 0x44, 0x5f, 0x57, 0x45, 0x45, 0x4b, 0x4c, 0x59, 0x10, 0x02,
	0x12, 0x21, 0x0a, 0x1d, 0x43, 0x41, 0x50, 0x41, 0x43, 0x49, 0x54, 0x59, 0x5f, 0x4c, 0x49, 0x4d,
	0x49, 0x54, 0x5f, 0x50, 0x45, 0x52, 0x49, 0x4f, 0x44, 0x5f, 0x4d, 0x4f, 0x4e, 0x54, 0x48, 0x4c,
	0x59, 0x10, 0x03, 0x32, 0xa9, 0x01, 0x0a, 0x10, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x45,
	0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x73, 0x12, 0x44, 0x0a, 0x11, 0x46, 0x65, 0x74, 0x63,
	0x68, 0x41, 0x75, 0x74, 0x68, 0x44, 0x61, 0x74, 0x61, 0x53, 0x79, 0x6e, 0x63, 0x12, 0x16, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41, 0x75,
	0x74, 0x68, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x4f,
	0x0a, 0x15, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x41, 0x75, 0x74, 0x68, 0x44, 0x61, 0x74, 0x61,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x73, 0x12, 0x1d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x41, 0x75, 0x74, 0x68, 0x44, 0x61, 0x74, 0x61, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41,
	0x75, 0x74, 0x68, 0x44, 0x61, 0x74, 0x61, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x30, 0x01, 0x42,
	0x09, 0x5a, 0x07, 0x2e, 0x3b, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
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

var file_envoy_auth_server_proto_gateway_endpoint_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes = make([]protoimpl.MessageInfo, 13)
var file_envoy_auth_server_proto_gateway_endpoint_proto_goTypes = []any{
	(CapacityLimitPeriod)(0),       // 0: proto.CapacityLimitPeriod
	(Auth_AuthType)(0),             // 1: proto.Auth.AuthType
	(*AuthDataRequest)(nil),        // 2: proto.AuthDataRequest
	(*AuthDataResponse)(nil),       // 3: proto.AuthDataResponse
	(*AuthDataUpdatesRequest)(nil), // 4: proto.AuthDataUpdatesRequest
	(*AuthDataUpdate)(nil),         // 5: proto.AuthDataUpdate
	(*GatewayEndpoint)(nil),        // 6: proto.GatewayEndpoint
	(*Auth)(nil),                   // 7: proto.Auth
	(*StaticAPIKey)(nil),           // 8: proto.StaticAPIKey
	(*JWT)(nil),                    // 9: proto.JWT
	(*RateLimiting)(nil),           // 10: proto.RateLimiting
	(*Metadata)(nil),               // 11: proto.Metadata
	(*Empty)(nil),                  // 12: proto.Empty
	nil,                            // 13: proto.AuthDataResponse.EndpointsEntry
	nil,                            // 14: proto.JWT.AuthorizedUsersEntry
}
var file_envoy_auth_server_proto_gateway_endpoint_proto_depIdxs = []int32{
	13, // 0: proto.AuthDataResponse.endpoints:type_name -> proto.AuthDataResponse.EndpointsEntry
	6,  // 1: proto.AuthDataUpdate.gateway_endpoint:type_name -> proto.GatewayEndpoint
	7,  // 2: proto.GatewayEndpoint.auth:type_name -> proto.Auth
	10, // 3: proto.GatewayEndpoint.rate_limiting:type_name -> proto.RateLimiting
	11, // 4: proto.GatewayEndpoint.metadata:type_name -> proto.Metadata
	1,  // 5: proto.Auth.auth_type:type_name -> proto.Auth.AuthType
	12, // 6: proto.Auth.no_auth:type_name -> proto.Empty
	8,  // 7: proto.Auth.static_api_key:type_name -> proto.StaticAPIKey
	9,  // 8: proto.Auth.jwt:type_name -> proto.JWT
	14, // 9: proto.JWT.authorized_users:type_name -> proto.JWT.AuthorizedUsersEntry
	0,  // 10: proto.RateLimiting.capacity_limit_period:type_name -> proto.CapacityLimitPeriod
	6,  // 11: proto.AuthDataResponse.EndpointsEntry.value:type_name -> proto.GatewayEndpoint
	12, // 12: proto.JWT.AuthorizedUsersEntry.value:type_name -> proto.Empty
	2,  // 13: proto.GatewayEndpoints.FetchAuthDataSync:input_type -> proto.AuthDataRequest
	4,  // 14: proto.GatewayEndpoints.StreamAuthDataUpdates:input_type -> proto.AuthDataUpdatesRequest
	3,  // 15: proto.GatewayEndpoints.FetchAuthDataSync:output_type -> proto.AuthDataResponse
	5,  // 16: proto.GatewayEndpoints.StreamAuthDataUpdates:output_type -> proto.AuthDataUpdate
	15, // [15:17] is the sub-list for method output_type
	13, // [13:15] is the sub-list for method input_type
	13, // [13:13] is the sub-list for extension type_name
	13, // [13:13] is the sub-list for extension extendee
	0,  // [0:13] is the sub-list for field type_name
}

func init() { file_envoy_auth_server_proto_gateway_endpoint_proto_init() }
func file_envoy_auth_server_proto_gateway_endpoint_proto_init() {
	if File_envoy_auth_server_proto_gateway_endpoint_proto != nil {
		return
	}
	file_envoy_auth_server_proto_gateway_endpoint_proto_msgTypes[5].OneofWrappers = []any{
		(*Auth_NoAuth)(nil),
		(*Auth_StaticApiKey)(nil),
		(*Auth_Jwt)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_envoy_auth_server_proto_gateway_endpoint_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   13,
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
