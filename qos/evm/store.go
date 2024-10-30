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

// Select returns an endpoint address matching an entry from the list of available endpoints.
// available endpoints are filtered based on their validity first.
// A random endpoint is then returned from the filtered list of valid endpoints.
func (es *EndpointStore) Select(availableEndpoints []relayer.Endpoint) (relayer.EndpointAddr, error) {
	logger := es.Logger.With("method", "Select")
	logger.With("total_endpoints", len(availableEndpoints)).Info().Msg("filtering available endpoints.")

	filteredEndpointsAddr, err := es.filterEndpoints(availableEndpoints)
	if err != nil {
		logger.Warn().Err(err).Msg("error filtering endpoints")
		return relayer.EndpointAddr(""), err
	}

	if len(filteredEndpointsAddr) == 0 {
		logger.Warn().Msg("all endpoints failed validation; selecting a random endpoint.")

		randomAvailableEndpoint := availableEndpoints[rand.Intn(len(availableEndpoints))]
		return randomAvailableEndpoint.Addr(), nil
	}

	logger.With(
		"total_endpoints", len(availableEndpoints),
		"endpoints_after_filtering", len(filteredEndpointsAddr),
	).Info().Msg("filtered endpoints")

	// TODO_FUTURE: consider ranking filtered endpoints, e.g. based on latency, rather than randomization.
	return filteredEndpointsAddr[rand.Intn(len(filteredEndpointsAddr))], nil
}

// filterEndpoints returns the subset of available endpoints that are valid according to previously processed observations.
func (es *EndpointStore) filterEndpoints(availableEndpoints []relayer.Endpoint) ([]relayer.EndpointAddr, error) {
	es.endpointsMu.RLock()
	defer es.endpointsMu.RUnlock()

	if len(availableEndpoints) == 0 {
		return nil, errors.New("select: received empty list of endpoints to select from")
	}

	logger := es.Logger.With("number_of_available_endpoints", fmt.Sprintf("%d", len(availableEndpoints)))
	logger.Info().Msg("select: processing available endpoints")

	var filteredEndpointsAddr []relayer.EndpointAddr
	// TODO_FUTURE: rank the endpoints based on some service-specific metric, e.g. latency, rather than making a single selection.
	for _, availableEndpoint := range availableEndpoints {
		logger := logger.With("endpoint", availableEndpoint.Addr())
		logger.Info().Msg("select: processing endpoint")

		endpoint, found := es.endpoints[availableEndpoint.Addr()]
		if !found {
			continue
		}

		if isEndpointValid(endpoint, es.Config.ChainID, es.blockHeight) {
			filteredEndpointsAddr = append(filteredEndpointsAddr, availableEndpoint.Addr())
		}

		logger.Info().Msg("select: invalid endpoint is filtered")
	}

	return filteredEndpointsAddr, nil
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
