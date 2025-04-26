package jsonrpc

import (
	"sync"
	"time"

	"github.com/buildwithgrove/path/protocol"
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// serviceState maintains the state for a QoS service.
// It provides methods for:
// - Updating state parameters using observations.
// - Updating endpoint results using observations
// - Reading state parameters and endpoints (Read Only) for:
//   - building endpoint results
//   - endpoint selection
//
type ServiceState struct {
	// logger for diagnostics
	logger polylog.Logger

	// mu protects the state map from concurrent access
	mu sync.RWMutex

	// stateParameters
	parameters map[string]*StateParameter

	endpointStore *endpointStore
}

func (s *ServiceState) GetStrParam(paramName string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	param, ok := s.parameters[paramName]
	if !ok {
		return "", false
	}

	return param.GetStr()
}

func (s *ServiceState) GetIntParam(paramName string) (int, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	param, ok := s.parameters[paramName]
	if !ok {
		return 0, false
	}

	return param.GetInt()

}

func (s *ServiceState) GetConsensusParam(paramName string) (map[string]int, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	param, ok := s.parameters[paramName]
	if !ok {
		return nil, false
	}

	return param.GetConsensus()
}

// Returns the stored Endpoint structs matching the endpoint queries.
func (s *serviceState) updateStoredEndpoints(endpointQueries []*endpointQuery) []Endpoint {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.endpointStore.updateEndpoints(endpointQueries)
}

func (s *ServiceState) updateParameters(updates StateParameterUpdateSet) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for paramName, param := range updates.Updates {
		s.parameters[paramName] = param
	}

	return nil
}
