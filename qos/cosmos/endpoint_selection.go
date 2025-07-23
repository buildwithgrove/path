package cosmos

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/selector"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
)

var (
	errEmptyResponseObs         = errors.New("endpoint is invalid: history of empty responses")
	errRecentInvalidResponseObs = errors.New("endpoint is invalid: recent invalid response")
	errEmptyEndpointListObs     = errors.New("received empty list of endpoints to select from")

	// CosmosSDK-specific validation errors
	errOutsideSyncAllowanceBlockNumberObs = errors.New("endpoint block number is outside sync allowance")
)

// TODO_UPNEXT(@adshmh): make the invalid response timeout duration configurable
// It is set to 30 minutes because that is the session time as of #321.
const invalidResponseTimeout = 30 * time.Minute

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
		With("chain_id", ss.serviceQoSConfig.getCosmosSDKChainID()).
		With("service_id", ss.serviceQoSConfig.GetServiceID())

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

// SelectMultiple returns multiple endpoint addresses from the list of valid endpoints.
// Valid endpoints are determined by filtering the available endpoints based on their
// validity criteria. If numEndpoints is 0, it defaults to 1.
func (ss *serviceState) SelectMultiple(allAvailableEndpoints protocol.EndpointAddrList, numEndpoints uint) (protocol.EndpointAddrList, error) {
	logger := ss.logger.With("method", "SelectMultiple").With("chain_id", ss.serviceQoSConfig.getCosmosSDKChainID()).With("num_endpoints", numEndpoints)
	logger.Info().Msgf("filtering %d available endpoints to select up to %d.", len(allAvailableEndpoints), numEndpoints)

	filteredEndpointsAddr, err := ss.filterValidEndpoints(allAvailableEndpoints)
	if err != nil {
		logger.Error().Err(err).Msg("error filtering endpoints")
		return nil, err
	}

	// Select random endpoints as fallback
	if len(filteredEndpointsAddr) == 0 {
		logger.Warn().Msg("SELECTING RANDOM ENDPOINTS because all endpoints failed validation.")
		return selector.RandomSelectMultiple(allAvailableEndpoints, numEndpoints), nil
	}

	// Select up to numEndpoints endpoints from filtered list
	logger.Info().Msgf("filtered %d endpoints from %d available endpoints", len(filteredEndpointsAddr), len(allAvailableEndpoints))
	return selector.SelectEndpointsWithDiversity(logger, filteredEndpointsAddr, numEndpoints), nil
}

// filterValidEndpoints returns the subset of available endpoints that are valid
// according to previously processed observations.
func (ss *serviceState) filterValidEndpoints(availableEndpoints protocol.EndpointAddrList) (protocol.EndpointAddrList, error) {
	ss.endpointStore.endpointsMu.RLock()
	defer ss.endpointStore.endpointsMu.RUnlock()

	logger := ss.logger.With("method", "filterValidEndpoints").With("qos_instance", "cosmossdk")

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
			logger.Error().Msgf("❓ SKIPPING endpoint %s because it was not found in PATH's endpoint store.", availableEndpointAddr)
			continue
		}

		if err := ss.basicEndpointValidation(endpoint); err != nil {
			logger.Error().Err(err).Msgf("❌ SKIPPING %s endpoint because it failed basic validation: %v", availableEndpointAddr, err)
			continue
		}

		filteredEndpointsAddr = append(filteredEndpointsAddr, availableEndpointAddr)
		logger.Info().Msgf("✅ endpoint %s passed validation", availableEndpointAddr)
	}

	return filteredEndpointsAddr, nil
}

// basicEndpointValidation returns an error if the supplied endpoint is not
// valid based on the perceived state of the CosmosSDK blockchain.
//
// It returns an error if:
// - The endpoint has returned an empty response in the past.
// - The endpoint has returned an unmarshaling error within the last 30 minutes.
// - The endpoint has returned an invalid response within the last 30 minutes.
// - The endpoint's response to a `/status` request indicates an invalid chain ID.
// - The endpoint's response to a `/status` request indicates it's catching up.
// - The endpoint's response to a `/status` request shows block height outside sync allowance.
// - The endpoint's response to a `/health` request indicates it's unhealthy.
func (ss *serviceState) basicEndpointValidation(endpoint endpoint) error {
	ss.serviceStateLock.RLock()
	defer ss.serviceStateLock.RUnlock()

	// Check if the endpoint has returned an empty response.
	if endpoint.hasReturnedEmptyResponse {
		return fmt.Errorf("empty response validation failed: %w", errEmptyResponseObs)
	}

	// Check if the endpoint has returned an unmarshaling error within the last 30 minutes.
	if endpoint.hasReturnedUnmarshalingError && endpoint.invalidResponseLastObserved != nil {
		timeSinceInvalidResponse := time.Since(*endpoint.invalidResponseLastObserved)
		if timeSinceInvalidResponse < invalidResponseTimeout {
			return fmt.Errorf("recent unmarshaling error validation failed (%.0f minutes ago): %w",
				timeSinceInvalidResponse.Minutes(), errRecentInvalidResponseObs)
		}
	}

	// Check if the endpoint has returned an invalid response within the invalid response timeout period.
	if endpoint.hasReturnedInvalidResponse && endpoint.invalidResponseLastObserved != nil {
		timeSinceInvalidResponse := time.Since(*endpoint.invalidResponseLastObserved)
		if timeSinceInvalidResponse < invalidResponseTimeout {
			return fmt.Errorf("recent response validation failed (%.0f minutes ago): %w",
				timeSinceInvalidResponse.Minutes(), errRecentInvalidResponseObs)
		}
	}

	// Get the RPC types supported by the CosmosSDK service.
	supportedAPIs := ss.serviceQoSConfig.getSupportedAPIs()

	// If the service supports CometBFT, validate the endpoint's CometBFT checks.
	if _, ok := supportedAPIs[sharedtypes.RPCType_COMET_BFT]; ok {
		if err := ss.validateEndpointCometBFTChecks(endpoint); err != nil {
			return fmt.Errorf("cometBFT validation failed: %w", err)
		}
	}

	// If the service supports CosmosSDK, validate the endpoint's CosmosSDK checks.
	if _, ok := supportedAPIs[sharedtypes.RPCType_REST]; ok {
		if err := ss.validateEndpointCosmosSDKChecks(endpoint); err != nil {
			return fmt.Errorf("cosmos SDK validation failed: %w", err)
		}
	}

	return nil
}

