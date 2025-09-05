package evm

import (
	"net/http"

	pathhttp "github.com/buildwithgrove/path/network/http"
)

const (
	// HTTP status code 400 bad request is used if the request cannot be deserialized into JSONRPC.
	httpStatusRequestValidationFailureUnmarshalFailure = http.StatusBadRequest

	// TODO_MVP(@adshmh): Remove the error below once the qos interface is updated to replace ParseHTTPRequest with ParseRequest, decoupling the QoS service from the HTTP request.
	// HTTP status code 500 internal server error is used if reading the HTTP request's body fails
	httpStatusRequestValidationFailureReadHTTPBodyFailure = http.StatusInternalServerError

	// HTTP status codes returned on response validation failure: no response received
	httpStatusResponseValidationFailureNoResponse = http.StatusInternalServerError

	// HTTP status codes returned on response validation failure: empty response received
	httpStatusResponseValidationFailureEmptyResponse = http.StatusInternalServerError
)

// httpHeadersApplicationJSON is the `Content-Type` HTTP header used in all EVM responses.
var httpHeadersApplicationJSON = map[string]string{
	"Content-Type": "application/json",
}

// httpResponse is used by the RequestContext to provide
// an EVM-specific implementation of gateway package's HTTPResponse.
var _ pathhttp.HTTPResponse = httpResponse{}

// httpResponse encapsulates an HTTP response to be returned to the client
// including payload data and status code.
type httpResponse struct {
	// responsePayload contains the serialized response body.
	responsePayload []byte

	// httpStatusCode is the HTTP status code to be returned.
	// If not explicitly set, defaults to http.StatusOK (200).
	httpStatusCode int
}

// GetPayload returns the response payload as a byte slice.
func (hr httpResponse) GetPayload() []byte {
	return hr.responsePayload
}

// GetHTTPStatusCode returns the HTTP status code for this response.
// If no status code was explicitly set, returns http.StatusOK (200).
// StatusOK is returned by default from EVM QoS because it is the responsibility of the QoS service to decide on the HTTP status code returned to the client.
func (hr httpResponse) GetHTTPStatusCode() int {
	// Return the custom status code if set, otherwise default to 200 OK
	if hr.httpStatusCode != 0 {
		return hr.httpStatusCode
	}

	// Default to 200 OK HTTP status code.
	return http.StatusOK
}

// GetHTTPHeaders returns the set of headers for the HTTP response.
// As of PR #72, the only header returned for EVM is `Content-Type`.
func (r httpResponse) GetHTTPHeaders() map[string]string {
	// EVM only uses the `Content-Type` HTTP header.
	return httpHeadersApplicationJSON
}
