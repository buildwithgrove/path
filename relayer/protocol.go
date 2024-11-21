package relayer

import (
	"github.com/buildwithgrove/path/health"
)

// Protocol defines the core functionality of a protocol from the perspective of a gateway.
// The gateway expects a protocol to build and return a request context for a particular service ID.
type Protocol interface {
	// BuildRequestContext builds and returns a ProtocolRequestContext interface for handling a single service
	// request, which matches the provided Service ID.
	BuildRequestContext(ServiceID) (ProtocolRequestContext, error)

	health.Check
}

// ProtocolRequestContext defines the functionality expected by the gateway from the protocol,
// for a particular service ID.
//
// These include but not limited to:
//  1. Listing the endpoints available for sending relays for a specific service.
//  2. Send a relay to a specific endpoint and return its response.
//
// The first two implementations of this interface are (as of writing) are:
//   - Morse: in the relayer/morse package, and
//   - Shannon: in the relayer/shannon package.
type ProtocolRequestContext interface {
	// TODO_TECHDEBT: any protocol/network-level errors should result in
	// the endpoint being dropped by the protocol instance from the returned
	// set of available endpoints.
	// e.g. an endpoint that is temporarily/permanently unavailable.
	SelectEndpoint(EndpointSelector) error

	// HandleServiceRequest sends the supplied payload to the endpoint selected using the above SelectEndpoint method,
	// and receives and verfieis the response.
	HandleServiceRequest(Payload) (Response, error)
}