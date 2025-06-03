package gateway

import (
	"context"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

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
	// 	- protocol: Morse
	// 	- observation: "endpoint maxed-out or over-serviced (i.e. onchain rate limiting)"
	// 	- result: skip the endpoint for a set time period until a new session begins.
	ApplyObservations(*protocolobservations.Observations) error

	// TODO_FUTURE(@adshmh): support specifying the app(s) used for sending/signing synthetic relay requests by the hydrator.
	// TODO_TECHDEBT: Enable the hydrator for gateway modes beyond Centralized only.
	//
	// ConfiguredServiceIDs returns the list of service IDs that the protocol instance is configured to serve.
	// For Morse:
	// 	- Returns the list of all service IDs with available configured AATs.
	// For Shannon:
	// 	- Returns the list of all service IDs for which the gateway is configured to serve.
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
// The first two implementations of this interface are (as of writing):
//   - Morse: in the relayer/morse package, and
//   - Shannon: in the relayer/shannon package.
type ProtocolRequestContext interface {
	// HandleServiceRequest sends the supplied payload to the endpoint selected using the above SelectEndpoint method,
	// and receives and verifies the response.
	HandleServiceRequest(protocol.Payload) (protocol.Response, error)

	// HandleWebsocketRequest handles a WebSocket connection request.
	// Only Shannon protocol supports WebSocket connections; requests to Morse will always return an error.
	HandleWebsocketRequest(polylog.Logger, *http.Request, http.ResponseWriter) error

	// GetObservations builds and returns the set of protocol-specific observations using the current context.
	//
	// Hypothetical illustrative example.
	//
	// If the context is:
	// 	- Protocol: Morse
	//	- SelectedEndpoint: `endpoint_101`
	//	- Event: HandleServiceRequest returned a "maxed-out endpoint" error
	//
	// Then the observation can be:
	//  - `maxed-out endpoint` on `endpoint_101`.
	GetObservations() protocolobservations.Observations
}
