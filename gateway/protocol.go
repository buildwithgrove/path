// gateway package defines the requirements and steps of sending relays from the perspective of:
// a) protocols, i.e. Morse and Shannon protocols, which provide:
// - a list of endpoints available for a service.
// - a function for sending a relay to a specific endpoint.
// b) gateways, which are required to provide a function for
// selecting an endpoint to which the relay is to be sent.
// c) Quality-of-Service (QoS) services: which provide:
// - interpretation of the user's request as the payload to be sent to an endpoint.
// - selection of the best endpoint for handling a user's request.
//
// TODO_MVP(@adshmh): add a README with a diagram of all the above.
// TODO_MVP(@adshmh): add a section for the following packages once they are added: Metrics, Message.
package gateway

import (
	"net/http"

	"github.com/buildwithgrove/path/health"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/protocol"
)

// Protocol defines the core functionality of a protocol from the perspective of a gateway.
// The gateway expects a protocol to build and return a request context for a particular service ID.
type Protocol interface {
	// BuildRequestContext builds and returns a ProtocolRequestContext interface for handling a single service
	// request, which matches the provided Service ID.
	BuildRequestContext(protocol.ServiceID, *http.Request) (ProtocolRequestContext, error)

	// SupportedGamewayModes returns the Gateway modes supported by the protocol instance.
	// See protocol/gateway_mode.go for more details.
	SupportedGatewayModes() []protocol.GatewayMode

	ApplyObservations(protocolobservations.ProtocolDetails) error
	health.Check
}

// ProtocolRequestContext defines the functionality expected by the gateway from the protocol,
// for a particular service ID.
//
// These include but not limited to:
//  1. Listing the endpoints available for sending relays for a specific service.
//  2. Send a relay to a specific endpoint and return its response.
//
// The first two implementations of this interface are (as of writing):
//   - Morse: in the relayer/morse package, and
//   - Shannon: in the relayer/shannon package.
type ProtocolRequestContext interface {
	// TODO_TECHDEBT(@adshmh): any protocol/network-level errors should result in
	// the endpoint being dropped by the protocol instance from the returned
	// set of available endpoints.
	// e.g. an endpoint that is temporarily/permanently unavailable.
	SelectEndpoint(protocol.EndpointSelector) error

	// HandleServiceRequest sends the supplied payload to the endpoint selected using the above SelectEndpoint method,
	// and receives and verfieis the response.
	HandleServiceRequest(protocol.Payload) (protocol.Response, error)

	// AvailableEndpoints returns the list of available endpoints matching both the service ID and the operation mode of the request context.
	// This is needed by the Endpooint Hydrator as an easy-to-read method of obtaining all available endpoints, rather than using the SelectEndpoint method.
	// This method is scoped to a specific ProtocolRequestContext, because different operation modes impact the available applications and endpoints.
	// See the Shannon package's operation_mode.go file for more details.
	AvailableEndpoints() ([]protocol.Endpoint, error)

	GetObservations() protocolobservations.ProtocolDetails
}
