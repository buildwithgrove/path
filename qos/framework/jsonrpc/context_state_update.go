package jsonrpc

import (

)

// StateUpdateContext provides context and helper methods for updating service state.
type StateUpdateContext struct {
	// The result data observations
	Results []*ResultData
	
	// Current service state (read-only copy)
	currentState map[string]string
	
	// New state being built
	newState map[string]string
}

// CopyCurrentState creates a copy of the current state as the starting point.
func (ctx *StateUpdateContext) CopyCurrentState() {
	ctx.newState = make(map[string]string, len(ctx.currentState))
	for k, v := range ctx.currentState {
		ctx.newState[k] = v
	}
}

// SetValue sets a value in the new state.
func (ctx *StateUpdateContext) SetValue(key, value string) *StateUpdateContext {
	ctx.newState[key] = value
	return ctx
}

// DeleteValue removes a value from the new state.
func (ctx *StateUpdateContext) DeleteValue(key string) *StateUpdateContext {
	delete(ctx.newState, key)
	return ctx
}

// GetValue gets a value from the current state.
func (ctx *StateUpdateContext) GetValue(key string) (string, bool) {
	value, exists := ctx.currentState[key]
	return value, exists
}

// GetState returns the new state map.
func (ctx *StateUpdateContext) GetState() map[string]string {
	return ctx.newState
}


