package shannon

import (
	"fmt"

	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
	sdk "github.com/pokt-network/shannon-sdk"

	"github.com/buildwithgrove/path/protocol"
)

// By fulfilling the protocol package Endpoint interface, the endpoint struct below allows
// all code outside this package to refer to a specific Shannon SupplierEndpoint as a uniquely identifiable entity
// that can serve relays.
var _ protocol.Endpoint = endpoint{}

// endpoint is used to fulfill a protocol package Endpoint using a Shannon SupplierEndpoint.
// An endpoint is identified by combining its supplier address and its URL, because
// in Shannon a supplier can have multiple endpoints for a service.
type endpoint struct {
	supplier string
	url      string
	// TODO_TECHDEBT(@commoddity): Investigate if we should allow supporting additional RPC type endpoints.
	websocketUrl string

	// TODO_IMPROVE: If the same endpoint is in the session of multiple apps at the same time,
	// the first app will be chosen. A randomization among the apps in this (unlikely) scenario
	// may be needed.
	// session is the active session corresponding to the app, of which the endpoint is a member.
	session sessiontypes.Session
}

// TODO_TECHDEBT(@adshmh): Refactor to cleanly separate the "fallback" logic from the endpoint.
// Example:
// Make endpoint an interface, implemented by:
// - A Shannon endpoint
// - A "fallback" URL with configurable fields: e.g. the Supplier set as "Grove"
//
// fallbackSupplier is a const value used to identify fallback endpoints.
// Fallback endpoints do not exist on the Shannon protocol and so do not have a supplier address.
// Instead, they are identified by the fallbackSupplier const value.
const fallbackSupplier = "fallback"

// isFallback returns true if the endpoint is a fallback endpoint.
func (e endpoint) isFallback() bool {
	return e.supplier == fallbackSupplier
}

// TODO_MVP(@adshmh): replace EndpointAddr with a URL; a single URL should be treated the same regardless of the app to which it is attached.
// For protocol-level concerns: the (app/session, URL) should be taken into account; e.g. a healthy endpoint may have been maxed out for a particular app.
// For QoS-level concerns: only the URL of the endpoint matters; e.g. an unhealthy endpoint should be skipped regardless of the app/session to which it is attached.
func (e endpoint) Addr() protocol.EndpointAddr {
	return protocol.EndpointAddr(fmt.Sprintf("%s-%s", e.supplier, e.url))
}

// PublicURL returns the URL of the endpoint.
func (e endpoint) PublicURL() string {
	return e.url
}

// WebsocketURL returns the URL of the endpoint.
func (e endpoint) WebsocketURL() (string, error) {
	if e.websocketUrl == "" {
		return "", fmt.Errorf("websocket URL is not set")
	}
	return e.websocketUrl, nil
}

// Session returns a pointer to the session associated with the endpoint.
func (e endpoint) Session() *sessiontypes.Session {
	return &e.session
}

// Supplier returns the supplier address of the endpoint.
func (e endpoint) Supplier() string {
	return e.supplier
}

// endpointsFromSession returns the list of all endpoints from a Shannon session.
// It returns a map for efficient lookup, as the main/only consumer of this function uses
// the return value for selecting an endpoint for sending a relay.
func endpointsFromSession(session sessiontypes.Session) (map[protocol.EndpointAddr]endpoint, error) {
	sf := sdk.SessionFilter{
		Session: &session,
	}

	// AllEndpoints will return a map of supplier address to a list of supplier endpoints.
	//
	// Each supplier address will have one or more endpoints, one per RPC-type.
	// For example, a supplier may have one endpoint for JSON-RPC and one for websocket.
	allEndpoints, err := sf.AllEndpoints()
	if err != nil {
		return nil, err
	}

	endpoints := make(map[protocol.EndpointAddr]endpoint)
	for _, supplierEndpoints := range allEndpoints {
		// All endpoints for a supplier will have the same supplier address & session,
		// so we can use the first item to set the supplier address & session.
		endpoint := endpoint{
			supplier: string(supplierEndpoints[0].Supplier()),
			// Set the session field on the endpoint for efficient lookup when sending relays.
			session: session,
		}

		// Set the URL of the endpoint based on the RPC type.
		// Each supplier endpoint may have multiple RPC types, so we need to set the URL for each.
		//
		// IMPORTANT: As of PATH PR #345 the only supported RPC types are:
		// 	- `JSON_RPC`
		// 	- `WEBSOCKET`
		//
		// References:
		// 	- PATH PR #345 - https://github.com/buildwithgrove/path/pull/345
		// 	- poktroll `RPCType` enum - https://github.com/pokt-network/poktroll/blob/main/x/shared/types/service.pb.go#L31
		for _, supplierRPCTypeEndpoint := range supplierEndpoints {
			switch supplierRPCTypeEndpoint.RPCType() {

			// If the endpoint is a `WEBSOCKET` RPC type endpoint, set the websocket URL.
			case sharedtypes.RPCType_WEBSOCKET:
				endpoint.websocketUrl = supplierRPCTypeEndpoint.Endpoint().Url

			// Currently only `WEBSOCKET` & `JSON_RPC` types are supported, so `JSON_RPC` is the default.
			default:
				endpoint.url = supplierRPCTypeEndpoint.Endpoint().Url
			}
		}

		endpoints[endpoint.Addr()] = endpoint
	}

	return endpoints, nil
}
