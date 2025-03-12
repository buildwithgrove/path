// request.go
package evm

import (
	"github.com/buildwithgrove/path/observation/qos"
)

// request defines methods needed for request metrics collection.
// Abstracts proto-specific details from metrics logic.
type request interface {
	// GetRequestValidationError returns the validation error if any.
	// A nil return value indicates the request is valid.
	// A non-nil value indicates the request is invalid, with the specific error type.
	GetRequestValidationError() *qos.EVMRequestValidationError

	// IsSuccessful checks if any endpoint provided a valid response to this request.
	IsSuccessful() bool
}

// requestAdapter implements the request interface for EVMRequestObservations
type requestAdapter struct {
	observations *qos.EVMRequestObservations
}

func (a requestAdapter) GetRequestValidationError() *qos.EVMRequestValidationError {
	// Check if this is an HTTP body read failure
	if httpBodyReadFailure := a.observations.GetEvmHttpBodyReadFailure(); httpBodyReadFailure != nil {
		return &httpBodyReadFailure.ValidationError
	}

	// Check if this is a request unmarshaling failure
	if unmarshalingFailure := a.observations.GetEvmRequestUnmarshalingFailure(); unmarshalingFailure != nil {
		return &unmarshalingFailure.ValidationError
	}

	// If neither failure type is present, the request is valid
	return nil
}

func (a requestAdapter) IsSuccessful() bool {
	// If there are no endpoint observations, the request was not successful
	if len(a.observations.GetEndpointObservations()) == 0 {
		return false
	}

	for _, observation := range a.observations.GetEndpointObservations() {
		if resp := extractEndpointResponseFromObservation(observation); resp != nil {
			// Response is valid if GetResponseValidationError returns nil
			if resp.GetResponseValidationError() == nil {
				return true
			}
		}
	}

	return false
}

// extractRequestStatus creates a request adapter from EVMRequestObservations.
// Returns nil if observations is nil.
func extractRequestStatus(observations *qos.EVMRequestObservations) request {
	if observations == nil {
		return nil
	}

	return requestAdapter{observations}
}
