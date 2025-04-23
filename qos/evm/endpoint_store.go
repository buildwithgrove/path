package evm

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
)

// endpointStore maintains QoS data on the set of available endpoints
// for an EVM-based blockchain service.
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
	evmObservations *qosobservations.EVMRequestObservations,
	archivalBlockHeight string,
) map[protocol.EndpointAddr]endpoint {
	es.endpointsMu.Lock()
	defer es.endpointsMu.Unlock()

	endpointObservations := evmObservations.GetEndpointObservations()

	logger := es.logger.With(
		"qos_instance", "evm",
		"method", "UpdateEndpointsFromObservations",
	)

	logger.Info().Msg(fmt.Sprintf("About to update endpoints from %d observations.", len(endpointObservations)))

	updatedEndpoints := make(map[protocol.EndpointAddr]endpoint)
	for _, observation := range endpointObservations {
		if observation == nil {
			logger.Info().Msg("EVM EndpointStore received a nil observation. Skipping...")
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
			archivalBlockHeight,
		)

		// If the observation did not mutate the endpoint, there is no need to update the stored endpoint entry.
		if !endpointWasMutated {
			logger.Info().Msg("endpoint was not mutated by observations. Skipping.")
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
// For archival balance observations:
// - Only updates the archival balance if the balance was observed at the specified archival block height
// - This ensures accurate historical balance validation at the specific block number
//
// TODO_TECHDEBT(@adshmh): add a method to distinguish the following two scenarios:
//   - an endpoint that returned in invalid response.
//   - an endpoint with no/incomplete observations.
func applyObservation(
	endpoint *endpoint,
	observation *qosobservations.EVMEndpointObservation,
	archivalBlockHeight string,
) (endpointWasMutated bool) {
	// If emptyResponse is not nil, the observation is for an empty response check.
	if observation.GetEmptyResponse() != nil {
		applyEmptyResponseObservation(endpoint)
		endpointWasMutated = true
		return
	}

	// If blockNumberResponse is not nil, the observation is for a blockNumber check.
	if observation.GetBlockNumberResponse() != nil {
		applyBlockNumberObservation(endpoint, observation.GetBlockNumberResponse())
		endpointWasMutated = true
		return
	}

	// If chainIDResponse is not nil, the observation is for a chainID check.
	if observation.GetChainIdResponse() != nil {
		applyChainIDObservation(endpoint, observation.GetChainIdResponse())
		endpointWasMutated = true
		return
	}

	// If getBalanceResponse is not nil, the observation is for a getBalance check (which may be an archival check).
	if getBalanceResponse := observation.GetGetBalanceResponse(); getBalanceResponse != nil {
		balanceBlockHeight := getBalanceResponse.GetBlockNumber()

		// Only update the archival balance if the balance was observed at the archival block height.
		if balanceBlockHeight == archivalBlockHeight {
			applyArchivalObservation(endpoint, getBalanceResponse)
			endpointWasMutated = true
			return
		}
	}

	return endpointWasMutated // endpoint was not mutated by the observation
}

// applyEmptyResponseObservation updates the empty response check if a valid observation is provided.
func applyEmptyResponseObservation(endpoint *endpoint) {
	endpoint.hasReturnedEmptyResponse = true
}

// applyBlockNumberObservation updates the block number check if a valid observation is provided.
func applyBlockNumberObservation(endpoint *endpoint, blockNumberResponse *qosobservations.EVMBlockNumberResponse) {
	parsedBlockNumberResponse := parseBlockNumberResponse(blockNumberResponse.GetBlockNumberResponse())

	endpoint.checkBlockNumber = endpointCheckBlockNumber{
		parsedBlockNumberResponse: parsedBlockNumberResponse,
	}
}

// parseBlockNumberResponse parses the block number response from a string to a uint64.
// eg. "0x3f8627c" -> 66609788
func parseBlockNumberResponse(response string) *uint64 {
	parsed, err := strconv.ParseUint(response, 0, 64)
	if err != nil {
		zero := uint64(0)
		return &zero
	}
	return &parsed
}

// applyChainIDObservation updates the chain ID check if a valid observation is provided.
func applyChainIDObservation(endpoint *endpoint, chainIDResponse *qosobservations.EVMChainIDResponse) {
	observedChainID := chainIDResponse.GetChainIdResponse()

	fmt.Println("DEBUG - observedChainID", observedChainID)

	endpoint.checkChainID = endpointCheckChainID{
		chainID:   &observedChainID,
		expiresAt: time.Now().Add(checkChainIDInterval),
	}
}

// applyArchivalObservation updates the archival check if a valid observation is provided.
func applyArchivalObservation(endpoint *endpoint, archivalResponse *qosobservations.EVMGetBalanceResponse) {
	endpoint.checkArchival = endpointCheckArchival{
		observedArchivalBalance: archivalResponse.GetBalance(),
		expiresAt:               time.Now().Add(checkArchivalInterval),
	}
}
