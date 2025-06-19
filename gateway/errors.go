package gateway

import (
	"errors"
)

var (
	// no service ID was provided by the user.
	GatewayErrNoServiceIDProvided = errors.New("no service ID provided")

	// QoS instance rejected the request.
	// e.g. HTTP payload could not be unmarshaled into a JSONRPC request.
	GatewayErrRejectedByQoS = errors.New("QoS instance rejected the request")
)
