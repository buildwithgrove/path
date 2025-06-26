package cometbft

import (
	"encoding/json"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const (
	// errCodeUnmarshaling is set as the JSON-RPC response's error code if the endpoint returns a malformed response.
	errCodeUnmarshaling = -32600

	// errMsgUnmarshaling is the generic message returned to the user if the endpoint returns a malformed response.
	errMsgUnmarshaling = "the response returned by the endpoint is not a valid JSON-RPC response"

	// errDataFieldRawBytes is the key of the entry in the JSON-RPC error response's "data" map which holds the endpoint's original response.
	errDataFieldRawBytes = "endpoint_response"

	// errDataFieldUnmarshalingErr is the key of the entry in the JSON-RPC error response's "data" map which holds the unmarshaling error.
	errDataFieldUnmarshalingErr = "unmarshaling_error"
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
// TODO_MVP(@adshmh): handle any unmarshaling errors and build a method-specific payload generator.
func (r responseGeneric) GetResponsePayload() []byte {
	bz, err := json.Marshal(r.jsonRPCResponse)
	if err != nil {
		// This should never happen: log an entry but return the response anyway.
		r.logger.Warn().Err(err).Msg("responseGeneric: Marshaling JSON-RPC response failed.")
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
		errDataFieldRawBytes:        string(malformedResponsePayload),
		errDataFieldUnmarshalingErr: err.Error(),
	}

	// CometBFT always returns a "1" ID for error responses.
	if response.ID.IsEmpty() {
		response.ID = errorID
	}

	return responseGeneric{
		jsonRPCResponse: jsonrpc.GetErrorResponse(response.ID, errCodeUnmarshaling, errMsgUnmarshaling, errData),
	}
}
