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
// It performs several tasks, most notably:
//
//	1- Endpoint selection based on the quality data available
//	2- Application of endpoints' observations to update the data on endpoints.
type EndpointStore struct {
	ServiceState *ServiceState
	Logger       polylog.Logger

	endpointsMu sync.RWMutex
	endpoints   map[protocol.EndpointAddr]endpoint
}

// Select returns an endpoint address matching an entry from the list of available endpoints.
// available endpoints are filtered based on their validity first.
// A random endpoint is then returned from the filtered list of valid endpoints.
func (es *EndpointStore) Select(availableEndpoints []protocol.Endpoint) (protocol.EndpointAddr, error) {
	filteredEndpointsAddr, err := es.filterEndpoints(availableEndpoints)
	if err != nil {
		return protocol.EndpointAddr(""), err
	}

	logger := es.Logger.With("number of available endpoints", len(availableEndpoints))
	if len(filteredEndpointsAddr) == 0 {
		logger.Warn().Msg("select: all endpoints failed validation; selecting a random endpoint.")
		randomAvailableEndpoint := availableEndpoints[rand.Intn(len(availableEndpoints))]
		return randomAvailableEndpoint.Addr(), nil
	}

	// TODO_FUTURE: consider ranking filtered endpoints, e.g. based on latency, rather than randomization.
	return filteredEndpointsAddr[rand.Intn(len(filteredEndpointsAddr))], nil
}

// filterEndpoints returns the subset of available endpoints that are valid according to previously processed observations.
func (es *EndpointStore) filterEndpoints(availableEndpoints []protocol.Endpoint) ([]protocol.EndpointAddr, error) {
	es.endpointsMu.RLock()
	defer es.endpointsMu.RUnlock()

	logger := es.Logger.With("method", "filterEndpoints").With("qos_instance", "solana")

	if len(availableEndpoints) == 0 {
		return nil, errors.New("received empty list of endpoints to select from")
	}

	logger.Info().Msg(fmt.Sprintf("About to filter through %d available endpoints", len(availableEndpoints)))

	// TODO_FUTURE: rank the endpoints based on some service-specific metric.
	// For example: latency rather than making a single selection.
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
