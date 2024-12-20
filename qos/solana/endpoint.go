package solana

import (
	"fmt"
)

const (
	// Expected value of the `result` field to a `getHealth` request.
	resultGetHealthOK = "ok"
)

var (
	// The errors below list all the possible basic validation errors on an endpoint.
	errNoGetHealthObs                   = fmt.Errorf("endpoint has not had an observation of its response to a %q request", methodGetHealth)
	errInvalidGetHealthObs              = fmt.Errorf("endpoint responded incorrectly to a %q request, expected: %q", methodGetHealth, resultGetHealthOK)
	errNoGetEpochInfoObs                = fmt.Errorf("endpoint has not had an observation of its response to a %q request", methodGetEpochInfo)
	errInvalidGetEpochInfoHeightZeroObs = fmt.Errorf("endpoint responded with blockHeight of 0 to a %q request, expected a blockHeight of > 0", methodGetEpochInfo)
	errInvalidGetEpochInfoEpochZeroObs  = fmt.Errorf("endpoint responded with epoch of 0 to a %q request, expected an epoch of > 0", methodGetEpochInfo)
)

// endpoint captures the details required to validate a Solana endpoint.
type endpoint struct {
	// GetHealthResult stores the result of processing the endpoint's response to a `getHealth` request.
	// A pointer is used to distinguish between the following scenarios:
	// A. There has NOT been an observation of the endpoint's response to a `getHealth` request, and
	// B. There has been an observation of the endpoint's response to a `getHealth` request.
	GetHealthResult *string

	// EpochInfo stores the result of processing the endpoint's response to a `getEpochInfo` request.
	// A pointer is used to distinguish between the following scenarios:
	// A. There has NOT been an observation of the endpoint's response to a `getEpochInfo` request, and
	// B. There has been an observation the endpoint's response to a `getEpochInfo` request.
	GetEpochInfoResult *epochInfo

	// TODO_FUTURE: support archival endpoints.
}

// ValidateBasic returns an error if the endpoint is invalid regardless of the state of the service.
// e.g. an endpoint without an observation of its response to a `GetHealth` request is not considered valid.
func (e endpoint) ValidateBasic() error {
	switch {
	case e.GetHealthResult == nil:
		return errNoGetHealthObs
	case *e.GetHealthResult != resultGetHealthOK:
		return fmt.Errorf("invalid response: %s :%w", *e.GetHealthResult, errInvalidGetHealthObs)
	case e.GetEpochInfoResult == nil:
		return errNoGetEpochInfoObs
	case e.GetEpochInfoResult.BlockHeight == 0:
		return errInvalidGetEpochInfoHeightZeroObs
	case e.GetEpochInfoResult.Epoch == 0:
		return errInvalidGetEpochInfoEpochZeroObs
	default:
		return nil
	}
}
