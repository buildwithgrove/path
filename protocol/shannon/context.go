package shannon

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"

	"github.com/buildwithgrove/path/gateway"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/websockets"
)

// requestContext provides all the functionality required by the gateway package
// for handling a single service request.
var _ gateway.ProtocolRequestContext = &requestContext{}

// RelayRequestSigner is used by the request context to sign the relay request.
// It takes an unsigned relay request and an application, and returns a relay request signed either by the gateway that has delegation from the app.
// If/when the Permissionless Gateway Mode is supported by the Shannon integration, the app's own private key may also be used for signing the relay request.
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
	var selectedEndpointAddr protocol.EndpointAddr
	if rc.selectedEndpoint != nil {
		selectedEndpointAddr = rc.selectedEndpoint.Addr()
	}

	response, err := rc.sendRelay(payload)
	if err != nil {
		return protocol.Response{EndpointAddr: selectedEndpointAddr},
			fmt.Errorf("relay: error sending relay for service %s endpoint %s: %w",
				rc.serviceID, selectedEndpointAddr, err,
			)
	}

	// The Payload field of the response received from the endpoint, i.e. the relay miner,
	// is a serialized http.Response struct. It needs to be deserialized into an HTTP Response struct
	// to access the Service's response body, status code, etc.
	relayResponse, err := deserializeRelayResponse(response.Payload)
	if err != nil {
		return protocol.Response{EndpointAddr: selectedEndpointAddr},
			fmt.Errorf("relay: error unmarshaling endpoint response into a POKTHTTP response for service %s endpoint %s: %w",
				rc.serviceID, selectedEndpointAddr, err,
			)
	}

	relayResponse.EndpointAddr = selectedEndpointAddr
	return relayResponse, nil
}

// HandleWebsocketRequest opens a persistent websocket connection to the selected endpoint.
// Satisfies the gateway.ProtocolRequestContext interface.
// TODO_HACK(@commoddity, #143): Utilize this method once the Shannon protocol supports websocket connections.
func (rc *requestContext) HandleWebsocketRequest(logger polylog.Logger, req *http.Request, w http.ResponseWriter) error {
	var selectedEndpointURL string
	if rc.selectedEndpoint != nil {
		selectedEndpointURL = rc.selectedEndpoint.PublicURL()
	}

	wsLogger := logger.With(
		"endpoint_url", selectedEndpointURL,
		"endpoint_addr", rc.selectedEndpoint.Addr(),
		"service_id", rc.serviceID,
	)

	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	clientConn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		wsLogger.Error().Err(err).Msg("Error upgrading websocket connection request")
		return err
	}

	session := rc.selectedEndpoint.session
	supplier := rc.selectedEndpoint.supplier
	bridge, err := websockets.NewBridge(
		wsLogger,
		selectedEndpointURL,
		session,
		supplier,
		rc.relayRequestSigner,
		rc.fullNode,
		clientConn,
	)
	if err != nil {
		wsLogger.Error().Err(err).Msg("Error creating websocket bridge")
		return err
	}

	go bridge.Run()

	wsLogger.Info().Msg("websocket connection established")

	return nil
}

// AvailableEndpoints returns the list of endpoints available under the request context, which is populated by the protocol instance
// at the time of creating the request context.
// It implements the gateway.ProtocolRequestContext interface.
func (rc *requestContext) AvailableEndpoints() ([]protocol.Endpoint, error) {
	var availableEndpoints []protocol.Endpoint

	for _, endpoint := range rc.endpoints {
		availableEndpoints = append(availableEndpoints, endpoint)
	}

	return availableEndpoints, nil
}

// TODO_MVP(@adshmh): implement the following method to return the MVP set of Shannon protocol-level observation.
// GetObservations returns the set of Shannon protocol-level observations for the current request context.
// The returned observations are used to:
// 1. Update the Shannon's endpoint store.
// 2. Report metrics on the operation of PATH (in the metrics package)
// 3. Share the observation on the messaging platform (NATS, REDIS, etc.) to be picked up by the data pipeline and any other interested entities.
//
// Implements the gateway.ProtocolRequestContext interface.
func (rc *requestContext) GetObservations() protocolobservations.Observations {
	return protocolobservations.Observations{}
}

// sendRelay sends a the supplied payload as a relay request to the endpoint selected for the request context through the SelectEndpoint method.
// It is required to fulfill the FullNode interface.
func (rc *requestContext) sendRelay(payload protocol.Payload) (*servicetypes.RelayResponse, error) {
	if rc.selectedEndpoint == nil {
		return nil, fmt.Errorf("sendRelay: no endpoint has been selected on service %s", rc.serviceID)
	}

	session := rc.selectedEndpoint.session
	if session.Application == nil {
		return nil, fmt.Errorf("sendRelay: nil app on session %s for service %s", session.SessionId, rc.serviceID)
	}
	app := *session.Application

	payloadBz := []byte(payload.Data)
	relayRequest, err := buildUnsignedRelayRequest(*rc.selectedEndpoint, session, payloadBz, payload.Path)
	if err != nil {
		return nil, err
	}

	signedRelayReq, err := rc.signRelayRequest(relayRequest, app)
	if err != nil {
		return nil, fmt.Errorf("sendRelay: error signing the relay request for app %s: %w", app.Address, err)
	}

	ctxWithTimeout, cancelFn := context.WithTimeout(context.Background(), time.Duration(payload.TimeoutMillisec)*time.Millisecond)
	defer cancelFn()

	responseBz, err := sendHttpRelay(ctxWithTimeout, rc.selectedEndpoint.url, signedRelayReq)
	if err != nil {
		return nil, fmt.Errorf("relay: error sending request to endpoint %s: %w", rc.selectedEndpoint.url, err)
	}

	// Validate the response
	response, err := rc.fullNode.ValidateRelayResponse(sdk.SupplierAddress(rc.selectedEndpoint.supplier), responseBz)
	if err != nil {
		return nil, fmt.Errorf("relay: error verifying the relay response for app %s, endpoint %s: %w", app.Address, rc.selectedEndpoint.url, err)
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

// buildUnsignedRelayRequest builds a ready-to-sign RelayRequest struct using the
// supplied endpoint, session, and payload.
// The returned RelayRequest is intended to be signed and sent to the endpoint to
// receive the endpoint's response.
func buildUnsignedRelayRequest(
	endpoint endpoint,
	session sessiontypes.Session,
	payload []byte,
	path string,
) (*servicetypes.RelayRequest, error) {
	// If the path is not empty (ie. for a REST service request), append it to the endpoint's URL
	url := endpoint.url
	if path != "" {
		url = fmt.Sprintf("%s%s", url, path)
	}

	// TODO_TECHDEBT: need to select the correct underlying request (HTTP, etc.) based on the selected service.
	jsonRpcHttpReq, err := shannonJsonRpcHttpRequest(payload, url)
	if err != nil {
		return nil, fmt.Errorf("error building a JSONRPC HTTP request for url %s: %w", url, err)
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
