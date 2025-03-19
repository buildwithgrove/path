package jsonrpc

import (
	"net/http"
)

// GetRecommendedHTTPStatusCode maps a JSON-RPC error response code to an HTTP status code.
// DEV_NOTE: This is an opinionated implementation pattern not strictly defined in the JSONRPC specification.
// See #179 for more details.
func (r Response) GetRecommendedHTTPStatusCode() int {
	// Return 200 OK if no error is present
	if r.Error == nil {
		return http.StatusOK
	}

	// Return 500 Internal Server Error if the error code is in the range of -32099 to -32000
	if r.Error.Code >= -32099 && r.Error.Code <= -32000 {
		return http.StatusInternalServerError
	}

	// Return the appropriate 4xx based on the JSON-RPC error code
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

	// Default to 200 OK HTTP status code
	return http.StatusOK
}
