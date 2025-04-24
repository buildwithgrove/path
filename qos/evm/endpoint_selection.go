package evm

import (
	"errors"
	"math"
	"math/rand"

	"github.com/buildwithgrove/path/protocol"
)

/* -------------------- QoS Endpoint Selector -------------------- */

// serviceState provides the endpoint selection capability required
// by the protocol package for handling a service request.
var _ protocol.EndpointSelector = &serviceState{}

// Select returns an endpoint address matching an entry from the list of available endpoints.
// The endpoints are filtered based on their validity and weighted based on latency.
func (ss *serviceState) Select(availableEndpoints []protocol.EndpointAddr) (protocol.EndpointAddr, error) {
	logger := ss.logger.With(
		"method", "Select",
		"total_endpoints", len(availableEndpoints),
	)
	logger.Info().Msg("selecting endpoint from available options")

	if len(availableEndpoints) == 0 {
		return "", errors.New("received empty list of endpoints from protocol")
	}

	// Select the best endpoint based on filtering and ranking
	selectedValidEndpoint, err := ss.selectValidEndpointByWeight(availableEndpoints)
	if err != nil {
		logger.Error().Err(err).Msg("all endpoints failed validation; selecting a random endpoint")
		return availableEndpoints[rand.Intn(len(availableEndpoints))], nil
	}

	return selectedValidEndpoint, nil
}

// selectValidEndpointByWeight filters valid endpoints and selects one based on latency ranking.
// Returns the selected endpoint, or empty string if no valid endpoints found.
func (ss *serviceState) selectValidEndpointByWeight(availableEndpoints []protocol.EndpointAddr) (protocol.EndpointAddr, error) {
	// Acquire lock once for the entire operation
	ss.endpointStore.endpointsMu.RLock()
	defer ss.endpointStore.endpointsMu.RUnlock()

	// Get valid endpoints with their weights based on latency
	validEndpointsWithWeights := ss.getValidEndpointsWithWeights(availableEndpoints)
	if len(validEndpointsWithWeights) == 0 {
		return "", errors.New("no valid endpoints found")
	}

	// Select a valid endpoint using weighted probability based on latency
	return ss.selectValidEndpointByLatency(validEndpointsWithWeights), nil
}

/* -------------------- QoS Latency Weight Calculation -------------------- */

// Configuration constants for latency-based endpoint selection
const (
	// latencyPower determines how strongly to favor lower latency endpoints
	// Higher values give more weight to faster endpoints
	// 1.0 = linear inverse relationship (1/latency)
	// 1.5 = standard setting, moderately favors lower latency
	// 2.0 = strongly favors lower latency
	//
	// TODO_IN_THIS_PR(@commoddity): make `latencyPower` configurable in config YAML
	latencyPower = 1.5

	// minLatencyMs prevents division by zero and excessive weighting
	// Any latency below this value will be capped to this minimum
	minLatencyMs = 1.0
)

// getValidEndpointsWithWeights returns a map of valid endpoints to their weights.
//  1. Endpoints that fail QoS check validation are filtered out.
//  2. Valid endpoints receive a weight based on their latency.
func (ss *serviceState) getValidEndpointsWithWeights(availableEndpoints []protocol.EndpointAddr) map[protocol.EndpointAddr]float64 {
	validEndpointsWithWeights := make(map[protocol.EndpointAddr]float64)

	// Process each endpoint - filtering and weight calculation in one pass
	for _, addr := range availableEndpoints {
		endpoint, found := ss.endpointStore.endpoints[addr]
		if !found {
			ss.logger.Debug().Msgf("endpoint %s not found in store", addr)
			continue
		}

		// validate endpoint passes all QoS checks
		if err := ss.validateEndpoint(endpoint); err != nil {
			ss.logger.Debug().Err(err).Msgf("endpoint %s failed validation", addr)
			continue
		}

		// Calculate and store the weight for valid endpoints
		validEndpointsWithWeights[addr] = ss.calculateLatencyWeight(endpoint, addr)
	}

	return validEndpointsWithWeights
}

// calculateLatencyWeight converts an endpoint's latency into a weight.
// Lower latency = higher weight = better chance of selection, while still
// allowing higher latency endpoints to be selected.
func (ss *serviceState) calculateLatencyWeight(endpoint endpoint, addr protocol.EndpointAddr) float64 {
	// Get latency with minimum threshold to prevent division by zero
	latency := max(endpoint.averageLatencyMs, minLatencyMs)

	// Weight = 1 / (latency ^ latencyPower)
	// The `latencyPower` parameter controls how much weight is given to latency differences.
	// Higher values make the weight more sensitive to latency differences,
	// meaning faster endpoints are more likely to be selected.
	weight := 1.0 / math.Pow(latency, latencyPower)

	ss.logger.Debug().
		Str("endpoint_addr", string(addr)).
		Float64("latency_ms", latency).
		Float64("weight", weight).
		Msg("calculated endpoint weight")

	return weight
}

// selectValidEndpointByLatency chooses an endpoint based on response time.
//
// How it works:
//   - If there's only one endpoint, it's automatically chosen
//   - Otherwise, we use a weighted lottery where faster endpoints
//     have better chances of being picked
func (ss *serviceState) selectValidEndpointByLatency(validEndpointsWithWeights map[protocol.EndpointAddr]float64) protocol.EndpointAddr {
	// Short circuit for empty or single-entry maps
	if len(validEndpointsWithWeights) == 0 {
		return ""
	}

	if len(validEndpointsWithWeights) == 1 {
		for addr := range validEndpointsWithWeights {
			return addr
		}
	}

	// Calculate total weight for probability distribution
	totalWeight := sumWeights(validEndpointsWithWeights)

	// Select an endpoint using weighted probability
	selected := ss.weightedRandomSelection(validEndpointsWithWeights, totalWeight)
	if selected != "" {
		return selected
	}

	// Fallback to any endpoint if selection fails
	for addr := range validEndpointsWithWeights {
		ss.logger.Warn().Msg("weighted selection fallback used")
		return addr
	}

	return ""
}

// sumWeights calculates the total of all weights in the map.
// This function simply adds up all the individual endpoint weights to get the total sum.
// The total weight is essential for:
// 1. Calculating the probability percentage of each endpoint being selected
// 2. Setting the upper bound for the random selection in weightedRandomSelection
// 3. Normalizing the weights so they can be interpreted as probabilities
func sumWeights(weights map[protocol.EndpointAddr]float64) float64 {
	var total float64
	for _, weight := range weights {
		total += weight
	}
	return total
}

// weightedRandomSelection picks an endpoint based on its weight.
// It works like a weighted lottery:
// 1. Generate a random number
// 2. Pick the endpoint whose weight range contains that number
//
// Faster endpoints (with lower latency) get larger weight ranges,
// making them more likely to be chosen, but slower endpoints
// still have some chance of selection.
//
// This helps spread the load while favoring better-performing endpoints.
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
