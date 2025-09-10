package jsonrpc

import (
	"net/http"
)

// GetRecommendedHTTPStatusCode maps a JSON-RPC error response code to an HTTP status code.
// DEV_NOTE: This is an opinionated implementation pattern not strictly defined in the JSONRPC specification.
// See #179 or docusaurus/docs/develop/path/http_status_code.md for more details.
func (r Response) GetRecommendedHTTPStatusCode() int {
	// Return 200 OK if no error is present
	if r.Error == nil {
		return http.StatusOK
	}

	// Map standard JSON-RPC error codes to HTTP status codes
	switch r.Error.Code {
	case -32700: // Parse error
		return http.StatusBadRequest // 400
	case -32600: // Invalid request
		return http.StatusBadRequest // 400
	case -32601: // Method not found
		return http.StatusNotFound // 404
	case -32602: // Invalid params
		return http.StatusBadRequest // 400
	case -32603: // Internal error
		return http.StatusInternalServerError // 500
	case -32098: // Timeout (used by some providers)
		return http.StatusGatewayTimeout // 504
	case -32097: // Rate limited (used by some providers)
		return http.StatusTooManyRequests // 429
	}

	// Server error range (-32000 to -32099)
	if r.Error.Code >= -32099 && r.Error.Code <= -32000 {
		return http.StatusInternalServerError // 500
	}

	// Application-defined errors
	if r.Error.Code > 0 {
		// Positive error codes typically indicate client-side issues
		return http.StatusBadRequest // 400
	} else if r.Error.Code < 0 {
		// Other negative error codes typically indicate server-side issues
		return http.StatusInternalServerError // 500
	}

	// This should never be reached, but as a fallback return 500
	return http.StatusInternalServerError // 500
}
