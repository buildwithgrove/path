package evm

import (
	"sync"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
)

// ServiceState keeps the expected current state of the EVM blockchain based on
// the endpoints' responses to different requests.
type ServiceState struct {
	logger polylog.Logger

	serviceStateLock sync.RWMutex

	// chainID is the expected value of the `Result` field in any endpoint's response to an `eth_chainId` request.
	//
	// See the following link for more details: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_chainid
	//
	// Chain IDs Reference: https://chainlist.org/
	chainID string

	// perceivedBlockNumber is the perceived current block number based on endpoints' responses to `eth_blockNumber` requests.
	// It is calculated as the maximum of block height reported by any of the endpoints.
	//
	// See the following link for more details:
	// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_blocknumber
	perceivedBlockNumber uint64

	// archivalState contains the current state of the EVM archival check for the service.
	archivalState archivalState
}

// TODO_FUTURE: add an endpoint ranking method which can be used to assign a rank/score to a valid endpoint to guide endpoint selection.
//
// ValidateEndpoint returns an error if the supplied endpoint is not valid based on the perceived state of the EVM blockchain.
func (s *ServiceState) ValidateEndpoint(endpoint endpoint, endpointAddr protocol.EndpointAddr) error {
	s.serviceStateLock.RLock()
	defer s.serviceStateLock.RUnlock()

	// Ensure the response is not empty.
	if err := endpoint.validateEmptyResponse(); err != nil {
		return err
	}

	// Ensure the chain ID is valid.
	if err := endpoint.validateChainID(s.chainID); err != nil {
		return err
	}

	// Ensure the service state's perceived block number is not ahead of the endpoint's block number.
	if err := endpoint.validateBlockNumber(s.perceivedBlockNumber); err != nil {
		return err
	}

	if err := endpoint.validateArchivalCheck(s.archivalState.getBalance()); err != nil {
		return err
	}

	// The service state
	return nil
}

// UpdateFromEndpoints updates the service state using estimation(s) derived from the set of updated endpoints.
// This only includes the set of endpoints for which an observation was received.
func (s *ServiceState) UpdateFromEndpoints(updatedEndpoints map[protocol.EndpointAddr]endpoint) error {
	s.serviceStateLock.Lock()
	defer s.serviceStateLock.Unlock()

	// Initialize consensus map if it doesn't exist.
	s.archivalState.initializeConsensusMap()

	for endpointAddr, endpoint := range updatedEndpoints {
		logger := s.logger.With(
			"endpoint_addr", endpointAddr,
			"perceived_block_number", s.perceivedBlockNumber,
		)

		// Validate the endpoint's chain ID; do not update the perceived block number if the chain ID is invalid.
		if err := endpoint.validateChainID(s.chainID); err != nil {
			logger.Info().Err(err).Msg("Skipping endpoint with invalid chain id")
			continue
		}

		// Retrieve the block number from the endpoint.
		blockNumber, err := endpoint.getBlockNumber()
		if err != nil {
			logger.Info().Err(err).Msg("Skipping endpoint with invalid block number")
			continue
		}

		// Update the perceived block number.
		s.perceivedBlockNumber = blockNumber

		// Attempt to retrieve the archival balance from the endpoint.
		balance, err := endpoint.getArchivalBalance()
		if err != nil {
			logger.Info().Err(err).Msg("Skipping endpoint without archival balance")
			continue
		}

		// Update the consensus map to determine the balance at the perceived block number.
		s.archivalState.updateConsensusMap(balance)
	}

	// If the archival block number is not yet set for the service, calculate it.
	// This requires that the perceived block number is set in order to determine the latest possible block number.
	if s.perceivedBlockNumber != 0 && s.archivalState.getBlockNumberHex() == "" {
		s.archivalState.calculateArchivalBlockNumber(s.perceivedBlockNumber)
	}

	// If the expected archival balance is not yet set for the service, set it.
	// This utilizes the consensus map to determine a source of truth for the archival balance.
	// If <archivalConsensusThreshold> endpoints report the same balance, it is considered the source of truth.
	if s.archivalState.getBalance() == "" {
		s.archivalState.updateArchivalBalance(archivalConsensusThreshold)
	}

	return nil
}
