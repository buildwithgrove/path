package evm

import (
	"sync"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
)

// serviceState keeps the expected current state of the EVM blockchain based on
// the endpoints' responses to different requests.
type serviceState struct {
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
//
// It returns an error if:
// - The endpoint has returned an empty response in the past.
// - The endpoint's response to an `eth_chainId` request is not the expected chain ID.
// - The endpoint's response to an `eth_blockNumber` request is greater than the perceived block number.
// - The endpoint's archival check is invalid, if enabled.
func (ss *serviceState) ValidateEndpoint(endpoint endpoint, endpointAddr protocol.EndpointAddr) error {
	ss.serviceStateLock.RLock()
	defer ss.serviceStateLock.RUnlock()

	// TODO_TECHDEBT(@commoddity): move the endpoint validation methods to the service state
	// and pass them the endpoint rather than passing the service state to each endpoint method.

	// Ensure the response is not empty.
	if err := endpoint.validateEmptyResponse(); err != nil {
		return err
	}

	// Ensure the chain ID is valid.
	if err := endpoint.validateChainID(ss.chainID); err != nil {
		return err
	}

	// Ensure the service state's perceived block number is not ahead of the endpoint's block number.
	if err := endpoint.validateBlockNumber(ss.perceivedBlockNumber); err != nil {
		return err
	}

	// Ensure the endpoint has returned an archival balance for the perceived block number.
	// If the service does not require an archival check, this will always return a nil error.
	if err := endpoint.validateArchivalCheck(ss.archivalState); err != nil {
		return err
	}
	if err := endpoint.validateBlockNumber(s.perceivedBlockNumber); err != nil {
		return err
	}
	if s.shouldPerformArchivalCheck() {
		if err := endpoint.validateArchivalCheck(s.archivalState.balance, endpointAddr); err != nil {
			return err
		}
	}
	return nil
}

// UpdateFromEndpoints updates the service state using estimation(s) derived from the set of updated endpoints.
// This only includes the set of endpoints for which an observation was received.
func (ss *serviceState) UpdateFromEndpoints(updatedEndpoints map[protocol.EndpointAddr]endpoint) error {
	ss.serviceStateLock.Lock()
	defer ss.serviceStateLock.Unlock()

	// Initialize consensus map if it doesn't exist.
	if s.archivalState.balanceConsensus == nil {
		s.archivalState.balanceConsensus = make(map[string]int)
	}

	for endpointAddr, endpoint := range updatedEndpoints {
		logger := ss.logger.With(
			"endpoint_addr", endpointAddr,
			"perceived_block_number", ss.perceivedBlockNumber,
		)

		// Validate the endpoint's chain ID; do not update the perceived block number if the chain ID is invalid.
		if err := endpoint.validateChainID(ss.chainID); err != nil {
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
		ss.perceivedBlockNumber = blockNumber
	}

	// If archival checks are enabled for the service, update the archival state.
	if ss.archivalState.isEnabled() {
		// Update the archival state based on the perceived block number.
		// Note that when the expected balance at the archival block number is known, this becomes a no-op.
		ss.archivalState.updateArchivalState(ss.perceivedBlockNumber, updatedEndpoints)
	}

	return nil
}
