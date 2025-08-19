package gateway

import (
	"context"
	"fmt"
	"net/http"

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
// - QoS validation and context building
// - Protocol context setup (single endpoint selection vs HTTP's multiple endpoints)
// - Message routing and observation (per-message vs HTTP's per-request)
// - Bridge lifecycle management
//
// Key differences from HTTP requestContext:
// - Single endpoint selection (websockets can't do parallel requests)
// - Per-message observations instead of per-request observations
// - Long-lived connection management vs one-shot request/response
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

	// // Bridge and connection management
	// bridge WebsocketsBridge

	// Channel for receiving message processing notifications from the bridge
	messageObservationsChan chan *observation.RequestResponseObservations
}

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
// For websockets, this replaces the previous BuildQoSContextFromWebsocket method.
// Following the TODO comment: ParseHTTPRequest should be the single entry point to QoS.
func (wrc *websocketRequestContext) BuildQoSContextFromHTTP(httpReq *http.Request) error {
	// Use ParseHTTPRequest as the single entry point to QoS for websocket requests
	// The QoS implementation should detect if this is a websocket subscription request
	// and validate it accordingly

	// TODO_TECHDEBT(@adshmh): Use ParseHTTPRequest as the single entry point to QoS, including for a WebSocket request.
	// - The ParseHTTPRequest method in QoS should:
	//   - Check the request payload
	//   - Detect it is a subscription request.
	//   - Validate the request, e.g. params field.
	//   - Reject invalid WebSocket requests, similar to HTTP requests.
	qosCtx, isValid := wrc.serviceQoS.ParseHTTPRequest(wrc.context, httpReq)
	wrc.qosCtx = qosCtx

	if !isValid {
		// Update gateway observations for websocket rejection
		wrc.updateGatewayObservations(fmt.Errorf("websocket request rejected by QoS"))
		wrc.logger.Info().Msg("Websocket request rejected by QoS")
		return fmt.Errorf("websocket request rejected by QoS")
	}

	return nil
}

// BuildProtocolContextFromHTTPRequest builds the Protocol context for the websocket request.
// Similar to requestContext but only creates a single protocol context.
func (wrc *websocketRequestContext) BuildProtocolContextFromHTTPRequest(httpReq *http.Request) error {
	logger := wrc.logger.With("method", "BuildProtocolContextFromHTTPRequest").With("service_id", wrc.serviceID)

	// Retrieve the list of available endpoints for the requested service.
	availableEndpoints, endpointLookupObs, err := wrc.protocol.AvailableEndpoints(wrc.context, wrc.serviceID, httpReq)
	if err != nil {
		wrc.broadcastWebsocketConnectionErrorObservation(err, &endpointLookupObs)
		logger.Error().Err(err).Msg("no available endpoints could be found for websocket request")
		return fmt.Errorf("no available endpoints for websocket request: %w", err)
	}

	// For websockets, select a single endpoint
	selectedEndpoints, err := wrc.qosCtx.GetEndpointSelector().SelectMultiple(availableEndpoints, 1)
	if err != nil || len(selectedEndpoints) == 0 {
		wrc.broadcastWebsocketConnectionErrorObservation(err, &endpointLookupObs)
		logger.Error().Msgf("no endpoints could be selected for websocket request from %d available endpoints", len(availableEndpoints))
		return fmt.Errorf("no endpoints could be selected for websocket request from %d available endpoints", len(availableEndpoints))
	}

	// Build protocol context for the selected endpoint
	protocolCtx, protocolCtxSetupErrObs, err := wrc.protocol.BuildRequestContextForEndpoint(wrc.context, wrc.serviceID, selectedEndpoints[0], httpReq)
	if err != nil {
		wrc.broadcastWebsocketConnectionErrorObservation(err, &protocolCtxSetupErrObs)
		logger.Error().Err(err).Str("endpoint_addr", string(selectedEndpoints[0])).Msg("Failed to build protocol context for websocket endpoint")
		return fmt.Errorf("failed to build protocol context for websocket endpoint: %w", err)
	}

	wrc.protocolCtx = protocolCtx
	logger.Info().Msgf("Successfully built protocol context for websocket endpoint: %s", selectedEndpoints[0])

	return nil
}

