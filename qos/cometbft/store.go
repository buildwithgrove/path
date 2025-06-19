package cometbft

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
// for a CometBFT-based blockchain service.
// It performs several tasks:
// - Endpoint selection based on the quality data available
// - Application of endpoints' observations to update the data on endpoints
type EndpointStore struct {
	logger polylog.Logger

	// ServiceState is the current perceived state of the CometBFT blockchain
	serviceState *ServiceState

	endpointsMu sync.RWMutex
	endpoints   map[protocol.EndpointAddr]endpoint
}

// Select returns a random endpoint address from the list of valid endpoints.
// Valid endpoints are determined by filtering the available endpoints based on their
// validity criteria.
func (es *EndpointStore) Select(availableEndpoints protocol.EndpointAddrList) (protocol.EndpointAddr, error) {
	logger := es.logger.With("method", "Select").With("chain_id", es.serviceState.chainID)
	logger.Info().Msgf("filtering %d available endpoints.", len(availableEndpoints))

	filteredEndpointsAddr, err := es.filterValidEndpoints(availableEndpoints)
	if err != nil {
		logger.Error().Err(err).Msg("error filtering endpoints")
		return protocol.EndpointAddr(""), err
	}

	if len(filteredEndpointsAddr) == 0 {
		logger.Warn().Msgf(
			"SELECTING A RANDOM ENDPOINT because all endpoints failed validation. Available endpoints: %s. Filtered endpoints: %s",
			availableEndpoints.String(), filteredEndpointsAddr.String())
		randomAvailableEndpointAddr := availableEndpoints[rand.Intn(len(availableEndpoints))]
		return randomAvailableEndpointAddr, nil
	}

	logger.Info().Msgf("filtered %d endpoints from %d available endpoints", len(filteredEndpointsAddr), len(availableEndpoints))

	// TODO_FUTURE: consider ranking filtered endpoints, e.g. based on latency, rather than randomization.
	selectedEndpointAddr := filteredEndpointsAddr[rand.Intn(len(filteredEndpointsAddr))]
	return selectedEndpointAddr, nil
}

// filterValidEndpoints returns the subset of available endpoints that are valid
// according to previously processed observations.
func (es *EndpointStore) filterValidEndpoints(allAvailableEndpoints protocol.EndpointAddrList) (protocol.EndpointAddrList, error) {
	es.endpointsMu.RLock()
	defer es.endpointsMu.RUnlock()

	logger := es.logger.With("method", "filterValidEndpoints", "qos_instance", "cometbft")

	if len(allAvailableEndpoints) == 0 {
		return nil, errors.New("received empty list of endpoints to select from")
	}

	logger.Info().Msgf("About to filter for valid endpoints through %d available endpoints", len(allAvailableEndpoints))

	// TODO_FUTURE: use service-specific metrics to add an endpoint ranking method
	// which can be used to assign a rank/score to a valid endpoint to guide endpoint selection.
	var filteredValidEndpointsAddr protocol.EndpointAddrList
	for _, availableEndpointAddr := range allAvailableEndpoints {
		logger := logger.With("endpoint_addr", availableEndpointAddr)

		logger.Info().Msg("processing endpoint")

		endpoint, found := es.endpoints[availableEndpointAddr]
		if !found {
			logger.Info().Msgf("endpoint %s not found in the store. Skipping...", availableEndpointAddr)
			continue
		}

		if err := es.serviceState.ValidateEndpoint(endpoint); err != nil {
			logger.Info().Err(err).Msgf("skipping endpoint that failed validation: %v", endpoint)
			continue
		}

		filteredValidEndpointsAddr = append(filteredValidEndpointsAddr, availableEndpointAddr)
		logger.Info().Msgf("endpoint %s passed validation", availableEndpointAddr)
	}

	return filteredValidEndpointsAddr, nil
}
