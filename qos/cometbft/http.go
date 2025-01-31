package cometbft

import (
	"github.com/buildwithgrove/path/gateway"
)

// httpHeadersApplicationJSON is the `Content-Type` HTTP header used in all CometBFT responses.
var httpHeadersApplicationJSON = map[string]string{
	"Content-Type": "application/json",
}

// httpResponse is used by the RequestContext to provide
// a CometBFT-specific implementation of gateway package's HTTPResponse.
var _ gateway.HTTPResponse = httpResponse{}

type httpResponse struct {
	responsePayload []byte
	responseStatus  int
}

// GetPayload returns the payload for the HTTP response.
func (hr httpResponse) GetPayload() []byte {
	return hr.responsePayload
}

// GetHTTPStatusCode returns the HTTP status code for the response.
func (hr httpResponse) GetHTTPStatusCode() int {
	return hr.responseStatus
}

// GetHTTPHeaders returns the set of headers for the HTTP response.
func (r httpResponse) GetHTTPHeaders() map[string]string {
	// TODO_TECHDEBT: Consider adding support for returning the response headers to the caller?
	// CometBFT only uses the `Content-Type` HTTP header.
	return httpHeadersApplicationJSON
}
