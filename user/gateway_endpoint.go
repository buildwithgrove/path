package user

// A unique identifier for a user application. It must be passed as the last path segment of the service request URL.
type EndpointID string

// A unique identifier for a user's account. Used for metering and billing purposes.
type AccountID string

// The pricing plan type for a user app's Account. Used for metering and billing purposes.
type PlanType string

// A GatewayEndpoint represents a user's application endpoint, which has two primary functions:
// 1. Identifying which user is making a service request, to enable metering, billing, and rate limiting
// 2. Allowing configuration of user-specific settings, such as API key authentication, etc.
//
// A GatewayEndpoint is associated to a single UserAccount. An Account can have multiple GatewayEndpoints.
// Settings related to service requests, such as enforcing API key authentication, are configured per GatewayEndpoint.
type GatewayEndpoint struct {
	// The unique identifier for a user's application endpoint, which must be passed
	// as the last path segment of the service request URL. eg `/v1/{endpoint_id}`
	// This is used to identify the GatewayEndpoint when making a service request.
	EndpointID EndpointID
	// The authentication settings for the GatewayEndpoint.
	Auth Auth
	// The UserAccount associated with the GatewayEndpoint.
	UserAccount UserAccount
	// The rate limiting settings for the GatewayEndpoint.
	RateLimiting RateLimiting
}

// The authentication settings for a GatewayEndpoint.
type Auth struct {
	// The API key for GatewayEndpoint. If APIKeyRequired is true, the API key
	// must be passed as the `Authentication` header in service requests.
	APIKey string
	// Whether the GatewayEndpoint requires an API key for authentication.
	// If not set, the GatewayEndpoint does not require an API key for authentication.
	APIKeyRequired bool
}

// A UserAccount represents a user's account, which holds the PlanType and groups GatewayEndpoints.
type UserAccount struct {
	// The unique identifier for a UserAccount, used for metering & billing purposes.
	AccountID AccountID
	// The plan type for a UserAccount, which identifies the pricing plan for the Account.
	PlanType PlanType
}

// The rate limiting settings for a GatewayEndpoint.
type RateLimiting struct {
	// ThroughputLimit refers to rate limiting per-second.
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
