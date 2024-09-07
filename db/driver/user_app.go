package driver

type UserAppID string
type AccountID string

type UserApp struct {
	ID                UserAppID
	AccountID         AccountID
	PlanType          string
	SecretKey         string
	SecretKeyRequired bool
	ThroughputLimit   int32
	Whitelists        map[string]map[string]struct{}
}

func (a *UserApp) IsWhitelisted(whitelistType, value string) bool {
	whitelistValues, ok := a.Whitelists[whitelistType]
	if !ok {
		return false
	}
	_, ok = whitelistValues[value]
	return ok
}
