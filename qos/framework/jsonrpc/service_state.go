package jsonrpc

import (
	"sync"
	"time"

	"github.com/buildwithgrove/path/protocol"
	"github.com/pokt-network/poktroll/pkg/polylog"
)


// TODO_IN_THIS_PR: Add a Consensus type as a StateParamValue type (and find better names):
//	e.g. ReadonlyServiceState.GetConsensusValue(key string) Consensus (map[string]int) ==> for Archival checks (and maybe blockNumber)

// TODO_IN_THIS_PR: find the best fitting file for this.
// ====> framework.go  ????
// EndpointSelector is called to select eligible endpoints based on the service state.
type EndpointSelector func(ctx *EndpointSelectionContext, []protocol.Endpoint) (protocol.EndpointAddr, error)

// EndpointResultBuilder is called to process a result based on the current service state.
type EndpointResultBuilder func(ctx *ResultContext) *ResultData

// StateUpdater is called to update the service state based on observed results.
type StateUpdater func(ctx *StateUpdateContext) *FullState

// TODO_IN_THIS_PR: choose a name for this struct.
type FullState struct {

}



// serviceState provides the functionality required for selecting an endpoint from the set of available ones.
var _ protocol.EndpointSelector = &serviceState{}


// serviceState maintains the state for a QoS service.
// It provides methods for updating endpoint result data, applying observations, and filtering endpoints based on the service's state.
// It implements the protocol.EndpointSelector interface.
type serviceState struct {
	// mu protects the state map from concurrent access
	mu sync.RWMutex

	// TODO_IN_THIS_PR: add the data type required for archival checks ONLY (e.g. a map[string]json.RawMessage)
	// state is a simple key-value store for the service state
	// This is intentionally kept as string-to-string for simplicity
	state map[string]string

	// logger for diagnostics
	logger polylog.Logger

	// Custom service callbacks
	resultProcessor  ResultProcessor
	stateUpdater     StateUpdater
	endpointSelector EndpointSelector


	endpointStore *endpointStore

		// Map of method-specific result builders, specified by the custom QoS service.
		// The framework will utilize a default result builder for JSONRC methods not specified by the custom QoS service.
		MethodResultBuilders map[jsonrpc.Method]EndpointResultBuilder

}


func (s *serviceState) buildResult(parsedResult *EndpointResult) *EndpointResult {
	// Create result context for the processor
	ctx := &EndpointResultContext{
		EndpointAddr: call.EndpointAddr,    ///   <<< remove: the Result should already have an endpointCall pointer in it.
		Request:      call.Request,
		Response:     result.parseResult.Response,

		// Process valid JSONRPC response based on method.
		// The context will use a default builder if a builder matching the method is not found.
		resultBuilder: s.methodResultBuilders[call.Request.Method]
	}

	return ctx.buildResult()
}


// ProcessResult calls the custom processor with the result data
// and the current service state, returning the updated result data.
func (s *ServiceState) ProcessResult(resultData *ResultData, endpointCtx *EndpointContext) *ResultData {
	stateCopy := s.GetStateCopy()

	// Create the result context
	ctx := &ResultContext{
		ResultData:  resultData,
		EndpointCtx: endpointCtx,
		state:       stateCopy,
	}

	// Call the custom processor with the context
	return s.resultProcessor(ctx)
}

// UpdateStateFromResults calls the custom updater with the results
// and the current service state, updating the service state.
func (s *ServiceState) UpdateStateFromResults(results []*ResultData) {
	// No need to process if no custom state updater is provided.
	if stateUpdater == nil {
		return
	}

	// No need to process if no results
	if len(results) == 0 {
		return
	}

	stateCopy := s.GetStateCopy()

	// Create the state update context
	ctx := &StateUpdateContext{
		Results:      results,
		currentState: stateCopy,
		newState:     make(map[string]string),
	}
	
	// Initialize with a copy of the current state
	ctx.CopyCurrentState()

	// Call the custom updater with the context
	updatedState := s.stateUpdater(ctx)

	// Acquire write lock to update the state
	s.mu.Lock()
	defer s.mu.Unlock()

	// Replace the state with the updated state
	s.state = updatedState
}

// Select initializes an instance of EndpointSelectionContext to which it delegates the endpoint selection.
func (s *ServiceState) Select(request *jsonrpc.Request, endpoints []protocol.Endpoint) (protocol.EndpointAddr, error) {
	stateCopy := s.GetStateCopy()

	// Create the endpoint selection context
	ctx := &EndpointSelectionContext{
		logger:    s.logger,
		Request:   request,
		Endpoints: unsanctionedEndpoints,
		State:     stateCopy,

		endpointStore: s.endpointStore,
		customSelector: s.endpointSelector,
		selected:  make([]Endpoint, 0),
	}

	return ctx.selectEndpoint()
}

// GetStateValue retrieves a value from the service state by key.
// Returns the value and a boolean indicating if the key was found.
func (s *ServiceState) GetStateValue(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	value, exists := s.state[key]
	return value, exists
}

// SetStateValue sets a value in the service state by key.
func (s *ServiceState) SetStateValue(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.state[key] = value
}

// DeleteStateValue removes a value from the service state by key.
func (s *ServiceState) DeleteStateValue(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.state, key)
}

// GetStateCopy returns a copy of the entire service state.
func (s *ServiceState) GetStateCopy() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	stateCopy := make(map[string]string, len(s.state))
	for k, v := range s.state {
		stateCopy[k] = v
	}
	
	return stateCopy
}
