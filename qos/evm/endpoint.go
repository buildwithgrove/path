package evm

import (
	"errors"
	"fmt"
	"strconv"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
)

// The errors below list all the possible validation errors on an endpoint.
var (
	// Shared error message templates
	errNoObservationStr   = "endpoint has not had an observation of its response to a %q request"
	errInvalidResponseStr = "endpoint returned an invalid response to a %q request"

	// empty response errors
	errHasReturnedEmptyResponse = errors.New("endpoint is invalid: history of empty responses")

	// chainID check errors
	errNoChainIDObs      = fmt.Errorf(errNoObservationStr, methodChainID)
	errInvalidChainIDObs = fmt.Errorf(errInvalidResponseStr, methodChainID)

	// block number check errors
	errNoBlockNumberObs      = fmt.Errorf(errNoObservationStr, methodBlockNumber)
	errInvalidBlockNumberObs = fmt.Errorf(errInvalidResponseStr, methodBlockNumber)
	errBlockNumberTooLow     = "endpoint has block height %d, perceived block height is %d"

	// archival check errors
	errNoArchivalBalanceObs      = fmt.Errorf("endpoint has not returned an archival balance response to a %q request", methodGetBalance)
	errInvalidArchivalBalanceObs = "endpoint has archival balance %s, expected archival balance %s"
)

// endpoint captures the details required to validate an EVM endpoint.
type endpoint struct {
	// TODO_TECHDEBT(@adshmh): Persist this state across restarts to maintain endpoint exclusions.
	//
	// hasReturnedEmptyResponse tracks endpoints that have returned empty responses.
	// These endpoints are excluded from selection until service restart.
	hasReturnedEmptyResponse bool

	// chainIDResponse stores the result of processing the endpoint's response to an `eth_chainId` request.
	// It is nil if there has NOT been an observation of the endpoint's response to an `eth_chainId` request.
	chainIDResponse *string

	// parsedBlockNumberResponse stores the result of processing the endpoint's response to an `eth_blockNumber` request.
	// It is nil if there has NOT been an observation of the endpoint's response to an `eth_blockNumber` request.
	parsedBlockNumberResponse *uint64

	// observedArchivalBalance stores the result of processing the endpoint's response
	// to an `eth_getBalance` request for a specific contract address at a specific block number.
	observedArchivalBalance string
}

func (e *endpoint) validateEmptyResponse() error {
	if e.hasReturnedEmptyResponse {
		return errHasReturnedEmptyResponse
	}
	return nil
}

func (e *endpoint) validateChainID(chainID string) error {
	if e.chainIDResponse == nil {
		return errNoChainIDObs
	}
	if *e.chainIDResponse != chainID {
		return fmt.Errorf("%s. expected: %s, got: %s", errInvalidChainIDObs, chainID, *e.chainIDResponse)
	}
	return nil
}

// validateBlockNumber checks if the perceived block number is valid.
// The perceived block number:
//   - Valid if it is less than or equal to the last observed block number by this endpoint.
//   - Invalid if it is greater than the last observed block number by this endpoint.
func (e *endpoint) validateBlockNumber(perceivedBlockNumber uint64) error {
	_, err := e.getBlockNumber()
	if err != nil {
		return err
	}
	// TODO_IMPROVE(@commoddity): implement Allowance for block number check to allow blocks within
	// a certain range of the current block number to still be considered valid to serve requests.
	if *e.parsedBlockNumberResponse < perceivedBlockNumber {
		return fmt.Errorf(errBlockNumberTooLow, e.parsedBlockNumberResponse, perceivedBlockNumber)
	}
	return nil
}

// getBlockNumber returns the parsed block number value for the endpoint.
func (e endpoint) getBlockNumber() (uint64, error) {
	if e.parsedBlockNumberResponse == nil {
		return 0, errNoBlockNumberObs
	}
	if *e.parsedBlockNumberResponse == 0 {
		return 0, errInvalidBlockNumberObs
	}
	return *e.parsedBlockNumberResponse, nil
}

// validateArchivalCheck validates that the endpoint has returned an archival balance for the perceived block number.

// If the archival check is not enabled for the service, this will always return a nil error.
func (e *endpoint) validateArchivalCheck(archivalState archivalState) error {
	// Return nil if the archival check is not enabled for the service.
	if archivalState.archivalCheckConfig.IsEmpty() {
		return nil
	}

	// No archival balance observation was made yet
	if e.observedArchivalBalance == "" {
		return errNoArchivalBalanceObs
	}

	// The archival balance observed does not match the expected archival balance
	if e.observedArchivalBalance != archivalState.expectedBalance {
		return fmt.Errorf(errInvalidArchivalBalanceObs, e.observedArchivalBalance, archivalState.expectedBalance)
	}

	return nil
}

// getArchivalBalance returns the observed archival balance for the endpoint at the archival block height.
// Returns an error if the endpoint hasn't yet returned an archival balance observation.
func (e endpoint) getArchivalBalance() (string, error) {
	if e.observedArchivalBalance == "" {
		return "", errNoArchivalBalanceObs
	}
	return e.observedArchivalBalance, nil
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
	if obs.GetEmptyResponse() != nil {
		e.hasReturnedEmptyResponse = true
		return true
	}

	if chainIDResponse := obs.GetChainIdResponse(); chainIDResponse != nil {
		observedChainID := chainIDResponse.GetChainIdResponse()
		e.chainIDResponse = &observedChainID
		return true
	}

	if blockNumberResponse := obs.GetBlockNumberResponse(); blockNumberResponse != nil {
		e.parsedBlockNumberResponse = parseBlockNumberResponse(blockNumberResponse.GetBlockNumberResponse())
		return true
	}

	if getBalanceResponse := obs.GetGetBalanceResponse(); getBalanceResponse != nil {
		// Only update the archival balance if the balance was observed at the archival block height.
		if balanceBlockHeight := getBalanceResponse.GetBlockNumber(); balanceBlockHeight == archivalBlockHeight {
			e.observedArchivalBalance = getBalanceResponse.GetBalance()
			return true
		}
	}

	return false
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
