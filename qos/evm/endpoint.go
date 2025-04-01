package evm

import (
	"strconv"
	"time"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
)

// endpoint captures the details required to validate an EVM endpoint.
// It contains all checks that should be run for the endpoint to validate
// it is providing a valid response to service requests.
type endpoint struct {
	checkEmptyResponse *endpointCheckEmptyResponse
	checkChainID       *endpointCheckChainID
	checkBlockNumber   *endpointCheckBlockNumber
	checkArchival      *endpointCheckArchival
}

// newEndpoint initializes a new endpoint with the checks that should be run for the endpoint.
func newEndpoint() endpoint {
	return endpoint{
		checkEmptyResponse: &endpointCheckEmptyResponse{},
		checkChainID:       &endpointCheckChainID{},
		checkBlockNumber:   &endpointCheckBlockNumber{},
		checkArchival:      &endpointCheckArchival{},
	}
}

// getChecks returns the list of checks that should be run for the endpoint.
// The pre-selected endpoint address is assigned to the request context in this method.
func (e *endpoint) getChecks(es *EndpointStore) []gateway.RequestQoSContext {
	var checks []gateway.RequestQoSContext

	if e.checkEmptyResponse.shouldRun() {
		checks = append(checks, getEndpointCheck(es, e.checkEmptyResponse))
	}
	if e.checkChainID.shouldRun() {
		checks = append(checks, getEndpointCheck(es, e.checkChainID))
	}
	if e.checkBlockNumber.shouldRun() {
		checks = append(checks, getEndpointCheck(es, e.checkBlockNumber))
	}
	if e.checkArchival.shouldRun() {
		checks = append(checks, getEndpointCheck(es, e.checkArchival))
	}

	return checks
}

// Validate returns an error if the endpoint is invalid.
// e.g. an endpoint without an observation of its response to an `eth_chainId` request is not considered valid.
func (e endpoint) Validate(serviceState *ServiceState) error {
	if err := e.checkEmptyResponse.isValid(serviceState); err != nil {
		return err
	}
	if err := e.checkChainID.isValid(serviceState); err != nil {
		return err
	}
	if err := e.checkBlockNumber.isValid(serviceState); err != nil {
		return err
	}
	if err := e.checkArchival.isValid(serviceState); err != nil {
		return err
	}
	return nil
}

// ApplyObservation updates the data stored regarding the endpoint using the supplied observation.
// It returns true if the observation was not unrecognized, i.e. mutated the endpoint.
// TODO_TECHDEBT(@adshmh): add a method to distinguish the following two scenarios:
//   - an endpoint that returned in invalid response.
//   - an endpoint with no/incomplete observations.
func (e *endpoint) ApplyObservation(obs *qosobservations.EVMEndpointObservation, shouldPerformArchivalCheck bool) bool {
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

	// If blockNumberResponse is not nil, the observation is for a blockNumber check.
	if blockNumberResponse := obs.GetBlockNumberResponse(); blockNumberResponse != nil {
		e.applyBlockNumberObservation(blockNumberResponse)
		return true
	}

	// shouldPerformArchivalCheck is true if the observation is for an archival check.
	if shouldPerformArchivalCheck {
		// If archivalResponse is not nil, the observation is for an archival check.
		if archivalResponse := obs.GetArchivalResponse(); archivalResponse != nil {
			e.applyArchivalObservation(archivalResponse)
			return true
		}
	}

	return false
}

// applyEmptyResponseObservation updates the empty response check if a valid observation is provided.
func (e *endpoint) applyEmptyResponseObservation() {
	e.checkEmptyResponse = &endpointCheckEmptyResponse{
		hasReturnedEmptyResponse: true, // An empty response is always invalid.
	}
}

// applyChainIDObservation updates the chain ID check if a valid observation is provided.
func (e *endpoint) applyChainIDObservation(chainIDResponse *qosobservations.EVMChainIDResponse) {
	observedChainID := chainIDResponse.GetChainIdResponse()

	e.checkChainID = &endpointCheckChainID{
		chainID:   &observedChainID,
		expiresAt: time.Now().Add(checkChainIDInterval),
	}
}

// applyBlockNumberObservation updates the block number check if a valid observation is provided.
func (e *endpoint) applyBlockNumberObservation(blockNumberResponse *qosobservations.EVMBlockNumberResponse) {
	e.checkBlockNumber = &endpointCheckBlockNumber{
		blockNumber: parseBlockNumberResponse(blockNumberResponse.GetBlockNumberResponse()),
		expiresAt:   time.Now().Add(checkBlockNumberInterval),
	}
}

// parseBlockNumberResponse parses the block number response from the endpoint.
// It returns a pointer to the parsed block number, or a pointer to a zero value if the response is invalid.
func parseBlockNumberResponse(response string) *uint64 {
	// base 0: use the string's prefix to determine its base.
	parsed, err := strconv.ParseUint(response, 0, 64)

	// The endpoint returned an invalid response to an `eth_blockNumber` request.
	// Explicitly set the parsedBlockNumberResponse to a zero value as the ParseUInt
	// does not guarantee returning a 0 on all error cases.
	if err != nil {
		zero := uint64(0)
		return &zero
	}
	return &parsed
}

// applyArchivalObservation updates the archival check if a valid observation is provided.
func (e *endpoint) applyArchivalObservation(archivalResponse *qosobservations.EVMArchivalResponse) {
	e.checkArchival = &endpointCheckArchival{
		archivalBalance: archivalResponse.GetBalance(),
		expiresAt:       time.Now().Add(checkArchivalInterval),
	}
}

// GetBlockNumber returns the parsed block number value for the endpoint.
func (e endpoint) getBlockNumber() (uint64, error) {
	if e.checkBlockNumber.blockNumber == nil {
		return 0, errNoBlockNumberObs
	}

	return *e.checkBlockNumber.blockNumber, nil
}

// getArchivalBalance returns the parsed archival balance value for the endpoint.
func (e endpoint) getArchivalBalance() (string, error) {
	if e.checkArchival.archivalBalance == "" {
		return "", errNoArchivalBalanceObs
	}

	return e.checkArchival.archivalBalance, nil
}
