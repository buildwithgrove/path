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
	endpoints map[relayer.EndpointAddr]endpoint
	serviceID relayer.ServiceID

	// selectedEndpoint is the endpoint that has been selected for sending a relay.
	// Sending a relay will fail if this field is not set through a call to the SelectEndpoint method.
	selectedEndpoint *endpoint
}

func (rc *requestContext) HandleServiceRequest(payload relayer.Payload) (relayer.Response, error) {
	if rc.selectedEndpoint == nil {
		return relayer.Response{}, fmt.Errorf("handleServiceRequest: no endpoint has been selected on service %s", rc.serviceID)
	}
	endpoint := rc.selectedEndpoint

	morseEndpoint, err := getEndpoint(endpoint.session, endpoint.Addr())
	if err != nil {
		return relayer.Response{}, fmt.Errorf("handleServiceRequest: error matching the selected endpoint %s against session's nodes: %w", endpoint.Addr(), err)
	}

	output, err := rc.sendRelay(
		string(rc.serviceID),
		morseEndpoint,
		endpoint.session,
		endpoint.app.aat,
		// TODO_IMPROVE: chain-specific timeouts
		0, // SDK to use the default timeout.
		payload,
	)

	return relayer.Response{
		EndpointAddr:   endpoint.Addr(),
		Bytes:          []byte(output.Response),
		HTTPStatusCode: output.StatusCode,
	}, err
}

func (rc *requestContext) SelectEndpoint(selector relayer.EndpointSelector) error {
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
		return provider.RelayOutput{}, fmt.Errorf("relay: received null output from the SDK")
	}

	// TODO_DISCUSS: do we need to verify the node/proof structs?
	return *output.RelayOutput, err
}
