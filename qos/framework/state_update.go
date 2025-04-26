package framework

// TODO_FUTURE(@adshmh): Support deleting StateParameters by adding a `ToDelete` field
type StateParameterUpdateSet struct {
	Updates map[string]*StateParameter
}
