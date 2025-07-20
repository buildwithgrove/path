package cosmos

import (
	"encoding/json"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const (
	// error codes to use as the JSONRPC response's error code if the endpoint returns a malformed response.
	// -32000 Error code will result in returning a 500 HTTP Status Code to the client.
	errCodeUnmarshaling  = -32000
	errCodeEmptyResponse = -32000

	// Error messages for JSONRPC response validation failures
	errMsgJSONRPCUnmarshaling  = "the JSONRPC response returned by the endpoint is not valid"
	errMsgJSONRPCEmptyResponse = "the endpoint returned an empty JSON-RPC response"

	// errDataFieldRawBytes is the key of the entry in the JSON-RPC error response's "data" map which holds the endpoint's original response.
	errDataFieldRawBytes = "endpoint_response"

	// errDataFieldUnmarshalingErr is the key of the entry in the JSON-RPC error response's "data" map which holds the unmarshaling error.
	errDataFieldUnmarshalingErr = "unmarshaling_error"
)

// unmarshalAsJSONRPCResponse converts raw endpoint bytes into a JSONRPC response struct.
// The second return value contains the validation failure, if any.
func unmarshalAsJSONRPCResponse(
	logger polylog.Logger,
	jsonrpcRequestID jsonrpc.ID,
	data []byte,
) (jsonrpc.Response, *qosobservations.CosmosResponseValidationError) {
	// Empty payload is invalid.
	if len(data) == 0 {
		errEmptyPayload := errors.New("Failed to unmarshal endpoint payload as JSONRPC: endpoint returned an empty response.")
		logger.With(
			"unmarshal_err", errEmptyPayload,
			"error_type", "empty_response",
		).Debug().Msg(errEmptyPayload.Error())

		// Create a geneirc JSONRPC response for the user.
		validationErr := CosmosResponseValidationError_COSMOS_RESPONSE_VALIDATION_ERROR_EMPTY
		return getGenericJSONRPCErrResponse(logger, jsonrpcRequestID, "", errors.New(errMsg)), &validationErr
	}

	// Unmarshal the raw response payload into a JSONRPC response.
	var jsonrpcResponse jsonrpc.Response
	err := json.Unmarshal(data, &jsonrpcResponse)
	if err != nil {
		// The response raw payload (e.g. as received from an endpoint) could not be unmarshalled as a JSONRC response.
		// Return a generic response to the user.
		payloadStr := string(data)
		logger.With(
			"unmarshal_err", err,
			"raw_payload", log.Preview(payloadStr),
			"error_type", "parse_error",
		).Debug().Msg("Failed to unmarshal endpoint payload as JSONRPC.")

		// Create a generic JSONRPC response for the user.
		validationErr := CosmosResponseValidationError_COSMOS_RESPONSE_VALIDATION_ERROR_UNMARSHAL
		return getGenericJSONRPCErrResponse(logger, jsonrpcReq.ID, data, err), &validationErr
	}

	// Validate the JSONRPC response.
	if err := jsonrpcResponse.Validate(jsonrpcRequestID); err != nil {
		payloadStr := string(data)
		logger.With(
			"unmarshal_err", err,
			"raw_payload", log.Preview(payloadStr),
			"error_type", "validation_error",
		).Debug().Msg("Failed to unmarshal endpoint payload as JSONRPC: JSONRPC response failed validation.")

		validationErr := CosmosResponseValidationError_COSMOS_RESPONSE_VALIDATION_ERROR_JSONRPC
		return getGenericJSONRPCErrResponse(logger, jsonrpcReq.ID, data, err), &validationErr
	}

	// JSONRPC response successfully validated.
	return jsonrpcResponse, nil
}

// getGenericJSONRPCErrResponse returns a generic response wrapped around a JSONRPC error response with the supplied ID, error, and the invalid payload in the "data" field.
func getGenericJSONRPCErrResponse(
	id jsonrpc.ID,
	malformedResponsePayload []byte,
	err error,
) responseGeneric {
	errData := map[string]string{
		errDataFieldRawBytes:        string(malformedResponsePayload),
		errDataFieldUnmarshalingErr: err.Error(),
	}

	return responseGeneric{
		Response: jsonrpc.GetErrorResponse(id, errCodeUnmarshaling, errMsgUnmarshaling, errData),
	}
}
