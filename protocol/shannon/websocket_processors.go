package shannon

import (
	"encoding/json"
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_IN_THIS_PR(@commoddity): Handle errors using QoS package - ie move JSON-RPC logic to the QoS package/
// This should be done as part of creating the QoS-level message processor.

// ---------- Shannon Error Handling ----------

// getClientErrorResponse sends a JSON-RPC error response to the client
// Attempts to extract the request ID from the original message for proper error formatting.
func getClientErrorResponse(
	logger polylog.Logger,
	originalMessage []byte,
	err error,
) ([]byte, error) {
	logger = logger.With("method", "getClientErrorResponse")

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
