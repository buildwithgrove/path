package solana

import (
	"net/http"

	"github.com/buildwithgrove/path/gateway"
)

const (
	// HTTP status code 400 bad request is used if the request cannot be deserialized into JSONRPC.
	httpStatusRequestValidationFailureUnmarshalFailure = http.StatusBadRequest

	// HTTP status code 500 internal server error is used if reading the HTTP request's body fails
	httpStatusRequestValidationFailureReadHTTPBodyFailure = http.StatusInternalServerError

	// HTTP status codes returned on response validation failure: no response received
	httpStatusResponseValidationFailureNoResponse = http.StatusInternalServerError
)

// httpResponse is used by the RequestContext to provide
// a Solana-specific implementation of gateway package's HTTPResponse.
var _ gateway.HTTPResponse = httpResponse{}

type httpResponse struct {
	responsePayload []byte
}

// GetPayload returns the payload for the HTTP response.
func (hr httpResponse) GetPayload() []byte {
	return hr.responsePayload
}

// TODO_MVP(@adshmh): Implement HTTP status code selection based on JSONRPC error codes. See qos/evm for reference.
// GetHTTPStatusCode returns the HTTP status code for the response.
func (hr httpResponse) GetHTTPStatusCode() int {
	// TODO_TECHDEBT: Default to 200 OK HTTP status code for Solana for now.
	return http.StatusOK
}

// GetHTTPHeaders returns the set of headers for the HTTP response.
// Solana does not need to set HTTP response headers.
func (r httpResponse) GetHTTPHeaders() map[string]string {
	return nil
}
