package evm

import (
	"net/http"

	"github.com/buildwithgrove/path/gateway"
)

// httpHeadersApplicationJSON is the `Content-Type` HTTP header used in all EVM responses.
var httpHeadersApplicationJSON = map[string]string{
	"Content-Type": "application/json",
}

// httpResponse is used by the RequestContext to provide
// an EVM-specific implementation of gateway package's HTTPResponse.
var _ gateway.HTTPResponse = httpResponse{}

type httpResponse struct {
	responsePayload []byte
	// allow over-riding the default HTTP status code of 200.
	httpStatusCode int
}

func (hr httpResponse) GetPayload() []byte {
	return hr.responsePayload
}

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
