package framework

// TODO_FUTURE(@adshmh): Support deleting StateParameters by adding a `ToDelete` field
type StateParameterUpdateSet struct {
	Updates map[string]*StateParameter
}

func (spu *StateParameterUpdateSet) Set(paramName string, param *StateParameter) {
	if spu.Updates == nil {
		spu.Updates = make(map[string]*StateParameter)
	}

	spu.Updates[paramName] = param
}
