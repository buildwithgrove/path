package evm

import (
	"fmt"
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// The serviceState struct maintains the expected current state of the EVM blockchain
// based on the endpoints' responses to different requests.
//
// It has three main responsibilities:
//  1. Generate QoS endpoint checks for the hydrator
//  2. Select a valid endpoint for a service request
//  3. Update the service state from endpoint observations
type serviceState struct {
	logger polylog.Logger

	serviceStateLock sync.RWMutex
	serviceConfig    EVMServiceQoSConfig

	// endpointStore maintains the set of available endpoints and their quality data
	endpointStore *endpointStore

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

/* -------------------- QoS Endpoint Check Generator -------------------- */

// serviceState provides the endpoint check generator required by
// the gateway package to augment endpoints' quality data,
// using synthetic service requests.
var _ gateway.QoSEndpointCheckGenerator = &serviceState{}

// GetRequiredQualityChecks returns the list of quality checks required for an endpoint.
// It is called in the `gateway/hydrator.go` file on each run of the hydrator.
func (ss *serviceState) GetRequiredQualityChecks(endpointAddr protocol.EndpointAddr) []gateway.RequestQoSContext {
	ss.endpointStore.endpointsMu.RLock()
	defer ss.endpointStore.endpointsMu.RUnlock()

	endpoint := ss.endpointStore.endpoints[endpointAddr]

	var checks = []gateway.RequestQoSContext{
		// Block number check should always run
		ss.getEndpointCheck(endpoint.checkBlockNumber.getRequest()),
	}

	// Chain ID check runs infrequently as an endpoint's EVM chain ID is very unlikely to change regularly.
	if ss.shouldChainIDCheckRun(endpoint.checkChainID) {
		checks = append(checks, ss.getEndpointCheck(endpoint.checkChainID.getRequest()))
	}

	// Archival check runs infrequently as the result of a request for an archival block is not expected to change regularly.
	// Additionally, this check will only run if the service is configured to perform archival checks.
	if ss.archivalState.shouldArchivalCheckRun(endpoint.checkArchival) {
		checks = append(
			checks,
			ss.getEndpointCheck(endpoint.checkArchival.getRequest(ss.archivalState)),
		)
	}

	return checks
}

// getEndpointCheck prepares a request context for a specific endpoint check.
// The pre-selected endpoint address is assigned to the request context in the `endpoint.getChecks` method.
// It is called in the individual `check_*.go` files to build the request context.
// getEndpointCheck prepares a request context for a specific endpoint check.
func (ss *serviceState) getEndpointCheck(jsonrpcReq jsonrpc.Request) *requestContext {
	return &requestContext{
		logger:       ss.logger,
		serviceState: ss,
		jsonrpcReq:   jsonrpcReq,
		// Set the chain and Service ID: this is required to generate observations with the correct chain ID.
		chainID:   ss.serviceConfig.getEVMChainID(),
		serviceID: ss.serviceConfig.GetServiceID(),
	}
}

// shouldChainIDCheckRun returns true if the chain ID check is not yet initialized or has expired.
func (ss *serviceState) shouldChainIDCheckRun(check endpointCheckChainID) bool {
	return check.expiresAt.IsZero() || check.expiresAt.Before(time.Now())
}

/* -------------------- QoS Endpoint Validation -------------------- */

// validateEndpoint returns an error if the supplied endpoint is not
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
		return fmt.Errorf("endpoint is invalid: history of empty responses")
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

/* -------------------- QoS Endpoint State Updater -------------------- */

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
