package evm

import (
	"errors"
	"math/rand"
	"sync"

	"github.com/buildwithgrove/path/relayer"
)

// EndpointStore provides the endpoint selection capability required
// by the relayer package for handling a service request.
var _ relayer.EndpointSelector = &EndpointStore{}

type EndpointStoreConfig struct {
	// TODO_TECHDEBT: apply the sync allowance when validating an endpoint's block height.
	// SyncAllowance specifies the maximum number of blocks an endpoint
	// can be behind, compared to the blockchain's estimated block height,
	// before being filtered out.
	SyncAllowance uint64

	// ChainID is the ID used by the corresponding blockchain.
	// It is used to verify responses to service requests with `eth_chainId` method.
	ChainID string
}

// EndpointStore
type EndpointStore struct {
	Config EndpointStoreConfig

	mutex       sync.RWMutex
	endpoints   map[relayer.EndpointAddr]endpoint
	blockHeight uint64
}

// TODO_UPNEXT(@adshmh): Update this method along with the relayer.EndpointSelector interface.
func (es *EndpointStore) Select(availableEndpoints map[relayer.AppAddr][]relayer.Endpoint) (relayer.AppAddr, relayer.EndpointAddr, error) {
	es.mutex.RLock()
	defer es.mutex.RUnlock()

	if len(availableEndpoints) == 0 {
		return relayer.AppAddr(""), relayer.EndpointAddr(""), errors.New("select: received empty list of endpoints to select from")
	}

	// TODO_INCOMPLETE: randomize the array of available endpoints, to avoid picking the same valid endpoint every time.

	// TODO_FUTURE: rank the endpoints based on some service-specific metric, e.g. latency, rather than making a single selection.
	for appAddr, endpoints := range availableEndpoints {
		for _, availableEndpoint := range endpoints {
			endpointAddr := availableEndpoint.Addr()
			endpoint, found := es.endpoints[endpointAddr]
			if !found {
				continue
			}

			if isEndpointValid(endpoint, es.Config.ChainID, es.blockHeight) {
				return appAddr, endpointAddr, nil
			}
		}
	}

	// TODO_INCOMPLETE: log a warning/info message to provide some visibility if endpoint selection
	// consistently reaches this point, resulting in potential service degradation, possibly due to a bug.

	// TODO_UPNEXT(@adshmh): Remove the app address hack once the relayer.EndpointSelector
	// interface is updated.
	// return a random endpoint if no endpoint has details in the store.
	for appAddr, appEndpoints := range availableEndpoints {
		return appAddr, appEndpoints[rand.Intn(len(appEndpoints))].Addr(), nil
	}

	return relayer.AppAddr(""), relayer.EndpointAddr(""), errors.New("select: all apps have empty endpoint lists.")
}

func isEndpointValid(endpoint endpoint, chainID string, blockHeight uint64) bool {
	endpointBlockHeight, err := endpoint.GetBlockHeight()
	if err != nil {
		return false
	}

	return endpoint.ChainID == chainID && endpointBlockHeight >= blockHeight
}
