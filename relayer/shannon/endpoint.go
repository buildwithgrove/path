package shannon

import (
	"fmt"

	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"

	"github.com/buildwithgrove/path/relayer"
)

// By fulfilling the relayer package Endpoint interface, the endpoint struct below allows
// all code outside this package to refer to a specific Shannon SupplierEndpoint as a uniquely identifiable entity
// that can serve relays.
var _ relayer.Endpoint = endpoint{}

// endpoint is used to fulfull a relayer package Endpoint using a Shannon SupplierEndpoint.
// An endpoint is identified by combining its supplier address and its URL, because
// in Shannon a supplier can have multiple endpoints for a service.
type endpoint struct {
	supplier string
	url      string
}

func (e endpoint) Addr() relayer.EndpointAddr {
	return relayer.EndpointAddr(fmt.Sprintf("%s-%s", e.supplier, e.url))
}

func (e endpoint) PublicURL() string {
	return e.url
}

// endpointsFromSession returns the list of all endpoints from a Shannon session.
func endpointsFromSession(session sessiontypes.Session) ([]endpoint, error) {
	sf := sdk.SessionFilter{
		Session: &session,
	}

	allEndpoints, err := sf.AllEndpoints()
	if err != nil {
		return nil, err
	}

	var endpoints []endpoint
	for _, supplierEndpoints := range allEndpoints {
		for _, supplierEndpoint := range supplierEndpoints {
			endpoints = append(endpoints, endpoint{
				supplier: string(supplierEndpoint.Supplier()),
				url:      supplierEndpoint.Endpoint().Url,
			})
		}
	}

	return endpoints, nil
}

// endpointFromSession returns the endpoint matching the input address from the list of all SupplierEndpoints of a Shannon session.
func endpointFromSession(session sessiontypes.Session, endpointAddr relayer.EndpointAddr) (endpoint, error) {
	endpoints, err := endpointsFromSession(session)
	if err != nil {
		return endpoint{}, fmt.Errorf("endpointFromSession: error getting all endpoints for session %s: %w", session.SessionId, err)
	}

	for _, e := range endpoints {
		if e.Addr() == endpointAddr {
			return e, nil
		}
	}

	return endpoint{}, fmt.Errorf("endpointFromSession: endpoint %s not found in the session", endpointAddr)
}
