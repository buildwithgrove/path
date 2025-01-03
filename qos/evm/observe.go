// evm package provides the support required for interacting
// with an EVM blockchain through the gateway.
package evm

import (
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
)

// UpdateEndpointsFromObservations creates/updates endpoint entries in the store based on the supplied observations.
// It returns the set of created/updated endpoints.
func (es *EndpointStore) UpdateEndpointsFromObservations(
	evmObservations *qosobservations.EVMObservations,
) map[protocol.EndpointAddr]endpoint {
	es.endpointsMu.Lock()
	defer es.endpointsMu.Unlock()

	if es.endpoints == nil {
		es.endpoints = make(map[protocol.EndpointAddr]endpoint)
	}
	endpointObservations := evmObservations.GetEndpointObservations()

	updatedEndpoints := make(map[protocol.EndpointAddr]endpoint)
	for _, observation := range endpointObservations {
		if observation == nil {
			continue
		}
		endpointAddr := protocol.EndpointAddr(observation.EndpointAddr)

		logger := es.Logger.With(
			"observations_count", len(endpointObservations),
			"endpoint", endpointAddr,
		)
		logger.Info().Msg("processing observation for endpoint.")

		// It is a valid scenario for an endpoint to not be present in the store.
		// e.g. when the first observation(s) are received for an endpoint.
		endpoint := es.endpoints[endpointAddr]

		isMutated := endpoint.ApplyObservation(observation)
		if !isMutated {
			continue
		}

		es.endpoints[endpointAddr] = endpoint
		updatedEndpoints[endpointAddr] = endpoint
	}

	return updatedEndpoints
}
