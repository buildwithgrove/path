package shannon

import (
	"encoding/json"
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sdk "github.com/pokt-network/shannon-sdk"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/observation"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/buildwithgrove/path/websockets"
)

// TODO_TECHDEBT(@adshmh): Move any logic unrelated to protocol out of this file:
// - Data publishing: gateway package.
// - Observation publishing: gateway package.
// - getClientErrorResponse: qos/* packages: protocol should not contain JSONRPC logic.
// - Separate and make fallback-related logic more explicit.
//
// ---------- Shannon Client Message Handler ----------

var _ websockets.WebSocketMessageHandler = &websocketClientMessageHandler{}

// websocketClientMessageHandler handles websocket client messages
type websocketClientMessageHandler struct {
	logger             polylog.Logger
	selectedEndpoint   endpoint
	relayRequestSigner RelayRequestSigner
	serviceID          protocol.ServiceID
}

// HandleMessage processes a message from the client.
func (h *websocketClientMessageHandler) HandleMessage(msg websockets.WebSocketMessage) ([]byte, error) {
	logger := h.logger.With("method", "HandleMessage")

	logger.Debug().Msgf("received message from client: %s", string(msg.Data))

	// If the selected endpoint is a fallback endpoint, skip protocol-level signing of the request.
	if h.selectedEndpoint.IsFallback() {
		return msg.Data, nil
	}

	// Sign the client message before sending it to the Endpoint on the network.
	clientMessageBz, err := h.signClientMessage(msg)
	if err != nil {
		logger.Error().Err(err).Msg("❌ failed to sign request")
		// If we fail to sign the request, we return a JSON-RPC error
		// response to the client instead of terminating the connection.
		return h.getClientErrorResponse(msg.Data, err)
	}

	return clientMessageBz, nil
}

// signClientMessage signs the client message in order to send it to the Endpoint
func (h *websocketClientMessageHandler) signClientMessage(msg websockets.WebSocketMessage) ([]byte, error) {
	logger := h.logger.With("method", "signClientMessage")

	unsignedRelayRequest := &servicetypes.RelayRequest{
		Meta: servicetypes.RelayRequestMetadata{
			SessionHeader:           h.selectedEndpoint.Session().GetHeader(),
			SupplierOperatorAddress: h.selectedEndpoint.Supplier(),
		},
		Payload: msg.Data,
	}

	app := h.selectedEndpoint.Session().GetApplication()
	if app == nil {
		logger.Error().Msg("❌ SHOULD NEVER HAPPEN: session application is nil")
		return nil, fmt.Errorf("session application is nil")
	}

	// Sign the unsigned relay request.
	signedRelayRequest, err := h.relayRequestSigner.SignRelayRequest(unsignedRelayRequest, *app)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errRelayRequestSigningFailed, err.Error())
	}

	relayRequestBz, err := signedRelayRequest.Marshal()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errRelayRequestSigningFailed, err.Error())
	}

	return relayRequestBz, nil
}

// ---------- Shannon Error Handling ----------

// getClientErrorResponse sends a JSON-RPC error response to the client
// Attempts to extract the request ID from the original message for proper error formatting.
func (h *websocketClientMessageHandler) getClientErrorResponse(
	originalMessage []byte,
	error error,
) ([]byte, error) {
	logger := h.logger.With("method", "getClientErrorResponse")

	// Try to extract the request ID from the original message
	var requestID jsonrpc.ID

	// Attempt to parse the original message to extract the ID
	var jsonrpcRequest jsonrpc.Request
	if err := json.Unmarshal(originalMessage, &jsonrpcRequest); err != nil {
		logger.Debug().Err(err).Msg("failed to parse original message for ID, using null ID")
		// Zero value creates a null ID. This adheres to the JSON-RPC 2.0 spec
		// when the original ID cannot be obtained from the original message.
		// https://www.jsonrpc.org/specification#response_object
		requestID = jsonrpc.ID{}
	} else {
		requestID = jsonrpcRequest.ID
	}

	errorMsg := fmt.Sprintf("Failed to sign request: %s", error.Error())

	// Create the JSON-RPC error response
	errorResponse := jsonrpc.GetErrorResponse(requestID, -32000, errorMsg, nil)

	errorResponseBytes, err := json.Marshal(errorResponse)
	if err != nil {
		logger.Error().Err(err).Msg("❌ SHOULD NEVER HAPPEN: failed to marshal error response")
		return nil, fmt.Errorf("failed to marshal error response: %w", err)
	}

	return errorResponseBytes, nil
}

// ---------- Shannon Endpoint Message Handler ----------

var _ websockets.WebSocketMessageHandler = &endpointMessageHandler{}

// endpointMessageHandler handles endpoint messages with Shannon-specific logic.
type endpointMessageHandler struct {
	logger           polylog.Logger
	selectedEndpoint endpoint
	fullNode         FullNode
	serviceID        protocol.ServiceID
}

