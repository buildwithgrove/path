package evm

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/buildwithgrove/path/protocol"
)

var (
	errEmptyResponseObs         = errors.New("endpoint is invalid: history of empty responses")
	errRecentInvalidResponseObs = errors.New("endpoint is invalid: recent invalid response")
	errEmptyEndpointListObs     = errors.New("received empty list of endpoints to select from")
)

// TODO_UPNEXT(@adshmh): make the invalid response timeout duration configurable
// It is set to 5 minutes because that is the session time as of #321.
const invalidResponseTimeout = 5 * time.Minute

/* -------------------- QoS Valid Endpoint Selector -------------------- */
// This section contains methods for the `serviceState` struct
// but are kept in a separate file for clarity and readability.

// serviceState provides the endpoint selection capability required
// by the protocol package for handling a service request.
var _ protocol.EndpointSelector = &serviceState{}

// Select returns an endpoint address matching an entry from the list of available endpoints.
// available endpoints are filtered based on their validity first.
// A random endpoint is then returned from the filtered list of valid endpoints.
func (ss *serviceState) Select(availableEndpoints protocol.EndpointAddrList) (protocol.EndpointAddr, error) {
	logger := ss.logger.With("method", "Select").
		With("chain_id", ss.serviceConfig.getEVMChainID()).
		With("service_id", ss.serviceConfig.GetServiceID())

	logger.Info().Msgf("filtering %d available endpoints.", len(availableEndpoints))

	filteredEndpointsAddr, err := ss.filterValidEndpoints(availableEndpoints)
	if err != nil {
		logger.Error().Err(err).Msg("error filtering endpoints")
		return protocol.EndpointAddr(""), err
	}

	if len(filteredEndpointsAddr) == 0 {
		logger.Warn().Msgf("SELECTING A RANDOM ENDPOINT because all endpoints failed validation from: %s", availableEndpoints.String())
		randomAvailableEndpointAddr := availableEndpoints[rand.Intn(len(availableEndpoints))]
		return randomAvailableEndpointAddr, nil
	}

	logger.Info().Msgf("filtered %d endpoints from %d available endpoints", len(filteredEndpointsAddr), len(availableEndpoints))

	// TODO_FUTURE: consider ranking filtered endpoints, e.g. based on latency, rather than randomization.
	selectedEndpointAddr := filteredEndpointsAddr[rand.Intn(len(filteredEndpointsAddr))]
	return selectedEndpointAddr, nil
}

// filterValidEndpoints returns the subset of available endpoints that are valid
// according to previously processed observations.
func (ss *serviceState) filterValidEndpoints(availableEndpoints protocol.EndpointAddrList) (protocol.EndpointAddrList, error) {
	ss.endpointStore.endpointsMu.RLock()
	defer ss.endpointStore.endpointsMu.RUnlock()

	logger := ss.logger.With("method", "filterValidEndpoints").With("qos_instance", "evm")

	if len(availableEndpoints) == 0 {
		return nil, errEmptyEndpointListObs
	}

	logger.Info().Msgf("About to filter through %d available endpoints", len(availableEndpoints))

	// TODO_FUTURE: use service-specific metrics to add an endpoint ranking method
	// which can be used to assign a rank/score to a valid endpoint to guide endpoint selection.
	var filteredEndpointsAddr protocol.EndpointAddrList
	for _, availableEndpointAddr := range availableEndpoints {
		logger := logger.With("endpoint_addr", availableEndpointAddr)
		logger.Info().Msg("processing endpoint")

		endpoint, found := ss.endpointStore.endpoints[availableEndpointAddr]
		if !found {
			logger.Error().Msgf("❓ SKIPPING endpoint because it was not found in PATH's endpoint store: %s", availableEndpointAddr)
			continue
		}

		if err := ss.basicEndpointValidation(endpoint); err != nil {
			logger.Error().Err(err).Msgf("❌ SKIPPING endpoint because it failed basic validation: %s", availableEndpointAddr)
			continue
		}

		filteredEndpointsAddr = append(filteredEndpointsAddr, availableEndpointAddr)
		logger.Info().Msgf("✅ endpoint passed validation: %s", availableEndpointAddr)
	}

	return filteredEndpointsAddr, nil
}

