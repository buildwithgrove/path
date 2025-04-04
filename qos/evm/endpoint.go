package evm

import (
	"errors"
	"strconv"
	"time"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
)

var errHasReturnedEmptyResponse = errors.New("endpoint is invalid: history of empty responses")

// endpoint captures the details required to validate an EVM endpoint.
// It contains all checks that should be run for the endpoint to validate
// it is providing a valid response to service requests.
type endpoint struct {
	hasReturnedEmptyResponse bool
	checkChainID             endpointCheckChainID
	checkBlockNumber         endpointCheckBlockNumber
	checkArchival            endpointCheckArchival
}

// newEndpoint initializes a new endpoint with the checks that should be run for the endpoint.
func newEndpoint() endpoint {
	return endpoint{
		checkChainID:     endpointCheckChainID{},
		checkBlockNumber: endpointCheckBlockNumber{},
		checkArchival:    endpointCheckArchival{},
	}
}

// getChecks returns the list of checks that should be run for the endpoint.
// The pre-selected endpoint address is assigned to the request context in this method.
func (e *endpoint) getChecks(es *EndpointStore) []gateway.RequestQoSContext {
	var checks []gateway.RequestQoSContext

	if e.checkChainID.shouldRun() {
		checks = append(checks, getEndpointCheck(es, e.checkChainID.getRequest()))
	}
	if e.checkBlockNumber.shouldRun() {
		checks = append(checks, getEndpointCheck(es, e.checkBlockNumber.getRequest()))
	}
	if e.checkArchival.shouldRun(es.serviceState.archivalState) {
		checks = append(checks, getEndpointCheck(es, e.checkArchival.getRequest(es.serviceState.archivalState)))
	}

	return checks
}

// ApplyObservation updates the data stored regarding the endpoint using the supplied observation.
// It Returns true if the observation was not unrecognized, i.e. mutated the endpoint.
//
// For archival balance observations:
// - Only updates the archival balance if the balance was observed at the specified archival block height
// - This ensures accurate historical balance validation at the specific block number
//
// TODO_TECHDEBT(@adshmh): add a method to distinguish the following two scenarios:
//   - an endpoint that returned in invalid response.
//   - an endpoint with no/incomplete observations.
func (e *endpoint) ApplyObservation(obs *qosobservations.EVMEndpointObservation, archivalBlockHeight string) bool {
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

	// If getBalanceResponse is not nil, the observation is for a getBalance check (which may be an archival check).
	if getBalanceResponse := obs.GetGetBalanceResponse(); getBalanceResponse != nil {
		// Only update the archival balance if the balance was observed at the archival block height.
		if balanceBlockHeight := getBalanceResponse.GetBlockNumber(); balanceBlockHeight == archivalBlockHeight {
			e.applyArchivalObservation(getBalanceResponse)
			return true
		}
	}

	return false
}

// applyEmptyResponseObservation updates the empty response check if a valid observation is provided.
func (e *endpoint) applyEmptyResponseObservation() {
	e.hasReturnedEmptyResponse = true
}

// applyChainIDObservation updates the chain ID check if a valid observation is provided.
func (e *endpoint) applyChainIDObservation(chainIDResponse *qosobservations.EVMChainIDResponse) {
	observedChainID := chainIDResponse.GetChainIdResponse()

	e.checkChainID = endpointCheckChainID{
		chainID:   &observedChainID,
		expiresAt: time.Now().Add(checkChainIDInterval),
	}
}

// applyBlockNumberObservation updates the block number check if a valid observation is provided.
func (e *endpoint) applyBlockNumberObservation(blockNumberResponse *qosobservations.EVMBlockNumberResponse) {
	e.checkBlockNumber = endpointCheckBlockNumber{
		parsedBlockNumberResponse: parseBlockNumberResponse(blockNumberResponse.GetBlockNumberResponse()),
		expiresAt:                 time.Now().Add(checkBlockNumberInterval),
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

// applyArchivalObservation updates the archival check if a valid observation is provided.
func (e *endpoint) applyArchivalObservation(archivalResponse *qosobservations.EVMGetBalanceResponse) {
	e.checkArchival = endpointCheckArchival{
		observedArchivalBalance: archivalResponse.GetBalance(),
		expiresAt:               time.Now().Add(checkArchivalInterval),
	}
}

// getBlockNumber returns the parsed block number value for the endpoint.
func (e endpoint) getBlockNumber() (uint64, error) {
	if e.checkBlockNumber.parsedBlockNumberResponse == nil {
		return 0, errNoBlockNumberObs
	}
	if *e.checkBlockNumber.parsedBlockNumberResponse == 0 {
		return 0, errInvalidBlockNumberObs
	}
	return *e.checkBlockNumber.parsedBlockNumberResponse, nil
}

// getArchivalBalance returns the observed archival balance for the endpoint at the archival block height.
// Returns an error if the endpoint hasn't yet returned an archival balance observation.
func (e endpoint) getArchivalBalance() (string, error) {
	if e.checkArchival.observedArchivalBalance == "" {
		return "", errNoArchivalBalanceObs
	}
	return e.checkArchival.observedArchivalBalance, nil
}
