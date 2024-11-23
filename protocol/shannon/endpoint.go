package shannon

import (
	"fmt"

	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"

	"github.com/buildwithgrove/path/protocol"
)

// By fulfilling the protocol package Endpoint interface, the endpoint struct below allows
// all code outside this package to refer to a specific Shannon SupplierEndpoint as a uniquely identifiable entity
// that can serve relays.
var _ protocol.Endpoint = endpoint{}

// endpoint is used to fulfull a protocol package Endpoint using a Shannon SupplierEndpoint.
// An endpoint is identified by combining its supplier address and its URL, because
// in Shannon a supplier can have multiple endpoints for a service.
type endpoint struct {
	supplier string
	url      string

	// TODO_IMPROVE: If the same endpoint is in the session of multiple apps at the same time,
	// the first app will be chosen. A randomization among the apps in this (unlikely) scenario
	// may be needed.
	// session is the active session corresponding to the app, of which the endpoint is a member.
	session sessiontypes.Session
}

// TODO_UPNEXT(@adshmh): replace EndpointAddr with a URL; a single URL should be treated the same regardless of the app to which it is attached.
// For protocol-level concerns: the (app/session, URL) should be taken into account; e.g. a healthy endpoint may have been maxed out for a particular app.
// For QoS-level concerns: only the URL of the endpoint matters; e.g. an unhealthy endpoint should be skipped regardless of the app/session to which it is attached.
func (e endpoint) Addr() protocol.EndpointAddr {
	return protocol.EndpointAddr(fmt.Sprintf("%s-%s", e.supplier, e.url))
}

func (e endpoint) PublicURL() string {
	return e.url
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
		for _, supplierEndpoint := range supplierEndpoints {
			endpoint := endpoint{
				supplier: string(supplierEndpoint.Supplier()),
				url:      supplierEndpoint.Endpoint().Url,
				// Set the session field on the endpoint for efficient lookup when sending relays.
				session: session,
			}
			endpoints[endpoint.Addr()] = endpoint
		}
	}

	return endpoints, nil
}
