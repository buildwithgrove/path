// solana package provides the support required for interacting
// with the Solana blockchain through the gateway.
package solana

import (
	"errors"

	"github.com/buildwithgrove/path/message"
	"github.com/buildwithgrove/path/protocol"
)

// observationSet provides all the functionality required
// by the message package's ObservationSet to handle the sharing
// of QoS data between PATH instances, and the updating of local
// PATH instance's QoS data on Solana endpoints.
var _ message.ObservationSet = observationSet{}

// observation captures the result of processing an endpoint's
// response to a service request.
// It provides details needed to establish the validity of
// an endpoint.
type observation interface {
	// Apply updates the endpoint based on the observation's contents.
	// e.g. an observation from a response to a `getHealth` request updates the IsHealthy field of an endpoint.
	Apply(*endpoint)
}

type observationSet struct {
	// TODO_IMPROVE: use an interface here.
	EndpointStore *EndpointStore
	ServiceState  *ServiceState

	Observations map[protocol.EndpointAddr][]observation
}

// TODO_UPNEXT(@adshmh): implement marshalling to allow the
// observation set to be shared among PATH instances.
func (os observationSet) MarshalJSON() ([]byte, error) {
	return nil, nil
}

func (os observationSet) Broadcast() error {
	if os.EndpointStore == nil {
		return errors.New("broadcast: endpoint store not set")
	}

	updatedEndpoints := os.EndpointStore.ProcessObservations(os.Observations)

	// update the (estimated) current state of the blockchain.
	return os.ServiceState.UpdateFromEndpoints(updatedEndpoints)
}

// TODO_TECHDEBT: factor-out any code that is common between the endpoint stores of diffrent QoS instances.
// Alternatively, have a ServiceState instance wrapped around an endpoint store: the ServiceState performs all
// endpoint selection/verification, using a minimal set of load/store operations from an endpoint store.
func (es *EndpointStore) ProcessObservations(endpointObservations map[protocol.EndpointAddr][]observation) map[protocol.EndpointAddr]*endpoint {
	es.endpointsMu.Lock()
	defer es.endpointsMu.Unlock()

	if es.endpoints == nil {
		es.endpoints = make(map[protocol.EndpointAddr]endpoint)
	}

	updatedEndpoints := make(map[protocol.EndpointAddr]*endpoint)
	for endpointAddr, observations := range endpointObservations {
		logger := es.Logger.With(
			"endpoint", endpointAddr,
			"observations count", len(observations),
		)
		logger.Info().Msg("processing observations for endpoint.")

		// It is a valid scenario for an endpoint to not be present in the store.
		// e.g. when the first observation(s) are received for an endpoint.
		endpoint := es.endpoints[endpointAddr]
		for _, observation := range observations {
			observation.Apply(&endpoint)
		}
		es.endpoints[endpointAddr] = endpoint

		updatedEndpoints[endpointAddr] = &endpoint
	}

	return updatedEndpoints
}
