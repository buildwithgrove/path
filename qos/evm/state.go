package evm

import (
	"fmt"
	"sync"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos"
)

// TODO_MVP(@adshmh): rename the EndpointStoreConfig struct below and use it in the `State` struct.
// The `EndpointStore` will only maintain data on the endpoints instead of how this data should be used to validate endpoints.
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

// ServiceState keeps the expected current state of the EVM blockchain based on the endpoints' responses to
// different requests.
type ServiceState struct {
	Logger polylog.Logger

	// ChainID is the expected value of the `Result` field in any endpoint's response to an `eth_chainId` request.
	ChainID string

	stateLock sync.RWMutex

	// perceivedBlockNumber is the perceived current block number based on endpoints' responses to `eth_blockNumber` requests.
	// It is calculated as the maximum of block height reported by any of the endpoints.
	//
	// See the following link for more details:
	// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_blocknumber
	perceivedBlockNumber uint64
}

// TODO_FUTURE: add an endpoint ranking method which can be used to assign a rank/score to a valid endpoint to guide endpoint selection.
//
// ValidateEndpoint returns an error if the supplied endpoint is not valid based on the perceived state of the EVM blockchain.
func (s *ServiceState) ValidateEndpoint(qosEndpoint qos.Endpoint) error {
	s.stateLock.RLock()
	defer s.stateLock.RUnlock()

	evmEndpoint, ok := qosEndpoint.(endpoint)
	if !ok {
		return fmt.Errorf("endpoint was not of type evm.endpoint")
	}

	if err := evmEndpoint.Validate(s.ChainID); err != nil {
		return err
	}

	if err := validateEndpointBlockNumber(evmEndpoint, s.perceivedBlockNumber); err != nil {
		return err
	}

	return nil
}

// UpdateFromObservations updates the service state using estimation(s) derived from the set of updated endpoints.
// This only includes the set of endpoints for which an observation was received.
func (s *ServiceState) UpdateFromEndpoints(updatedEndpoints map[protocol.EndpointAddr]qos.Endpoint) error {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	for endpointAddr, updatedEndpoint := range updatedEndpoints {
		logger := s.Logger.With(
			"endpoint_addr", endpointAddr,
			"perceived_block_number", s.perceivedBlockNumber,
		)

		evmEndpoint, ok := updatedEndpoint.(endpoint)
		if !ok {
			logger.Info().Msg("endpoint was not of type evm.endpoint. Skipping.")
			continue
		}

		// Do NOT use the endpoint for updating the perceived state of the EVM blockchain if the endpoint is not considered valid.
		// e.g. an endpoint with an invalid response to `eth_chainId` will not be used to update the perceived block number.
		if err := evmEndpoint.Validate(s.ChainID); err != nil {
			logger.Info().Err(err).Msg("Skipping endpoint with invalid chain id")
			continue
		}

		// TODO_TECHDEBT: use a more resilient method for updating block height.
		// e.g. one endpoint returning a very large number as block height should
		// not result in all other endpoints being marked as invalid.
		blockNumber, err := evmEndpoint.getBlockNumber()
		if err != nil {
			logger.Info().Err(err).Msg("Skipping endpoint with invalid block number")
			continue
		}

		s.perceivedBlockNumber = blockNumber

		logger.With("endpoint_block_number", blockNumber).Info().Msg("Updating latest block height")
	}

	return nil
}

// validateEndpointBlockNumber validates the supplied endpoint against the supplied perceived block number for the EVM blockchain.
func validateEndpointBlockNumber(evmEndpoint endpoint, perceivedBlockNumber uint64) error {
	blockNumber, err := evmEndpoint.getBlockNumber()
	if err != nil {
		return err
	}

	if blockNumber < perceivedBlockNumber {
		return fmt.Errorf("endpoint has block height %d, perceived block height is %d", blockNumber, perceivedBlockNumber)
	}

	return nil
}
