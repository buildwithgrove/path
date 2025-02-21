package evm

import (
	"strconv"
	"time"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
)

// endpoint captures the details required to validate an EVM endpoint.
type endpoint struct {
	checks map[endpointCheckName]evmEndpointCheck
}

func newEndpoint() endpoint {
	return endpoint{
		checks: map[endpointCheckName]evmEndpointCheck{
			endpointCheckNameEmptyResponse: &endpointCheckEmptyResponse{},
			endpointCheckNameChainID:       &endpointCheckChainID{},
			endpointCheckNameBlockHeight:   &endpointCheckBlockNumber{},
		},
	}
}

// Validate returns an error if the endpoint is invalid.
// e.g. an endpoint without an observation of its response to an `eth_chainId` request is not considered valid.
func (e endpoint) Validate(serviceState *ServiceState) error {
	for _, check := range e.checks {
		if err := check.IsValid(serviceState); err != nil {
			return err
		}
	}
	return nil
}

// ApplyObservation updates the data stored regarding the endpoint using the supplied observation.
// It Returns true if the observation was not unrecognized, i.e. mutated the endpoint.
// TODO_TECHDEBT(@adshmh): add a method to distinguish the following two scenarios:
//   - an endpoint that returned in invalid response.
//   - an endpoint with no/incomplete observations.
func (e *endpoint) ApplyObservation(obs *qosobservations.EVMEndpointObservation) bool {
	// If emptyResponse is not nil, the observation is for an empty response check.
	if obs.GetEmptyResponse() != nil {
		return e.applyEmptyResponseObservation()
	}

	// If chainIDResponse is not nil, the observation is for a chainID check.
	if chainIDResponse := obs.GetChainIdResponse(); chainIDResponse != nil {
		return e.applyChainIDObservation(chainIDResponse)
	}

	// If blockNumberResponse is not nil, the obbservation is for a blockNumber check.
	if blockNumberResponse := obs.GetBlockNumberResponse(); blockNumberResponse != nil {
		return e.applyBlockNumberObservation(blockNumberResponse)
	}

	return false
}

// applyEmptyResponseObservation updates the empty response check if a valid observation is provided.
func (e *endpoint) applyEmptyResponseObservation() bool {
	e.checks[endpointCheckNameEmptyResponse] = &endpointCheckEmptyResponse{
		hasReturnedEmptyResponse: true, // An empty response is always invalid.
		expiresAt:                time.Now().Add(emptyResponseCheckInterval),
	}
	return true
}

// applyChainIDObservation updates the chain ID check if a valid observation is provided.
func (e *endpoint) applyChainIDObservation(chainIDResponse *qosobservations.EVMChainIDResponse) bool {
	observedChainID := chainIDResponse.GetChainIdResponse()
	e.checks[endpointCheckNameChainID] = &endpointCheckChainID{
		chainID:   &observedChainID,
		expiresAt: time.Now().Add(chainIDCheckInterval),
	}
	return true
}

// applyBlockNumberObservation updates the block number check if a valid observation is provided.
func (e *endpoint) applyBlockNumberObservation(blockNumberResponse *qosobservations.EVMBlockNumberResponse) bool {
	// base 0: use the string's prefix to determine its base.
	parsedBlockNumber, err := strconv.ParseUint(blockNumberResponse.GetBlockNumberResponse(), 0, 64)
	// The endpoint returned an invalid response to an `eth_blockNumber` request.
	// Explicitly set the parsedBlockNumberResponse to a zero value as the ParseUInt does not guarantee returning a 0 on all error cases.
	if err != nil {
		zero := uint64(0)
		e.checks[endpointCheckNameBlockHeight] = &endpointCheckBlockNumber{
			blockHeight: &zero,
			expiresAt:   time.Now().Add(blockHeightCheckInterval),
		}
		return true
	}

	e.checks[endpointCheckNameBlockHeight] = &endpointCheckBlockNumber{
		blockHeight: &parsedBlockNumber,
		expiresAt:   time.Now().Add(blockHeightCheckInterval),
	}
	return true
}

// GetBlockNumber returns the parsed block number value for the endpoint.
func (e endpoint) GetBlockNumber() (uint64, error) {
	observedBlockHeight, ok := e.checks[endpointCheckNameBlockHeight].(*endpointCheckBlockNumber)
	if !ok {
		return 0, errNoBlockNumberObs
	}
	if observedBlockHeight.blockHeight == nil {
		return 0, errNoBlockNumberObs
	}
	return *observedBlockHeight.blockHeight, nil
}
