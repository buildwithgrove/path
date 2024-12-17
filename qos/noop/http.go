package noop

import (
	"github.com/buildwithgrove/path/gateway"
)

var _ gateway.HTTPResponse = &HTTPResponse{}

type HTTPResponse struct {
	payload        []byte
	httpStatusCode int
}

func (h *HTTPResponse) GetPayload() []byte {
	return h.payload
}

func (h *HTTPResponse) GetHTTPStatusCode() int {
	return h.httpStatusCode
}

func (h *HTTPResponse) GetHTTPHeaders() map[string]string {
	return nil
}
