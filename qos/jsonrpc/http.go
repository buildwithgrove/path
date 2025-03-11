package jsonrpc

import (
	"net/http"
)

// GetRecommendedHTTPStatusCode returns the recommended HTTP status code for the JSONRPC response.
// DEV_NOTE: this is only a recommendation based on a common practice. It does NOT necessarily reflect the JSONRPC spec.
func (r Response) GetRecommendedHTTPStatusCode() int {
	if r.Error == nil {
		return http.StatusOK
	}

	if r.Error.Code >= -32099 && r.Error.Code <= -32000 {
		return http.StatusInternalServerError
	}

	switch r.Error.Code {
	case -32700:
		return http.StatusBadRequest
	case -32600:
		return http.StatusBadRequest
	case -32601:
		return http.StatusNotFound
	case -32602:
		return http.StatusBadRequest
	case -32603:
		return http.StatusInternalServerError
	}

	return http.StatusOK
}
