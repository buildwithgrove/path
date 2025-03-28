package morse

import (
	"fmt"

	"github.com/pokt-foundation/pocket-go/provider"

	"github.com/buildwithgrove/path/protocol"
)

// The relayer package's Endpoint interface is fulfilled by the endpoint struct below, which allows
// all code outside this package to uniquely identify any Morse endpoint, e.g. for the purpose of selecting
// the best endpoint when sending a relay.
var _ protocol.Endpoint = endpoint{}

// endpoint is used to convert a Morse endpoint, i.e. an entity that can serve relay requests,
// to the Endpoint defined in the relayer package.
// endpoint contains the address and URL of a Morse endpoint.
// The address is used to uniquely identify the endpoint, and the URL is used for sending relay.
type endpoint struct {
	address string
	url     string

	// session holds the session to which the endpoint belongs for the purpose of sending relays.
	// this is used for sending relays to the endpoint.
	session provider.Session

	// app holds the app corresponding to the session to which the endpoint belongs.
	// this is used for sending relays to the endpoint.
	app app
}

func (e endpoint) IsEmpty() bool {
	return e.address == "" || e.url == "" || len(e.session.Nodes) == 0 || e.app.IsEmpty()
}

func (e endpoint) Addr() protocol.EndpointAddr {
	return protocol.EndpointAddr(e.address)
}

func (e endpoint) PublicURL() string {
	return e.url
}

func getEndpointsFromAppSession(app app, session provider.Session) []endpoint {
	endpoints := make([]endpoint, len(session.Nodes))
	for i, sessionNode := range session.Nodes {
		endpoints[i] = endpoint{
			address: sessionNode.Address,
			url:     sessionNode.ServiceURL,
			session: session,
			app:     app,
		}
	}

	return endpoints
}

// getEndpoint returns a Morse endpoint, from the provided session's list of endpoints, which matches the input address.
// Note: this function is intentionally named getEndpoint to reflect its generality even though it returns
// a provider.Node struct. This is a legacy structure used and required to send a relay to a Morse endpoint.
func getEndpoint(session provider.Session, endpointAddr protocol.EndpointAddr) (provider.Node, error) {
	for _, node := range session.Nodes {
		if node.Address == string(endpointAddr) {
			return node, nil
		}
	}

	return provider.Node{}, fmt.Errorf("endpoint with address %s not in session", endpointAddr)
}
