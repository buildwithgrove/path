package evm

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
)

// EndpointStore provides the endpoint selection capability required
// by the protocol package for handling a service request.
var _ protocol.EndpointSelector = &EndpointStore{}

// TODO_MVP(@adshmh): rename the EndpointStoreConfig struct below and use it in the `State` struct.
// The `EndpointStore` will only maintain data on the endpoints instead of how this data should be used
// to validate endpoints.
//
// EndpointStoreConfig captures the modifiable settings of the EndpointStore.
// This will enable `EndpointStore` to be used as part of QoS for other EVM-based
// blockchains which may have different desired QoS properties.
// e.g. different blockchains QoS instances could have different tolerance levels
// for deviation from the current block height.
type EndpointStoreConfig struct {
	// TODO_TECHDEBT: apply the sync allowance when validating an endpoint's block height.
	// SyncAllowance specifies the maximum number of blocks an endpoint
	// can be behind, compared to the blockchain's perceived block height,
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
	Logger polylog.Logger

	// ServiceState is the current perceived state of the EVM blockchain.
	*ServiceState

	Config EndpointStoreConfig

	endpointsMu sync.RWMutex
	endpoints   map[protocol.EndpointAddr]endpoint
}

// Select returns an endpoint address matching an entry from the list of available endpoints.
// available endpoints are filtered based on their validity first.
// A random endpoint is then returned from the filtered list of valid endpoints.
func (es *EndpointStore) Select(availableEndpoints []protocol.Endpoint) (protocol.EndpointAddr, error) {
	logger := es.Logger.With("method", "Select")
	logger.With("total_endpoints", len(availableEndpoints)).Info().Msg("filtering available endpoints.")

	filteredEndpointsAddr, err := es.filterEndpoints(availableEndpoints)
	if err != nil {
		logger.Warn().Err(err).Msg("error filtering endpoints")
		return protocol.EndpointAddr(""), err
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
func (es *EndpointStore) filterEndpoints(availableEndpoints []protocol.Endpoint) ([]protocol.EndpointAddr, error) {
	es.endpointsMu.RLock()
	defer es.endpointsMu.RUnlock()

	logger := es.Logger.With("method", "filterEndpoints").With("qos_instance", "evm")

	if len(availableEndpoints) == 0 {
		return nil, errors.New("received empty list of endpoints to select from")
	}

	logger.Info().Msg(fmt.Sprintf("About to filter through %d available endpoints", len(availableEndpoints)))

	// TODO_FUTURE: rank the endpoints based on some service-specific metric.
	// For example: latency rather than making a single selection.
	var filteredEndpointsAddr []protocol.EndpointAddr
	for _, availableEndpoint := range availableEndpoints {
		endpointAddr := availableEndpoint.Addr()

		logger := logger.With("endpoint", endpointAddr)
		logger.Info().Msg("processing endpoint")

		endpoint, found := es.endpoints[endpointAddr]
		if !found {
			logger.Info().Msg(fmt.Sprintf("endpoint %s not found in the store. Skipping...", endpointAddr))
			continue
		}

		if err := es.ServiceState.ValidateEndpoint(endpoint); err != nil {
			logger.Info().Err(err).Msg(fmt.Sprintf("skipping endpoint that failed validation: %v", endpoint))
			continue
		}

		filteredEndpointsAddr = append(filteredEndpointsAddr, availableEndpoint.Addr())
		logger.Info().Msg(fmt.Sprintf("endpoint %s passed validation", endpointAddr))
	}

	return filteredEndpointsAddr, nil
}
