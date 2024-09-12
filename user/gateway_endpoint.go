package user

// A unique identifier for a user account's endpoint. It must be passed as the last path segment of the service request URL.
type EndpointID string

// A unique identifier for a user's account, which identifies the account that owns the GatewayEndpoint.
type AccountID string

// The pricing plan type for a user app's Account. Used for metering and billing purposes.
type PlanType string

// A GatewayEndpoint represents a user account's endpoint, which has two primary functions:
// 1. Identifying which endpoint is being used to make a service request.
// 2. Allowing configuration of endpoint-specific settings, such as API key authentication, etc.
//
// A GatewayEndpoint is associated to a single UserAccount. A UserAccount can have multiple GatewayEndpoints.
// Settings related to service requests, such as enforcing API key authentication, are configured per GatewayEndpoint.
type GatewayEndpoint struct {
	// The unique identifier for a user account's endpoint, which must be passed
	// as the last path segment of the service request URL. eg `/v1/{endpoint_id}`
	// This is used to identify the GatewayEndpoint when making a service request.
	EndpointID EndpointID
	// The authentication settings for the GatewayEndpoint.
	Auth Auth
	// The UserAccount that the GatewayEndpoint belongs to, including the PlanType.
	UserAccount UserAccount
	// The rate limiting settings for the GatewayEndpoint, which includes both
	// the throughput (TPS) limit and the capacity (longer period) limit.
	RateLimiting RateLimiting
}

// The authentication settings for a GatewayEndpoint.
type Auth struct {
	// The API key for GatewayEndpoint. If APIKeyRequired is true, the API key
	// must be passed as the `Authorization` HTTP header in service requests.
	APIKey string
	// Whether the GatewayEndpoint requires an API key for authentication.
	// If not set, the GatewayEndpoint does not require an API key to be passed
	// as the `Authorization` header and all requests using the endpoint will be allowed.
	APIKeyRequired bool
}

// A UserAccount contains the PlanType and may have multiple GatewayEndpoints.
type UserAccount struct {
	// The unique identifier for a UserAccount.
	AccountID AccountID
	// The plan type for a UserAccount, which identifies the pricing plan for the Account.
	PlanType PlanType
}

// The rate limiting settings for a GatewayEndpoint.
type RateLimiting struct {
	// ThroughputLimit refers to rate limiting per-second (TPS).
	ThroughputLimit int
	// CapacityLimit refers to rate limiting over longer periods, such as a day, week or month.
	CapacityLimit int
	// The period over which the CapacityLimit is enforced. One of `daily`, `weekly` or `monthly`.
	CapacityLimitPeriod CapacityLimitPeriod
}

// CapacityLimitPeriod is the period over which the capacity limit is enforced.
// For example: CapacityLimit=`100,000` and CapacityLimitPeriod=`daily`
// enforces a rate limit of 100,000 requests per day.
type CapacityLimitPeriod string

const (
	CapacityLimitPeriodDaily   CapacityLimitPeriod = "daily"
	CapacityLimitPeriodWeekly  CapacityLimitPeriod = "weekly"
	CapacityLimitPeriodMonthly CapacityLimitPeriod = "monthly"
)

func (e *GatewayEndpoint) GetAuth() (string, bool) {
	return e.Auth.APIKey, e.Auth.APIKeyRequired
}

func (e *GatewayEndpoint) GetThroughputLimit() int {
	return e.RateLimiting.ThroughputLimit
}

func (e *GatewayEndpoint) GetCapacityLimit() (int, CapacityLimitPeriod) {
	return e.RateLimiting.CapacityLimit, e.RateLimiting.CapacityLimitPeriod
}
