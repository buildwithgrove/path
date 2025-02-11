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
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/buildwithgrove/path/observation"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/websockets"
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

	// WebsocketEndpoints is a temporary workaround to allow PATH to enable websocket
	// connections to a single user-provided websocket-enabled endpoint URL per service ID.
	// TODO_HACK(@commoddity, #143): Remove this field once the Shannon protocol supports websocket connections.
	WebsocketEndpoints map[protocol.ServiceID]string
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
	// Determine the type of service request and handle it accordingly.
	switch determineServiceRequestType(httpReq) {
	case websocketServiceRequest:
		g.handleWebSocketRequest(ctx, httpReq, w)
	default:
		g.handleHTTPServiceRequest(ctx, httpReq, w)
	}
}

// handleHTTPRequest handles a standard HTTP service request.
func (g Gateway) handleHTTPServiceRequest(ctx context.Context, httpReq *http.Request, w http.ResponseWriter) {
	// build a gatewayRequestContext with components necessary to process HTTP requests.
	gatewayRequestCtx := &requestContext{
		logger: g.Logger,

		gatewayObservations: getUserRequestGatewayObservations(),
		protocol:            g.Protocol,
		httpRequestParser:   g.HTTPRequestParser,
		metricsReporter:     g.MetricsReporter,
		dataReporter:        g.DataReporter,
		// TODO_MVP(@adshmh): build the gateway observation data and pass it to the request context.
		// TODO_MVP(@adshmh): build the HTTP request observation data and pass it to the request context.
	}

	defer func() {
		// Write the user-facing HTTP response.
		gatewayRequestCtx.WriteHTTPUserResponse(w)
		// Broadcast all observations, e.g. protocol-level, QoS-level, etc. contained in the gateway request context.
		gatewayRequestCtx.BroadcastAllObservations()
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

// getUserRequestGatewayObservations returns gateway-level observations for an organic request.
// Example: request originated from a user.
func getUserRequestGatewayObservations() observation.GatewayObservations {
	return observation.GatewayObservations{
		RequestType:  observation.RequestType_REQUEST_TYPE_ORGANIC,
		ReceivedTime: timestamppb.Now(),
	}
}

// handleWebsocketRequest handles WebSocket connection requests by directly connecting
// to the provided websocket endpoint URL.
//
// Current Implementation:
// - Bypasses protocol layer entirely as a temporary workaround
// - Directly uses provided WebSocket endpoint URL
// - Allows PATH to pass WebSocket messages without protocol support
//
// TODO_HACK(@commoddity, #143): Remove temporary workaround when Shannon protocol
// supports WebSocket connections. Changes will:
// - Utilize existing context system for endpoint selection
// - Select from available Shannon protocol service endpoints
// - Match HTTP request handling pattern
// - Use HandleWebsocketRequest method defined on gateway.Protocol
func (g Gateway) handleWebSocketRequest(ctx context.Context, httpReq *http.Request, w http.ResponseWriter) {
	// Upgrade HTTP to websocket connection first to enable error reporting
	// via websocket close messages for easier debugging
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	clientConn, err := upgrader.Upgrade(w, httpReq, nil)
	if err != nil {
		g.Logger.Error().Msg("handleWebsocketRequest: error upgrading websocket connection request")
		return
	}

	// Check if there are any websocket endpoint URLs set for the service ID in the config.
	if len(g.WebsocketEndpoints) == 0 {
		handleWebsocketError(g.Logger, clientConn, "handleWebsocketRequest: no websocket endpoint URLs are set in config")
		return
	}

	// Get service ID from HTTP request in order to select the correct websocket endpoint URL.
	serviceID, _, err := g.HTTPRequestParser.GetQoSService(ctx, httpReq)
	if err != nil {
		handleWebsocketError(g.Logger, clientConn, "handleWebsocketRequest: error getting QoS service")
		return
	}

	// Get the websocket endpoint URL for the service ID.
	endpointURL := g.WebsocketEndpoints[serviceID]
	if endpointURL == "" {
		errMsg := fmt.Sprintf("handleWebsocketRequest: websocket endpoint URL is not set in  config for service ID %s", serviceID)
		handleWebsocketError(g.Logger, clientConn, errMsg)
		return
	}

	// Create a websocket bridge to handle the websocket connection
	// between the Client and the websocket Endpoint.
	bridge, err := websockets.NewBridge(g.Logger, endpointURL, clientConn)
	if err != nil {
		handleWebsocketError(g.Logger, clientConn, "handleWebsocketRequest: error creating websocket bridge")
		return
	}

	// Run the websocket bridge in a separate goroutine.
	go bridge.Run()

	g.Logger.Info().Str("ws_endpoints_urls", endpointURL).Msg("handleWebsocketRequest: websocket connection established")
}

// handleWebsocketError logs errors and sends close message to the websocket client.
func handleWebsocketError(logger polylog.Logger, clientConn *websocket.Conn, errorMsg string) {
	logger.Error().Msg(errorMsg)

	closeMessage := websocket.FormatCloseMessage(websocket.CloseNormalClosure, errorMsg)

	if err := clientConn.WriteMessage(websocket.CloseMessage, closeMessage); err != nil {
		logger.Error().Msg("handleWebsocketError: error writing websocket close message")
	}

	clientConn.Close()
}
