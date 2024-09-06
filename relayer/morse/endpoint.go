package morse

import (
	"fmt"

	"github.com/pokt-foundation/pocket-go/provider"

	"github.com/pokt-foundation/portal-middleware/relayer"
)

// The relayer package's Endpoint interface is fulfilled by the endpoint struct below, which allows
// all code outside this package to uniquely identify any Morse endpoint, e.g. for the purpose of selecting
// the best endpoint when sending a relay.
var _ relayer.Endpoint = endpoint{}

// endpoint is used to convert a Morse endpoint, i.e. an entity that can serve relay requests,
// to the Endpoint defined in the relayer package.
// endpoint contains the address and URL of a Morse endpoint.
// The address is used to uniquely identify the endpoint, and the URL is used for sending relay.
type endpoint struct {
	address string
	url     string
}

func (e endpoint) Addr() relayer.EndpointAddr {
	return relayer.EndpointAddr(e.address)
}

func (e endpoint) PublicURL() string {
	return e.url
}

func endpointsFromSession(session provider.Session) []relayer.Endpoint {
	endpoints := make([]relayer.Endpoint, len(session.Nodes))
	for i, sessionNode := range session.Nodes {
		endpoints[i] = endpoint{
			address: sessionNode.Address,
			url:     sessionNode.ServiceURL,
		}
	}

	return endpoints
}

// getEndpoint returns a Morse endpoint, from the provided session's list of endpoints, which matches the input address.
// Note: this function is intentionally named getEndpoint to reflect its generality even though it returns
// a provider.Node struct. This is a legacy structure used and required to send a relay to a Morse endpoint.
func getEndpoint(session provider.Session, endpointAddr relayer.EndpointAddr) (provider.Node, error) {
	for _, node := range session.Nodes {
		if node.Address == string(endpointAddr) {
			return node, nil
		}
	}

	return provider.Node{}, fmt.Errorf("endpoint with address %s not in session", endpointAddr)
}
