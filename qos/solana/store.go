package solana

import (
	"errors"
	"math/rand"
	"sync"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
)

// EndpointStore provides the endpoint selection capability required
// by the protocol package for handling a service request.
var _ protocol.EndpointSelector = &EndpointStore{}

// EndpointStore maintains QoS data on the set of available endpoints
// for the Solana blockchain service.
// It performs several tasks:
// - Endpoint selection based on the quality data available
// - Application of endpoints' observations to update the data on endpoints.
type EndpointStore struct {
	logger polylog.Logger

	serviceState *ServiceState

	endpointsMu sync.RWMutex
	endpoints   map[protocol.EndpointAddr]endpoint
}

// Select returns a random endpoint address from the list of valid endpoints.
// Valid endpoints are determined by filtering the available endpoints based on their
// validity criteria.
func (es *EndpointStore) Select(allAvailableEndpoints protocol.EndpointAddrList) (protocol.EndpointAddr, error) {
	logger := es.logger.With(
		"qos", "Solana",
		"method", "Select",
		"num_endpoints", len(allAvailableEndpoints),
	)

	logger.Debug().Msg("filtering available endpoints.")

	filteredEndpointsAddr, err := es.filterValidEndpoints(allAvailableEndpoints)
	if err != nil {
		logger.Error().Err(err).Msg("error filtering endpoints: service request will fail.")
		return protocol.EndpointAddr(""), err
	}

	// No valid endpoints -> select a random endpoint
	if len(filteredEndpointsAddr) == 0 {
		logger.Warn().Msg("SELECTING A RANDOM ENDPOINT because all endpoints failed validation.")
		randomAvailableEndpointAddr := allAvailableEndpoints[rand.Intn(len(allAvailableEndpoints))]
		return randomAvailableEndpointAddr, nil
	}

	// TODO_FUTURE: consider ranking filtered endpoints, e.g. based on latency, rather than randomization.
	return filteredEndpointsAddr[rand.Intn(len(filteredEndpointsAddr))], nil
}

// SelectMultiple returns multiple endpoint addresses from the list of valid endpoints.
// Valid endpoints are determined by filtering the available endpoints based on their
// validity criteria. If maxCount is 0, it defaults to 1.
func (es *EndpointStore) SelectMultiple(allAvailableEndpoints protocol.EndpointAddrList, maxCount int) (protocol.EndpointAddrList, error) {
	logger := es.logger.With(
		"qos", "Solana",
		"method", "SelectMultiple",
		"num_endpoints", len(allAvailableEndpoints),
		"max_count", maxCount,
	)

	if maxCount <= 0 {
		maxCount = 1
	}

	logger.Debug().Msgf("filtering available endpoints to select up to %d.", maxCount)

	filteredEndpointsAddr, err := es.filterValidEndpoints(allAvailableEndpoints)
	if err != nil {
		logger.Error().Err(err).Msg("error filtering endpoints: service request will fail.")
		return nil, err
	}

	// No valid endpoints -> select random endpoints
	if len(filteredEndpointsAddr) == 0 {
		logger.Warn().Msg("SELECTING RANDOM ENDPOINTS because all endpoints failed validation.")
		countToSelect := maxCount
		if countToSelect > len(allAvailableEndpoints) {
			countToSelect = len(allAvailableEndpoints)
		}

		// Create a copy to avoid modifying the original slice
		availableCopy := make(protocol.EndpointAddrList, len(allAvailableEndpoints))
		copy(availableCopy, allAvailableEndpoints)

		// Fisher-Yates shuffle for random selection without replacement
		var selectedEndpoints protocol.EndpointAddrList
		for i := 0; i < countToSelect; i++ {
			j := rand.Intn(len(availableCopy)-i) + i
			availableCopy[i], availableCopy[j] = availableCopy[j], availableCopy[i]
			selectedEndpoints = append(selectedEndpoints, availableCopy[i])
		}
		return selectedEndpoints, nil
	}

	// Select up to maxCount endpoints from filtered list
	countToSelect := maxCount
	if countToSelect > len(filteredEndpointsAddr) {
		countToSelect = len(filteredEndpointsAddr)
	}

	// Create a copy to avoid modifying the original slice
	filteredCopy := make(protocol.EndpointAddrList, len(filteredEndpointsAddr))
	copy(filteredCopy, filteredEndpointsAddr)

	// Fisher-Yates shuffle for random selection without replacement
	var selectedEndpoints protocol.EndpointAddrList
	for i := 0; i < countToSelect; i++ {
		j := rand.Intn(len(filteredCopy)-i) + i
		filteredCopy[i], filteredCopy[j] = filteredCopy[j], filteredCopy[i]
		selectedEndpoints = append(selectedEndpoints, filteredCopy[i])
	}

	return selectedEndpoints, nil
}

// filterValidEndpoints returns the subset of available endpoints that are valid according to previously processed observations.
func (es *EndpointStore) filterValidEndpoints(allAvailableEndpoints protocol.EndpointAddrList) (protocol.EndpointAddrList, error) {
	es.endpointsMu.RLock()
	defer es.endpointsMu.RUnlock()

	logger := es.logger.With(
		"method", "filterEndpoints",
		"qos_instance", "solana",
		"num_endpoints", len(allAvailableEndpoints),
	)

	if len(allAvailableEndpoints) == 0 {
		return nil, errors.New("received empty list of endpoints to select from")
	}

	logger.Debug().Msg("About to filter available endpoints.")

	// TODO_FUTURE: rank the endpoints based on some service-specific metric.
	// For example: latency rather than making a single selection.
	var filteredEndpointsAddr protocol.EndpointAddrList
	for _, availableEndpointAddr := range allAvailableEndpoints {
		logger := logger.With("endpoint_addr", availableEndpointAddr)

		logger.Debug().Msg("Processing endpoint")

		endpoint, found := es.endpoints[availableEndpointAddr]
		if !found {
			logger.Warn().Msgf("❓ Skipping endpoint because it was not found in PATH's endpoint store: %s", availableEndpointAddr)
			continue
		}

		if err := es.serviceState.ValidateEndpoint(endpoint); err != nil {
			logger.Error().Err(err).Msgf("❌ Skipping endpoint because it failed validation: %s", availableEndpointAddr)
			continue
		}

		filteredEndpointsAddr = append(filteredEndpointsAddr, availableEndpointAddr)
		logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msgf("✅ endpoint passed validation: %s", availableEndpointAddr)
	}

	return filteredEndpointsAddr, nil
}
