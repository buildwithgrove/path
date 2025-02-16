package evm

import (
	"errors"
	"fmt"
	"strconv"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
)

// The errors below list all the possible validation errors on an endpoint.
var (
	errNoChainIDObs             = fmt.Errorf("endpoint has not had an observation of its response to a %q request", methodChainID)
	errInvalidChainIDObs        = fmt.Errorf("endpoint returned an invalid response to a %q request", methodChainID)
	errNoBlockNumberObs         = fmt.Errorf("endpoint has not had an observation of its response to a %q request", methodBlockNumber)
	errInvalidBlockNumberObs    = fmt.Errorf("endpoint returned an invalid response to a %q request", methodBlockNumber)
	errHasReturnedEmptyResponse = errors.New("endpoint is invalid: history of empty responses")
)

// endpoint captures the details required to validate an EVM endpoint.
type endpoint struct {
	// hasReturnedEmptyResponse indicates if the endpoint has ever returned an empty response.
	// Endpoints that return empty responses are marked invalid and excluded from selection.
	hasReturnedEmptyResponse bool

	// chainIDResponse stores the result of processing the endpoint's response to an `eth_chainId` request.
	// It is nil if there has NOT been an observation of the endpoint's response to an `eth_chainId` request.
	chainIDResponse *string

	// parsedBlockNumberResponse stores the result of processing the endpoint's response to an `eth_blockNumber` request.
	// It is nil if there has NOT been an observation of the endpoint's response to an `eth_blockNumber` request.
	parsedBlockNumberResponse *uint64

	// TODO_FUTURE: support archival endpoints.
}

// Validate returns an error if the endpoint is invalid.
// e.g. an endpoint without an observation of its response to an `eth_chainId` request is not considered valid.
func (e endpoint) Validate(chainID string) error {
	switch {
	case e.hasReturnedEmptyResponse:
		return errHasReturnedEmptyResponse
	case e.chainIDResponse == nil:
		return errNoChainIDObs
	case *e.chainIDResponse != chainID:
		return fmt.Errorf("invalid response: %s expected %s :%w", *e.chainIDResponse, chainID, errInvalidChainIDObs)
	case e.parsedBlockNumberResponse == nil:
		return errNoBlockNumberObs
	case *e.parsedBlockNumberResponse == 0:
		return errInvalidBlockNumberObs
	default:
		return nil
	}
}

// ApplyObservation updates the data stored regarding the endpoint using the supplied observation.
// It Returns true if the observation was not unrecognized, i.e. mutated the endpoint.
// TODO_TECHDEBT(@adshmh): add a method to distinguish the following two scenarios:
//   - an endpoint that returned in invalid response.
//   - an endpoint with no/incomplete observations.
func (e *endpoint) ApplyObservation(obs *qosobservations.EVMEndpointObservation) bool {
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
		// base 0: use the string's prefix to determine its base.
		parsedBlockNumber, err := strconv.ParseUint(blockNumberResponse.GetBlockNumberResponse(), 0, 64)
		// The endpoint returned an invalid response to an `eth_blockNumber` request.
		// Explicitly set the parsedBlockNumberResponse to a zero value as the ParseUInt does not guarantee returning a 0 on all error cases.
		if err != nil {
			zero := uint64(0)
			e.parsedBlockNumberResponse = &zero
			return true
		}

		e.parsedBlockNumberResponse = &parsedBlockNumber
		return true
	}

	return false
}

// GetBlockNumber returns the parsed block number value for the endpoint.
func (e endpoint) GetBlockNumber() (uint64, error) {
	if e.parsedBlockNumberResponse == nil {
		return 0, errNoBlockNumberObs
	}

	return *e.parsedBlockNumberResponse, nil
}
