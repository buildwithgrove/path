// solana package provides the support required for interacting
// with the Solana blockchain through the gateway.
package solana

import (
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/relayer"
)

// TODO_TECHDEBT: factor-out any code that is common between the endpoint stores of diffrent QoS instances.
// Alternatively, have a ServiceState instance wrapped around an endpoint store: the ServiceState performs all
// endpoint selection/verification, using a minimal set of load/store operations from an endpoint store.

// UpdateEndpointsFromObservations creates/updates endpoint entries in the store based on the supplied observations.
// It returns the set of created/updated endpoints.
func (es *EndpointStore) UpdateEndpointsFromObservations(
	solanaObservations *qosobservations.SolanaDetails,
) map[relayer.EndpointAddr]endpoint {
	es.endpointsMu.Lock()
	defer es.endpointsMu.Unlock()

	if es.endpoints == nil {
		es.endpoints = make(map[relayer.EndpointAddr]endpoint)
	}

	logger := es.Logger.With(
		"observations_count", len(solanaObservations.EndpointDetails),
	)

	updatedEndpoints := make(map[relayer.EndpointAddr]endpoint)
	for _, observation := range solanaObservations.EndpointDetails {
		if observation == nil {
			continue
		}

		logger := logger.With("endpoint", observation.EndpointAddr)
		logger.Info().Msg("processing observation for endpoint.")

		// It is a valid scenario for an endpoint to not be present in the store.
		// e.g. when the first observation(s) are received for an endpoint.
		ep := es.endpoints[relayer.EndpointAddr(observation.EndpointAddr)]

		isMutated := ep.ApplyObservation(observation)
		if !isMutated {
			continue
		}

		es.endpoints[relayer.EndpointAddr(observation.EndpointAddr)] = ep
		updatedEndpoints[relayer.EndpointAddr(observation.EndpointAddr)] = ep
	}

	return updatedEndpoints
}
