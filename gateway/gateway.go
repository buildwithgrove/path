// Package gateway implements components for operating a gateway service.
//
// Protocols (Morse, Shannon):
// - Provide available endpoints for a service
// - Send relays to specific endpoints
//
// Gateways:
// - Select endpoints for relay transmission
//
// QoS Services:
// - Interpret user requests into endpoint payloads
// - Select optimal endpoints for request handling
//
// TODO_MVP(@adshmh): add a README with a diagram of all the above.
// TODO_MVP(@adshmh): add a section for the following packages once they are added: Metrics, Message.
package gateway

import (
	"context"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/buildwithgrove/path/observation"
)

// Gateway handles end-to-end service requests via HandleHTTPServiceRequest:
// - Receives user request
// - Processes request
// - Returns response
//
// TODO_FUTURE: Current HTTP-only format supports JSONRPC, REST, Websockets
// and gRPC. May expand to other formats in future.
type Gateway struct {
	Logger polylog.Logger

	// HTTPRequestParser is used by the gateway instance to
	// interpret an HTTP request as a pair of service ID and
	// its corresponding QoS instance.
	HTTPRequestParser

	// The Protocol instance is used to fulfill the
	// service requests received by the gateway through
	// sending the service payload to an endpoint.
	Protocol

	// MetricsReporter is used to export metrics based on observations made in handling service requests.
	MetricsReporter RequestResponseReporter

	// DataReporter is used to export, to the data pipeline, observations made in handling service requests.
	// It is declared separately from the `MetricsReporter` to be consistent with the gateway package's role
	// of explicitly defining PATH gateway's components and their interactions.
	DataReporter RequestResponseReporter
}

// HandleHTTPServiceRequest implements PATH gateway's HTTP request processing:
//
// This is written as a template method to allow customization of steps.
// Template pattern allows customization of service steps:
// - Establishing QoS context
// - Sending payload via relay protocols
// Reference: https://en.wikipedia.org/wiki/Template_method_pattern
//
// TODO_FUTURE: Refactor when adding other protocols (e.g. gRPC):
// - Extract generic processing into common method
// - Keep HTTP-specific details separate
func (g Gateway) HandleServiceRequest(ctx context.Context, httpReq *http.Request, w http.ResponseWriter) {
	// build a gatewayRequestContext with components necessary to process requests.
	gatewayRequestCtx := &requestContext{
		logger:              g.Logger,
		gatewayObservations: getUserRequestGatewayObservations(),
		protocol:            g.Protocol,
		httpRequestParser:   g.HTTPRequestParser,
		metricsReporter:     g.MetricsReporter,
		dataReporter:        g.DataReporter,
		// TODO_MVP(@adshmh): build the gateway observation data and pass it to the request context.
		// TODO_MVP(@adshmh): build the HTTP request observation data and pass it to the request context.
	}

	defer func() {
		// Broadcast all observations, e.g. protocol-level, QoS-level, etc. contained in the gateway request context.
		gatewayRequestCtx.BroadcastAllObservations()
	}()

	// Initialize the GatewayRequestContext struct using the HTTP request.
	// e.g. extract the target service ID from the HTTP request.
	err := gatewayRequestCtx.InitFromHTTPRequest(httpReq)
	if err != nil {
		return
	}

	// Determine the type of service request and handle it accordingly.
	switch determineServiceRequestType(httpReq) {
	case websocketServiceRequest:
		g.handleWebSocketRequest(ctx, httpReq, gatewayRequestCtx, w)
	default:
		g.handleHTTPServiceRequest(ctx, httpReq, gatewayRequestCtx, w)
	}
}

// handleHTTPRequest handles a standard HTTP service request.
func (g Gateway) handleHTTPServiceRequest(ctx context.Context, httpReq *http.Request, gatewayRequestCtx *requestContext, w http.ResponseWriter) {
	defer func() {
		// Write the user-facing HTTP response. This is deliberately not called for websocket requests as they do not return an HTTP response.
		gatewayRequestCtx.WriteHTTPUserResponse(w)
	}()

	// Initialize the GatewayRequestContext struct using the HTTP request.
	// e.g. extract the target service ID from the HTTP request.
	err := gatewayRequestCtx.InitFromHTTPRequest(httpReq)
	if err != nil {
		return
	}

	// Build the QoS context for the target service ID using the HTTP request's payload.
	err = gatewayRequestCtx.BuildQoSContextFromHTTP(ctx, httpReq)
	if err != nil {
		return
	}

	// Build the protocol context for the HTTP request.
	err = gatewayRequestCtx.BuildProtocolContextFromHTTP(httpReq)
	if err != nil {
		return
	}

	// Use the gateway request context to process the relay(s) corresponding to the HTTP request.
	// Any returned errors are ignored here and processed by the gateway context in the deferred calls.
	// See the `BroadcastAllObservations` method of `gateway.requestContext` struct for details.
	_ = gatewayRequestCtx.HandleRelayRequest()
}

// handleWebsocketRequest handles WebSocket connection requests
func (g Gateway) handleWebSocketRequest(ctx context.Context, httpReq *http.Request, gatewayRequestCtx *requestContext, w http.ResponseWriter) {
	// Build the QoS context for the target service ID using the HTTP request's payload.
	err := gatewayRequestCtx.BuildQoSContextFromWebsocket(ctx, httpReq)
	if err != nil {
		return
	}

	// Build the protocol context for the HTTP request.
	err = gatewayRequestCtx.BuildProtocolContextFromHTTP(httpReq)
	if err != nil {
		return
	}

	// Use the gateway request context to process the websocket connection request.
	// Any returned errors are ignored here and processed by the gateway context in the deferred calls.
	// See the `BroadcastAllObservations` method of `gateway.requestContext` struct for details.
	_ = gatewayRequestCtx.HandleWebsocketRequest(httpReq, w)
}

// getUserRequestGatewayObservations returns gateway-level observations for an organic request.
// Example: request originated from a user.
func getUserRequestGatewayObservations() observation.GatewayObservations {
	return observation.GatewayObservations{
		RequestType:  observation.RequestType_REQUEST_TYPE_ORGANIC,
		ReceivedTime: timestamppb.Now(),
	}
}
