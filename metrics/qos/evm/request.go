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

	// GetHTTPStatusCode returns the HTTP status code that should be returned to the user.
	// For invalid requests, this returns the status code from the validation error.
	// Returns 0 if no specific status code is available.
	GetHTTPStatusCode() int
}

var _ request = requestAdapter{}

// requestAdapter implements the request interface for EVMRequestObservations
type requestAdapter struct {
	observations *qos.EVMRequestObservations
}

// GetRequestValidationError satisfies the request interface
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

// IsSuccessful satisfies the request interface.
func (a requestAdapter) IsSuccessful() bool {
	// If there are no endpoint observations, the request was not successful
	if len(a.observations.GetEndpointObservations()) == 0 {
		return false
	}

	// Iterate through all endpoint observations and return true if any (i.e. at least one) are valid
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

// GetHTTPStatusCode satisfies the request interface.
func (a requestAdapter) GetHTTPStatusCode() int {
	// Check if this is an HTTP body read failure
	if httpBodyReadFailure := a.observations.GetEvmHttpBodyReadFailure(); httpBodyReadFailure != nil {
		return int(httpBodyReadFailure.GetHttpStatusCode())
	}

	// Check if this is a request unmarshaling failure
	if unmarshalingFailure := a.observations.GetEvmRequestUnmarshalingFailure(); unmarshalingFailure != nil {
		return int(unmarshalingFailure.GetHttpStatusCode())
	}

	// No specific HTTP status code for this request was recorded, found or determined otherwise
	return 0
}

// newRequestAdapter creates a request adapter from EVMRequestObservations.
// Returns nil if observations is nil.
func newRequestAdapter(observations *qos.EVMRequestObservations) request {
	if observations == nil {
		return nil
	}

	return requestAdapter{observations}
}
