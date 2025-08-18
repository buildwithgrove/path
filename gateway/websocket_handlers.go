package gateway

import (
	"encoding/json"
	"fmt"

	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/websockets"
)

// ---------- Shannon Client Message Handler ----------

var _ websockets.WebSocketMessageHandler = &websocketClientMessageHandler{}

// websocketClientMessageHandler handles websocket client messages
type websocketClientMessageHandler struct {
	logger      polylog.Logger
	protocolCtx ProtocolRequestContext
	serviceID   protocol.ServiceID
}

// HandleMessage processes a message from the client.
func (h *websocketClientMessageHandler) HandleMessage(msgData []byte) ([]byte, error) {
	logger := h.logger.With("method", "HandleMessage")

	logger.Debug().Msgf("received message from client: %s", string(msgData))

	clientMessageBz, err := h.protocolCtx.SignClientWebsocketMessage(msgData)
	if err != nil {
		logger.Error().Err(err).Msg("❌ failed to sign request")
		return h.getClientErrorResponse(msgData, err)
	}

	return clientMessageBz, nil
}

// TODO_IN_THIS_PR(@commoddit): Handle errors using QoS package - ie move JSON-RPC logic to the QoS package.

// ---------- Shannon Error Handling ----------

// getClientErrorResponse sends a JSON-RPC error response to the client
// Attempts to extract the request ID from the original message for proper error formatting.
func (h *websocketClientMessageHandler) getClientErrorResponse(
	originalMessage []byte,
	err error,
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

	errorMsg := fmt.Sprintf("Failed to sign request: %s", err.Error())

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

var _ websockets.WebSocketMessageHandler = &websocketEndpointMessageHandler{}

// websocketEndpointMessageHandler handles endpoint messages with Shannon-specific logic.
type websocketEndpointMessageHandler struct {
	logger      polylog.Logger
	protocolCtx ProtocolRequestContext
	serviceID   protocol.ServiceID
}

// HandleMessage processes a message from the endpoint.
func (h *websocketEndpointMessageHandler) HandleMessage(msgData []byte) ([]byte, error) {
	logger := h.logger.With("method", "HandleMessage")

	// Validate the relay response using the Shannon FullNode
	validatedEndpointMessage, err := h.protocolCtx.ValidateEndpointWebsocketMessage(msgData)
	if err != nil {
		logger.Error().Err(err).Msg("❌ failed to validate relay response")
		return nil, err
	}

	logger.Debug().Msgf("received message from protocol endpoint: %s", string(validatedEndpointMessage))

	return validatedEndpointMessage, nil
}
