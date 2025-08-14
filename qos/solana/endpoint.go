package solana

import (
	"fmt"
	"time"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
)

// Expected value of the `result` field to a `getHealth` request.
const resultGetHealthOK = "ok"

const (
	// TODO_TECHDEBT(@adshmh): Make this configurable.
	validationErrorWindow = 30 * time.Minute
)

// The errors below list all the possible basic validation errors on an endpoint.
var (
	errNoGetHealthObs                   = fmt.Errorf("endpoint has not had an observation of its response to a %q request", methodGetHealth)
	errInvalidGetHealthObs              = fmt.Errorf("endpoint responded incorrectly to a %q request, expected: %q", methodGetHealth, resultGetHealthOK)
	errNoGetEpochInfoObs                = fmt.Errorf("endpoint has not had an observation of its response to a %q request", methodGetEpochInfo)
	errInvalidGetEpochInfoHeightZeroObs = fmt.Errorf("endpoint responded with blockHeight of 0 to a %q request, expected a blockHeight of > 0", methodGetEpochInfo)
	errInvalidGetEpochInfoEpochZeroObs  = fmt.Errorf("endpoint responded with epoch of 0 to a %q request, expected an epoch of > 0", methodGetEpochInfo)
	errRecentValidationError            = fmt.Errorf("endpoint has recent JSON-RPC validation errors")
)

// TODO_TECHDEBT(@adshmh): Include a Sanctions mechanism to handle endpoints with dishonest behavior, e.g. using public RPCs.
//
// endpoint captures details required to validate a Solana endpoint.
type endpoint struct {
	// SolanaGetHealthResponse stores result of processing endpoint's `getHealth` response.
	// Pointer distinguishes between no observation vs. observed response scenarios.
	*qosobservations.SolanaGetHealthResponse

	// SolanaGetEpochInfoResponse stores result of processing endpoint's `getEpochInfo` response.
	// Pointer distinguishes between no observation vs. observed response scenarios.
	*qosobservations.SolanaGetEpochInfoResponse

	// latestValidationError tracks most recent JSON-RPC response validation error
	latestValidationError *qosobservations.JsonRpcResponseValidationError

	// TODO_FUTURE: support archival endpoints.
}

// validateBasic checks if endpoint has required observations to be valid.
// Returns error if necessary responses are lacking, invalid, or have recent validation errors.
func (e endpoint) validateBasic() error {
	// Check for recent validation errors first
	if e.hasRecentValidationErrors() {
		return errRecentValidationError
	}

	switch {
	case e.SolanaGetHealthResponse == nil:
		return errNoGetHealthObs

	case e.Result != resultGetHealthOK:
		return fmt.Errorf("❌Invalid solana health response: %s :%w", e.Result, errInvalidGetHealthObs)

	case e.SolanaGetEpochInfoResponse == nil:
		return errNoGetEpochInfoObs

	case e.BlockHeight == 0:
		return errInvalidGetEpochInfoHeightZeroObs

	case e.Epoch == 0:
		return errInvalidGetEpochInfoEpochZeroObs

	default:
		return nil
	}
}

// hasRecentValidationErrors checks if endpoint has validation error within the configured window.
func (e endpoint) hasRecentValidationErrors() bool {
	if e.latestValidationError == nil {
		return false
	}

	cutoff := time.Now().Add(-validationErrorWindow)
	return e.latestValidationError.Timestamp.AsTime().After(cutoff)
}

// applyObservation updates endpoint data using provided observation.
// Returns true if observation was recognized.
// IMPORTANT: This function mutates the endpoint.
func (e *endpoint) applyObservation(obs *qosobservations.SolanaEndpointObservation) bool {
	if epochInfoResponse := obs.GetGetEpochInfoResponse(); epochInfoResponse != nil {
		e.SolanaGetEpochInfoResponse = epochInfoResponse
		return true
	}

	if getHealthResponse := obs.GetGetHealthResponse(); getHealthResponse != nil {
		e.SolanaGetHealthResponse = getHealthResponse
		return true
	}

	if unrecognizedResponse := obs.GetUnrecognizedResponse(); unrecognizedResponse != nil {
		// Update latest validation error if observation contains more recent error
		if validationError := unrecognizedResponse.ValidationError; validationError != nil {
			if e.latestValidationError == nil ||
				validationError.Timestamp.AsTime().After(e.latestValidationError.Timestamp.AsTime()) {
				e.latestValidationError = validationError
			}
		}
		return true
	}

	return false
}
