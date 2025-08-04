package gateway

import (
	"errors"
)

var (
	// no service ID was provided by the user.
	ErrGatewayNoServiceIDProvided = errors.New("no service ID provided")

	// QoS instance rejected the request.
	// e.g. HTTP payload could not be unmarshaled into a JSONRPC request.
	errGatewayRejectedByQoS = errors.New("QoS instance rejected the request")

	// Error building protocol contexts from HTTP request.
	errBuildProtocolContextsFromHTTPRequest = errors.New("error building protocol contexts from HTTP request")

	// Fallback request errors
	errFallbackRequestCreationFailed = errors.New("failed to create HTTP request for fallback URL")
	errFallbackRequestSendFailed     = errors.New("failed to send fallback request")
	errFallbackResponseReadFailed    = errors.New("failed to read fallback response body")
)
