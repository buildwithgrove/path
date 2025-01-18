// solana package provides the support required for interacting
// with the Solana blockchain through the gateway.
package solana

import (
	"fmt"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
)

// TODO_TECHDEBT: factor-out any code that is common between the endpoint stores of diffrent QoS instances.
// Alternatively, have a ServiceState instance wrapped around an endpoint store: the ServiceState performs all
// endpoint selection/verification, using a minimal set of load/store operations from an endpoint store.
//
// UpdateEndpointsFromObservations creates/updates endpoint entries in the store based on the supplied observations.
// It returns the set of created/updated endpoints.
func (es *EndpointStore) UpdateEndpointsFromObservations(
	solanaObservations *qosobservations.SolanaRequestObservations,
) map[protocol.EndpointAddr]endpoint {
	es.endpointsMu.Lock()
	defer es.endpointsMu.Unlock()

	if es.endpoints == nil {
		es.endpoints = make(map[protocol.EndpointAddr]endpoint)
	}

	endpointObservations := solanaObservations.GetEndpointObservations()

	logger := es.Logger.With(
		"qos_instance", "solana",
		"method", "UpdateEndpointsFromObservations",
	)
	logger.Info().Msg(fmt.Sprintf("About to update endpoints from %d observations.", len(endpointObservations)))

	updatedEndpoints := make(map[protocol.EndpointAddr]endpoint)
	for _, observation := range endpointObservations {
		if observation == nil {
			logger.Info().Msg("Solana EndpointStore received a nil observation. Skipping...")
			continue
		}

		endpointAddr := protocol.EndpointAddr(observation.EndpointAddr)

		logger := logger.With("endpoint", endpointAddr)
		logger.Info().Msg("processing observation for endpoint.")

		// It is a valid scenario for an endpoint to not be present in the store.
		// e.g. when the first observation(s) are received for an endpoint.
		endpoint := es.endpoints[endpointAddr]

		isMutated := endpoint.ApplyObservation(observation)
		// If the observation did not mutate the endpoint, there is no need to update the stored endpoint entry.
		if !isMutated {
			logger.Info().Msg("endpoint was not mutated by observations. Skipping.")
			continue
		}

		es.endpoints[endpointAddr] = endpoint
		updatedEndpoints[endpointAddr] = endpoint
	}

	return updatedEndpoints
}
