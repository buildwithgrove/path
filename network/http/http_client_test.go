package http

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsureHTTPSuccess(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		expectError bool
	}{
		// 2xx Success codes - should pass
		{
			name:        "200 OK",
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:        "201 Created",
			statusCode:  http.StatusCreated,
			expectError: false,
		},
		{
			name:        "202 Accepted",
			statusCode:  http.StatusAccepted,
			expectError: false,
		},
		{
			name:        "204 No Content",
			statusCode:  http.StatusNoContent,
			expectError: false,
		},
		{
			name:        "299 Last 2xx",
			statusCode:  299,
			expectError: false,
		},
		// Non-2xx codes - should fail
		{
			name:        "100 Continue",
			statusCode:  http.StatusContinue,
			expectError: true,
		},
		{
			name:        "300 Multiple Choices",
			statusCode:  http.StatusMultipleChoices,
			expectError: true,
		},
		{
			name:        "301 Moved Permanently",
			statusCode:  http.StatusMovedPermanently,
			expectError: true,
		},
		{
			name:        "400 Bad Request",
			statusCode:  http.StatusBadRequest,
			expectError: true,
		},
		{
			name:        "401 Unauthorized",
			statusCode:  http.StatusUnauthorized,
			expectError: true,
		},
		{
			name:        "403 Forbidden",
			statusCode:  http.StatusForbidden,
			expectError: true,
		},
		{
			name:        "404 Not Found",
			statusCode:  http.StatusNotFound,
			expectError: true,
		},
		{
			name:        "429 Too Many Requests",
			statusCode:  http.StatusTooManyRequests,
			expectError: true,
		},
		{
			name:        "500 Internal Server Error",
			statusCode:  http.StatusInternalServerError,
			expectError: true,
		},
		{
			name:        "502 Bad Gateway",
			statusCode:  http.StatusBadGateway,
			expectError: true,
		},
		{
			name:        "503 Service Unavailable",
			statusCode:  http.StatusServiceUnavailable,
			expectError: true,
		},
		{
			name:        "504 Gateway Timeout",
			statusCode:  http.StatusGatewayTimeout,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := EnsureHTTPSuccess(tc.statusCode)
			if tc.expectError {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrRelayEndpointHTTPError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
