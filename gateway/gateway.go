// Package gateway implements components for operating a gateway service.
//
// Protocol (Shannon):
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

// HandleServiceRequest implements PATH gateway's service request processing:
//
// This method acts as a request router that determines the type of incoming request
// (HTTP or WebSocket) and delegates to the appropriate handler. This separation
// allows for different processing flows while maintaining a unified entry point.
//
// Request Flow:
// 1. Determine request type (HTTP vs WebSocket upgrade)
// 2. Route to appropriate handler:
//   - WebSocket: Long-lived bidirectional connection with message-based observations
//   - HTTP: Request-response cycle with single observation broadcast
//
// TODO_FUTURE: Refactor when adding other protocols (e.g. gRPC):
//   - Extract generic processing into common method
//   - Keep protocol-specific details separate
func (g Gateway) HandleServiceRequest(
	ctx context.Context,
	httpReq *http.Request,
	responseWriter http.ResponseWriter,
) {
	// Determine the type of service request and handle it accordingly.
	switch determineServiceRequestType(httpReq) {
	case websocketServiceRequest:
		g.handleWebSocketRequest(ctx, httpReq, responseWriter)
	default:
		g.handleHTTPServiceRequest(ctx, httpReq, responseWriter)
	}
}

// handleHTTPServiceRequest handles a standard HTTP service request.
func (g Gateway) handleHTTPServiceRequest(
	ctx context.Context,
	httpReq *http.Request,
	responseWriter http.ResponseWriter,
) {
	logger := g.Logger.With("method", "handleHTTPServiceRequest")

	// Build a gatewayRequestContext with components necessary to process requests.
	gatewayRequestCtx := &requestContext{
		logger:              g.Logger,
		context:             ctx,
		gatewayObservations: getUserRequestGatewayObservations(httpReq),
		protocol:            g.Protocol,
		httpRequestParser:   g.HTTPRequestParser,
		metricsReporter:     g.MetricsReporter,
		dataReporter:        g.DataReporter,
	}

	// Initialize the GatewayRequestContext struct using the HTTP request.
	// e.g. extract the target service ID from the HTTP request.
	err := gatewayRequestCtx.InitFromHTTPRequest(httpReq)
	if err != nil {
		return
	}

	defer func() { // Broadcast all observations, e.g. protocol-level, QoS-level, etc. contained in the gateway request context.
		gatewayRequestCtx.BroadcastAllObservations()
		// Write the user-facing HTTP response. This is deliberately not called for websocket requests as they do not return an HTTP response.
		gatewayRequestCtx.WriteHTTPUserResponse(responseWriter)
	}()

	// TODO_TECHDEBT(@adshmh): Pass the context with deadline to QoS once it can handle deadlines.
	// Build the QoS context for the target service ID using the HTTP request's payload.
	err = gatewayRequestCtx.BuildQoSContextFromHTTP(httpReq)
	if err != nil {
		logger.Error().Err(err).Msg("❌ Error building QoS context for HTTP request")
		return
	}

	// TODO_MVP(@adshmh): Enhance the protocol interface used by the gateway to provide explicit error classification.
	// Implementation should:
	//   1. Differentiate between user errors (e.g., invalid Service ID in request) and system errors (e.g., endpoint timeout)
	//   2. Add error type field to protocol response structure
	//   3. Pass specific error codes from the protocol back to QoS service
	// This will allow the QoS service to return more helpful diagnostic messages and enable better metrics collection for different failure modes.
	//
	// Build the protocol context for the HTTP request.
	err = gatewayRequestCtx.BuildProtocolContextsFromHTTPRequest(httpReq)
	if err != nil {
		logger.Error().Err(err).Msg("❌ Error building protocol context for HTTP request")
		return
	}

	// Use the gateway request context to process the relay(s) corresponding to the HTTP request.
	// Any returned errors are ignored here and processed by the gateway context in the deferred calls.
	// See the `BroadcastAllObservations` method of `gateway.requestContext` struct for details.
	err = gatewayRequestCtx.HandleRelayRequest()
	if err != nil {
		logger.Error().Err(err).Msg("❌ Error processing relay request")
		return
	}
}

// handleWebSocketRequest handles WebSocket connection requests.
func (g Gateway) handleWebSocketRequest(
	ctx context.Context,
	httpReq *http.Request,
	w http.ResponseWriter,
) {
	logger := g.Logger.With("method", "handleWebSocketRequest")

	// Build a websocketRequestContext with components necessary to process websocket requests.
	websocketRequestCtx := &websocketRequestContext{
		logger:              g.Logger.With("component", "websocket_request_context"),
		context:             ctx,
		gatewayObservations: getUserRequestGatewayObservations(httpReq),
		protocol:            g.Protocol,
		httpRequestParser:   g.HTTPRequestParser,
		metricsReporter:     g.MetricsReporter,
		dataReporter:        g.DataReporter,
		messageSuccessChan:  make(chan struct{}, 100),
		messageErrorChan:    make(chan error, 100),
	}

	// Initialize the websocket request context using the HTTP request.
	err := websocketRequestCtx.InitFromHTTPRequest(httpReq)
	if err != nil {
		logger.Error().Err(err).Msg("❌ Error initializing websocket request context")
		return
	}

	// Build the QoS context for the target service ID using the HTTP request.
	// This replaces BuildQoSContextFromWebsocket and uses ParseHTTPRequest as single entry point.
	err = websocketRequestCtx.BuildQoSContextFromHTTP(httpReq)
	if err != nil {
		logger.Error().Err(err).Msg("❌ Error building QoS context for websocket request")
		return
	}

	// Build the protocol context for the websocket request.
	err = websocketRequestCtx.BuildProtocolContextFromHTTPRequest(httpReq)
	if err != nil {
		logger.Error().Err(err).Msg("❌ Error building protocol context for websocket request")
		return
	}

	// Handle the websocket connection request using the websocket request context.
	err = websocketRequestCtx.HandleWebsocketRequest(httpReq, w)
	if err != nil {
		logger.Error().Err(err).Msg("❌ Error processing websocket request")
		return
	}

	// For websockets, we don't immediately broadcast observations since the connection
	// will be long-lived and observations will be handled per-message basis.
	// The websocketRequestContext will handle observations during its lifecycle.
	logger.Info().Msg("✅ Successfully established websocket connection and started bridge")
}
