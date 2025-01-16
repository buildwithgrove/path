package evm

import (
	"fmt"
	"sync"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
)

// ServiceState keeps the expected current state of the EVM blockchain based on the endpoints' responses to
// different requests.
type ServiceState struct {
	// ChainID is the expected value of the `Result` field in any endpoint's response to an `eth_chainId` request.
	ChainID string

	Logger polylog.Logger

	stateLock sync.RWMutex
	// estimatedBlockNumber is the estimated current block number based on endpoints' responses to `eth_blockNumber` requests.
	// It is calculated as the maximum of block height reported by any of the endpoints.
	//
	// See the following link for more details:
	// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_blocknumber
	estimatedBlockNumber uint64
}

// TODO_FUTURE: add an endpoint ranking method which can be used to assign a rank/score to a valid endpoint to guide endpoint selection.
//
// ValidateEndpoint returns an error if the supplied endpoint is not valid based on the estimated state of the EVM blockchain.
func (s *ServiceState) ValidateEndpoint(endpoint endpoint) error {
	s.stateLock.RLock()
	defer s.stateLock.RUnlock()

	if err := endpoint.Validate(s.ChainID); err != nil {
		return err
	}

	blockNumber, err := endpoint.GetBlockNumber()
	if err != nil {
		return err
	}

	if blockNumber < s.estimatedBlockNumber {
		return fmt.Errorf("endpoint has block height %d, estimated block height is %d", blockNumber, s.estimatedBlockNumber)
	}

	return nil
}

// UpdateFromObservations updates the service state using estimation(s) derived from the set of updated endpoints.
// This only includes the set of endpoints for which an observation was received.
func (s *ServiceState) UpdateFromEndpoints(updatedEndpoints map[protocol.EndpointAddr]endpoint) error {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	for endpointAddr, endpoint := range updatedEndpoints {
		// Do NOT use the endpoint for updating the estimated state of the EVM blockchain if the endpoint is not considered valid.
		// e.g. an endpoint with an invalid response to `eth_chainId` will not be used to update the estimated block number.
		if err := endpoint.Validate(s.ChainID); err != nil {
			continue
		}

		// TODO_TECHDEBT: use a more resilient method for updating block height.
		// e.g. one endpoint returning a very large number as block height should
		// not result in all other endpoints being marked as invalid.
		blockNumber, err := endpoint.GetBlockNumber()
		if err != nil {
			continue
		}

		logger := s.Logger.With(
			"endpoint", endpointAddr,
			"estimated_block_number", s.estimatedBlockNumber,
			"endpoint_block_number", blockNumber,
		)

		s.estimatedBlockNumber = blockNumber

		logger.Info().Msg("Updating latest block height")
	}

	return nil
}
