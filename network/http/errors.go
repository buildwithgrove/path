package http

import "errors"

// Endpoint's backend service returned a non 2xx HTTP status code.
var ErrRelayEndpointHTTPError = errors.New("endpoint returned non 2xx HTTP status code")
