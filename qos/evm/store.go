package evm

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/relayer"
)

// EndpointStore provides the endpoint selection capability required
// by the relayer package for handling a service request.
var _ relayer.EndpointSelector = &EndpointStore{}

// EndpointStoreConfig captures the modifiable settings of the EndpointStore.
// This will enable `EndpointStore` to be used as part of QoS for other EVM-based
// blockchains which may have different desired QoS properties.
// e.g. different blockchains QoS instances could have different tolerance levels
// for deviation from the current block height.
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

// EndpointStore maintains QoS data on the set of available endpoints
// for an EVM-based blockchain service.
// It performs several tasks, most notable:
//
//	1- Endpoint selection based on the quality data available
//	2- Application of endpoints' observations to update the data on endpoints.
type EndpointStore struct {
	Config EndpointStoreConfig
	Logger polylog.Logger

	endpointsMu sync.RWMutex
	endpoints   map[relayer.EndpointAddr]endpoint
	// blockHeight is the expected latest block height on the blockchain.
	// It is calculated as the maximum of block height reported by any of the endpoints.
	blockHeight uint64
}

// TODO_UPNEXT(@adshmh): Update this method along with the relayer.EndpointSelector interface.
func (es *EndpointStore) Select(availableEndpoints map[relayer.AppAddr][]relayer.Endpoint) (relayer.AppAddr, relayer.EndpointAddr, error) {
	es.endpointsMu.RLock()
	defer es.endpointsMu.RUnlock()

	if len(availableEndpoints) == 0 {
		return relayer.AppAddr(""), relayer.EndpointAddr(""), errors.New("select: received empty list of endpoints to select from")
	}

	// TODO_UPNEXT(@adshmh): Use a randomization, e.g. standard library's
	// random.Shuffle() method, once the Protocol interface is updated,
	// rather than relying on map's range operator randomization
	uniqueEndpoints := make(map[relayer.EndpointAddr]relayer.AppAddr)
	for appAddr, endpoints := range availableEndpoints {
		for _, endpoint := range endpoints {
			uniqueEndpoints[endpoint.Addr()] = appAddr
		}
	}

	logger := es.Logger.With("number of unique endpoints", fmt.Sprintf("%d", len(uniqueEndpoints)))
	logger.Info().Msg("select: processing available endpoints")

	// TODO_FUTURE: rank the endpoints based on some service-specific metric, e.g. latency, rather than making a single selection.
	for endpointAddr, appAddr := range uniqueEndpoints {
		logger := logger.With("endpoint", endpointAddr)
		logger.Info().Msg("select: processing endpoint")

		endpoint, found := es.endpoints[endpointAddr]
		if !found {
			continue
		}

		if isEndpointValid(endpoint, es.Config.ChainID, es.blockHeight) {
			return appAddr, endpointAddr, nil
		}

		logger.Info().Msg("select: invalid endpoint is filtered")
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

// isEndpointValid returns true if the input endpoint is valid for the passed
// chain ID and query block height.
func isEndpointValid(endpoint endpoint, chainID string, queryBlockHeight uint64) bool {
	endpointBlockHeight, err := endpoint.GetBlockHeight()
	if err != nil {
		return false
	}

	return endpoint.ChainID == chainID && endpointBlockHeight >= queryBlockHeight
}
