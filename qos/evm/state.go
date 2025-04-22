package evm

import (
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
)

// serviceState keeps the expected current state of the EVM blockchain
// based on the endpoints' responses to different requests.
type serviceState struct {
	logger polylog.Logger

	serviceStateLock sync.RWMutex
	serviceConfig    EVMServiceQoSConfig

	// perceivedBlockNumber is the perceived current block number
	// based on endpoints' responses to `eth_blockNumber` requests.
	// It is calculated as the maximum of block height reported by
	// any of the endpoints for the service.
	//
	// See the following link for more details:
	// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_blocknumber
	perceivedBlockNumber uint64

	// archivalState contains the current state of the EVM archival check for the service.
	archivalState archivalState
}

// TODO_FUTURE: add an endpoint ranking method which can be used to
// assign a rank/score to a valid endpoint to guide endpoint selection.
//
// ValidateEndpoint returns an error if the supplied endpoint is not
// valid based on the perceived state of the EVM blockchain.
//
// It returns an error if:
// - The endpoint has returned an empty response in the past.
// - The endpoint's response to an `eth_chainId` request is not the expected chain ID.
// - The endpoint's response to an `eth_blockNumber` request is greater than the perceived block number.
// - The endpoint's archival check is invalid, if enabled.
func (ss *serviceState) validateEndpoint(endpoint endpoint) error {
	ss.serviceStateLock.RLock()
	defer ss.serviceStateLock.RUnlock()

	// Ensure the endpoint has not returned an empty response.
	if endpoint.hasReturnedEmptyResponse {
		return errHasReturnedEmptyResponse
	}

	// Ensure the endpoint's block number is not more than the sync allowance behind the perceived block number.
	if err := ss.isBlockNumberValid(endpoint.checkBlockNumber); err != nil {
		return err
	}

	// Ensure the endpoint's EVM chain ID matches the expected chain ID.
	if err := ss.isChainIDValid(endpoint.checkChainID); err != nil {
		return err
	}

	// Ensure the endpoint has returned an archival balance for the perceived block number.
	if err := ss.archivalState.isArchivalBalanceValid(endpoint.checkArchival); err != nil {
		return err
	}

	return nil
}

// isValid returns an error if the endpoint's block height is less
// than the perceived block height minus the sync allowance.
func (ss *serviceState) isBlockNumberValid(check endpointCheckBlockNumber) error {
	if ss.perceivedBlockNumber == 0 {
		return errNoBlockNumberObs
	}

	// If the endpoint's block height is less than the perceived block height minus the sync allowance,
	// then the endpoint is behind the chain and should be filtered out.
	minAllowedBlockNumber := ss.perceivedBlockNumber - ss.serviceConfig.getSyncAllowance()

	if *check.parsedBlockNumberResponse < minAllowedBlockNumber {
		return errInvalidBlockNumberObs
	}

	return nil
}

// isChainIDValid returns an error if the endpoint's chain ID does not
// match the expected chain ID in the service state.
func (ss *serviceState) isChainIDValid(check endpointCheckChainID) error {
	if check.chainID == nil {
		return errNoChainIDObs
	}
	if *check.chainID != ss.serviceConfig.getEVMChainID() {
		return errInvalidChainIDObs
	}
	return nil
}

// shouldChainIDCheckRun returns true if the chain ID check is not yet initialized or has expired.
func (ss *serviceState) shouldChainIDCheckRun(check endpointCheckChainID) bool {
	return check.expiresAt.IsZero() || check.expiresAt.Before(time.Now())
}

// updateFromEndpoints updates the service state using estimation(s) derived from the set of updated
// endpoints. This only includes the set of endpoints for which an observation was received.
func (ss *serviceState) updateFromEndpoints(updatedEndpoints map[protocol.EndpointAddr]endpoint) error {
	ss.serviceStateLock.Lock()
	defer ss.serviceStateLock.Unlock()

	for endpointAddr, endpoint := range updatedEndpoints {
		logger := ss.logger.With(
			"endpoint_addr", endpointAddr,
			"perceived_block_number", ss.perceivedBlockNumber,
		)

		// Do not update the perceived block number if the chain ID is invalid.
		if err := ss.isChainIDValid(endpoint.checkChainID); err != nil {
			logger.Info().Err(err).Msg("Skipping endpoint with invalid chain id")
			continue
		}

		// Retrieve the block number from the endpoint.
		blockNumber, err := endpoint.checkBlockNumber.getBlockNumber()
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
		// When the expected balance at the archival block number is known, this becomes a no-op.
		ss.archivalState.updateArchivalState(ss.perceivedBlockNumber, updatedEndpoints)
	}

	return nil
}
