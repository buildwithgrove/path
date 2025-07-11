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

	//
	errBuildProtocolContextsFromHTTPRequest = errors.New("error building protocol contexts from HTTP request")
)
