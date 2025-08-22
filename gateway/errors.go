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

	// WebSocket request was rejected by QoS instance.
	// e.g. WebSocket subscription request validation failed.
	errWebsocketRequestRejectedByQoS = errors.New("WebSocket request rejected by QoS instance")

	// WebSocket connection establishment failed.
	// e.g. Failed to upgrade HTTP connection to WebSocket or connect to endpoint.
	errWebsocketConnectionFailed = errors.New("WebSocket connection establishment failed")
)
