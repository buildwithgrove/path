package morse

import (
	"context"
	"fmt"
	"time"

	"github.com/pokt-foundation/pocket-go/provider"
	sdkrelayer "github.com/pokt-foundation/pocket-go/relayer"

	"github.com/buildwithgrove/path/gateway"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/protocol"
)

// requestContext provides all the functionality required by the gateway package
// for handling a single service request.
var _ gateway.ProtocolRequestContext = &requestContext{}

// requestContext captures all the data required for handling a single service request.
type requestContext struct {
	fullNode  FullNode
	serviceID protocol.ServiceID

	// endpoints contains all the candidate endpoints available for processing a service request.
	endpoints map[protocol.EndpointAddr]endpoint
	// selectedEndpoint is the endpoint that has been selected for sending a relay.
	// NOTE: Sending a relay will fail if this field is not set through a call to the SelectEndpoint method.
	selectedEndpoint *endpoint
}

// AvailableEndpoints returns the list of available endpoints for the current request context.
// This list is populated by the Morse protocol instance when building the request context.
// This method implements the gateway.ProtocolRequestContext interface.
func (rc *requestContext) AvailableEndpoints() ([]protocol.Endpoint, error) {
	var availableEndpoints []protocol.Endpoint

	for _, endpoint := range rc.endpoints {
		availableEndpoints = append(availableEndpoints, endpoint)
	}

	return availableEndpoints, nil
}

// HandleServiceRequest satisfies the gateway package's ProtocolRequestContext interface.
// It uses the supplied payload to send a relay request to an endpoint, and verifies and returns the response.
func (rc *requestContext) HandleServiceRequest(payload protocol.Payload) (protocol.Response, error) {
	if rc.selectedEndpoint == nil {
		return protocol.Response{}, fmt.Errorf("HandleServiceRequest: no endpoint has been selected on service %s", rc.serviceID)
	}

	morseEndpoint, err := getEndpoint(rc.selectedEndpoint.session, rc.selectedEndpoint.Addr())
	if err != nil {
		return protocol.Response{},
			fmt.Errorf("HandleServiceRequest: error matching the selected endpoint %s against session's nodes: %w", rc.selectedEndpoint.Addr(), err)
	}

	output, err := rc.sendRelay(
		string(rc.serviceID),
		morseEndpoint,
		rc.selectedEndpoint.session,
		rc.selectedEndpoint.app.aat,
		// TODO_IMPROVE: chain-specific timeouts
		0, // SDK to use the default timeout.
		payload,
	)

	return protocol.Response{
		EndpointAddr:   rc.selectedEndpoint.Addr(),
		Bytes:          []byte(output.Response),
		HTTPStatusCode: output.StatusCode,
	}, err
}

// SelectEndpoint satisfies the gateway package's ProtocolRequestContext interface.
// It uses the supplied selector to select an endpoint from the request context's set of candidate endpoints
// for handling a service request.
func (rc *requestContext) SelectEndpoint(selector protocol.EndpointSelector) error {
	var endpoints []protocol.Endpoint
	for _, endpoint := range rc.endpoints {
		endpoints = append(endpoints, endpoint)
	}
	if len(endpoints) == 0 {
		return fmt.Errorf("SelectEndpoint: No endpoints found to select from on service %s", rc.serviceID)
	}

	selectedEndpointAddr, err := selector.Select(endpoints)
	if err != nil {
		return fmt.Errorf("SelectEndpoint: selector returned an error for service %s: %w", rc.serviceID, err)
	}

	selectedEndpoint, found := rc.endpoints[selectedEndpointAddr]
	if !found {
		return fmt.Errorf("SelectEndpoint: endpoint address %q does not match any available endpoints on service %s", selectedEndpointAddr, rc.serviceID)
	}

	rc.selectedEndpoint = &selectedEndpoint
	return nil
}

// TODO_MVP(@adshmh): implement the following method to return the MVP set of Shannon protocol-level observation.
// GetObservations returns the set of Shannon protocol-level observations for the current request context.
// The returned observations are used to:
// 1. Update the Shannon's endpoint store.
// 2. Report metrics on the operation of PATH (in the metrics package)
// 3. Share the observation on the messaging platform (NATS, REDIS, etc.) to be picked up by the data pipeline and any other interested entities.
//
// This method implements the gateway.ProtocolRequestContext interface.
func (rc *requestContext) GetObservations() protocolobservations.Observations {
	return protocolobservations.Observations{}
}

// sendRelay is a helper function for handling the low-level details of a Morse relay.
func (rc *requestContext) sendRelay(
	chainID string,
	node provider.Node,
	session provider.Session,
	aat provider.PocketAAT,
	timeoutMillisec int,
	payload protocol.Payload,
) (provider.RelayOutput, error) {
	fullNodeInput := &sdkrelayer.Input{
		Blockchain: chainID,
		Node:       &node,
		Session:    &session,
		PocketAAT:  &aat,
		Data:       payload.Data,
		Method:     payload.Method,
		Path:       payload.Path,
	}

	timeout := timeoutMillisec
	if timeout == 0 {
		timeout = defaultRelayTimeoutMillisec
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	output, err := rc.fullNode.SendRelay(ctx, fullNodeInput)
	if output.RelayOutput == nil {
		return provider.RelayOutput{}, fmt.Errorf("relay: received null RelayOutput field in the relay response from the SDK")
	}

	// TODO_TECHDEBT: complete the following items regarding the node and proof structs
	// 1. Verify their correctness
	// 2. Pass a logger to the request context to log them in debug mode.
	if output.Node == nil {
		return provider.RelayOutput{}, fmt.Errorf("relay: received null Node field in the relay response from the SDK")
	}

	if output.Proof == nil {
		return provider.RelayOutput{}, fmt.Errorf("relay: received null Proof field in the relay response from the SDK")
	}

	return *output.RelayOutput, err
}
