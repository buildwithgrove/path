package shannon

import (
	"encoding/json"
	"fmt"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/observation"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/buildwithgrove/path/websockets"
	"github.com/pokt-network/poktroll/pkg/polylog"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sdk "github.com/pokt-network/shannon-sdk"
)

// ---------- Shannon Client Message Handler ----------

var _ websockets.MessageHandler = &shannonClientMessageHandler{}

// shannonClientMessageHandler handles client messages with Shannon-specific logic.
type shannonClientMessageHandler struct {
	logger             polylog.Logger
	selectedEndpoint   endpoint
	relayRequestSigner RelayRequestSigner
	serviceID          protocol.ServiceID
}

// HandleMessage processes a message from the client.
func (h *shannonClientMessageHandler) HandleMessage(msg websockets.Message) ([]byte, error) {
	h.logger.Debug().Msgf("received message from client: %s", string(msg.Data))

	// If the selected endpoint is a fallback endpoint, skip protocol-level signing of the request.
	if h.selectedEndpoint.IsFallback() {
		return msg.Data, nil
	}

	// Sign the client message before sending it to the Endpoint
	clientMessageBz, err := h.signClientMessage(msg)
	if err != nil {
		return h.getClientErrorResponse(msg, -32000, fmt.Sprintf("Failed to sign request: %s", err.Error()))
	}

	return clientMessageBz, nil
}

// signClientMessage signs the client message in order to send it to the Endpoint
func (h *shannonClientMessageHandler) signClientMessage(msg websockets.Message) ([]byte, error) {
	unsignedRelayRequest := &servicetypes.RelayRequest{
		Meta: servicetypes.RelayRequestMetadata{
			SessionHeader:           h.selectedEndpoint.Session().GetHeader(),
			SupplierOperatorAddress: h.selectedEndpoint.Supplier(),
		},
		Payload: msg.Data,
	}

	app := h.selectedEndpoint.Session().GetApplication()
	// SHOULD NEVER HAPPEN: `signClientMessage` is called only for protocol endpoints.
	// But we guard against it anyway.
	if app == nil {
		return nil, fmt.Errorf("session application is nil")
	}

	signedRelayRequest, err := h.relayRequestSigner.SignRelayRequest(unsignedRelayRequest, *app)
	if err != nil {
		return nil, fmt.Errorf("error signing client message: %s", err.Error())
	}

	relayRequestBz, err := signedRelayRequest.Marshal()
	if err != nil {
		return nil, fmt.Errorf("error marshalling signed client message: %s", err.Error())
	}

	return relayRequestBz, nil
}

// ---------- Shannon Error Handling ----------

// getClientErrorResponse creates a JSON-RPC error response to the client.
// Attempts to extract the request ID from the original message for proper error formatting.
func (h *shannonClientMessageHandler) getClientErrorResponse(originalMessage websockets.Message, errorCode int, errorMessage string) ([]byte, error) {
	// Try to extract the request ID from the original message
	var requestID jsonrpc.ID

	// Attempt to parse the original message to extract the ID
	var jsonrpcRequest struct {
		ID jsonrpc.ID `json:"id"`
	}

	if err := json.Unmarshal(originalMessage.Data, &jsonrpcRequest); err != nil {
		// If we can't parse the original message, use null ID (zero value of ID struct)
		h.logger.Warn().Err(err).Msg("Failed to parse original message for ID, using null ID")
		requestID = jsonrpc.ID{} // Zero value creates a null ID
	} else {
		requestID = jsonrpcRequest.ID
	}

	// Create the JSON-RPC error response
	errorResponse := jsonrpc.GetErrorResponse(requestID, errorCode, errorMessage, nil)

	// Marshal the error response
	errorResponseBytes, err := json.Marshal(errorResponse)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to marshal error response")
		return nil, fmt.Errorf("failed to marshal error response: %w", err)
	}

	return errorResponseBytes, nil
}

// ---------- Shannon Endpoint Message Handler ----------

var _ websockets.MessageHandler = &shannonEndpointMessageHandler{}

// shannonEndpointMessageHandler handles endpoint messages with Shannon-specific logic.
type shannonEndpointMessageHandler struct {
	logger           polylog.Logger
	selectedEndpoint endpoint
	fullNode         FullNode
	serviceID        protocol.ServiceID
}

// HandleMessage processes a message from the endpoint.
func (h *shannonEndpointMessageHandler) HandleMessage(msg websockets.Message) ([]byte, error) {
	h.logger.Debug().Msgf("received message from endpoint: %s", string(msg.Data))

	// If the selected endpoint is a fallback endpoint, skip protocol-level validation of the relay response.
	if h.selectedEndpoint.IsFallback() {
		return msg.Data, nil
	}

	// Validate the relay response using the Shannon FullNode
	relayResponse, err := h.fullNode.ValidateRelayResponse(sdk.SupplierAddress(h.selectedEndpoint.Supplier()), msg.Data)
	if err != nil {
		return nil, fmt.Errorf("error validating relay response: %w", err)
	}

	return relayResponse.Payload, nil
}

// ---------- Shannon Observation Publisher ----------

var _ websockets.ObservationPublisher = &shannonObservationPublisher{}

// shannonObservationPublisher handles publishing Shannon-specific observations.
type shannonObservationPublisher struct {
	serviceID            protocol.ServiceID
	protocolObservations *protocolobservations.Observations
	gatewayObservations  *observation.GatewayObservations
	dataReporter         gateway.RequestResponseReporter
}

// SetObservationContext sets the gateway observations and data reporter.
// This is called by the gateway when starting the bridge.
func (p *shannonObservationPublisher) SetObservationContext(
	gatewayObservations *observation.GatewayObservations,
	dataReporter gateway.RequestResponseReporter,
) {
	p.gatewayObservations = gatewayObservations
	p.dataReporter = dataReporter
}

// PublishObservations publishes the Shannon-specific observations for this websocket message.
func (p *shannonObservationPublisher) PublishObservations() {
	if p.dataReporter == nil || p.gatewayObservations == nil {
		return
	}

	// Create the request-response observations for this websocket message
	observations := &observation.RequestResponseObservations{
		ServiceId: string(p.serviceID),
		Gateway:   p.gatewayObservations,
		Protocol:  p.protocolObservations,
	}

	p.dataReporter.Publish(observations)
}
