package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/buildwithgrove/path/observation"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/websockets"
)

var _ websockets.WebsocketMessageProcessor = &websocketRequestContext{}

// websocketRequestContext is responsible for orchestrating the flow of websocket messages
// between client and endpoint. It handles:
//   - QoS validation and context building
//   - Protocol context setup (single endpoint selection vs HTTP's multiple endpoints)
//   - Message routing and observation (per-message vs HTTP's per-request)
//   - Bridge lifecycle management
//
// Key differences from HTTP requestContext:
//   - Single endpoint selection (websockets can't do parallel requests)
//   - Per-message observations instead of per-request observations
//   - Long-lived connection management vs one-shot request/response
type websocketRequestContext struct {
	logger polylog.Logger

	// Enforces request completion deadline.
	context context.Context

	// httpRequestParser is used by the request context to interpret an HTTP request as a pair of:
	// 	1. service ID
	// 	2. The service ID's corresponding QoS instance.
	httpRequestParser HTTPRequestParser

	// metricsReporter is used to export metrics based on observations made in handling service requests.
	metricsReporter RequestResponseReporter

	// dataReporter is used to export, to the data pipeline, observations made in handling service requests.
	dataReporter RequestResponseReporter

	// QoS related request context
	serviceID  protocol.ServiceID
	serviceQoS QoSService
	qosCtx     RequestQoSContext

	// Protocol related request context
	protocol Protocol
	// For websockets, we only use a single protocol context
	protocolCtx ProtocolRequestContext

	// gatewayObservations stores gateway related observations.
	gatewayObservations *observation.GatewayObservations

	// Channel for receiving message processing notifications from the bridge
	messageObservationsChan chan *observation.RequestResponseObservations

	// connectionStartTime tracks when the WebSocket connection was established
	// Used to calculate ConnectionDurationMs in observations
	connectionStartTime *time.Time
}

// ---------- Websocket Connection Establishment ----------

// InitFromHTTPRequest builds the required context for serving a WebSocket request.
// Similar to requestContext.InitFromHTTPRequest but for websockets.
func (wrc *websocketRequestContext) InitFromHTTPRequest(httpReq *http.Request) error {
	wrc.logger = wrc.getWSRequestLogger(httpReq)

	// Extract the service ID and find the target service's corresponding QoS instance.
	serviceID, serviceQoS, err := wrc.httpRequestParser.GetQoSService(wrc.context, httpReq)
	wrc.serviceID = serviceID
	if err != nil {
		// Update gateway observations
		wrc.updateGatewayObservations(err)
		wrc.logger.Error().Err(err).Msg("HTTP request rejected by parser for websocket connection")
		return fmt.Errorf("websocket request rejected by parser: %w", err)
	}

	wrc.serviceQoS = serviceQoS
	return nil
}

// BuildQoSContextFromHTTP builds the QoS context instance using the supplied HTTP request.
func (wrc *websocketRequestContext) BuildQoSContextFromHTTP(httpReq *http.Request) error {
	// TODO_TECHDEBT(@adshmh,@commoddity): ParseHTTPRequest (eg for EVM) currently
	// assumes that the request is a JSON-RPC request, which is not the case for WebSocket
	// connection requests, which is an HTTP request with no body and sepcialized headers.
	//
	// We should update QoS packages to either:
	//   - Add a new method for parsing WebSocket connection requests.
	//   - Update ParseHTTPRequest to handle WebSocket connection requests.
	//
	// TODO_TECHDEBT(@adshmh,@commoddity): Use ParseHTTPRequest as the single entry point to QoS for websocket requests.
	qosCtx, isValid := wrc.serviceQoS.ParseHTTPRequest(wrc.context, httpReq)
	wrc.qosCtx = qosCtx

	// Reject invalid WebSocket requests.
	if !isValid {
		// Update gateway observations for websocket rejection
		wrc.updateGatewayObservations(errWebsocketRequestRejectedByQoS)
		wrc.logger.Info().Msg("WebSocket request rejected by QoS")
		return errWebsocketRequestRejectedByQoS
	}

	return nil
}

