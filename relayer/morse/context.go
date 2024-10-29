package morse

import (
	"context"
	"fmt"
	"time"

	"github.com/pokt-foundation/pocket-go/provider"
	sdkrelayer "github.com/pokt-foundation/pocket-go/relayer"

	"github.com/buildwithgrove/path/relayer"
)

// requestContext provides all the functionality required by the relayer package
// for handling a single service request.
var _ relayer.ProtocolRequestContext = &requestContext{}

// requestContext captures all the data required for handling a single service request.
type requestContext struct {
	fullNode  FullNode
	serviceID relayer.ServiceID

	// endpoints contains all the candidate endpoints available for processing a service request.
	endpoints map[relayer.EndpointAddr]endpoint
	// selectedEndpoint is the endpoint that has been selected for sending a relay.
	// Sending a relay will fail if this field is not set through a call to the SelectEndpoint method.
	selectedEndpoint *endpoint
}

// HandleServiceRequest satisfies the relayer package's ProtocolRequestContext interface.
// It uses the supplied payload to send a relay request to an endpoint, and verifies and returns the response.
func (rc *requestContext) HandleServiceRequest(payload relayer.Payload) (relayer.Response, error) {
	if rc.selectedEndpoint == nil {
		return relayer.Response{}, fmt.Errorf("HandleServiceRequest: no endpoint has been selected on service %s", rc.serviceID)
	}

	morseEndpoint, err := getEndpoint(rc.selectedEndpoint.session, rc.selectedEndpoint.Addr())
	if err != nil {
		return relayer.Response{},
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

	return relayer.Response{
		EndpointAddr:   rc.selectedEndpoint.Addr(),
		Bytes:          []byte(output.Response),
		HTTPStatusCode: output.StatusCode,
	}, err
}

// SelectEndpoint satisfies the relayer package's ProtocolRequestContext interface.
// It uses the supplied selector to select an endpoint from the request context's set of candidate endpoints
// for handling a service request.
func (rc *requestContext) SelectEndpoint(selector relayer.EndpointSelector) error {
	var endpoints []relayer.Endpoint
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

// sendRelay is a helper function for handling the low-level details of a Morse relay.
func (rc *requestContext) sendRelay(
	chainID string,
	node provider.Node,
	session provider.Session,
	aat provider.PocketAAT,
	timeoutMillisec int,
	payload relayer.Payload,
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
