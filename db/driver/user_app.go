package driver

type UserAppID string
type AccountID string
type WhitelistType string
type WhitelistValue string

type UserApp struct {
	ID                UserAppID
	AccountID         AccountID
	PlanType          string
	SecretKey         string
	SecretKeyRequired bool
	ThroughputLimit   int32
	Whitelists        map[WhitelistType]map[WhitelistValue]struct{}
}

func (a *UserApp) IsWhitelisted(whitelistType WhitelistType, value WhitelistValue) bool {
	whitelistValues, ok := a.Whitelists[whitelistType]
	if !ok {
		return false
	}
	_, ok = whitelistValues[value]
	return ok
}
