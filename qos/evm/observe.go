// evm package provides the support required for interacting
// with an EVM blockchain through the gateway.
package evm

import (
	"errors"

	"github.com/buildwithgrove/path/message"
	"github.com/buildwithgrove/path/relayer"
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
// e.g. an observation can be the block height reported
// by an endpoint of an EVM-based blockchain service.
type observation struct {
	// TODO_IMPROVE: use a custom type here.
	ChainID string
	// This is intentionally a string to allow validation
	// of an endpoint's response.
	BlockHeight string
}

type observationSet struct {
	// TODO_IMPROVE: use an interface here.
	EndpointStore *EndpointStore

	Observations map[relayer.EndpointAddr][]observation
}

// TODO_UPNEXT(@adshmh): implement marshalling to allow the
// observation set to be processed, e.g. by the corresponding QoS instance.
func (os observationSet) MarshalJSON() ([]byte, error) {
	return nil, nil
}

func (os observationSet) Broadcast() error {
	if os.EndpointStore == nil {
		return errors.New("notifyStakeHolders: endpoint store not set")
	}

	return os.EndpointStore.ProcessObservations(os.Observations)
}

// TODO_IMPROVE: use a separate function/struct here, instead of splitting
// the EndpointStore's methods across multiple files.
func (es *EndpointStore) ProcessObservations(endpointObservations map[relayer.EndpointAddr][]observation) error {
	es.endpointsMu.Lock()
	defer es.endpointsMu.Unlock()

	for endpointAddr, observations := range endpointObservations {
		// It is a valid scenario for an endpoint to not be present in the store.
		// e.g. when the first observation(s) are received for an endpoint.
		endpoint := es.endpoints[endpointAddr]
		endpoint.Process(observations)
		es.endpoints[endpointAddr] = endpoint

		if err := endpoint.Validate(es.Config.ChainID); err != nil {
			continue
		}

		endpointBlockHeight, err := endpoint.GetBlockHeight()
		if err != nil {
			continue
		}

		// TODO_TECHDEBT: use a more resilient method for updating block height.
		// e.g. one endpoint returning a very large number as block height should
		// not result in all other endpoints being marked as invalid.
		if endpointBlockHeight > es.blockHeight {
			es.Logger.With(
				"block height", endpointBlockHeight,
				"endpoint", endpointAddr,
			).Info().Msg("Updating latest block height")
			es.blockHeight = endpointBlockHeight
		}
	}

	return nil
}
