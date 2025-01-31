package qos

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
)

// EndpointStore provides the endpoint selection capability required
// by the protocol package for handling a service request.
var _ protocol.EndpointSelector = &EndpointStore{}

// EndpointStore provides the endpoint check generator required by
// the gateway package to augment endpoints' quality data,
// using synthetic service requests.
var _ gateway.QoSEndpointCheckGenerator = &EndpointStore{}

// EndpointStore maintains QoS data on the set of available endpoints for any service
// It performs several tasks, most notable:
//
//	1- Endpoint selection based on the quality data available
//	2- Application of endpoints' observations to update the data on endpoints.
type EndpointStore struct {
	Logger polylog.Logger

	// ServiceState is the current perceived state of the
	// service and varies between service QoS implementations.
	ServiceState ServiceState

	// RequiredQualityChecks is the set of quality checks performed by the Hydrator
	// that must be satisfied for an endpoint to be considered valid.
	// This slice is initialized by the service-specific implementation of NewQoSInstance.
	RequiredQualityChecks []gateway.RequestQoSContext

	// endpoints is the set of currently stored endpoints for a given service and
	// is set by applying Hydrator observations to the to a service's available
	// endpoints in each service QoS implementation's ApplyObservations method.
	endpoints   map[protocol.EndpointAddr]Endpoint
	endpointsMu sync.RWMutex
}

// ServiceState is the current perceived state of the service and
type ServiceState interface {
	ValidateEndpoint(endpoint Endpoint) error
}

// Endpoint is the interface that must be satisfied by any endpoint for a given service.
type Endpoint interface{}

// Select returns an endpoint address matching an entry from the list of available endpoints.
// available endpoints are filtered based on their validity first.
// A random endpoint is then returned from the filtered list of valid endpoints.
func (es *EndpointStore) Select(availableEndpoints []protocol.Endpoint) (protocol.EndpointAddr, error) {
	logger := es.Logger.With("method", "Select")
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

// GetEndpoints returns all currently stored endpoints.
func (es *EndpointStore) GetEndpoints() map[protocol.EndpointAddr]Endpoint {
	es.endpointsMu.RLock()
	defer es.endpointsMu.RUnlock()

	return es.endpoints
}

// UpdateEndpointsFromObservations creates/updates endpoint entries in the store based on the supplied observations.
// The process of applying observations to the endpoints is handled by the service-specific implementation of ApplyObservations as structs provided by the `observation` package are service-specific.
func (es *EndpointStore) UpdateEndpointsFromObservations(updatedEndpoints map[protocol.EndpointAddr]Endpoint,
) {
	es.endpointsMu.Lock()
	defer es.endpointsMu.Unlock()

	es.endpoints = updatedEndpoints
}

// filterEndpoints returns the subset of available endpoints that are valid according to previously processed observations.
func (es *EndpointStore) filterEndpoints(availableEndpoints []protocol.Endpoint) ([]protocol.EndpointAddr, error) {
	es.endpointsMu.RLock()
	defer es.endpointsMu.RUnlock()

	logger := es.Logger.With("method", "filterEndpoints").With("qos_instance", "cometbft")

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

func (es *EndpointStore) GetRequiredQualityChecks() []gateway.RequestQoSContext {
	return es.RequiredQualityChecks
}
