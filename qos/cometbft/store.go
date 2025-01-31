package cometbft

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
// for a CometBFT-based blockchain service.
// It performs several tasks:
// - Endpoint selection based on the quality data available
// - Application of endpoints' observations to update the data on endpoints
type EndpointStore struct {
	logger polylog.Logger

	// ServiceState is the current perceived state of the CometBFT blockchain
	*ServiceState

	endpointsMu sync.RWMutex
	endpoints   map[protocol.EndpointAddr]endpoint
}

// Select returns an endpoint address matching an entry from the list of available endpoints.
// Available endpoints are filtered based on their validity first.
// A random endpoint is then returned from the filtered list of valid endpoints.
//
// TODO_TECHDEBT(@commoddity): Look into refactoring and reusing specific components
// that play identical roles across QoS packages in order to reduce code duplication.
// For example, the EndpointStore is a great candidate for refactoring.
func (es *EndpointStore) Select(availableEndpoints []protocol.Endpoint) (protocol.EndpointAddr, error) {
	logger := es.logger.With("method", "Select")
	logger.With("total_endpoints", len(availableEndpoints)).Info().Msg("filtering available endpoints.")

	filteredEndpointsAddr, err := es.filterEndpoints(availableEndpoints)
	if err != nil {
		logger.Warn().Err(err).Msg("error filtering endpoints")
		return protocol.EndpointAddr(""), err
	}

	if len(filteredEndpointsAddr) == 0 {
		logger.Warn().Msg("all endpoints failed validation; selecting a random endpoint.")
		randomAvailableEndpoint := availableEndpoints[rand.Intn(len(availableEndpoints))]
		return randomAvailableEndpoint.Addr(), nil
	}

	logger.With(
		"total_endpoints", len(availableEndpoints),
		"endpoints_after_filtering", len(filteredEndpointsAddr),
	).Info().Msg("filtered endpoints")

	// TODO_FUTURE: consider ranking filtered endpoints, e.g. based on latency, rather than randomization.
	return filteredEndpointsAddr[rand.Intn(len(filteredEndpointsAddr))], nil
}

// filterEndpoints returns the subset of available endpoints that are valid
// according to previously processed observations.
func (es *EndpointStore) filterEndpoints(availableEndpoints []protocol.Endpoint) ([]protocol.EndpointAddr, error) {
	es.endpointsMu.RLock()
	defer es.endpointsMu.RUnlock()

	logger := es.logger.With("method", "filterEndpoints", "qos_instance", "cometbft")

	if len(availableEndpoints) == 0 {
		return nil, errors.New("received empty list of endpoints to select from")
	}

	logger.Info().Msg(fmt.Sprintf("About to filter through %d available endpoints", len(availableEndpoints)))

	// TODO_FUTURE: rank the endpoints based on some service-specific metric.
	var filteredEndpointsAddr []protocol.EndpointAddr
	for _, availableEndpoint := range availableEndpoints {
		endpointAddr := availableEndpoint.Addr()
		logger := logger.With("endpoint", endpointAddr)
		logger.Info().Msg("processing endpoint")

		endpoint, found := es.endpoints[endpointAddr]
		if !found {
			logger.Info().Msg(fmt.Sprintf("endpoint %s not found in the store. Skipping...", endpointAddr))
			continue
		}

		if err := es.ServiceState.ValidateEndpoint(endpoint); err != nil {
			logger.Info().Err(err).Msg(fmt.Sprintf("skipping endpoint that failed validation: %v", endpoint))
			continue
		}

		filteredEndpointsAddr = append(filteredEndpointsAddr, availableEndpoint.Addr())
		logger.Info().Msg(fmt.Sprintf("endpoint %s passed validation", endpointAddr))
	}

	return filteredEndpointsAddr, nil
}
