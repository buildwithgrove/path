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

	allEndpoints, err := sf.AllEndpoints()
	if err != nil {
		return nil, err
	}

	endpoints := make(map[protocol.EndpointAddr]endpoint)
	for _, supplierEndpoints := range allEndpoints {
		endpoint := endpoint{
			supplier: string(supplierEndpoints[0].Supplier()),
			// Set the session field on the endpoint for efficient lookup when sending relays.
			session: session,
		}

		for _, supplierEndpoint := range supplierEndpoints {
			switch supplierEndpoint.RPCType() {
			// If the endpoint is a websocket RPC type endpoint, set the websocket URL.
			case sharedtypes.RPCType_WEBSOCKET:
				endpoint.websocketUrl = supplierEndpoint.Endpoint().Url
			// For now, only websocket & JSON-RPC types are supported, so JSON-RPC is the default.
			default:
				endpoint.url = supplierEndpoint.Endpoint().Url
			}
		}

		endpoints[endpoint.Addr()] = endpoint
	}

	return endpoints, nil
}