// BuildProtocolContextFromHTTPRequest builds the Protocol context for the websocket request.
// Similar to requestContext but only creates a single protocol context.
// Returns protocol observations that should be broadcast if the context creation fails.
func (wrc *websocketRequestContext) BuildProtocolContextFromHTTPRequest(
	httpReq *http.Request,
) (*protocolobservations.Observations, error) {
	logger := wrc.logger.With("method", "BuildProtocolContextFromHTTPRequest").With("service_id", wrc.serviceID)

	// Retrieve the list of available endpoints for the requested service.
	// endpointLookupObs will capture the details of the endpoint lookup, including whether it is an error or success.
	availableEndpoints, endpointLookupObs, err := wrc.protocol.AvailableEndpoints(wrc.context, wrc.serviceID, httpReq)
	if err != nil {
		logger.Error().Err(err).Msg("no available endpoints could be found for websocket request")
		return &endpointLookupObs, fmt.Errorf("no available endpoints for websocket request: %w", err)
	}

	// For websockets, select a single endpoint
	selectedEndpoint, err := wrc.qosCtx.GetEndpointSelector().Select(availableEndpoints)
	if err != nil {
		logger.Error().Msgf("no endpoints could be selected for websocket request from %d available endpoints", len(availableEndpoints))
		return &endpointLookupObs, fmt.Errorf("no endpoints could be selected for websocket request from %d available endpoints", len(availableEndpoints))
	}

	// Build protocol context for the selected endpoint
	protocolCtx, protocolCtxSetupErrObs, err := wrc.protocol.BuildRequestContextForEndpoint(wrc.context, wrc.serviceID, selectedEndpoint, httpReq)
	if err != nil {
		logger.Error().Err(err).Str("endpoint_addr", string(selectedEndpoint)).Msg("Failed to build protocol context for websocket endpoint")
		return &protocolCtxSetupErrObs, fmt.Errorf("failed to build protocol context for websocket endpoint: %w", err)
	}

	wrc.protocolCtx = protocolCtx
	logger.Info().Msgf("Successfully built protocol context for websocket endpoint: %s", selectedEndpoint)

	// Return endpoint lookup observations for success case
	return &endpointLookupObs, nil
}

// HandleWebsocketRequest establishes the websocket connection and starts the bridge,
// which handles the message processing loop and sends message observations to the gateway.
func (wrc *websocketRequestContext) HandleWebsocketRequest(
	httpRequest *http.Request,
	httpResponseWriter http.ResponseWriter,
) error {
	// Create the websocket bridge and start it asynchronously to avoid blocking the main thread.
	if err := wrc.stateWebSocketBridge(httpRequest, httpResponseWriter); err != nil {
		wrc.logger.Error().Err(err).Msg("❌ Failed to create websocket bridge.")
		return err
	}

	return nil
}

// stateWebSocketBridge creates a websocket bridge and starts it in a goroutine.
// It also starts a goroutine to listen for message processing notifications from the bridge.
func (wrc *websocketRequestContext) stateWebSocketBridge(
	httpRequest *http.Request,
	httpResponseWriter http.ResponseWriter,
) error {
	// Get the websocket-specific URL from the selected endpoint.
	websocketEndpointURL, err := wrc.protocolCtx.GetWebsocketEndpointURL()
	if err != nil {
		// Wrap the endpoint URL error with our specific error type
		endpointErr := fmt.Errorf("%w: selected endpoint does not support websocket RPC type: %s", errWebsocketConnectionFailed, err.Error())
		wrc.updateGatewayObservations(endpointErr)
		wrc.logger.Error().Err(err).Msg("❌ Selected endpoint does not support websocket RPC type")
		return endpointErr
	}
	wrc.logger = wrc.logger.With("websocket_url", websocketEndpointURL)

	// Get the headers for the websocket connection that will be sent to the endpoint.
	endpointConnectionHeaders, err := wrc.protocolCtx.GetWebsocketConnectionHeaders()
	if err != nil {
		// Wrap the connection headers error with our specific error type
		headersErr := fmt.Errorf("%w: failed to get websocket connection headers: %s", errWebsocketConnectionFailed, err.Error())
		wrc.updateGatewayObservations(headersErr)
		wrc.logger.Error().Err(err).Msg("❌ Failed to get websocket connection headers")
		return headersErr
	}

	// Start the websockets bridge in a goroutine to avoid blocking the main thread.
	// The bridge uses the websocket request context as the message processor to
	// perform both protocol-level and QoS-level message processing.
	// Pass the shared WebSocket context so both bridge and gateway use the same lifecycle.
	if err := websockets.StartBridge(
		wrc.context, // Pass the shared WebSocket context
		wrc.logger,
		httpRequest,
		httpResponseWriter,
		websocketEndpointURL,
		endpointConnectionHeaders,
		wrc,
		wrc.messageObservationsChan,
	); err != nil {
		// Wrap the WebSocket bridge startup error with our specific error type
		bridgeErr := fmt.Errorf("%w: %s", errWebsocketConnectionFailed, err.Error())
		wrc.updateGatewayObservations(bridgeErr)
		wrc.logger.Error().Err(err).Msg("Failed to start WebSocket bridge")
		return bridgeErr
	}

	// Record the connection establishment time for duration tracking
	now := time.Now()
	wrc.connectionStartTime = &now

	// Start listening for message processing notifications from the bridge.
	go wrc.listenForMessageNotifications()

	return nil
}

