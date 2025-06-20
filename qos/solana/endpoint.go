package solana

import (
	"fmt"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
)

// Expected value of the `result` field to a `getHealth` request.
const resultGetHealthOK = "ok"

// The errors below list all the possible basic validation errors on an endpoint.
var (
	errNoGetHealthObs                   = fmt.Errorf("endpoint has not had an observation of its response to a %q request", methodGetHealth)
	errInvalidGetHealthObs              = fmt.Errorf("endpoint responded incorrectly to a %q request, expected: %q", methodGetHealth, resultGetHealthOK)
	errNoGetEpochInfoObs                = fmt.Errorf("endpoint has not had an observation of its response to a %q request", methodGetEpochInfo)
	errInvalidGetEpochInfoHeightZeroObs = fmt.Errorf("endpoint responded with blockHeight of 0 to a %q request, expected a blockHeight of > 0", methodGetEpochInfo)
	errInvalidGetEpochInfoEpochZeroObs  = fmt.Errorf("endpoint responded with epoch of 0 to a %q request, expected an epoch of > 0", methodGetEpochInfo)
)

// endpoint captures the details required to validate a Solana endpoint.
type endpoint struct {
	// SolanaGetHealthResponse stores the result of processing the endpoint's response to a `getHealth` request.
	// A pointer is used to distinguish between the following scenarios:
	// 	1. There has NOT been an observation of the endpoint's response to a `getHealth` request, and
	// 	2. There has been an observation of the endpoint's response to a `getHealth` request.
	*qosobservations.SolanaGetHealthResponse

	// SolanaGetEpochInfoResponse stores the result of processing the endpoint's response to a `getEpochInfo` request.
	// A pointer is used to distinguish between the following scenarios:
	// 	1. There has NOT been an observation of the endpoint's response to a `getEpochInfo` request
	// 	2. There has been an observation of the endpoint's response to a `getEpochInfo` request
	*qosobservations.SolanaGetEpochInfoResponse

	// TODO_FUTURE: support archival endpoints.
}

// ValidateBasic checks if the endpoint has the required observations to be considered valid.
// Returns an error if the necessary responses are either lacking or invalid.
func (e endpoint) ValidateBasic() error {
	switch {

	case e.SolanaGetHealthResponse == nil:
		return errNoGetHealthObs

	case e.SolanaGetHealthResponse.Result != resultGetHealthOK:
		return fmt.Errorf("❌Invalid solana health response: %s :%w", e.SolanaGetHealthResponse.Result, errInvalidGetHealthObs)

	case e.SolanaGetEpochInfoResponse == nil:
		return errNoGetEpochInfoObs

	case e.SolanaGetEpochInfoResponse.BlockHeight == 0:
		return errInvalidGetEpochInfoHeightZeroObs

	case e.SolanaGetEpochInfoResponse.Epoch == 0:
		return errInvalidGetEpochInfoEpochZeroObs

	default:
		return nil
	}
}

// ApplyObservation updates the endpoint data using the provided observation.
// Returns true if the observation was recognized.
// IMPORTANT: This function mutates the endpoint.
func (e *endpoint) ApplyObservation(obs *qosobservations.SolanaEndpointObservation) bool {
	if epochInfoResponse := obs.GetGetEpochInfoResponse(); epochInfoResponse != nil {
		e.SolanaGetEpochInfoResponse = epochInfoResponse
		return true
	}

	if getHealthResponse := obs.GetGetHealthResponse(); getHealthResponse != nil {
		e.SolanaGetHealthResponse = getHealthResponse
		return true
	}

	return false
}
