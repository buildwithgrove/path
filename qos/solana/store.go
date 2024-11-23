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

	if len(availableEndpoints) == 0 {
		return nil, errors.New("select: received empty list of endpoints to select from")
	}

	logger := es.Logger.With("number of available endpoints", fmt.Sprintf("%d", len(availableEndpoints)))
	logger.Info().Msg("select: processing available endpoints")

	var filteredEndpointsAddr []protocol.EndpointAddr
	// TODO_FUTURE: rank the endpoints based on some service-specific metric, e.g. latency, rather than making a single selection.
	for _, availableEndpoint := range availableEndpoints {
		logger := logger.With("endpoint", availableEndpoint.Addr())
		logger.Info().Msg("select: processing endpoint")

		endpoint, found := es.endpoints[availableEndpoint.Addr()]
		if !found {
			continue
		}

		if err := es.ServiceState.ValidateEndpoint(endpoint); err != nil {
			logger.Info().Err(err).Msg("select: invalid endpoint is filtered")
			continue
		}

		filteredEndpointsAddr = append(filteredEndpointsAddr, availableEndpoint.Addr())
	}

	return filteredEndpointsAddr, nil
}