// basicEndpointValidation returns an error if the supplied endpoint is not
// valid based on the perceived state of the EVM blockchain.
//
// It returns an error if:
// - The endpoint has returned an empty response in the past.
// - The endpoint has returned an invalid response within the last 30 minutes.
// - The endpoint's response to an `eth_chainId` request is not the expected chain ID.
// - The endpoint's response to an `eth_blockNumber` request is greater than the perceived block number.
// - The endpoint's archival check is invalid, if enabled.
func (ss *serviceState) basicEndpointValidation(endpoint endpoint) error {
	ss.serviceStateLock.RLock()
	defer ss.serviceStateLock.RUnlock()

	// Check if the endpoint has returned an empty response.
	if endpoint.hasReturnedEmptyResponse {
		return fmt.Errorf("empty response validation failed: %w", errEmptyResponseObs)
	}

	// Check if the endpoint has returned an invalid response within the invalid response timeout period.
	if endpoint.hasReturnedInvalidResponse && endpoint.invalidResponseLastObserved != nil {
		timeSinceInvalidResponse := time.Since(*endpoint.invalidResponseLastObserved)
		if timeSinceInvalidResponse < invalidResponseTimeout {
			return fmt.Errorf("recent invalid response validation failed (%.0f minutes ago): %w. Empty response: %t. Response validation error: %s",
				timeSinceInvalidResponse.Minutes(), errRecentInvalidResponseObs, endpoint.hasReturnedEmptyResponse, endpoint.invalidResponseError)
		}
	}

	// Check if the endpoint's block number is not more than the sync allowance behind the perceived block number.
	if err := ss.isBlockNumberValid(endpoint.checkBlockNumber); err != nil {
		return fmt.Errorf("block number validation failed: %w", err)
	}

	// Check if the endpoint's EVM chain ID matches the expected chain ID.
	if err := ss.isChainIDValid(endpoint.checkChainID); err != nil {
		return fmt.Errorf("chain ID validation failed: %w", err)
	}

	// Check if the endpoint has returned an archival balance for the perceived block number.
	if err := ss.archivalState.isArchivalBalanceValid(endpoint.checkArchival); err != nil {
		return fmt.Errorf("archival balance validation failed: %w", err)
	}

	return nil
}

// isBlockNumberValid returns an error if:
//   - The endpoint has not had an observation of its response to a `eth_blockNumber` request.
//   - The endpoint's block height is less than the perceived block height minus the sync allowance.
func (ss *serviceState) isBlockNumberValid(check endpointCheckBlockNumber) error {
	if ss.perceivedBlockNumber == 0 {
		return errNoBlockNumberObs
	}

	if check.parsedBlockNumberResponse == nil {
		return errNoBlockNumberObs
	}

	// Dereference pointer to show actual block number instead of memory address in error logs
	parsedBlockNumber := *check.parsedBlockNumberResponse

	// If the endpoint's block height is less than the perceived block height minus the sync allowance,
	// then the endpoint is behind the chain and should be filtered out.
	syncAllowance := ss.serviceConfig.getSyncAllowance()
	minAllowedBlockNumber := ss.perceivedBlockNumber - syncAllowance
	if parsedBlockNumber < minAllowedBlockNumber {
		return fmt.Errorf("%w: block number %d is outside the sync allowance relative to min allowed block number %d and sync allowance %d",
			errOutsideSyncAllowanceBlockNumberObs, parsedBlockNumber, minAllowedBlockNumber, syncAllowance)
	}

	return nil
}

// isChainIDValid returns an error if:
//   - The endpoint has not had an observation of its response to a `eth_chainId` request.
//   - The endpoint's chain ID does not match the expected chain ID in the service state.
func (ss *serviceState) isChainIDValid(check endpointCheckChainID) error {
	if check.chainID == nil {
		return errNoChainIDObs
	}

	// Dereference pointer to show actual chain ID instead of memory address in error logs
	chainID := *check.chainID

	expectedChainID := ss.serviceConfig.getEVMChainID()
	if chainID != expectedChainID {
		return fmt.Errorf("%w: chain ID %s does not match expected chain ID %s",
			errInvalidChainIDObs, chainID, expectedChainID)
	}
	return nil
}
