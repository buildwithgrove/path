// TODO_REFACTOR(@adshmh): Extract patterns from this package into a shared location to enable reuse across other observation interpreters (e.g., solana, cometbft).
// This would establish a consistent interpretation pattern across all QoS services while maintaining service-specific handlers.
package qos

import (
	"errors"
)

var (
	ErrEVMNoObservations              = errors.New("no observations available")
	ErrEVMNoEndpointObservationsFound = errors.New("no endpoint observations listed")
)

// EVMObservationInterpreter provides interpretation helpers for EVM QoS observations.
type EVMObservationInterpreter struct {
	Observations *EVMRequestObservations
}

// GetRequestMethod returns the JSON-RPC method from the request
// Returns method and true if extraction succeeded
// Returns empty string and false if request is invalid or missing
func (i *EVMObservationInterpreter) GetRequestMethod() (string, bool) {
	if i.Observations == nil {
		return "", false
	}

	// Check for validation failures using the shared method
	if _, reqError := i.checkRequestValidationFailures(); reqError != nil {
		return "", false
	}

	req := i.Observations.GetJsonrpcRequest()
	if req == nil {
		return "", false
	}

	method := req.GetMethod()
	if method == "" {
		return "", false
	}

	return method, true
}

// GetChainID returns the chain ID associated with the EVM observations
// Returns chain ID and true if extraction succeeded
// Returns empty string and false if not available
//
// DEV_NOTE: If adapting this for other QoS observations, chainID may need to be
// renamed to ServiceID for non-blockchain services.
func (i *EVMObservationInterpreter) GetChainID() (string, bool) {
	if i.Observations == nil {
		return "", false
	}

	chainID := i.Observations.GetChainId()
	if chainID == "" {
		return "", false
	}

	return chainID, true
}

// GetRequestStatus returns the request status information:
// - httpStatusCode: the suggested HTTP status code to return to the client
// - requestError: error details (nil if successful)
// - err: error if interpreter cannot determine status (e.g., nil observations)
func (i *EVMObservationInterpreter) GetRequestStatus() (httpStatusCode int, requestError *EVMRequestError, err error) {
	if i.Observations == nil {
		return 0, nil, ErrEVMNoObservations
	}

	// Check for request validation failures first
	if httpStatusCode, requestError := i.checkRequestValidationFailures(); requestError != nil {
		return httpStatusCode, requestError, nil
	}

	// Then check endpoint responses
	return i.getEndpointResponseStatus()
}

// GetEndpointObservations returns endpoint observations and a boolean indicating success
// Returns nil, false if observations are missing or validation failed
func (i *EVMObservationInterpreter) GetEndpointObservations() ([]*EVMEndpointObservation, bool) {
	if i.Observations == nil {
		return nil, false
	}

	// Check for validation failures using the shared method
	if _, reqError := i.checkRequestValidationFailures(); reqError != nil {
		return nil, false
	}

	observations := i.Observations.GetEndpointObservations()
	if len(observations) == 0 {
		return nil, false
	}

	return observations, true
}

// checkRequestValidationFailures examines observations for request validation failures
// Returns (httpStatusCode, requestError) where requestError is non-nil if a validation failure was found
func (i *EVMObservationInterpreter) checkRequestValidationFailures() (int, *EVMRequestError) {
	// Check for HTTP body read failure
	if failure := i.Observations.GetEvmHttpBodyReadFailure(); failure != nil {
		errType := EVMRequestValidationError_EVM_REQUEST_VALIDATION_ERROR_HTTP_BODY_READ_FAILURE
		return int(failure.GetHttpStatusCode()), &EVMRequestError{
			requestValidationError: &errType,
		}
	}

	// Check for unmarshaling failure
	if failure := i.Observations.GetEvmRequestUnmarshalingFailure(); failure != nil {
		errType := EVMRequestValidationError_EVM_REQUEST_VALIDATION_ERROR_REQUEST_UNMARSHALING_FAILURE
		return int(failure.GetHttpStatusCode()), &EVMRequestError{
			requestValidationError: &errType,
		}
	}

	// No validation failures found
	return 0, nil
}

// getEndpointResponseStatus extracts status info about endpoint responses
// Returns (httpStatusCode, requestError, error) tuple
func (i *EVMObservationInterpreter) getEndpointResponseStatus() (int, *EVMRequestError, error) {
	observations := i.Observations.GetEndpointObservations()

	// No endpoint observations indicates no responses were received
	if len(observations) == 0 {
		return 0, nil, ErrEVMNoEndpointObservationsFound
	}

	// Use only the last observation (latest response)
	lastObs := observations[len(observations)-1]
	handler, err := getEVMResponseHandler(lastObs)
	if err != nil {
		return 0, nil, err
	}

	// Extract the status code and error type
	statusCode, errType := handler.ExtractValidityStatus(lastObs)
	if errType == nil {
		return statusCode, nil, nil
	}

	// Create appropriate EVmRequestError based on the observed error type
	reqError := &EVMRequestError{
		responseValidationError: errType,
	}

	return statusCode, reqError, nil
}
