package cosmos

import (
	"strconv"
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
)

// endpointStore maintains QoS data on the set of available endpoints
// for a CosmosSDK-based blockchain service.
//
// It performs two key tasks:
//  1. Storing the set of endpoints and their quality data.
//  2. Application of endpoints' observations to update the data on endpoints.
type endpointStore struct {
	logger polylog.Logger

	endpointsMu sync.RWMutex
	endpoints   map[protocol.EndpointAddr]endpoint
}

// updateEndpointsFromObservations creates/updates endpoint entries in the store based
// on the supplied observations. It returns the set of created/updated endpoints.
func (es *endpointStore) updateEndpointsFromObservations(
	cosmosObservations *qosobservations.CosmosSDKRequestObservations,
) map[protocol.EndpointAddr]endpoint {
	es.endpointsMu.Lock()
	defer es.endpointsMu.Unlock()

	endpointObservations := cosmosObservations.GetEndpointObservations()

	logger := es.logger.With(
		"qos_instance", "cosmos",
		"method", "UpdateEndpointsFromObservations",
	)

	logger.Info().Msgf("About to update endpoints from %d observations.", len(endpointObservations))

	updatedEndpoints := make(map[protocol.EndpointAddr]endpoint)
	for _, observation := range endpointObservations {
		if observation == nil {
			logger.Info().Msg("CosmosSDK EndpointStore received a nil observation. Skipping...")
			continue
		}

		endpointAddr := protocol.EndpointAddr(observation.EndpointAddr)

		logger := logger.With("endpoint_addr", endpointAddr)
		logger.Info().Msg("processing observation for endpoint.")

		// It is a valid scenario for an endpoint to not be present in the store.
		// e.g. when the first observation(s) are received for an endpoint.
		storedEndpoint := es.endpoints[endpointAddr]

		endpointWasMutated := applyObservation(
			&storedEndpoint,
			observation,
		)

		// If the observation did not mutate the endpoint, there is no need to update the stored endpoint entry.
		if !endpointWasMutated {
			logger.Info().Msg("endpoint was not mutated by observations. Skipping update of internal endpoint store.")
			continue
		}

		es.endpoints[endpointAddr] = storedEndpoint
		updatedEndpoints[endpointAddr] = storedEndpoint
	}

	return updatedEndpoints
}

// applyObservation updates the data stored regarding the endpoint using the supplied observation.
// It returns true if the observation was not unrecognized, i.e. mutated the endpoint.
//
// TODO_TECHDEBT(@adshmh): add a method to distinguish the following two scenarios:
//   - an endpoint that returned in invalid response.
//   - an endpoint with no/incomplete observations.
func applyObservation(
	endpoint *endpoint,
	observation *qosobservations.CosmosSDKEndpointObservation,
) (endpointWasMutated bool) {
	// If emptyResponse is not nil, the observation is for an empty response check.
	if observation.GetEmptyResponse() != nil {
		applyEmptyResponseObservation(endpoint)
		endpointWasMutated = true
		return
	}

	// If cometbftHealthResponse is not nil, the observation is for a cometbft health check.
	if observation.GetCometbftHealthResponse() != nil {
		applyCometbftHealthObservation(endpoint, observation.GetCometbftHealthResponse())
		endpointWasMutated = true
		return
	}

	// If cometbftStatusResponse is not nil, the observation is for a cometbft status check.
	if observation.GetCometbftStatusResponse() != nil {
		applyCometbftStatusObservation(endpoint, observation.GetCometbftStatusResponse())
		endpointWasMutated = true
		return
	}

	// If cosmosStatusResponse is not nil, the observation is for a cosmos status check.
	if observation.GetCosmosStatusResponse() != nil {
		applyCosmosStatusObservation(endpoint, observation.GetCosmosStatusResponse())
		endpointWasMutated = true
		return
	}

	// If unrecognizedResponse is not nil, the observation is for an unrecognized response.
	if unrecognizedResponse := observation.GetUnrecognizedResponse(); unrecognizedResponse != nil {
		applyUnrecognizedResponseObservation(endpoint, unrecognizedResponse)
		endpointWasMutated = true
		return
	}

	return endpointWasMutated // endpoint was not mutated by the observation
}

// applyEmptyResponseObservation updates the empty response check if a valid observation is provided.
func applyEmptyResponseObservation(endpoint *endpoint) {
	endpoint.hasReturnedEmptyResponse = true
	now := time.Now()
	endpoint.invalidResponseLastObserved = &now
}

// applyCometbftHealthObservation updates the health check if a valid observation is provided.
func applyCometbftHealthObservation(endpoint *endpoint, healthResponse *qosobservations.CometBFTHealthResponse) {
	healthy := healthResponse.GetHealthStatusResponse()
	endpoint.checkCometbftHealth = endpointCheckHealth{
		healthy:   &healthy,
		expiresAt: time.Now().Add(checkHealthInterval),
	}
}

// applyCometbftStatusObservation updates the status check if a valid observation is provided.
func applyCometbftStatusObservation(endpoint *endpoint, statusResponse *qosobservations.CometBFTStatusResponse) {
	chainID := statusResponse.GetChainIdResponse()
	catchingUp := statusResponse.GetCatchingUpResponse()
	blockHeight := parseBlockHeightResponse(statusResponse.GetLatestBlockHeightResponse())

	endpoint.checkCometbftStatus = endpointCheckStatus{
		chainID:           &chainID,
		catchingUp:        &catchingUp,
		latestBlockHeight: &blockHeight,
		expiresAt:         time.Now().Add(checkStatusInterval),
	}
}

// applyCosmosStatusObservation updates the cosmos status check if a valid observation is provided.
func applyCosmosStatusObservation(endpoint *endpoint, cosmosStatusResponse *qosobservations.CosmosSDKStatusResponse) {
	blockHeight := cosmosStatusResponse.GetLatestBlockHeightResponse()

	endpoint.checkCosmosStatus = endpointCheckCosmosStatus{
		latestBlockHeight: &blockHeight,
		expiresAt:         time.Now().Add(checkCosmosStatusInterval),
	}
}

// parseBlockHeightResponse parses the block height response from a string to a uint64.
// CosmosSDK returns block height as a string, so we need to parse it.
func parseBlockHeightResponse(response string) uint64 {
	if response == "" {
		return 0
	}

	// Convert string to uint64 - CosmosSDK returns block height as decimal string
	parsed, err := strconv.ParseUint(response, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

// applyUnrecognizedResponseObservation updates the invalid response check if a validation error is present.
func applyUnrecognizedResponseObservation(endpoint *endpoint, unrecognizedResponse *qosobservations.CosmosSDKUnrecognizedResponse) {
	// Check if the unrecognized response has a validation error set to something other than UNSPECIFIED
	// Note: For CosmosSDK, we don't have a validation error field in the unrecognized response,
	// so we'll just mark it as an invalid response
	endpoint.hasReturnedInvalidResponse = true
	now := time.Now()
	endpoint.invalidResponseLastObserved = &now
}
