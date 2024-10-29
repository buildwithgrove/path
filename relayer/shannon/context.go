package shannon

import (
	"fmt"

	"github.com/buildwithgrove/path/relayer"
)

// requestContext provides all the functionality required by the relayer package
// for handling a single service request.
var _ relayer.ProtocolRequestContext = &requestContext{}

// requestContext captures all the data required for handling a single service request.
type requestContext struct {
	fullNode  FullNode
	endpoints map[relayer.EndpointAddr]endpoint
	serviceID relayer.ServiceID

	// selectedEndpoint is the endpoint that has been selected for sending a relay.
	// Sending a relay will fail if this field is not set through a call to the SelectEndpoint method.
	selectedEndpoint *endpoint
}

func (rc *requestContext) SelectEndpoint(selector relayer.EndpointSelector) error {
	// Convert the map of endpoints to a list for easier business logic.
	var endpoints []relayer.Endpoint
	for _, endpoint := range rc.endpoints {
		endpoints = append(endpoints, endpoint)
	}
	if len(endpoints) == 0 {
		return fmt.Errorf("selectEndpoint: No endpoints found to select from on service %s", rc.serviceID)
	}

	selectedEndpointAddr, err := selector.Select(endpoints)
	if err != nil {
		return fmt.Errorf("selectEndpoint: selector returned an error for service %s: %w", rc.serviceID, err)
	}

	selectedEndpoint, found := rc.endpoints[selectedEndpointAddr]
	if !found {
		return fmt.Errorf("selectEndpoint: endpoint address %q does not match any available endpoints on service %s", selectedEndpointAddr, rc.serviceID)
	}

	rc.selectedEndpoint = &selectedEndpoint
	return nil
}

func (rc *requestContext) HandleServiceRequest(payload relayer.Payload) (relayer.Response, error) {
	if rc.selectedEndpoint == nil {
		return relayer.Response{}, fmt.Errorf("handleServiceRequest: no endpoint has been selected on service %s", rc.serviceID)
	}
	endpoint := rc.selectedEndpoint

	session := endpoint.session
	if session.Application == nil {
		return relayer.Response{}, fmt.Errorf("handleServiceRequest: nil app on session %s for service %s", session.SessionId, rc.serviceID)
	}
	app := *session.Application

	response, err := rc.fullNode.SendRelay(app, session, *rc.selectedEndpoint, payload)
	if err != nil {
		return relayer.Response{EndpointAddr: endpoint.Addr()},
			fmt.Errorf("relay: error sending relay for service %s endpoint %s: %w",
				rc.serviceID, endpoint.Addr(), err,
			)
	}

	// The Payload field of the response received from the endpoint, i.e. the relay miner,
	// is a serialized http.Response struct. It needs to be deserialized into an HTTP Response struct
	// to access the Service's response body, status code, etc.
	relayResponse, err := deserializeRelayResponse(response.Payload)
	if err != nil {
		return relayer.Response{EndpointAddr: endpoint.Addr()},
			fmt.Errorf("relay: error unmarshalling endpoint response into a POKTHTTP response for service %s app %s endpoint %s: %w",
				rc.serviceID, app.Address, endpoint.Addr(), err,
			)
	}

	relayResponse.EndpointAddr = endpoint.Addr()
	return relayResponse, nil
}
