package http

import (
	"errors"
	"fmt"
	"net/http"
)

// Endpoint's backend service returned a non 2xx HTTP status code.
var ErrRelayEndpointHTTPError = errors.New("endpoint returned non 2xx HTTP status code")

// EnsureHTTPSuccess returns an error if the status code is not a 2xx successful status code.
// Otherwise returns nil.
func EnsureHTTPSuccess(statusCode int) error {
	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%w: %d", ErrRelayEndpointHTTPError, statusCode)
	}
	return nil
}
