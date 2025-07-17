package cosmos

import (
	"encoding/json"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const (
	// -32000 Error code will result in returning a 500 HTTP Status Code to the client.
	errCodeUnmarshaling = -32000

	// errMsgUnmarshaling is the generic message returned to the user if the endpoint returns a malformed response.
	errMsgUnmarshaling = "the response returned by the endpoint is not a valid JSON-RPC response"

	// errCodeEmptyResponse is the error code returned to the user if the endpoint returns an empty response.
	errCodeEmptyResponse = -32001

	// errMsgEmptyResponse is the message returned to the user if the endpoint returns an empty response.
	errMsgEmptyResponse = "the response returned by the endpoint is empty"

	// errDataFieldRawBytes is the key of the entry in the JSON-RPC error response's "data" map which holds the endpoint's original response.
	errDataFieldRawBytes = "endpoint_response"

	// errDataFieldUnmarshalingErr is the key of the entry in the JSON-RPC error response's "data" map which holds the unmarshaling error.
	errDataFieldUnmarshalingErr = "unmarshaling_error"
)

// responseGeneric represents the standard response structure for CometBFT-based blockchain requests.
// Used as a fallback when:
// - No validation/observation is needed for the JSON-RPC method
// - No specific unmarshallers/structs exist for the request method
// - REST responses that don't conform to JSON-RPC standards
type responseGeneric struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSON-RPC response parsed from an endpoint's response bytes.
	// Will be empty for REST responses.
	jsonRPCResponse jsonrpc.Response

	// rawData stores the raw response data for REST responses
	rawData []byte

	// isRestResponse indicates whether this is a REST response (true) or JSON-RPC response (false)
	isRestResponse bool
}

// GetObservation returns an observation that is NOT used in validating endpoints.
// This allows sharing data with other entities, e.g. a data pipeline.
// Implements the response interface.
func (r responseGeneric) GetObservation() qosobservations.CosmosSDKEndpointObservation {
	var responseID string
	if r.isRestResponse {
		// For REST responses, we don't have a JSON-RPC ID
		responseID = ""
	} else {
		responseID = r.jsonRPCResponse.ID.String()
	}

	return qosobservations.CosmosSDKEndpointObservation{
		ResponseObservation: &qosobservations.CosmosSDKEndpointObservation_UnrecognizedResponse{
			UnrecognizedResponse: &qosobservations.CosmosSDKUnrecognizedResponse{
				JsonrpcResponse: &qosobservations.JsonRpcResponse{
					Id: responseID,
				},
			},
		},
	}
}

// GetResponsePayload returns the payload for the response.
// For REST responses, returns the raw data as-is.
// For JSON-RPC responses, marshals the JSON-RPC response.
// Implements the response interface.
func (r responseGeneric) GetResponsePayload() []byte {
	if r.isRestResponse {
		// For REST responses, return the raw data as-is
		return r.rawData
	}

	// For JSON-RPC responses, marshal the JSON-RPC response
	bz, err := json.Marshal(r.jsonRPCResponse)
	if err != nil {
		// This should never happen: log an entry but return the response anyway.
		r.logger.Warn().Err(err).Msg("responseGeneric: Marshaling JSON-RPC response failed.")
	}
	return bz
}

// GetResponseStatusCode returns the appropriate HTTP status code.
// For REST responses, defaults to 200 OK (assuming valid JSON).
// For JSON-RPC responses, follows CometBFT response codes:
// - 200: Success
// - 500: Error
// Reference: https://docs.cometbft.com/v1.0/spec/rpc/
// Implements the response interface.
func (r responseGeneric) GetResponseStatusCode() int {
	if r.isRestResponse {
		// For REST responses, default to 200 OK
		return http.StatusOK
	}

	// For JSON-RPC responses, check for errors
	if r.jsonRPCResponse.IsError() {
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

// GetHTTPResponse builds and returns the httpResponse matching the responseGeneric instance.
// Implements the response interface.
func (r responseGeneric) GetHTTPResponse() httpResponse {
	return httpResponse{
		responsePayload: r.GetResponsePayload(),
		httpStatusCode:  r.GetResponseStatusCode(),
	}
}

// responseUnmarshallerGeneric processes raw response data into a responseGeneric struct.
// It handles both JSON-RPC and REST responses.
// If the provided jsonrpcResp is empty/invalid, it treats the raw data as a REST response.
func responseUnmarshallerGeneric(
	logger polylog.Logger,
	jsonrpcResp jsonrpc.Response,
	data []byte,
	isJSONRPC bool,
) (response, error) {
	// If the jsonrpcResp is empty (indicating JSON-RPC unmarshaling failed),
	// treat this as a response to a REST-like request
	if jsonrpcResp.ID.IsEmpty() {
		// If the data is empty, return an error response
		if len(data) == 0 {
			return getEmptyJSONRPCErrResponse(logger, jsonrpcResp, isJSONRPC), nil
		}

		// If the data is not empty, validate that it is valid JSON for REST responses
		var jsonData any
		if err := json.Unmarshal(data, &jsonData); err != nil {
			// If it's not valid JSON, return an error response
			return getGenericJSONRPCErrResponse(logger, jsonrpcResp, data, isJSONRPC, err), nil
		}

		// If the data is valid JSON, treat it as a REST response.
		return responseGeneric{
			logger:         logger,
			rawData:        data,
			isRestResponse: true,
		}, nil
	}

	// If data is provided and we have a valid JSON-RPC response, use it
	if len(data) > 0 {
		var response jsonrpc.Response
		if err := json.Unmarshal(data, &response); err != nil {
			return getGenericJSONRPCErrResponse(logger, response, data, isJSONRPC, err), nil
		}

		return responseGeneric{
			logger:          logger,
			jsonRPCResponse: response,
			isRestResponse:  false,
		}, nil
	}

	// Use the provided JSON-RPC response
	return responseGeneric{
		logger:          logger,
		jsonRPCResponse: jsonrpcResp,
		isRestResponse:  false,
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
	isJSONRPC bool,
	err error,
) responseGeneric {
	errData := map[string]string{
		errDataFieldRawBytes:        string(malformedResponsePayload),
		errDataFieldUnmarshalingErr: err.Error(),
	}

	// If the response ID is empty and the request is not JSON-RPC, use the restLikeResponseID
	if !isJSONRPC && response.ID.IsEmpty() {
		response.ID = restLikeResponseID
	}

	return responseGeneric{
		jsonRPCResponse: jsonrpc.GetErrorResponse(response.ID, errCodeUnmarshaling, errMsgUnmarshaling, errData),
	}
}

// getEmptyJSONRPCErrResponse creates a generic response containing:
// - JSON-RPC error
// - Empty response payload
func getEmptyJSONRPCErrResponse(
	_ polylog.Logger,
	response jsonrpc.Response,
	isJSONRPC bool,
) responseGeneric {
	// If the response ID is empty and the request is not JSON-RPC, use the restLikeResponseID
	if !isJSONRPC && response.ID.IsEmpty() {
		response.ID = restLikeResponseID
	}

	return responseGeneric{
		jsonRPCResponse: jsonrpc.GetErrorResponse(response.ID, errCodeEmptyResponse, errMsgEmptyResponse, nil),
	}
}
