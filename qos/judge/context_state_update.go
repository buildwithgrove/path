package judge

// StateUpdateContext provides context and helper methods for updating service state.
type StateUpdateContext struct {
	// Current service state (read-only copy)
	// Provides direct access to the read-only state
	*ServiceState

	// custom state updater function to be called from the context.
	stateUpdater StateUpdater

	// updated endpoints on which the state update should be based.
	updatedEndpoints []*Endpoint

	// tracks the set of params set for update through the context.
	paramsToUpdate *StateParameterUpdateSet
}

func (ctx *StateUpdateContext) updateFromEndpoints(updatedEndpoints []*Endpoint) error {
	// get the list of params to update by calling the custom state updater.
	paramsToUpdate := ctx.stateUpdater(ctx)

	// Update the state parameters through the service state.
	return ctx.ServiceState.updateParameters(paramsToUpdate)
}

func (ctx *StateUpdateContext) GetUpdatedEndpoints() []*Endpoint {
	return ctx.updatedEndpoints
}

func (ctx *StateUpdateContext) SetIntParam(paramName string, value int) {
	param := &StateParameter{
		intValue: &value,
	}

	ctx.paramsToUpdate.Set(paramName, param)
}

func (ctx *StateUpdateContext) SetStrParam(paramName, value string) {
	param := &StateParameter{
		strValue: &value,
	}

	ctx.paramsToUpdate.Set(paramName, param)
}

// TODO_IN_THIS_PR: copy the map to prevent reference leaks
func (ctx *StateUpdateContext) SetConsensusParam(paramName string, consensusValues map[string]int) {
	param := &StateParameter{
		consensusValues: consensusValues,
	}

	ctx.paramsToUpdate.Set(paramName, param)
}

// Return the set of updated state parameters.
func (ctx *StateUpdateContext) BuildStateParameterUpdateSet() *StateParameterUpdateSet {
	return ctx.paramsToUpdate
}
