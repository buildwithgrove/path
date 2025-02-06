package cometbft

import (
	"encoding/json"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const (
	// errCodeUnmarshalling is set as the JSON-RPC response's error code if the endpoint returns a malformed response.
	errCodeUnmarshalling = -32600

	// errMsgUnmarshalling is the generic message returned to the user if the endpoint returns a malformed response.
	errMsgUnmarshalling = "the response returned by the endpoint is not a valid JSON-RPC response"

	// errDataFieldRawBytes is the key of the entry in the JSON-RPC error response's "data" map which holds the endpoint's original response.
	errDataFieldRawBytes = "endpoint_response"

	// errDataFieldUnmarshallingErr is the key of the entry in the JSON-RPC error response's "data" map which holds the unmarshalling error.
	errDataFieldUnmarshallingErr = "unmarshalling_error"
)

// responseGeneric represents the standard response structure for CometBFT-based blockchain requests.
// Used as a fallback when:
// - No validation/observation is needed for the JSON-RPC method
// - No specific unmarshallers/structs exist for the request method
type responseGeneric struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSON-RPC response parsed from an endpoint's response bytes.
	jsonRPCResponse jsonrpc.Response
}

// GetObservation returns an observation that is NOT used in validating endpoints.
// This allows sharing data with other entities, e.g. a data pipeline.
// Implements the response interface.
func (r responseGeneric) GetObservation() qosobservations.CometBFTEndpointObservation {
	return qosobservations.CometBFTEndpointObservation{
		ResponseObservation: &qosobservations.CometBFTEndpointObservation_UnrecognizedResponse{
			UnrecognizedResponse: &qosobservations.CometBFTUnrecognizedResponse{
				JsonrpcResponse: &qosobservations.JsonRpcResponse{
					Id: r.jsonRPCResponse.ID.String(),
				},
			},
		},
	}
}

// GetResponsePayload returns the payload for the response to a `/health` request.
// Implements the response interface.
//
// TODO_MVP(@adshmh): handle any unmarshalling errors and build a method-specific payload generator.
func (r responseGeneric) GetResponsePayload() []byte {
	bz, err := json.Marshal(r.jsonRPCResponse)
	if err != nil {
		// This should never happen: log an entry but return the response anyway.
		r.logger.Warn().Err(err).Msg("responseGeneric: Marshalling JSON-RPC response failed.")
	}
	return bz
}

// CometBFT response codes:
// - 200: Success
// - 500: Error
// Reference: https://docs.cometbft.com/v0.38/rpc/
// Implements the response interface.
func (r responseGeneric) GetResponseStatusCode() int {
	if r.jsonRPCResponse.IsError() {
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

// responseUnmarshallerGeneric processes raw response data into a responseGeneric struct.
// It extracts and stores any data needed for generating a response payload.
func responseUnmarshallerGeneric(
	logger polylog.Logger,
	_ jsonrpc.Response,
	data []byte,
) (response, error) {
	var response jsonrpc.Response
	if err := json.Unmarshal(data, &response); err != nil {
		return getGenericJSONRPCErrResponse(logger, response, data, err), nil
	}

	return responseGeneric{
		logger:          logger,
		jsonRPCResponse: response,
	}, nil
}

// getGenericJSONRPCErrResponse creates a generic response containing:
// - JSON-RPC error with supplied ID
// - Error details
// - Invalid payload in the "data" field
func getGenericJSONRPCErrResponse(
	_ polylog.Logger,
	response jsonrpc.Response,
	malformedResponsePayload []byte,
	err error,
) responseGeneric {
	errData := map[string]string{
		errDataFieldRawBytes:         string(malformedResponsePayload),
		errDataFieldUnmarshallingErr: err.Error(),
	}

	// CometBFT always returns a "1" ID for error responses.
	if response.ID.IsEmpty() {
		response.ID = errorID
	}

	return responseGeneric{
		jsonRPCResponse: jsonrpc.GetErrorResponse(response.ID, errCodeUnmarshalling, errMsgUnmarshalling, errData),
	}
}
