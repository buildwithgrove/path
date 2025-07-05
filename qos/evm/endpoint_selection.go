package evm

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	shannonmetrics "github.com/buildwithgrove/path/metrics/protocol/shannon"
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

// SelectMultiple returns multiple endpoint addresses from the list of available endpoints.
// Available endpoints are filtered based on their validity first.
// Endpoints are selected with TLD diversity preference when possible.
// If maxCount is 0, it defaults to 1. If maxCount is greater than available endpoints, it returns all valid endpoints.
func (ss *serviceState) SelectMultiple(availableEndpoints protocol.EndpointAddrList, maxCount int) (protocol.EndpointAddrList, error) {
	logger := ss.logger.With("method", "SelectMultiple").
		With("chain_id", ss.serviceConfig.getEVMChainID()).
		With("service_id", ss.serviceConfig.GetServiceID()).
		With("max_count", maxCount)

	if maxCount <= 0 {
		maxCount = 1
	}

	logger.Info().Msgf("filtering %d available endpoints to select up to %d.", len(availableEndpoints), maxCount)

	filteredEndpointsAddr, err := ss.filterValidEndpoints(availableEndpoints)
	if err != nil {
		logger.Error().Err(err).Msg("error filtering endpoints")
		return nil, err
	}

	if len(filteredEndpointsAddr) == 0 {
		logger.Warn().Msgf("SELECTING RANDOM ENDPOINTS because all endpoints failed validation from: %s", availableEndpoints.String())
		// Select random endpoints as fallback
		var randomEndpoints protocol.EndpointAddrList
		countToSelect := maxCount
		if countToSelect > len(availableEndpoints) {
			countToSelect = len(availableEndpoints)
		}

		// Create a copy to avoid modifying the original slice
		availableCopy := make(protocol.EndpointAddrList, len(availableEndpoints))
		copy(availableCopy, availableEndpoints)

		// Fisher-Yates shuffle for random selection without replacement
		for i := 0; i < countToSelect; i++ {
			j := rand.Intn(len(availableCopy)-i) + i
			availableCopy[i], availableCopy[j] = availableCopy[j], availableCopy[i]
			randomEndpoints = append(randomEndpoints, availableCopy[i])
		}
		return randomEndpoints, nil
	}

	logger.Info().Msgf("filtered %d endpoints from %d available endpoints", len(filteredEndpointsAddr), len(availableEndpoints))

	// Use the diversity-aware selection
	return ss.selectEndpointsWithDiversity(filteredEndpointsAddr, maxCount), nil
}

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

// selectEndpointsWithDiversity selects endpoints with TLD diversity preference.
// This method is now used internally by SelectMultiple to ensure endpoint diversity.
func (ss *serviceState) selectEndpointsWithDiversity(availableEndpoints protocol.EndpointAddrList, maxCount int) protocol.EndpointAddrList {
	// Get endpoint URLs to extract TLD information
	endpointTLDs := ss.getEndpointTLDs(availableEndpoints)

	// Count unique TLDs for logging
	uniqueTLDs := make(map[string]bool)
	for _, tld := range endpointTLDs {
		if tld != "" {
			uniqueTLDs[tld] = true
		}
	}

	ss.logger.Debug().Msgf("Endpoint selection: %d available endpoints across %d unique TLDs, selecting up to %d endpoints",
		len(availableEndpoints), len(uniqueTLDs), maxCount)

	var selectedEndpoints protocol.EndpointAddrList
	usedTLDs := make(map[string]bool)
	remainingEndpoints := make(protocol.EndpointAddrList, len(availableEndpoints))
	copy(remainingEndpoints, availableEndpoints)

	// First pass: Try to select endpoints with different TLDs
	for i := 0; i < maxCount && len(remainingEndpoints) > 0; i++ {
		var selectedEndpoint protocol.EndpointAddr
		var err error

		// Try to find an endpoint with a different TLD
		if i > 0 && len(usedTLDs) > 0 {
			selectedEndpoint, err = ss.selectEndpointWithDifferentTLD(remainingEndpoints, endpointTLDs, usedTLDs)
			if err != nil {
				// Fallback to random selection if no different TLD found
				selectedEndpoint = remainingEndpoints[rand.Intn(len(remainingEndpoints))]
				err = nil
			}
		} else {
			// First endpoint: use random selection
			selectedEndpoint = remainingEndpoints[rand.Intn(len(remainingEndpoints))]
		}

		if err != nil {
			ss.logger.Warn().Err(err).Msgf("Failed to select endpoint %d, stopping selection", i+1)
			break
		}

		selectedEndpoints = append(selectedEndpoints, selectedEndpoint)

		// Track the TLD of the selected endpoint
		if tld, exists := endpointTLDs[selectedEndpoint]; exists {
			usedTLDs[tld] = true
			ss.logger.Debug().Msgf("Selected endpoint with TLD: %s (endpoint: %s)", tld, selectedEndpoint)
		}

		// Remove the selected endpoint from the remaining pool
		newRemainingEndpoints := make(protocol.EndpointAddrList, 0, len(remainingEndpoints)-1)
		for _, endpoint := range remainingEndpoints {
			if endpoint != selectedEndpoint {
				newRemainingEndpoints = append(newRemainingEndpoints, endpoint)
			}
		}
		remainingEndpoints = newRemainingEndpoints
	}

	// Count fallback selections (endpoints without TLD diversity)
	fallbackSelections := 0
	for _, endpoint := range selectedEndpoints {
		if tld, exists := endpointTLDs[endpoint]; exists && tld != "" {
			// Count how many endpoints use this TLD
			tldCount := 0
			for _, otherEndpoint := range selectedEndpoints {
				if otherTLD, exists := endpointTLDs[otherEndpoint]; exists && otherTLD == tld {
					tldCount++
				}
			}
			if tldCount > 1 {
				fallbackSelections++
			}
		}
	}

	ss.logger.Info().Msgf("Selected %d endpoints across %d different TLDs (diversity: %.1f%%, fallback selections: %d)",
		len(selectedEndpoints), len(usedTLDs),
		float64(len(usedTLDs))/float64(len(selectedEndpoints))*100, fallbackSelections)
	return selectedEndpoints
}

