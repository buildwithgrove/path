package evm

import (
	"strconv"
	"time"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
)

// endpoint captures the details required to validate an EVM endpoint.
type endpoint struct {
	checks map[endpointCheckName]*evmQualityCheck
}

// newEndpoint initializes a new endpoint with the checks that should be run for the endpoint.
func newEndpoint(es *EndpointStore) endpoint {
	return endpoint{
		checks: map[endpointCheckName]*evmQualityCheck{
			checkNameEmptyResponse: {
				// TODO_MVP(@commoddity): should we provide for a mechanism to un-sanction an endpoint that has returned an empty response?
				requestContext: nil, // An empty response disqualifies an endpoint for an entire session.
				check:          &endpointCheckEmptyResponse{},
			},
			checkNameChainID: {
				requestContext: getEndpointCheck(es, withChainIDCheck),
				check:          &endpointCheckChainID{},
			},
			checkNameBlockNumber: {
				requestContext: getEndpointCheck(es, withBlockNumberCheck),
				check:          &endpointCheckBlockNumber{},
			},
		},
	}
}

// getChecks returns the list of checks that should be run for the endpoint.
// The pre-selected endpoint address is assigned to the request context in this method.
func (e *endpoint) getChecks(endpointAddr protocol.EndpointAddr) []gateway.RequestQoSContext {
	var checks []gateway.RequestQoSContext

	for _, check := range e.checks {
		// The check should run if both are true:
		// 1. The check has a non-nil request context.
		// 2. The check was just initialized or has expired.
		if check.shouldRun() {
			requestContext := check.getRequestContext()
			requestContext.setPreSelectedEndpointAddr(endpointAddr)
			checks = append(checks, requestContext)
		}
	}

	return checks
}

// Validate returns an error if the endpoint is invalid.
// e.g. an endpoint without an observation of its response to an `eth_chainId` request is not considered valid.
func (e endpoint) Validate(serviceState *ServiceState) error {
	for _, check := range e.checks {
		if err := check.isValid(serviceState); err != nil {
			return err
		}
	}
	return nil
}

// ApplyObservation updates the data stored regarding the endpoint using the supplied observation.
// It returns true if the observation was not unrecognized, i.e. mutated the endpoint.
// TODO_TECHDEBT(@adshmh): add a method to distinguish the following two scenarios:
//   - an endpoint that returned in invalid response.
//   - an endpoint with no/incomplete observations.
func (e *endpoint) ApplyObservation(obs *qosobservations.EVMEndpointObservation) bool {
	// If emptyResponse is not nil, the observation is for an empty response check.
	if obs.GetEmptyResponse() != nil {
		e.applyEmptyResponseObservation()
		return true
	}

	// If chainIDResponse is not nil, the observation is for a chainID check.
	if chainIDResponse := obs.GetChainIdResponse(); chainIDResponse != nil {
		e.applyChainIDObservation(chainIDResponse)
		return true
	}

	// If blockNumberResponse is not nil, the obbservation is for a blockNumber check.
	if blockNumberResponse := obs.GetBlockNumberResponse(); blockNumberResponse != nil {
		e.applyBlockNumberObservation(blockNumberResponse)
		return true
	}

	return false
}

// applyEmptyResponseObservation updates the empty response check if a valid observation is provided.
func (e *endpoint) applyEmptyResponseObservation() {
	e.checks[checkNameEmptyResponse].check = &endpointCheckEmptyResponse{
		hasReturnedEmptyResponse: true, // An empty response is always invalid.
	}
}

// applyChainIDObservation updates the chain ID check if a valid observation is provided.
func (e *endpoint) applyChainIDObservation(chainIDResponse *qosobservations.EVMChainIDResponse) {
	observedChainID := chainIDResponse.GetChainIdResponse()

	e.checks[checkNameChainID].check = &endpointCheckChainID{
		chainID:   &observedChainID,
		expiresAt: time.Now().Add(checkChainIDInterval),
	}
}

// applyBlockNumberObservation updates the block number check if a valid observation is provided.
func (e *endpoint) applyBlockNumberObservation(blockNumberResponse *qosobservations.EVMBlockNumberResponse) {
	// base 0: use the string's prefix to determine its base.
	parsedBlockNumber, err := strconv.ParseUint(blockNumberResponse.GetBlockNumberResponse(), 0, 64)

	// The endpoint returned an invalid response to an `eth_blockNumber` request.
	// Explicitly set the parsedBlockNumberResponse to a zero value as the ParseUInt
	// does not guarantee returning a 0 on all error cases.
	if err != nil {
		zero := uint64(0)
		e.checks[checkNameBlockNumber].check = &endpointCheckBlockNumber{
			blockNumber: &zero,
			expiresAt:   time.Now().Add(checkBlockNumberInterval),
		}
		return
	}

	e.checks[checkNameBlockNumber].check = &endpointCheckBlockNumber{
		blockNumber: &parsedBlockNumber,
		expiresAt:   time.Now().Add(checkBlockNumberInterval),
	}
}

// GetBlockNumber returns the parsed block number value for the endpoint.
func (e endpoint) GetBlockNumber() (uint64, error) {
	blockNumberCheck, ok := e.checks[checkNameBlockNumber]
	if !ok {
		return 0, errNoBlockNumberObs
	}

	observedBlockNumber := blockNumberCheck.check.(*endpointCheckBlockNumber)
	if observedBlockNumber.blockNumber == nil {
		return 0, errNoBlockNumberObs
	}

	return *observedBlockNumber.blockNumber, nil
}
