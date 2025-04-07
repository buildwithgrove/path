// cometbft package provides the support required for interacting
// with a CometBFT blockchain through the gateway.
package cometbft

import (
	"fmt"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
)

// UpdateEndpointsFromObservations CRUDs endpoint entries in the store based on the supplied observations.
// It returns the set of created/updated endpoints.
func (es *EndpointStore) UpdateEndpointsFromObservations(
	cometbftObservations *qosobservations.CometBFTRequestObservations,
) map[protocol.EndpointAddr]endpoint {
	es.endpointsMu.Lock()
	defer es.endpointsMu.Unlock()

	if es.endpoints == nil {
		es.endpoints = make(map[protocol.EndpointAddr]endpoint)
	}

	endpointObservations := cometbftObservations.GetEndpointObservations()

	logger := es.logger.With(
		"qos_instance", "cometbft",
		"method", "UpdateEndpointsFromObservations",
	)
	logger.Info().Msg(fmt.Sprintf("About to update endpoints from %d observations.", len(endpointObservations)))

	updatedEndpoints := make(map[protocol.EndpointAddr]endpoint)
	for _, observation := range endpointObservations {
		if observation == nil {
			logger.Info().Msg("CometBFT EndpointStore received a nil observation. Skipping...")
			continue
		}

		endpointAddr := protocol.EndpointAddr(observation.EndpointAddr)

		logger := logger.With("endpoint_addr", endpointAddr)
		logger.Info().Msg("processing observation for endpoint.")

		// It is a valid scenario for an endpoint to not be present in the store.
		// E.g. when the first observation(s) are received for an endpoint.
		endpoint := es.endpoints[endpointAddr]

		// If the observation did not mutate the endpoint, there is no need to update the stored endpoint entry.
		isMutated := endpoint.ApplyObservation(observation)
		if !isMutated {
			logger.Info().Msg("endpoint was not mutated by observations. Skipping.")
			continue
		}

		es.endpoints[endpointAddr] = endpoint
		updatedEndpoints[endpointAddr] = endpoint
	}

	return updatedEndpoints
}