// getEndpointTLDs extracts TLD information from endpoint addresses
func (ss *serviceState) getEndpointTLDs(endpoints protocol.EndpointAddrList) map[protocol.EndpointAddr]string {
	endpointTLDs := make(map[protocol.EndpointAddr]string)

	// extractTLDFromEndpointAddr extracts effective TLD+1 from endpoint address
	extractTLDFromEndpointAddr := func(addr string) string {
		// Try direct URL parsing first
		if etld, err := shannonmetrics.ExtractEffectiveTLDPlusOne(addr); err == nil {
			return etld
		}

		// Handle embedded URLs (e.g., "supplier-https://example.com")
		if idx := strings.Index(addr, "http"); idx != -1 {
			if etld, err := shannonmetrics.ExtractEffectiveTLDPlusOne(addr[idx:]); err == nil {
				return etld
			}
		}

		// Fallback: try adding https:// prefix for domain-like strings
		parts := strings.FieldsFunc(addr, func(r rune) bool {
			return r == '-' || r == '_' || r == ' '
		})

		for _, part := range parts {
			if strings.Contains(part, ".") && !strings.HasPrefix(part, "http") {
				if etld, err := shannonmetrics.ExtractEffectiveTLDPlusOne("https://" + part); err == nil {
					return etld
				}
			}
		}

		return ""
	}

	for _, endpointAddr := range endpoints {
		if tld := extractTLDFromEndpointAddr(string(endpointAddr)); tld != "" {
			endpointTLDs[endpointAddr] = tld
		}
	}

	return endpointTLDs
}

// selectEndpointWithDifferentTLD attempts to select an endpoint with a TLD that hasn't been used yet
func (ss *serviceState) selectEndpointWithDifferentTLD(
	availableEndpoints protocol.EndpointAddrList,
	endpointTLDs map[protocol.EndpointAddr]string,
	usedTLDs map[string]bool,
) (protocol.EndpointAddr, error) {
	// Filter endpoints to only those with different TLDs
	var endpointsWithDifferentTLDs protocol.EndpointAddrList

	for _, endpoint := range availableEndpoints {
		if tld, exists := endpointTLDs[endpoint]; exists {
			if !usedTLDs[tld] {
				endpointsWithDifferentTLDs = append(endpointsWithDifferentTLDs, endpoint)
			}
		} else {
			// If we can't determine TLD, include it anyway
			endpointsWithDifferentTLDs = append(endpointsWithDifferentTLDs, endpoint)
		}
	}

	if len(endpointsWithDifferentTLDs) == 0 {
		return "", fmt.Errorf("no endpoints with different TLDs available")
	}

	// Select a random endpoint from the filtered list
	return endpointsWithDifferentTLDs[rand.Intn(len(endpointsWithDifferentTLDs))], nil
}
