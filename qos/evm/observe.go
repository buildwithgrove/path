// evm package provides the support required for interacting
// with an EVM blockchain through the gateway.
package evm

import (
	"context"
)

// observationSet provides all the functionality required
// by the message package's ObservationSet to handle the sharing
// of QoS data between PATH instances, and the updating of local
// PATH instance's QoS data on EVM endpoints.
var _ message.ObservationSet = observationSet{}

// observation captures the result of processing an endpoint's
// response to a service request.
// It provides details needed to establish the validity of
// an endpoint.
type observation struct {
	// TODO_IMPROVE: use a custom type here.
	ChainID     string
	BlockHeight uint64
}

type observationSet struct {
	// TODO_IMPROVE: use an interface here.
	EndpointStore *EndpointStore

	Observations map[relayer.EndpointAddr][]observation
}

// TODO_IN_THIS_COMMIT: implement marshalling
func (os observationSet) MarshalJSON() ([]byte, error) {
	return nil, nil
}

func (os observationSet) NotifyStakeHolders() error {
	if os.EndpointStore == nil {
		return errors.New("notifyStakeHolders: endpoint store not set")
	}

	return os.EndpointStore.ApplyObservations(os.Observations)
}

// TODO_IMPROVE: use a separate function/struct here, instead of splitting
// the EndpointStore's methods across multiple files.
func (es *EndpointStore) ApplyObservations(endpointObservations map[relayer.EndpointAddr]observation) error {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	for endpointAddr, observations := range endpointObservations {
		// It is a valid scenario for an endpoint to not be present in the store.
		// e.g. when the first observation(s) are received for an endpoint.
		endpoint := es.endpoints[endpointAddr]
		endpoint.Apply(observations)
		es.endpoints[endpointAddr] = endpoint

		if err := endpoint.Validate(es.Config.ChainID); err != nil {
			continue
		}

		// TODO_TECHDEBT: use a more resilient method for updating block height.
		// e.g. one endpoint returning a very large number as block height should
		// not result in all other endpoints being marked as invalid.
		if endpoint.BlockHeight > es.blockHeight {
			es.blockHeight = endpoint.BlockHeight
		}
	}
}
