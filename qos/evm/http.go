package evm

import (
	"net/http"

	"github.com/buildwithgrove/path/gateway"
)

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

func (r httpResponse) GetHTTPHeaders() map[string]string {
	// EVM does not need to set HTTP response headers.
	return nil
}
