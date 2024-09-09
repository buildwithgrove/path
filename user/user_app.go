package user

type UserAppID string
type AccountID string
type AllowlistType string

const (
	AllowlistTypeContracts  AllowlistType = "contracts"
	AllowlistTypeMethods    AllowlistType = "methods"
	AllowlistTypeOrigins    AllowlistType = "origins"
	AllowlistTypeServices   AllowlistType = "services"
	AllowlistTypeUserAgents AllowlistType = "user_agents"
)

type UserApp struct {
	ID                  UserAppID
	AccountID           AccountID
	PlanType            string
	SecretKey           string
	SecretKeyRequired   bool
	RateLimitThroughput int
	Allowlists          map[AllowlistType]map[string]struct{}
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
