package shannon

import (
	"context"
	"errors"
	"fmt"
	"time"

	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
)

// requestContext provides all the functionality required by the gateway package
// for handling a single service request.
var _ gateway.ProtocolRequestContext = &requestContext{}

// RelayRequestSigner is used by the request context to sign the relay request.
type RelayRequestSigner interface {
	SignRelayRequest(req *servicetypes.RelayRequest, app apptypes.Application) (*servicetypes.RelayRequest, error)
}

// requestContext captures all the data required for handling a single service request.
type requestContext struct {
	fullNode           FullNode
	serviceID          protocol.ServiceID
	relayRequestSigner RelayRequestSigner

	// endpoints contains all the candidate endpoints available for processing a service request.
	endpoints map[protocol.EndpointAddr]endpoint
	// selectedEndpoint is the endpoint that has been selected for sending a relay.
	// Sending a relay will fail if this field is not set through a call to the SelectEndpoint method.
	selectedEndpoint *endpoint
}

// SelectEndpoint satisfies the gateway package's ProtocolRequestContext interface.
// It uses the supplied selector to select an endpoint from the request context's set of candidate endpoints
// for handling a service request.
func (rc *requestContext) SelectEndpoint(selector protocol.EndpointSelector) error {
	// Convert the map of endpoints to a list for easier business logic.
	var endpoints []protocol.Endpoint
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

// HandleServiceRequest satisfies the gateway package's ProtocolRequestContext interface.
// It uses the supplied payload to send a relay request to an endpoint, and verifies and returns the response.
func (rc *requestContext) HandleServiceRequest(payload protocol.Payload) (protocol.Response, error) {
	if rc.selectedEndpoint == nil {
		return protocol.Response{}, fmt.Errorf("handleServiceRequest: no endpoint has been selected on service %s", rc.serviceID)
	}
	endpoint := rc.selectedEndpoint

	session := endpoint.session
	if session.Application == nil {
		return protocol.Response{}, fmt.Errorf("handleServiceRequest: nil app on session %s for service %s", session.SessionId, rc.serviceID)
	}
	app := *session.Application

	response, err := rc.sendRelay(app, session, *rc.selectedEndpoint, payload)
	if err != nil {
		return protocol.Response{EndpointAddr: endpoint.Addr()},
			fmt.Errorf("relay: error sending relay for service %s endpoint %s: %w",
				rc.serviceID, endpoint.Addr(), err,
			)
	}

	// The Payload field of the response received from the endpoint, i.e. the relay miner,
	// is a serialized http.Response struct. It needs to be deserialized into an HTTP Response struct
	// to access the Service's response body, status code, etc.
	relayResponse, err := deserializeRelayResponse(response.Payload)
	if err != nil {
		return protocol.Response{EndpointAddr: endpoint.Addr()},
			fmt.Errorf("relay: error unmarshalling endpoint response into a POKTHTTP response for service %s app %s endpoint %s: %w",
				rc.serviceID, app.Address, endpoint.Addr(), err,
			)
	}

	relayResponse.EndpointAddr = endpoint.Addr()
	return relayResponse, nil
}

// AvailableEndpoints returns the pre-set list of available endpoints.
// It implements the gateway.ProtocolRequestContext interface.
func (rc *requestContext) AvailableEndpoints() ([]protocol.Endpoint, error) {
	var availableEndpoints []protocol.Endpoint

	for _, endpoint := range rc.endpoints {
		availableEndpoints = append(availableEndpoints, endpoint)
	}

	return availableEndpoints, nil
}

// sendRelay sends a the supplied payload as a relay request to the supplied endpoint.
// It is required to fulfill the FullNode interface.
func (rc *requestContext) sendRelay(
	app apptypes.Application,
	session sessiontypes.Session,
	endpoint endpoint,
	payload protocol.Payload,
) (*servicetypes.RelayResponse, error) {
	relayRequest, err := buildRelayRequest(endpoint, session, []byte(payload.Data))
	if err != nil {
		return nil, err
	}

	signedRelayReq, err := rc.signRelayRequest(relayRequest, app)
	if err != nil {
		return nil, fmt.Errorf("relay: error signing the relay request for app %s: %w", app.Address, err)
	}

	ctxWithTimeout, cancelFn := context.WithTimeout(context.Background(), time.Duration(payload.TimeoutMillisec)*time.Millisecond)
	defer cancelFn()

	responseBz, err := sendHttpRelay(ctxWithTimeout, endpoint.url, signedRelayReq)
	if err != nil {
		return nil, fmt.Errorf("relay: error sending request to endpoint %s: %w", endpoint.url, err)
	}

	// Validate the response
	response, err := rc.fullNode.ValidateRelayResponse(sdk.SupplierAddress(endpoint.supplier), responseBz)
	if err != nil {
		return nil, fmt.Errorf("relay: error verifying the relay response for app %s, endpoint %s: %w", app.Address, endpoint.url, err)
	}

	return response, nil
}

func (rc *requestContext) signRelayRequest(unsignedRelayReq *servicetypes.RelayRequest, app apptypes.Application) (*servicetypes.RelayRequest, error) {
	// Verify the relay request's metadata, specifically the session header.
	// Note: cannot use the RelayRequest's ValidateBasic() method here, as it looks for a signature in the struct, which has not been added yet at this point.
	meta := unsignedRelayReq.GetMeta()

	if meta.GetSessionHeader() == nil {
		return nil, errors.New("signRelayRequest: relay request is missing session header")
	}

	sessionHeader := meta.GetSessionHeader()
	if err := sessionHeader.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("signRelayRequest: relay request session header is invalid: %w", err)
	}

	// Sign the relay request using the selected app's private key
	return rc.relayRequestSigner.SignRelayRequest(unsignedRelayReq, app)
}

// buildRelayRequest builds a ready-to-sign RelayRequest struct using the supplied endpoint, session, and payload.
// The returned RelayRequest can be signed and sent to the endpoint to receive the endpoint's response.
func buildRelayRequest(endpoint endpoint, session sessiontypes.Session, payload []byte) (*servicetypes.RelayRequest, error) {
	// TODO_TECHDEBT: need to select the correct underlying request (HTTP, etc.) based on the selected service.
	jsonRpcHttpReq, err := shannonJsonRpcHttpRequest(payload, endpoint.url)
	if err != nil {
		return nil, fmt.Errorf("error building a JSONRPC HTTP request for url %s: %w", endpoint.url, err)
	}

	relayRequest, err := embedHttpRequest(jsonRpcHttpReq)
	if err != nil {
		return nil, fmt.Errorf("error embedding a JSONRPC HTTP request for url %s: %w", endpoint.url, err)
	}

	// TODO_MVP(@adshmh): use the new `FilteredSession` struct provided by the Shannon SDK to get the session and the endpoint.
	relayRequest.Meta = servicetypes.RelayRequestMetadata{
		SessionHeader:           session.Header,
		SupplierOperatorAddress: string(endpoint.supplier),
	}

	return relayRequest, nil
}
