// solana package provides the support required for interacting
// with the Solana blockchain through the gateway.
package solana

import (
	"errors"

	"github.com/buildwithgrove/path/message"
	"github.com/buildwithgrove/path/observation"
	"github.com/buildwithgrove/path/relayer"
)

// TODO_TECHDEBT: factor-out any code that is common between the endpoint stores of diffrent QoS instances.
// Alternatively, have a ServiceState instance wrapped around an endpoint store: the ServiceState performs all
// endpoint selection/verification, using a minimal set of load/store operations from an endpoint store.

// UpdateEndpointsFromObservations creates/updates endpoint entries in the store based on the supplied observations.
// It returns the set of created/updated endpoints.
func (es *EndpointStore) UpdateEndpointsFromObservations(
	solanaObservations *observation.qos.SolanaDetails,
) map[relayer.EndpointAddr]*endpoint {
	es.endpointsMu.Lock()
	defer es.endpointsMu.Unlock()

	if es.endpoints == nil {
		es.endpoints = make(map[relayer.EndpointAddr]endpoint)
	}

	updatedEndpoints := make(map[relayer.EndpointAddr]*endpoint)
	for _, observation := range solanaObservations {
		logger := es.Logger.With(
			"endpoint", endpointAddr,
			"observations count", len(observations),
		)
		logger.Info().Msg("processing observations for endpoint.")

		// It is a valid scenario for an endpoint to not be present in the store.
		// e.g. when the first observation(s) are received for an endpoint.
		endpoint, found := es.endpoints[observation.EndpointAddr]
		if !found {
			endpoint = &endpoint{}
		}

		isMutated := endpoint.Apply(observation)
		if !isMutated {
			continue
		}

		es.endpoints[observation.EndpointAddr] = endpoint
		updatedEndpoints[observation.EndpointAddr] = endpoint
	}

	return updatedEndpoints
}
