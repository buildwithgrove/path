// response.go
package evm

import (
	"github.com/buildwithgrove/path/observation/qos"
)

// response defines methods needed for response metrics collection.
// Abstracts proto-specific details from metrics logic.
// DEV_NOTE: You MUST update this when adding new metric requirements.
type response interface {
	// GetResponseValidationError returns the validation error if any.
	// A nil return value indicates the response is valid.
	// A non-nil value indicates the response is invalid, with the specific error type.
	GetResponseValidationError() *qos.EVMResponseValidationError

	// GetHTTPStatusCode returns the HTTP status code from the response.
	// Returns 0 if no HTTP status code is available.
	GetHTTPStatusCode() int
}

// responseAdapter is an implementation of the response interface
// that wraps different response types and provides a common interface.
type responseAdapter struct {
	validationError *qos.EVMResponseValidationError
	httpStatusCode  int
}

func (a responseAdapter) GetResponseValidationError() *qos.EVMResponseValidationError {
	if a.validationError == nil {
		return nil
	}

	// If the error is UNSPECIFIED, treat it as valid (return nil)
	if *a.validationError == qos.EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_UNSPECIFIED {
		return nil
	}

	return a.validationError
}

func (a responseAdapter) GetHTTPStatusCode() int {
	return a.httpStatusCode
}

// extractEndpointResponseFromObservation extracts the response data from an endpoint observation.
// Returns nil if no response is present.
// DEV_NOTE: You MUST update this when adding new metric requirements.
func extractEndpointResponseFromObservation(observation *qos.EVMEndpointObservation) response {
	if observation == nil {
		return nil
	}

	// handle chain_id response
	if chainIDResp := observation.GetChainIdResponse(); chainIDResp != nil {
		return responseAdapter{
			validationError: chainIDResp.ResponseValidationError,
			httpStatusCode:  int(chainIDResp.GetHttpStatusCode()),
		}
	}

	// handle block_number response
	if blockNumResp := observation.GetBlockNumberResponse(); blockNumResp != nil {
		return responseAdapter{
			validationError: blockNumResp.ResponseValidationError,
			httpStatusCode:  int(blockNumResp.GetHttpStatusCode()),
		}
	}

	// handle unrecognized response
	if unrecognizedResp := observation.GetUnrecognizedResponse(); unrecognizedResp != nil {
		return responseAdapter{
			validationError: unrecognizedResp.ResponseValidationError,
			httpStatusCode:  int(unrecognizedResp.GetHttpStatusCode()),
		}
	}

	// handle empty response
	if emptyResp := observation.GetEmptyResponse(); emptyResp != nil {
		// Empty responses are always invalid
		err := emptyResp.ResponseValidationError
		return responseAdapter{
			validationError: &err,
			httpStatusCode:  int(emptyResp.GetHttpStatusCode()),
		}
	}

	// handle no response
	if noResp := observation.GetNoResponse(); noResp != nil {
		// No responses are always invalid
		err := noResp.ResponseValidationError
		return responseAdapter{
			validationError: &err,
			httpStatusCode:  int(noResp.GetHttpStatusCode()),
		}
	}

	return nil
}

// getEndpointResponseValidationFailureReason returns why the endpoint response failed QoS validation.
// Returns the validation error from the first endpoint observation, or an empty string if none exist
// or if the response was valid.
func getEndpointResponseValidationFailureReason(observations *qos.EVMRequestObservations) string {
	// First check if we have any endpoint observations
	if len(observations.GetEndpointObservations()) == 0 {
		return ""
	}

	// Look for the first invalid response and return its validation error
	for _, observation := range observations.GetEndpointObservations() {
		resp := extractEndpointResponseFromObservation(observation)
		if resp == nil {
			continue
		}

		if validationErr := resp.GetResponseValidationError(); validationErr != nil {
			return validationErr.String()
		}
	}

	return ""
}
