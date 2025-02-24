package evm

import (
	"sync"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
)

// EndpointStoreConfig captures the modifiable settings of the EndpointStore.
// This will enable `EndpointStore` to be used as part of QoS for other EVM-based
// blockchains which may have different desired QoS properties.
// e.g. different blockchains QoS instances could have different tolerance levels
// for deviation from the current block height.
type serviceStateConfig struct {
	// syncAllowance specifies the maximum number of blocks an endpoint
	// can be behind, compared to the blockchain's perceived block height,
	// before being filtered out.
	syncAllowance uint64

	// chainID is the expected value of the `Result` field in any endpoint's response to an `eth_chainId` request.
	// See the following link for more details: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_chainid
	// Chain IDs Reference: https://chainlist.org/
	chainID string
}

// ServiceState keeps the expected current state of the EVM blockchain based on the endpoints' responses to
// different requests.
type ServiceState struct {
	logger polylog.Logger

	serviceStateLock sync.RWMutex

	// config captures the modifiable settings of the ServiceState.
	config serviceStateConfig

	// perceivedBlockNumber is the perceived current block number based on endpoints' responses to `eth_blockNumber` requests.
	// It is calculated as the maximum of block height reported by any of the endpoints.
	//
	// See the following link for more details:
	// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_blocknumber
	perceivedBlockNumber uint64
}

// UpdateFromEndpoints updates the service state using estimation(s) derived from the set of updated endpoints.
// This only includes the set of endpoints for which an observation was received.
func (s *ServiceState) UpdateFromEndpoints(updatedEndpoints map[protocol.EndpointAddr]endpoint) error {
	s.serviceStateLock.Lock()
	defer s.serviceStateLock.Unlock()

	for endpointAddr, endpoint := range updatedEndpoints {
		logger := s.logger.With(
			"endpoint_addr", endpointAddr,
			"perceived_block_number", s.perceivedBlockNumber,
		)

		// DO NOT use the endpoint for updating the perceived state of the EVM blockchain if the endpoint is not considered valid.
		// e.g. an endpoint with an invalid response to `eth_chainId` will not be used to update the perceived block number.
		if err := endpoint.Validate(s); err != nil {
			logger.Info().Err(err).Msg("Skipping endpoint with invalid chain id")
			continue
		}

		// TODO_TECHDEBT: use a more resilient method for updating block height.
		// E.g. one endpoint returning a very large number as block height should
		// not result in all other endpoints being marked as invalid.
		blockNumber, err := endpoint.GetBlockNumber()
		if err != nil {
			logger.Info().Err(err).Msg("Skipping endpoint with invalid block number")
			continue
		}

		s.perceivedBlockNumber = blockNumber

		logger.With("endpoint_block_number", blockNumber).Info().Msg("Updating latest block height")
	}

	return nil
}
