package noop

import (
	"github.com/buildwithgrove/path/gateway"
)

// HTTPResponse provides all the functionality required by the gate.HTTPResponse interface.
var _ gateway.HTTPResponse = &HTTPResponse{}

// HTTPResponse stores the data required for building and returning a user-facing HTTP response
// based on the response received from an endpoint to a service request.
type HTTPResponse struct {
	// payload is the raw payload received from the endpoint servicing the request.
	payload []byte
	// httpStatus is the HTTP status code to be returned to the user.
	httpStatusCode int
}

// GetPayload returns the payload of the user-facing HTTP response.
// This method implements the gateway.HTTPResponse interface.
func (h *HTTPResponse) GetPayload() []byte {
	return h.payload
}

// GetHTTPStatusCode returns the HTTP status code of the user-facing HTTP response.
// This method implements the gateway.HTTPResponse interface.
func (h *HTTPResponse) GetHTTPStatusCode() int {
	return h.httpStatusCode
}

// GetHTTPHeaders always returns nil, as HTTP headers are not used by noop QoS as of PR #106.
// This method implements the gateway.HTTPResponse interface.
func (h *HTTPResponse) GetHTTPHeaders() map[string]string {
	return nil
}
