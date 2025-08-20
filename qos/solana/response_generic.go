package solana

import (
	"encoding/json"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"google.golang.org/protobuf/types/known/timestamppb"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_UPNEXT(@adshmh): Update solana and cometbft QoS to detect and sanction malformed endpoint responses to any request.
// See evm implementation in #321 for reference.

const (
	// errCodeUnmarshaling is set as the JSON-RPC response's error code if the endpoint returns a malformed response.
	// -32000 Error code will result in returning a 500 HTTP Status Code to the client.
	errCodeUnmarshaling = -32000

	// errMsgUnmarshaling is the generic message returned to the user if the endpoint returns a malformed response.
	errMsgUnmarshaling = "the response returned by the endpoint is not a valid JSON-RPC response"

	// errDataFieldRawBytes is the key of the entry in the JSON-RPC error response's "data" map which holds the endpoint's original response.
	errDataFieldRawBytes = "endpoint_response"

	// errDataFieldUnmarshalingErr is the key of the entry in the JSON-RPC error response's "data" map which holds the unmarshaling error.
	errDataFieldUnmarshalingErr = "unmarshaling_error"
)

// responseUnmarshallerGeneric processes raw response data into a responseGeneric struct.
// It extracts and stores any data needed for generating a response payload.
func responseUnmarshallerGeneric(
	logger polylog.Logger,
	jsonrpcReq jsonrpc.Request,
	data []byte,
) response {
	var response jsonrpc.Response
	if err := json.Unmarshal(data, &response); err != nil {
		return getGenericJSONRPCErrResponse(logger, jsonrpcReq.ID, data, err)
	}

	return responseGeneric{
		Logger:   logger,
		Response: response,
	}
}

// responseGeneric captures fields for any Solana blockchain response.
// Used when no validation/observation applies to the request's JSON-RPC method.
type responseGeneric struct {
	Logger polylog.Logger
	jsonrpc.Response
	// jsonrpcResponseValidationError tracks JSON-RPC validation errors if response unmarshaling failed
	jsonrpcResponseValidationError *qosobservations.JsonRpcResponseValidationError
}

// GetObservation returns observation NOT used for endpoint validation.
// Shares data with other entities (e.g., data pipeline).
// Default catchall for responses other than `getHealth` and `getEpochInfo`.
func (r responseGeneric) GetObservation() qosobservations.SolanaEndpointObservation {
	// Build an observation from the stored JSONRPC response.
	unrecognizedResponse := &qosobservations.SolanaUnrecognizedResponse{
		JsonrpcResponse: r.Response.GetObservation(),
	}

	// Include validation error if present
	if r.jsonrpcResponseValidationError != nil {
		unrecognizedResponse.ValidationError = r.jsonrpcResponseValidationError
	}

	return qosobservations.SolanaEndpointObservation{
		// Set the HTTP status code using the JSONRPC Response
		HttpStatusCode: r.Response.GetRecommendedHTTPStatusCode(),
		ResponseObservation: &qosobservations.SolanaEndpointObservation_UnrecognizedResponse{
			UnrecognizedResponse: unrecognizedResponse,
		},
	}
}

// GetJSONRPCResponse returns response payload.
func (r responseGeneric) GetJSONRPCResponse() jsonrpc.Response {
	return r.Response
}

// getGenericJSONRPCErrResponse returns generic response with JSON-RPC error and validation error observation.
// Includes supplied ID, error, and invalid payload in "data" field.
func getGenericJSONRPCErrResponse(
	logger polylog.Logger,
	id jsonrpc.ID,
	malformedResponsePayload []byte,
	err error,
) responseGeneric {
	errData := map[string]string{
		errDataFieldRawBytes:        string(malformedResponsePayload),
		errDataFieldUnmarshalingErr: err.Error(),
	}

	// Create validation error observation
	jsonrpcResponseValidationError := &qosobservations.JsonRpcResponseValidationError{
		ErrorType: qosobservations.JsonRpcValidationErrorType_JSON_RPC_VALIDATION_ERROR_TYPE_NON_JSONRPC_RESPONSE,
		Timestamp: timestamppb.New(time.Now()),
	}

	return responseGeneric{
		Response:                       jsonrpc.GetErrorResponse(id, errCodeUnmarshaling, errMsgUnmarshaling, errData),
		jsonrpcResponseValidationError: jsonrpcResponseValidationError,
	}
}
