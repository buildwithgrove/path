package qos

// EVMRequestError represents a failure in processing an EVM request or response
// Contains information extracted from error types defined in evm.proto
type EVMRequestError struct {
	// For request validation errors: non-nil indicates a request error
	requestValidationError *EVMRequestValidationError
	// For response validation errors: non-nil indicates a response error
	responseValidationError *EVMResponseValidationError
}

// GetError returns the error type string representation.
// As of #186, this is limited to request, response or unknown error.
func (e *EVMRequestError) GetError() string {
	// Request error
	if e.IsRequestError() {
		return e.requestValidationError.String()
	}

	// Response error
	if e.IsResponseError() {
		return e.responseValidationError.String()
	}

	// This should never happen.
	return "UNKNOWN_ERROR"
}

// IsRequestError returns true if this is a request validation error
func (e *EVMRequestError) IsRequestError() bool {
	return e.requestValidationError != nil
}

// IsResponseError returns true if this is a response validation error
func (e *EVMRequestError) IsResponseError() bool {
	return e.responseValidationError != nil
}

// String returns a string representation of the error
func (e *EVMRequestError) String() string {
	return e.GetError()
}

// Error implements the error interface
func (e *EVMRequestError) Error() string {
	return e.GetError()
}