// listenForMessageNotifications listens for message processing notifications from
// the bridge and publishes observations for each message processed.
//
// This method runs in a goroutine and handles:
//   - Success notifications: Broadcast successful message observations
//   - Error notifications: Broadcast error observations with details
//   - Context cancellation: Clean shutdown when connection is closed
func (wrc *websocketRequestContext) listenForMessageNotifications() {
	for {
		select {
		case messageObservations := <-wrc.messageObservationsChan:
			// Message was processed successfully
			wrc.BroadcastMessageObservations(messageObservations)
		case <-wrc.context.Done():
			// Context cancelled, stop listening
			wrc.logger.Debug().Msg("Message notification listener stopped due to context cancellation")
			return
		}
	}
}

// ---------- Websocket Message Processing ----------

// ProcessClientWebsocketMessage processes a message from the client.
// It performs both Protocol-level and QoS-level message processing.
// If an error occurs, it is returned and the message is not forwarded to the endpoint.
func (wrc *websocketRequestContext) ProcessClientWebsocketMessage(msgData []byte) ([]byte, error) {
	logger := wrc.logger.With("method", "ProcessClientWebsocketMessage")

	logger.Debug().Msgf("received message from client: %s", string(msgData))

	// Process the client message using the protocol context.
	clientMessageBz, err := wrc.protocolCtx.ProcessProtocolClientWebsocketMessage(msgData)
	if err != nil {
		logger.Error().Err(err).Msg("❌ failed to perform protocol-level client message processing")
		return nil, err
	}

	return clientMessageBz, nil
}

// ProcessEndpointWebsocketMessage processes a message from the endpoint.
// It performs both Protocol-level and QoS-level message processing.
// If an error occurs, it is returned and the message is not forwarded to the client.
func (wrc *websocketRequestContext) ProcessEndpointWebsocketMessage(msgData []byte) ([]byte, *observation.RequestResponseObservations, error) {
	logger := wrc.logger.With("method", "ProcessEndpointWebsocketMessage")

	messageObservations := wrc.initializeMessageObservations()

	// Process the endpoint message using the protocol context and update the message observations.
	endpointMessageBz, protocolObservations, err := wrc.protocolCtx.ProcessProtocolEndpointWebsocketMessage(msgData)
	if err != nil {
		logger.Error().Err(err).Msg("❌ failed to perform protocol-level endpoint message processing")
		return nil, nil, err
	}

	// Add connection duration to WebSocket message observations
	wrc.updateWebsocketObservationsWithConnectionDuration(&protocolObservations)

	// Add protocol observations to the message observations.
	messageObservations.Protocol = &protocolObservations

	// TODO_TECHDEBT(@commoddity): process message using QoS context and update the message observations.
	// For example, for JSON-RPC method send through WebSocket, the QoS context should:
	//   - Check the request payload.
	//   - Detect it is a subscription request.
	//   - Validate the request, e.g. params field.
	//   - Reject invalid WebSocket requests, similar to HTTP requests. (e.g. invalid params field)
	// messageObservations.Qos = wrc.qosCtx.ProcessProtocolEndpointWebsocketMessage(msgData)

	return endpointMessageBz, messageObservations, nil
}

