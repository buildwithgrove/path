// gateway package defines the components and their interactions necessary for operating a gateway.
// It defines the requirements and steps of sending relays from the perspective of:
// a) protocols, i.e. Morse and Shannon protocols, which provide:
// - a list of endpoints available for a service.
// - a function for sending a relay to a specific endpoint.
// b) gateways, which are required to provide a function for
// selecting an endpoint to which the relay is to be sent.
// c) Quality-of-Service (QoS) services: which provide:
// - interpretation of the user's request as the payload to be sent to an endpoint.
// - selection of the best endpoint for handling a user's request.
//
// TODO_MVP(@adshmh): add a README with a diagram of all the above.
// TODO_MVP(@adshmh): add a section for the following packages once they are added: Metrics, Message.
package gateway

import (
	"context"
	"net/http"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/websockets"
	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// Gateway performs end-to-end handling of all service requests
// through a single function, i.e. HandleHTTPServiceRequest,
// which starts from the point of receiving a user request,
// and ends once a response has been returned to the user.
// TODO_FUTURE: Currently, the only supported format for both the
// request and the response is HTTP as it is sufficient for JSONRPC,
// REST, Websockets and gRPC but may expand in the future.
type Gateway struct {
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

	// WebsocketEndpointURLs is a temporary workaround to allow PATH to enable websocket 
	// connections to a single user-provided websocket-enabled endpoint URL per service ID.
	// TODO_FUTURE(@commoddity)[WebSockets]: Remove this field once the Shannon protocol supports websocket connections.
	WebsocketEndpointURLs map[protocol.ServiceID]string

	Logger polylog.Logger
}

// HandleHTTPServiceRequest defines the steps the PATH gateway takes to
// handle a service request. It is currently limited in scope to
// service requests received over HTTP, to avoid adding any abstraction
// layers that are not necessary yet.
// TODO_FUTURE: Once other service request protocols, e.g. GRPC, are
// within scope, the HandleHTTPServiceRequest needs to be
// refactored to keep HTTP-specific details and move the generic service
// request processing steps into a common method.
//
// HandleServiceRequest is written as a template method to allow the customization of steps
// invovled in serving a service request, e.g.:
//   - establishing a QoS context for the HTTP request.
//   - sending the service payload through a relaying protocol, etc.
//
// See the following link for more details:
// https://en.wikipedia.org/wiki/Template_method_pattern
func (g Gateway) HandleServiceRequest(ctx context.Context, httpReq *http.Request, w http.ResponseWriter) {
	// Determine the type of service request and handle it accordingly.
	switch determineServiceRequestType(httpReq) {
	case websocketServiceRequest:
		g.handleWebsocketRequest(ctx, httpReq, w)
	default:
		g.handleHTTPServiceRequest(ctx, httpReq, w)
	}
}

// handleHTTPRequest handles a standard HTTP service request.
func (g Gateway) handleHTTPServiceRequest(ctx context.Context, httpReq *http.Request, w http.ResponseWriter) {
	// build a gatewayRequestContext with components necessary to process HTTP requests.
	gatewayRequestCtx := &requestContext{
		protocol:          g.Protocol,
		httpRequestParser: g.HTTPRequestParser,
		metricsReporter:   g.MetricsReporter,
		dataReporter:      g.DataReporter,
		logger:            g.Logger,
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
	_ = gatewayRequestCtx.HandleRelayRequest()
}

// handleWebsocketRequest handles a WebSocket connection request direct to the provided websocket endpoint URL.
// NOTE: As a temporary workaround, websocket connections currently bypass the protocol entirely and utilize the
// provided websocket endpoint URL to send and receive messages. This allows PATH to pass websocket messages until
// the Shannon protocol supports websocket connections, which will enable onchain websocket support.
//
// TODO_FUTURE(@commoddity)[WebSockets]: Remove this temporary workaround once the Shannon protocol supports websocket connections.
// This will entail utilizing the existing system of contexts to select an endpoint to serve the websocket connection
// from among the available service endpoints on the Shannon protocol in the same way that HTTP requests are handled.
// A method `HandleWebsocketRequest` is defined on the `gateway.Protocol` interface for this purpose.
func (g Gateway) handleWebsocketRequest(ctx context.Context, httpReq *http.Request, w http.ResponseWriter) {
	if len(g.WebsocketEndpointURLs) == 0 {
		g.Logger.Error().Msg("no websocket endpoint URLs are set")
		return
	}

	// Get service ID from HTTP request in order to select the correct websocket endpoint URL.
	serviceID, _, err := g.HTTPRequestParser.GetQoSService(ctx, httpReq)
	if err != nil {
		g.Logger.Error().Msg("error getting QoS service")
		return
	}

	// Get the websocket endpoint URL for the service ID.
	endpointURL := g.WebsocketEndpointURLs[serviceID]
	if endpointURL == "" {
		g.Logger.Error().Msgf("websocket endpoint URL is not set for service ID %s", serviceID)
		return
	}

	// Upgrade the HTTP request to a websocket connection.
	var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	clientConn, err := upgrader.Upgrade(w, httpReq, nil)
	if err != nil {
		g.Logger.Error().Msg("error upgrading websocket connection request")
		return
	}

	// Create a websocket bridge to handle the websocket connection
	// between the Client and the websocket Endpoint.
	bridge, err := websockets.NewBridge(g.Logger, endpointURL, clientConn)
	if err != nil {
		g.Logger.Error().Msg("error creating websocket bridge")
		return
	}

	// Run the websocket bridge in a separate goroutine.
	go bridge.Run()

	g.Logger.Info().Str("websocket_endpoint_url", endpointURL).Msg("websocket connection established")
}
