package gateway

import (
	"context"
	"net/http"

	"github.com/buildwithgrove/path/health"
	"github.com/buildwithgrove/path/metrics/devtools"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/protocol"
)

// Protocol defines the core functionality of a protocol from the perspective of a gateway.
// The gateway expects a protocol to build and return a request context for a particular service ID.
type Protocol interface {
	// AvailableEndpoints returns the list of available endpoints matching both the service ID
	//
	// (Shannon only: in Delegated mode, the staked application is passed in the request header, which
	// filters the list of available endpoints. In all other modes, *http.Request will be nil.)
	// Context may contain a deadline that protocol should respect on best-effort basis.
	// Return observation if endpoint lookup fails.
	// Used as protocol observation for the request when no protocol context exists.
	AvailableEndpoints(
		context.Context,
		protocol.ServiceID,
		*http.Request,
	) (protocol.EndpointAddrList, protocolobservations.Observations, error)

	// BuildRequestContextForEndpoint builds and returns a ProtocolRequestContext containing a single selected endpoint.
	// One `ProtocolRequestContext` correspond to a single request, which is sent to a single endpoint.
	//
	// (Shannon only: in Delegated mode, the staked application is passed in the request header, which
	// filters the list of available endpoints. In all other modes, *http.Request will be nil.)
	// Context may contain a deadline that protocol should respect on best-effort basis.
	// Return observation if the context setup fails.
	// Used as protocol observation for the request when no protocol context exists.
	BuildRequestContextForEndpoint(
		context.Context,
		protocol.ServiceID,
		protocol.EndpointAddr,
		*http.Request,
	) (ProtocolRequestContext, protocolobservations.Observations, error)

	// SupportedGatewayModes returns the Gateway modes supported by the protocol instance.
	// See protocol/gateway_mode.go for more details.
	SupportedGatewayModes() []protocol.GatewayMode

	// ApplyObservations applies the supplied observations to the protocol instance's internal state.
	// Hypothetical example (for illustrative purposes only):
	// 	- protocol: Shannon
	// 	- observation: "endpoint maxed-out or over-serviced (i.e. onchain rate limiting)"
	// 	- result: skip the endpoint for a set time period until a new session begins.
	ApplyObservations(*protocolobservations.Observations) error

	// TODO_FUTURE(@adshmh): support specifying the app(s) used for sending/signing synthetic relay requests by the hydrator.
	// TODO_TECHDEBT: Enable the hydrator for gateway modes beyond Centralized only.
	//
	// ConfiguredServiceIDs returns the list of service IDs that the protocol instance is configured to serve.
	ConfiguredServiceIDs() map[protocol.ServiceID]struct{}

	// GetTotalServiceEndpointsCount returns the count of all unique endpoints for a service ID
	// without filtering sanctioned endpoints.
	GetTotalServiceEndpointsCount(protocol.ServiceID, *http.Request) (int, error)

	// HydrateDisqualifiedEndpointsResponse hydrates the disqualified endpoint response with the protocol-specific data.
	HydrateDisqualifiedEndpointsResponse(protocol.ServiceID, *devtools.DisqualifiedEndpointResponse)

	// health.Check interface is used to verify protocol instance's health status.
	health.Check
}

// ProtocolRequestContext defines the functionality expected by the gateway from the protocol,
// for a particular service ID.
//
// These include but not limited to:
//  1. Listing the endpoints available for sending relays for a specific service.
//  2. Send a relay to a specific endpoint and return its response.
//
// The implementation of this interface is in the relayer/shannon package.
type ProtocolRequestContext interface {
	// HandleServiceRequest sends the supplied payload to the endpoint selected using the above SelectEndpoint method,
	// and receives and verifies the response.
	HandleServiceRequest([]protocol.Payload) ([]protocol.Response, error)

	// GetObservations builds and returns the set of protocol-specific observations using the current context.
	//
	// Hypothetical illustrative example.
	//
	// If the context is:
	// 	- Protocol: Shannon
	//	- SelectedEndpoint: `endpoint_101`
	//	- Event: HandleServiceRequest returned a "maxed-out endpoint" error
	//
	// Then the observation can be:
	//  - `maxed-out endpoint` on `endpoint_101`.
	GetObservations() protocolobservations.Observations

	// TODO_TECHDEBT(@commodity, @adshmh): Revisit all the Websocket specific functions
	// in ProtocolRequestContext.
	// - Too many websocket specific functions are exposed explicitly implying a poor interface.
	// - Revisit the need for exposing these at all through a refactor?
	//
	// TODO_TECHDEBT(@commodity, @adshmh): Revisit casing of websocket vs Websocket vs WebSocket through.
	ProtocolRequestContextWebsocket
}

// ProtocolRequestContextWebsocket defines the functionality expected by the gateway from the protocol,
// specifically for websocket requests
type ProtocolRequestContextWebsocket interface {
	// GetWebsocketConnectionHeaders returns protocol-specific headers needed for websocket connections.
	// These headers contain protocol-specific information like session data, service IDs, etc.
	GetWebsocketConnectionHeaders() (http.Header, error)

	// GetWebsocketEndpointURL returns the websocket URL for the selected endpoint.
	// This URL is used to establish the websocket connection to the endpoint.
	GetWebsocketEndpointURL() (string, error)

	// ProcessProtocolClientWebsocketMessage processes a message from the client.
	ProcessProtocolClientWebsocketMessage([]byte) ([]byte, error)

	// ProcessProtocolEndpointWebsocketMessage processes a message from the endpoint.
	ProcessProtocolEndpointWebsocketMessage([]byte) ([]byte, protocolobservations.Observations, error)
}
