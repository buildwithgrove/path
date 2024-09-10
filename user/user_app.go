package user

type UserAppID string // A unique identifier for a user's application.
type AccountID string // A unique identifier for a user's account.
type PlanType string  // The plan type for a user's application.
// The allowlist type for a user's application.
// One of: contracts, methods, origins, services, user_agent
type AllowlistType string

// TODO_DISCUSS: which allowlist types are needed in PATH?
const (
	// AllowlistTypeContracts allows the UserApp to specify which blockchain contracts are allowed to be passed in service requests.
	AllowlistTypeContracts AllowlistType = "contracts"
	// AllowlistTypeMethods allows the UserApp to specify which methods are allowed to be used for service requests.
	AllowlistTypeMethods AllowlistType = "methods"
	// AllowlistTypeOrigins allows the UserApp to specify which request origins are allowed to make service requests.
	AllowlistTypeOrigins AllowlistType = "origins"
	// AllowlistTypeServices allows the UserApp to specify which services (eg. blockchains relay IDs) are allowed to make service requests.
	AllowlistTypeServices AllowlistType = "services"
	// AllowlistTypeUserAgents allows the UserApp to specify which user agents are allowed to make service requests.
	AllowlistTypeUserAgents AllowlistType = "user_agents"
)

// A UserApp represents a user's application, which has two primary functions:
// 1. Identifying which user is making a service request, to enabled metering, billing, and rate limiting
// 2. Allowing configuration of user-specific settings, such as secret key authentication, allowlists, etc.
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
	// The secret key for UserApp, used for authenticating requests (if required).
	SecretKey string
	// Whether the UserApp requires a secret key for authentication.
	SecretKeyRequired bool
	// The throughput rate limit for UserApp, in requests per second.
	RateLimitThroughput int
	// The allowlists for the UserApp, which allow the UserApp to specify certain security settings,
	// for example, which contracts, methods, origins, services, or user agents are allowed to make
	// service requests through the Gateway with the UserApp.
	Allowlists map[AllowlistType]map[string]struct{}
}

func (a *UserApp) IsContractAllowed(contractID string) bool {
	allowlistValues, ok := a.Allowlists[AllowlistTypeContracts]
	if !ok {
		return false
	}
	_, ok = allowlistValues[contractID]
	return ok
}

func (a *UserApp) IsMethodAllowed(method string) bool {
	allowlistValues, ok := a.Allowlists[AllowlistTypeMethods]
	if !ok {
		return false
	}
	_, ok = allowlistValues[method]
	return ok
}

func (a *UserApp) IsOriginAllowed(origin string) bool {
	allowlistValues, ok := a.Allowlists[AllowlistTypeOrigins]
	if !ok {
		return false
	}
	_, ok = allowlistValues[origin]
	return ok
}

func (a *UserApp) IsServiceAllowed(service string) bool {
	allowlistValues, ok := a.Allowlists[AllowlistTypeServices]
	if !ok {
		return false
	}
	_, ok = allowlistValues[service]
	return ok
}

func (a *UserApp) IsUserAgentAllowed(userAgent string) bool {
	allowlistValues, ok := a.Allowlists[AllowlistTypeUserAgents]
	if !ok {
		return false
	}
	_, ok = allowlistValues[userAgent]
	return ok
}
