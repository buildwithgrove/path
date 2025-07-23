package cosmos

import (
	"encoding/json"
	"errors"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/log"
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

	// errDataFieldUnmarshalingErr is the key of the entry in the JSON-RPC error response's "data" map which holds the unmarshaling error.
	errDataFieldUnmarshalingErr = "unmarshaling_error"
)

// A jsonrpcResponseValidator takes a parsed JSONRPC response and verifies its contents against the expected result.
// e.g. The response validator for "status" method verifies the result field against the expected status info struct.
type jsonrpcResponseValidator func(polylog.Logger, jsonrpc.Response) response

var (
	// All response types must implement the response interface.
	_ jsonrpcResponseValidator = responseValidatorCometBFTHealth
	_ jsonrpcResponseValidator = responseValidatorCometBFTStatus

	// Maps JSONRPC requests to their corresponding response validators, based on the JSONRPC method.
	jsonrpcRequestEndpointResponseValidators = map[string]jsonrpcResponseValidator{
		"health":      responseValidatorCometBFTHealth,
		"status":      responseValidatorCometBFTStatus,
		"eth_chainId": responseValidatorEVMChainID,
	}
)

// unmarshalJSONRPCRequestEndpointResponse parses the supplied raw byte slice from an endpoint.
// The raw byte is returned by an endpoint in response to a JSONRPC request.
func unmarshalJSONRPCRequestEndpointResponse(
	logger polylog.Logger,
	jsonrpcReq jsonrpc.Request,
	data []byte,
) response {
	// Parse and validate the raw payload as a JSONRPC response.
	jsonrpcResponse, responseValidationErr := unmarshalAsJSONRPCResponse(logger, jsonrpcReq.ID, data)

	// TODO_TECHDEBT(@adshmh): Separate User-response, which could be a generic response indicating an endpoint error, from the parsed response.
	// Endpoint response failed validation.
	// Return a generic response to the user.
	if responseValidationErr != nil {
		return &jsonrpcUnrecognizedResponse{
			logger: logger,
			// The generic user-facing response indicating an endpoint error.
			jsonrpcResponse: jsonrpcResponse,
			validationErr:   *responseValidationErr,
		}
	}

	// Lookup the JSONRPC method-specific validator for the response.
	jsonrpcRequestMethod := string(jsonrpcReq.Method)
	validator, found := jsonrpcRequestEndpointResponseValidators[jsonrpcRequestMethod]
	if found {
		return validator(logger, jsonrpcResponse)
	}

	// Default to a generic response if no method-specific response is found.
	return &jsonrpcUnrecognizedResponse{
		logger:          logger,
		jsonrpcResponse: jsonrpcResponse,
		validationErr:   qosobservations.CosmosResponseValidationError_COSMOS_RESPONSE_VALIDATION_ERROR_UNSPECIFIED,
	}
}

// unmarshalAsJSONRPCResponse converts raw endpoint bytes into a JSONRPC response struct.
// The second return value contains the validation failure, if any.
func unmarshalAsJSONRPCResponse(
	logger polylog.Logger,
	jsonrpcRequestID jsonrpc.ID,
	data []byte,
) (jsonrpc.Response, *qosobservations.CosmosResponseValidationError) {
	// Empty payload is invalid.
	if len(data) == 0 {
		errEmptyPayload := errors.New("failed to unmarshal endpoint payload as JSONRPC: endpoint returned an empty response")
		logger.With(
			"unmarshal_err", errEmptyPayload,
			"error_type", "empty_response",
		).Debug().Msg(errEmptyPayload.Error())

		// Create a generic JSONRPC response for the user.
		validationErr := qosobservations.CosmosResponseValidationError_COSMOS_RESPONSE_VALIDATION_ERROR_EMPTY
		return getGenericJSONRPCErrResponse(logger, jsonrpcRequestID, errCodeEmptyResponse, errMsgJSONRPCEmptyResponse), &validationErr
	}

	// Unmarshal the raw response payload into a JSONRPC response.
	var jsonrpcResponse jsonrpc.Response
	err := json.Unmarshal(data, &jsonrpcResponse)
	if err != nil {
		// The response raw payload (e.g. as received from an endpoint) could not be unmarshalled as a JSONRPC response.
		// Return a generic response to the user.
		payloadStr := string(data)
		logger.With(
			"unmarshal_err", err,
			"raw_payload", log.Preview(payloadStr),
			"error_type", "parse_error",
		).Debug().Msg("Failed to unmarshal endpoint payload as JSONRPC.")

		// Create a generic JSONRPC response for the user.
		validationErr := qosobservations.CosmosResponseValidationError_COSMOS_RESPONSE_VALIDATION_ERROR_UNMARSHAL
		return getGenericJSONRPCErrResponse(logger, jsonrpcRequestID, errCodeUnmarshaling, errMsgJSONRPCUnmarshaling), &validationErr
	}

	// Validate the JSONRPC response.
	if err := jsonrpcResponse.Validate(jsonrpcRequestID); err != nil {
		payloadStr := string(data)
		logger.With(
			"unmarshal_err", err,
			"raw_payload", log.Preview(payloadStr),
			"error_type", "validation_error",
		).Debug().Msg("Failed to unmarshal endpoint payload as JSONRPC: JSONRPC response failed validation.")

		validationErr := qosobservations.CosmosResponseValidationError_COSMOS_RESPONSE_VALIDATION_ERROR_FORMAT_MISMATCH
		return getGenericJSONRPCErrResponse(logger, jsonrpcRequestID, errCodeUnmarshaling, errMsgJSONRPCUnmarshaling), &validationErr
	}

	// JSONRPC response successfully validated.
	return jsonrpcResponse, nil
}

// getGenericJSONRPCErrResponse returns a generic response wrapped around a JSONRPC error response with the supplied ID, error, and the invalid payload in the "data" field.
func getGenericJSONRPCErrResponse(
	logger polylog.Logger,
	id jsonrpc.ID,
	errCode int,
	errMsg string,
) jsonrpc.Response {
	errData := map[string]string{
		errDataFieldUnmarshalingErr: errMsg,
	}

	return jsonrpc.GetErrorResponse(id, errCode, errMsg, errData)
}
