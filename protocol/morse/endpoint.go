package morse

import (
	"fmt"

	"github.com/pokt-foundation/pocket-go/provider"

	"github.com/buildwithgrove/path/protocol"
)

// The relayer package's Endpoint interface is fulfilled by the endpoint struct below, which:
// - Allows all code outside this package to uniquely identify any Morse endpoint
// - Enables selecting the best endpoint when sending a relay
var _ protocol.Endpoint = endpoint{}

// endpoint is used to convert a Morse endpoint to an Endpoint defined in the relayer package.
// An endpoint is considered to be an entity that can serve relay requests.
type endpoint struct {
	// Uniquely identifies the endpoint (i.e. onchain address).
	address string

	// URL of where to send the relay.
	url string

	// session to which the endpoint belongs for the purpose of sending relays.
	session provider.Session

	// app corresponding to the session to which the endpoint belongs and where to send relays
	app app
}

// IsValidForRelay checks if any of the required endpoint fields required for sending a relay are empty.
func (e endpoint) IsValidForRelay() bool {
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
