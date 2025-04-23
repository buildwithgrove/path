package solana

import (
	"errors"
	"fmt"
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
func (es *EndpointStore) Select(allAvailableEndpoints []protocol.EndpointAddr) (protocol.EndpointAddr, error) {
	logger := es.logger.With("method", "Select")
	logger.With("total_endpoints", len(allAvailableEndpoints)).Info().Msg("filtering available endpoints.")

	filteredEndpointsAddr, err := es.filterValidEndpoints(allAvailableEndpoints)
	if err != nil {
		logger.Warn().Err(err).Msg("error filtering endpoints")
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

// filterValidEndpoints returns the subset of available endpoints that are valid according to previously processed observations.
func (es *EndpointStore) filterValidEndpoints(allAvailableEndpoints []protocol.EndpointAddr) ([]protocol.EndpointAddr, error) {
	es.endpointsMu.RLock()
	defer es.endpointsMu.RUnlock()

	logger := es.logger.With("method", "filterEndpoints").With("qos_instance", "solana")

	if len(allAvailableEndpoints) == 0 {
		return nil, errors.New("received empty list of endpoints to select from")
	}

	logger.Info().Msg(fmt.Sprintf("About to filter through %d available endpoints", len(allAvailableEndpoints)))

	// TODO_FUTURE: rank the endpoints based on some service-specific metric.
	// For example: latency rather than making a single selection.
	var filteredEndpointsAddr []protocol.EndpointAddr
	for _, availableEndpointAddr := range allAvailableEndpoints {
		logger := logger.With("endpoint_addr", availableEndpointAddr)

		logger.Info().Msg("processing endpoint")

		endpoint, found := es.endpoints[availableEndpointAddr]
		if !found {
			logger.Info().Msg(fmt.Sprintf("endpoint %s not found in the store. Skipping...", availableEndpointAddr))
			continue
		}

		if err := es.serviceState.ValidateEndpoint(endpoint); err != nil {
			logger.Info().Err(err).Msg(fmt.Sprintf("skipping endpoint that failed validation: %v", endpoint))
			continue
		}

		filteredEndpointsAddr = append(filteredEndpointsAddr, availableEndpointAddr)
		logger.Info().Msg(fmt.Sprintf("endpoint %s passed validation", availableEndpointAddr))
	}

	return filteredEndpointsAddr, nil
}