// validateEndpointCometBFTChecks validates the endpoint's CometBFT checks.
// Checks:
//   - Health status
//   - Status information
func (ss *serviceState) validateEndpointCometBFTChecks(endpoint endpoint) error {
	// Check if the endpoint's health status is valid.
	if err := ss.isCometBFTHealthValid(endpoint.checkCometBFTHealth); err != nil {
		return fmt.Errorf("cometBFT health validation failed: %w", err)
	}

	// Check if the endpoint's status information is valid.
	if err := ss.isCometBFTStatusValid(endpoint.checkCometBFTStatus); err != nil {
		return fmt.Errorf("cometBFT status validation failed: %w", err)
	}

	return nil
}

// isCometBFTHealthValid returns an error if:
//   - The endpoint has not had an observation of its response to a `/health` request.
//   - The endpoint's health check indicates it's unhealthy.
func (ss *serviceState) isCometBFTHealthValid(check endpointCheckCometBFTHealth) error {
	healthy, err := check.GetHealthy()
	if err != nil {
		return fmt.Errorf("%w: %v", errNoHealthObs, err)
	}

	if !healthy {
		return fmt.Errorf("%w: endpoint reported unhealthy status", errInvalidHealthObs)
	}

	return nil
}

// isCometBFTStatusValid returns an error if:
//   - The endpoint has not had an observation of its response to a `/status` request.
//   - The endpoint's chain ID does not match the expected chain ID.
//   - The endpoint is catching up to the network.
//   - The endpoint's block height is outside the sync allowance.
func (ss *serviceState) isCometBFTStatusValid(check endpointCheckCometBFTStatus) error {
	// Check chain ID
	chainID, err := check.GetChainID()
	if err != nil {
		return fmt.Errorf("%w: %v", errNoStatusObs, err)
	}

	expectedChainID := ss.serviceQoSConfig.getCosmosSDKChainID()
	if chainID != expectedChainID {
		return fmt.Errorf("%w: chain ID %s does not match expected chain ID %s",
			errInvalidChainIDObs, chainID, expectedChainID)
	}

	// Check if the endpoint is catching up to the network.
	catchingUp, err := check.GetCatchingUp()
	if err != nil {
		return fmt.Errorf("%w: %v", errNoStatusObs, err)
	}

	if catchingUp {
		return fmt.Errorf("%w: endpoint is catching up to the network", errCatchingUpObs)
	}

	// Check if the endpoint's block height is within the sync allowance.
	latestBlockHeight, err := check.GetLatestBlockHeight()
	if err != nil {
		return fmt.Errorf("%w: %v", errNoStatusObs, err)
	}
	if err := ss.validateBlockHeightSyncAllowance(latestBlockHeight); err != nil {
		return fmt.Errorf("cometBFT block height sync allowance validation failed: %w", err)
	}

	return nil
}

// validateEndpointCosmosSDKChecks validates the endpoint's CosmosSDK checks.
// Checks:
//   - Status information (block height)
func (ss *serviceState) validateEndpointCosmosSDKChecks(endpoint endpoint) error {
	// Check if the endpoint's Cosmos SDK status information is valid.
	if err := ss.isCosmosStatusValid(endpoint.checkCosmosStatus); err != nil {
		return fmt.Errorf("cosmos SDK status validation failed: %w", err)
	}

	return nil
}

// isCosmosStatusValid returns an error if:
//   - The endpoint has not had an observation of its response to a `/cosmos/base/node/v1beta1/status` request.
//   - The endpoint's block height is outside the sync allowance.
func (ss *serviceState) isCosmosStatusValid(check endpointCheckCosmosStatus) error {
	// Check if the endpoint's block height is within the sync allowance.
	latestBlockHeight, err := check.GetHeight()
	if err != nil {
		return fmt.Errorf("%w: %v", errNoStatusObs, err)
	}
	if err := ss.validateBlockHeightSyncAllowance(latestBlockHeight); err != nil {
		return fmt.Errorf("cosmos SDK block height sync allowance validation failed: %w", err)
	}

	return nil
}

// validateBlockHeightSyncAllowance returns an error if:
//   - The endpoint's block height is outside the latest block height minus the sync allowance.
func (ss *serviceState) validateBlockHeightSyncAllowance(latestBlockHeight uint64) error {
	syncAllowance := ss.serviceQoSConfig.getSyncAllowance()
	minAllowedBlockNumber := ss.perceivedBlockNumber - syncAllowance
	if latestBlockHeight < minAllowedBlockNumber {
		return fmt.Errorf("%w: block number %d is outside the sync allowance relative to min allowed block number %d and sync allowance %d",
			errOutsideSyncAllowanceBlockNumberObs, latestBlockHeight, minAllowedBlockNumber, syncAllowance)
	}

	return nil
}
