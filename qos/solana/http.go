package solana

import (
	"net/http"

	"github.com/buildwithgrove/path/gateway"
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

// GetHTTPStatusCode returns the HTTP status code for the response.
func (hr httpResponse) GetHTTPStatusCode() int {
	// Solana always returns a 200 HTTP status code.
	return http.StatusOK
}

// GetHTTPHeaders returns the set of headers for the HTTP response.
// Solana does not need to set HTTP response headers.
func (r httpResponse) GetHTTPHeaders() map[string]string {
	return nil
}