// ---------- WebSocket Message Observations ----------

// BroadcastMessageObservations delivers the collected details regarding all aspects
// of the websocket message to all the interested parties.
func (wrc *websocketRequestContext) BroadcastMessageObservations(messageObservations *observation.RequestResponseObservations) {
	// Safety check: don't process nil observations
	if messageObservations == nil {
		wrc.logger.Warn().Msg("Received nil messageObservations, skipping broadcast")
		return
	}

	// observation-related tasks are called in Goroutines to avoid potentially blocking the handler.
	go func() {
		if protocolObservations := messageObservations.GetProtocol(); protocolObservations != nil {
			err := wrc.protocol.ApplyObservations(protocolObservations)
			if err != nil {
				wrc.logger.Warn().Err(err).Msg("error applying protocol observations for websocket.")
			}
		}

		// Apply QoS observations
		if qosObservations := messageObservations.GetQos(); qosObservations != nil {
			if err := wrc.serviceQoS.ApplyObservations(qosObservations); err != nil {
				wrc.logger.Warn().Err(err).Msg("error applying QoS observations for websocket.")
			}
		}

		// Prepare and publish observations to both the metrics and data reporters.
		observations := &observation.RequestResponseObservations{
			Gateway:  wrc.gatewayObservations,
			Protocol: messageObservations.Protocol,
			Qos:      messageObservations.Qos,
		}
		if wrc.metricsReporter != nil {
			wrc.metricsReporter.Publish(observations)
		}
		if wrc.dataReporter != nil {
			wrc.dataReporter.Publish(observations)
		}
	}()
}

// initializeMessageObservations creates a copy of observations for each websocket message.
// This ensures per-message observations as required.
func (wrc *websocketRequestContext) initializeMessageObservations() *observation.RequestResponseObservations {
	return &observation.RequestResponseObservations{
		ServiceId: string(wrc.serviceID),
		Gateway:   wrc.gatewayObservations,
	}
}

// ---------- WebSocket Connection Observations ----------

// BroadcastWebsocketConnectionRequestObservations broadcasts a single connection-level observation.
// This method combines protocol observations with gateway observations and publishes them.
// This method should be called from a defer in handleWebSocketRequest.
func (wrc *websocketRequestContext) BroadcastWebsocketConnectionRequestObservations(protocolObs *protocolobservations.Observations) {
	wrc.updateGatewayObservations(nil)

	// Add connection duration to WebSocket connection observations if available
	if protocolObs != nil {
		wrc.updateWebsocketObservationsWithConnectionDuration(protocolObs)
	}

	// Combine all observations into the standard RequestResponseObservations format
	observations := &observation.RequestResponseObservations{
		Gateway:  wrc.gatewayObservations,
		Protocol: protocolObs,
		Qos:      nil, // QoS is not applicable for websocket connection observations as they have not been processed yet.
	}

	// Broadcast the combined observations
	if wrc.metricsReporter != nil {
		wrc.metricsReporter.Publish(observations)
	}
	if wrc.dataReporter != nil {
		wrc.dataReporter.Publish(observations)
	}
}

// getConnectionDurationMs calculates the WebSocket connection duration in milliseconds.
// Returns nil if the connection start time hasn't been recorded yet.
func (wrc *websocketRequestContext) getConnectionDurationMs() *int64 {
	if wrc.connectionStartTime == nil {
		return nil
	}
	durationMs := time.Since(*wrc.connectionStartTime).Milliseconds()
	return &durationMs
}

// updateWebsocketObservationsWithConnectionDuration adds connection duration to WebSocket observations.
// This updates both WebSocket endpoint and message observations with the calculated connection duration.
func (wrc *websocketRequestContext) updateWebsocketObservationsWithConnectionDuration(protocolObs *protocolobservations.Observations) {
	if protocolObs == nil || protocolObs.Shannon == nil {
		return
	}

	connectionDurationMs := wrc.getConnectionDurationMs()
	if connectionDurationMs == nil {
		// Connection duration not available (e.g., connection not established yet)
		return
	}

	// Update all Shannon observations with connection duration
	for _, obs := range protocolObs.Shannon.Observations {
		if obs == nil {
			continue
		}

		// Update based on observation type
		switch obsData := obs.ObservationData.(type) {
		case *protocolobservations.ShannonRequestObservations_WebsocketEndpointObservation:
			if obsData.WebsocketEndpointObservation != nil {
				obsData.WebsocketEndpointObservation.ConnectionDurationMs = connectionDurationMs
			}
		case *protocolobservations.ShannonRequestObservations_WebsocketMessageObservation:
			if obsData.WebsocketMessageObservation != nil {
				obsData.WebsocketMessageObservation.ConnectionDurationMs = connectionDurationMs
			}
		}
	}
}

