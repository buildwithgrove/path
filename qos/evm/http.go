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
}

func (hr httpResponse) GetPayload() []byte {
	return hr.responsePayload
}

func (hr httpResponse) GetHTTPStatusCode() int {
	// EVM always returns a 200 HTTP status code.
	return http.StatusOK
}

// GetHTTPHeaders returns the set of headers for the HTTP response.
// As of PR #72, the only header returned for EVM is `Content-Type`.
func (r httpResponse) GetHTTPHeaders() map[string]string {
	// EVM only uses the `Content-Type` HTTP header.
	return httpHeadersApplicationJSON 
}
