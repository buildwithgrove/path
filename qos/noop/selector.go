package noop

import (
	"errors"
	"math/rand"

	"github.com/buildwithgrove/path/protocol"
)

// RandomEndpointSelector provides the functionality defined by the protocol.EndpointSelector interface.
var _ protocol.EndpointSelector = RandomEndpointSelector{}

// RandomEndpointSelector returns a randomly selected endpoint from the set of available ones, when its
// Select method is called.
// It has no fields, since the endpoint selection is random, i.e. no data is kept on the endpoints.
type RandomEndpointSelector struct{}

// Select returns a randomly selected endpoint from the set of supplied endpoints.
// This method fulfills the protocol.EndpointSelector interface.
func (_ RandomEndpointSelector) Select(endpoints []protocol.Endpoint) (protocol.EndpointAddr, error) {
	if len(endpoints) == 0 {
		return protocol.EndpointAddr(""), errors.New("RandomEndpointSelector: an empty endpoint list was supplied to the selector")
	}

	return endpoints[rand.Intn(len(endpoints))].Addr(), nil
}
