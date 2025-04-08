package evm

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
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
	}
}

// shouldChainIDCheckRun returns true if the chain ID check is not yet initialized or has expired.
func (ss *serviceState) shouldChainIDCheckRun(check endpointCheckChainID) bool {
	return check.expiresAt.IsZero() || check.expiresAt.Before(time.Now())
}

/* -------------------- QoS Endpoint Selector -------------------- */

// serviceState provides the endpoint selection capability required
// by the protocol package for handling a service request.
var _ protocol.EndpointSelector = &serviceState{}

// Select returns an endpoint address matching an entry from the list of available endpoints.
// The endpoints are filtered based on their validity and weighted based on latency.
func (ss *serviceState) Select(availableEndpoints []protocol.EndpointAddr) (protocol.EndpointAddr, error) {
	logger := ss.logger.With("method", "Select")
	logger.With("total_endpoints", len(availableEndpoints)).Info().Msg("selecting endpoint from available options")

	if len(availableEndpoints) == 0 {
		return "", errors.New("received empty list of endpoints to select from")
	}

	// Select the best endpoint based on filtering and ranking
	selectedEndpoint := ss.filterAndRankEndpoints(availableEndpoints)
	if selectedEndpoint != "" {
		return selectedEndpoint, nil
	}

	// If no valid endpoint found, select a random one as fallback
	logger.Warn().Msg("no valid endpoints found; using fallback endpoint")
	return availableEndpoints[rand.Intn(len(availableEndpoints))], nil
}

// filterAndRankEndpoints filters valid endpoints and selects one based on latency ranking.
// Returns the selected endpoint, or empty string if no valid endpoints found.
func (ss *serviceState) filterAndRankEndpoints(availableEndpoints []protocol.EndpointAddr) protocol.EndpointAddr {
	// Acquire lock once for the entire operation
	ss.endpointStore.endpointsMu.RLock()
	defer ss.endpointStore.endpointsMu.RUnlock()

	logger := ss.logger.With("method", "filterAndRankEndpoints")

	// Get valid endpoints with their weights based on latency
	endpointWeights := ss.getWeightedValidEndpoints(availableEndpoints)

	// If no valid endpoints found, return empty string
	if len(endpointWeights) == 0 {
		logger.Warn().Msg("no valid endpoints found")
		return ""
	}

	// Select an endpoint using weighted probability based on latency
	return ss.selectEndpointByLatency(endpointWeights)
}

/* -------------------- QoS Latency Weight Calculation -------------------- */

// Configuration constants for latency-based endpoint selection
const (
	// latencyPower determines how strongly to favor lower latency endpoints
	// Higher values give more weight to faster endpoints
	// 1.0 = linear inverse relationship (1/latency)
	// 1.5 = standard setting, moderately favors lower latency
	// 2.0 = strongly favors lower latency
	// TODO_IN_THIS_PR(@commoddity): make `latencyPower` configurable in config YAML
	latencyPower = 1.2

	// minLatencyMs prevents division by zero and excessive weighting
	// Any latency below this value will be capped to this minimum
	minLatencyMs = 1.0
)

// getWeightedValidEndpoints returns a map of valid endpoints to their weights.
// Endpoints with lower latency receive higher weights.
func (ss *serviceState) getWeightedValidEndpoints(endpoints []protocol.EndpointAddr) map[protocol.EndpointAddr]float64 {
	endpointsWithWeights := make(map[protocol.EndpointAddr]float64)

	// Process each endpoint - filtering and weight calculation in one pass
	for _, addr := range endpoints {
		endpoint, found := ss.endpointStore.endpoints[addr]
		if !found {
			ss.logger.Debug().Msgf("endpoint %s not found in store", addr)
			continue
		}

		// Endpoint validation happens in service_state.go
		if err := ss.validateEndpoint(endpoint); err != nil {
			ss.logger.Debug().Err(err).Msgf("endpoint %s failed validation", addr)
			continue
		}

		// Calculate and store the weight for valid endpoints
		endpointsWithWeights[addr] = ss.calculateLatencyWeight(endpoint, addr)
	}

	return endpointsWithWeights
}

// calculateLatencyWeight converts an endpoint's latency to a selection weight.
// Lower latency produces higher weight, making the endpoint more likely to be selected.
func (ss *serviceState) calculateLatencyWeight(endpoint endpoint, addr protocol.EndpointAddr) float64 {
	// Get latency with minimum threshold to prevent division by zero
	latency := max(endpoint.averageLatencyMs, minLatencyMs)

	// Weight = 1 / (latency ^ latencyPower)
	// Higher power = more aggressive favoring of lower latency
	weight := 1.0 / math.Pow(latency, latencyPower)

	ss.logger.Debug().
		Str("endpoint_addr", string(addr)).
		Float64("latency_ms", latency).
		Float64("weight", weight).
		Msg("calculated endpoint weight")

	return weight
}

// selectEndpointByLatency picks an endpoint based on weighted probability.
// Endpoints with lower latency (higher weights) have higher probability of selection.
func (ss *serviceState) selectEndpointByLatency(endpointWeights map[protocol.EndpointAddr]float64) protocol.EndpointAddr {
	// Short circuit for empty or single-entry maps
	if len(endpointWeights) == 0 {
		return ""
	}

	if len(endpointWeights) == 1 {
		for addr := range endpointWeights {
			return addr
		}
	}

	// Calculate total weight for probability distribution
	totalWeight := sumWeights(endpointWeights)

	// Log probability distribution
	ss.logWeightDistribution(endpointWeights, totalWeight)

	// Select an endpoint using weighted probability
	selected := ss.weightedRandomSelection(endpointWeights, totalWeight)
	if selected != "" {
		return selected
	}

	// Fallback to any endpoint if selection fails
	for addr := range endpointWeights {
		ss.logger.Warn().Msg("weighted selection fallback used")
		return addr
	}

	return ""
}

// sumWeights calculates the total of all weights in the map.
func sumWeights(weights map[protocol.EndpointAddr]float64) float64 {
	var total float64
	for _, weight := range weights {
		total += weight
	}
	return total
}

// logWeightDistribution logs the probability distribution of all endpoints.
func (ss *serviceState) logWeightDistribution(weights map[protocol.EndpointAddr]float64, totalWeight float64) {
	for addr, weight := range weights {
		probability := (weight / totalWeight) * 100

		ss.logger.Debug().
			Str("endpoint_addr", string(addr)).
			Float64("weight", weight).
			Float64("probability", probability).
			Msg("endpoint weight distribution")
	}
}

// weightedRandomSelection implements the weighted probability selection algorithm.
// Returns the selected endpoint or empty string if selection fails.
func (ss *serviceState) weightedRandomSelection(weights map[protocol.EndpointAddr]float64, totalWeight float64) protocol.EndpointAddr {
	// Generate a random number between 0 and totalWeight
	r := rand.Float64() * totalWeight

	// Find the endpoint whose weight range contains the random number
	var cumulativeWeight float64
	for addr, weight := range weights {
		cumulativeWeight += weight
		if r <= cumulativeWeight {
			ss.logger.Debug().
				Str("endpoint_addr", string(addr)).
				Msg("selected endpoint based on latency ranking")

			return addr
		}
	}

	// If we reach here, no endpoint was selected (rare, due to floating point imprecision)
	return ""
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