// HandleMessage processes a message from the endpoint.
func (h *endpointMessageHandler) HandleMessage(msg websockets.WebSocketMessage) ([]byte, error) {
	logger := h.logger.With("method", "HandleMessage")

	// If the selected endpoint is a fallback endpoint, skip protocol-level validation of the relay response.
	if h.selectedEndpoint.IsFallback() {
		logger.Debug().Msgf("received message from fallback endpoint: %s", string(msg.Data))
		return msg.Data, nil
	}

	// Validate the relay response using the Shannon FullNode
	relayResponse, err := h.fullNode.ValidateRelayResponse(sdk.SupplierAddress(h.selectedEndpoint.Supplier()), msg.Data)
	if err != nil {
		logger.Error().Err(err).Msg("❌ failed to validate relay response")
		return nil, fmt.Errorf("%w: %s", errRelayResponseValidationFailed, err.Error())
	}

	logger.Debug().Msgf("received message from protocol endpoint: %s", string(relayResponse.Payload))

	return relayResponse.Payload, nil
}

// ---------- Shannon Observation Publisher ----------

var _ websockets.ObservationPublisher = &observationPublisher{}

// observationPublisher handles publishing Shannon-specific observations.
type observationPublisher struct {
	logger polylog.Logger

	serviceID protocol.ServiceID

	// protocolObservations in the websocket bridge contain values that will be static for the duration
	// of the specific Bridge's operation.
	//
	// For example, the selected endpoint.
	//
	// As a websocket bridge represents a single persistent connection to a single endpoint,
	// the protocolObservations will be the same for the duration of the bridge's operation.
	protocolObservations *protocolobservations.Observations

	// gatewayObservations in the websocket bridge contain values that will be static for the duration
	// of the specific Bridge's operation.
	//
	// For example, the RequestAuth used to authorize the request.
	//
	// As a websocket bridge represents a single persistent connection to a single endpoint,
	// the gatewayObservations will be the same for the duration of the bridge's operation.
	gatewayObservations *observation.GatewayObservations

	// dataReporter is used to export, to the data pipeline, observations made in handling websocket messages.
	dataReporter gateway.RequestResponseReporter
}

// SetObservationContext sets the gateway observations and data reporter.
// This is called by the gateway when starting the bridge.
func (p *observationPublisher) SetObservationContext(
	gatewayObservations *observation.GatewayObservations,
	dataReporter gateway.RequestResponseReporter,
) {
	p.gatewayObservations = gatewayObservations
	p.dataReporter = dataReporter
}

// InitializeMessageObservations initializes the message observations for the current message
// with the static observation fields created by the Bridge.
// Message observations are updated in case of error.
func (p *observationPublisher) InitializeMessageObservations() *observation.RequestResponseObservations {
	return &observation.RequestResponseObservations{
		ServiceId: string(p.serviceID),
		Gateway:   p.gatewayObservations,
		Protocol:  p.protocolObservations,
	}
}

// UpdateMessageObservationsFromSuccess updates the observations for the current message
// if the message handler does not return an error.
func (p *observationPublisher) UpdateMessageObservationsFromSuccess(
	observations *observation.RequestResponseObservations,
) {
	// Get the websocket endpoint observation to update
	endpointObs, err := p.getWebsocketEndpointObservation(observations)
	if err != nil {
		p.logger.Error().Err(err).Msg("❌ SHOULD NEVER HAPPEN: failed to get websocket endpoint observation")
		return
	}

	buildWebsocketMessageSuccessObservation(endpointObs)
}

// UpdateMessageObservationsFromError updates the observations for the current message
// if the message handler returns an error.
func (p *observationPublisher) UpdateMessageObservationsFromError(
	observations *observation.RequestResponseObservations,
	messageError error,
) {
	// Get the websocket endpoint observation to update
	endpointObs, err := p.getWebsocketEndpointObservation(observations)
	if err != nil {
		p.logger.Error().Err(err).Msg("❌ SHOULD NEVER HAPPEN: failed to get websocket endpoint observation")
		return
	}

	buildWebsocketMessageErrorObservation(p.logger, endpointObs, messageError)
}

// getWebsocketEndpointObservation safely retrieves the websocket
// endpoint observation from the request-response observations.
//
// This method is primarily a sanity check as Bridge obervations should
// always have only one request observation with one endpoint observation.
func (p *observationPublisher) getWebsocketEndpointObservation(
	observations *observation.RequestResponseObservations,
) (*protocolobservations.ShannonEndpointObservation, error) {
	// Validate observation structure
	if observations == nil ||
		observations.Protocol == nil ||
		observations.Protocol.Shannon == nil {
		return nil, fmt.Errorf("observations are nil")
	}

	shannonObs := observations.Protocol.Shannon

	// For websocket connections, we expect exactly one request observation
	if len(shannonObs.Observations) != 1 {
		return nil, fmt.Errorf("observations have more than one request observation")
	}

	requestObs := shannonObs.Observations[0]

	// Each websocket connection should have exactly one endpoint observation
	if len(requestObs.EndpointObservations) != 1 {
		return nil, fmt.Errorf("request observation has more than one endpoint observation")
	}

	return requestObs.EndpointObservations[0], nil
}

// PublishObservations publishes the Shannon-specific observations for this websocket message.
func (p *observationPublisher) PublishMessageObservations(
	observations *observation.RequestResponseObservations,
) {
	// If the data reporter is not configured using the `data_reporter_config` field
	// in the config YAML, then the data reporter will be nil so we need to check for that.
	// For example, when running the Gateway in a local environment, the data reporter may be nil.
	if p.dataReporter == nil {
		return
	}

	p.dataReporter.Publish(observations)
}
