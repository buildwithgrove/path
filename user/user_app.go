package user

// A unique identifier for a user application. It must be passed as the last path segment of the service request URL.
type UserAppID string

// A unique identifier for a user's account. Used for metering and billing purposes.
type AccountID string

// The pricing plan type for a user app's Account. Used for metering and billing purposes.
type PlanType string

// A UserApp represents a user's application, which has two primary functions:
// 1. Identifying which user is making a service request, to enabled metering, billing, and rate limiting
// 2. Allowing configuration of user-specific settings, such as secret key authentication, etc.
//
// A UserApp is associated to a single Account. An Account can have multiple UserApps and one PlanType.
// Settings related to service requests, such as enforcing secret key authentication, are configured per UserApp.
// Pricing plan and team management settings are configured per Account.
type UserApp struct {

	// The unique identifier for a user's application, which must be passed
	// as the last path segment of the service request URL. eg `/v1/{user_app_id}`
	// This is used to identify the UserApp when making a service request.
	ID UserAppID
	// The unique identifier for a UserApp's Account, used for metering & billing purposes.
	AccountID AccountID
	// The plan type for a UserApp's Account, which identifies the pricing plan for the Account.
	PlanType PlanType
	// The throughput rate limit for UserApp, in requests per second.
	RateLimitThroughput int

	// The secret key for UserApp. If SecretKeyRequired is true, the secret key
	// must be passed as the `Authentication` header in service requests.
	SecretKey string
	// Whether the UserApp requires a secret key for authentication.
	// If not set, the UserApp does not require a secret key for authentication.
	SecretKeyRequired bool
}