// updateGatewayObservations updates the gateway-level observations in the websocket request context.
func (wrc *websocketRequestContext) updateGatewayObservations(err error) {
	// set the service ID on the gateway observations.
	wrc.gatewayObservations.ServiceId = string(wrc.serviceID)

	// Update the request completion time on the gateway observation
	wrc.gatewayObservations.CompletedTime = timestamppb.Now()

	// No errors: skip.
	if err == nil {
		return
	}

	// Request error already set: skip.
	if wrc.gatewayObservations.GetRequestError() != nil {
		return
	}

	// Classify WebSocket-specific errors based on error type
	switch {
	// Service ID not specified
	case errors.Is(err, ErrGatewayNoServiceIDProvided):
		wrc.logger.Error().Err(err).Msg("No service ID specified in the HTTP headers. WebSocket request will fail.")
		wrc.gatewayObservations.RequestError = &observation.GatewayRequestError{
			ErrorKind: observation.GatewayRequestErrorKind_GATEWAY_REQUEST_ERROR_KIND_MISSING_SERVICE_ID,
			Details:   err.Error(),
		}

	// WebSocket request was rejected by QoS instance
	case errors.Is(err, errWebsocketRequestRejectedByQoS):
		wrc.logger.Error().Err(err).Msg("QoS instance rejected the WebSocket request. Request will fail.")
		wrc.gatewayObservations.RequestError = &observation.GatewayRequestError{
			ErrorKind: observation.GatewayRequestErrorKind_GATEWAY_REQUEST_ERROR_KIND_WEBSOCKET_REJECTED_BY_QOS,
			Details:   err.Error(),
		}

	// WebSocket connection establishment failed
	case errors.Is(err, errWebsocketConnectionFailed):
		wrc.logger.Error().Err(err).Msg("WebSocket connection establishment failed. Request will fail.")
		wrc.gatewayObservations.RequestError = &observation.GatewayRequestError{
			ErrorKind: observation.GatewayRequestErrorKind_GATEWAY_REQUEST_ERROR_KIND_WEBSOCKET_CONNECTION_FAILED,
			Details:   err.Error(),
		}

	// Generic QoS rejection (fallback for backward compatibility)
	case errors.Is(err, errGatewayRejectedByQoS):
		wrc.logger.Error().Err(err).Msg("QoS instance rejected the request. Request will fail.")
		wrc.gatewayObservations.RequestError = &observation.GatewayRequestError{
			ErrorKind: observation.GatewayRequestErrorKind_GATEWAY_REQUEST_ERROR_KIND_REJECTED_BY_QOS,
			Details:   err.Error(),
		}

	default:
		wrc.logger.Warn().Err(err).Msg("SHOULD NEVER HAPPEN: unrecognized WebSocket gateway-level request error.")
		// Set a generic request error observation
		wrc.gatewayObservations.RequestError = &observation.GatewayRequestError{
			// unspecified error kind: this should not happen
			ErrorKind: observation.GatewayRequestErrorKind_GATEWAY_REQUEST_ERROR_KIND_UNSPECIFIED,
			Details:   err.Error(),
		}
	}
}

// getWSRequestLogger returns a logger with attributes set using the supplied HTTP request.
func (wrc *websocketRequestContext) getWSRequestLogger(httpReq *http.Request) polylog.Logger {
	var urlStr string
	if httpReq.URL != nil {
		urlStr = httpReq.URL.String()
	}

	return wrc.logger.With(
		"ws_req_url", urlStr,
		"ws_req_host", httpReq.Host,
		"ws_req_remote_addr", httpReq.RemoteAddr,
		"ws_req_content_length", httpReq.ContentLength,
		"request_type", "websocket",
	)
}