// HandleWebsocketRequest establishes the websocket connection and starts the message handling loop.
func (wrc *websocketRequestContext) HandleWebsocketRequest(
	httpRequest *http.Request,
	httpResponseWriter http.ResponseWriter,
) error {
	// Create the websocket bridge in the gateway package and start the bridge asynchronously.
	if err := wrc.initializeWebsocketBridge(httpRequest, httpResponseWriter); err != nil {
		wrc.logger.Error().Err(err).Msg("❌ Failed to create websocket bridge.")
		return err
	}

	return nil
}

// createWebsocketBridge creates a websocket bridge using protocol-specific components.
// This moves the bridge creation logic from Shannon to the gateway level.
func (wrc *websocketRequestContext) initializeWebsocketBridge(
	httpRequest *http.Request,
	httpResponseWriter http.ResponseWriter,
) error {
	// Get the websocket-specific URL from the selected endpoint.
	websocketEndpointURL, err := wrc.protocolCtx.GetWebsocketEndpointURL()
	if err != nil {
		wrc.logger.Error().Err(err).Msg("❌ Selected endpoint does not support websocket RPC type")
		return err
	}
	wrc.logger = wrc.logger.With("websocket_url", websocketEndpointURL)

	// Get the headers for the websocket connection that will be sent to the endpoint.
	endpointConnectionHeaders, err := wrc.protocolCtx.GetWebsocketConnectionHeaders()
	if err != nil {
		wrc.logger.Error().Err(err).Msg("❌ Failed to get websocket connection headers")
		return err
	}

	// Start the websockets bridge in a goroutine to avoid blocking the main thread.
	// The bridge uses the websocket request context as the message processor to
	// perform both protocol-level and QoS-level message processing.
	if err := websockets.StartBridge(
		wrc.logger,
		httpRequest,
		httpResponseWriter,
		websocketEndpointURL,
		endpointConnectionHeaders,
		wrc,
		wrc.messageObservationsChan,
	); err != nil {
		// TODO_IN_THIS_PR(@commoddity): Handle updating protocol observations on websocket connection error.
		return err
	}

	// Start listening for message processing notifications from the bridge.
	go wrc.listenForMessageNotifications()

	return nil
}

// listenForMessageNotifications listens for message processing notifications from the bridge
// and publishes observations for each message processed.
//
// This method runs in a goroutine and handles:
// - Success notifications: Broadcast successful message observations
// - Error notifications: Broadcast error observations with details
// - Context cancellation: Clean shutdown when connection is closed
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

// ProcessEndpointMessage processes a message from the endpoint.
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
	messageObservations.Protocol = protocolObservations

	// TODO_IN_THIS_PR(@commoddity): process message using QoS context and update the message observations.
	// messageObservations.Qos = wrc.qosCtx.ProcessProtocolEndpointWebsocketMessage(msgData)

	return endpointMessageBz, messageObservations, nil
}

// ---------- Message Observations ----------

// TODO_IN_THIS_PR(@commoddity): handle correctly updating message observations from protocol and QoS.

// BroadcastAllObservations delivers the collected details regarding all aspects
// of the websocket request to all the interested parties.
func (wrc *websocketRequestContext) BroadcastMessageObservations(messageObservations *observation.RequestResponseObservations) {
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

// broadcastWebsocketConnectionErrorObservation broadcasts the protocol-level observations for websocket connection errors.
// This method is called only if a protocol-level error occurs during the websocket connection setup.
func (wrc *websocketRequestContext) broadcastWebsocketConnectionErrorObservation(
	err error,
	protocolErrorObservations *protocolobservations.Observations,
) {
	wrc.updateGatewayObservations(err)

	if protocolErrorObservations != nil {
		err := wrc.protocol.ApplyObservations(protocolErrorObservations)
		if err != nil {
			wrc.logger.Warn().Err(err).Msg("error applying protocol observations for websocket.")
		}
	}

	observations := &observation.RequestResponseObservations{
		Gateway:  wrc.gatewayObservations,
		Protocol: protocolErrorObservations,
	}
	if wrc.metricsReporter != nil {
		wrc.metricsReporter.Publish(observations)
	}
	if wrc.dataReporter != nil {
		wrc.dataReporter.Publish(observations)
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

	// Set websocket-specific error observations
	wrc.logger.Error().Err(err).Msg("Websocket request error occurred")
	wrc.gatewayObservations.RequestError = &observation.GatewayRequestError{
		ErrorKind: observation.GatewayRequestErrorKind_GATEWAY_REQUEST_ERROR_KIND_REJECTED_BY_QOS,
		Details:   err.Error(),
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
