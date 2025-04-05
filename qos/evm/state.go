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
	serviceConfig    EVMServiceQoSConfig

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
func (ss *serviceState) ValidateEndpoint(endpoint endpoint) error {
	ss.serviceStateLock.RLock()
	defer ss.serviceStateLock.RUnlock()

	// Ensure the endpoint has not returned an empty response.
	if endpoint.hasReturnedEmptyResponse {
		return errHasReturnedEmptyResponse
	}

	// Ensure the endpoint's EVM chain ID matches the expected chain ID.
	evmChainID := ss.serviceConfig.getEVMChainID()
	if err := endpoint.checkChainID.isValid(evmChainID); err != nil {
		return err
	}

	// Ensure the endpoint's block number is not more than the sync allowance behind the perceived block number.
	perceivedBlockNumber := ss.perceivedBlockNumber
	syncAllowance := ss.serviceConfig.getSyncAllowance()
	if err := endpoint.checkBlockNumber.isValid(perceivedBlockNumber, syncAllowance); err != nil {
		return err
	}

	// Ensure the endpoint has returned an archival balance for the perceived block number.
	if err := endpoint.checkArchival.isValid(ss.archivalState); err != nil {
		return err
	}

	return nil
}

// UpdateFromEndpoints updates the service state using estimation(s) derived from the set of updated endpoints.
// This only includes the set of endpoints for which an observation was received.
func (ss *serviceState) UpdateFromEndpoints(updatedEndpoints map[protocol.EndpointAddr]endpoint) error {
	ss.serviceStateLock.Lock()
	defer ss.serviceStateLock.Unlock()

	for endpointAddr, endpoint := range updatedEndpoints {
		logger := ss.logger.With(
			"endpoint_addr", endpointAddr,
			"perceived_block_number", ss.perceivedBlockNumber,
		)

		// Validate the endpoint's chain ID; do not update the perceived block number if the chain ID is invalid.
		if err := endpoint.checkChainID.isValid(ss.serviceConfig.getEVMChainID()); err != nil {
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
