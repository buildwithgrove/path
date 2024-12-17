package noop

import (
	"errors"
	"math/rand"

	"github.com/buildwithgrove/path/protocol"
)

var _ protocol.EndpointSelector = RandomEndpointSelector{}

type RandomEndpointSelector struct{}

func (r RandomEndpointSelector) Select(endpoints []protocol.Endpoint) (protocol.EndpointAddr, error) {
	if len(endpoints) == 0 {
		return protocol.EndpointAddr(""), errors.New("RandomEndpointSelector: an empty endpoint list was supplied to the selector")
	}

	return endpoints[rand.Intn(len(endpoints))].Addr(), nil
}
