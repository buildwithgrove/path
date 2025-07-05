package gateway

import (
	"errors"
)

var (
	// no service ID was provided by the user.
	ErrNoServiceIDProvided = errors.New("no service ID provided")

	// QoS instance rejected the request.
	// e.g. HTTP payload could not be unmarshaled into a JSONRPC request.
	ErrRejectedByQoS = errors.New("QoS instance rejected the request")
)
