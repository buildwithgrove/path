package cosmos

import (
	"encoding/json"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const (
	// errCodeJSONRPCUnmarshaling is set as the JSON-RPC response's error code if the endpoint returns a malformed response.
	// -32000 Error code will result in returning a 500 HTTP Status Code to the client.
	errCodeJSONRPCUnmarshaling = -32000

	// errCodeJSONRPCEmptyResponse is set as the JSON-RPC response's error code if the endpoint returns an empty response.
	// -32000 Error code will result in returning a 500 HTTP Status Code to the client.
	errCodeJSONRPCEmptyResponse = -32000

	// errMsgJSONRPCUnmarshaling is the generic message returned to the user if the endpoint returns a malformed response.
	errMsgJSONRPCUnmarshaling = "the response returned by the endpoint is not a valid JSON-RPC response"

	// errDataFieldJSONRPCRawBytes is the key of the entry in the JSON-RPC error response's "data" map which holds the endpoint's original response.
	errDataFieldJSONRPCRawBytes = "endpoint_response"

	// errDataFieldJSONRPCUnmarshalingErr is the key of the entry in the JSON-RPC error response's "data" map which holds the unmarshaling error.
	errDataFieldJSONRPCUnmarshalingErr = "unmarshaling_error"
)

// responseUnmarshalerRESTGeneric processes raw response data into a responseRESTGeneric struct.
// It extracts and stores any data needed for generating a response payload.
// Always returns a valid response interface, never returns an error.
func responseUnmarshalerRESTGeneric(
	logger polylog.Logger,
	data []byte,
) response {
	logger = logger.With("response_processor", "generic")

	// Handle empty responses
	if len(data) == 0 {
		logger.Error().
			Str("endpoint_type", "REST").
			Msg("Received empty response from REST endpoint")

		return getRESTGenericEmptyErrorResponse(logger)
	}

	// Try to unmarshal as JSON-RPC response (for REST endpoints that return JSON-RPC)
	var jsonrpcResponse jsonrpc.Response
	if err := json.Unmarshal(data, &jsonrpcResponse); err != nil {
		logger.Error().
			Err(err).
			Str("raw_payload", string(data)).
			Msg("Failed to unmarshal REST response as JSON-RPC")

		return getRESTGenericUnmarshalErrorResponse(logger, err)
	}

	logger.Debug().
		Str("jsonrpc_id", jsonrpcResponse.ID.String()).
		Bool("has_error", jsonrpcResponse.IsError()).
		Msg("Successfully parsed generic REST response as JSON-RPC")

	return responseRESTGeneric{
		logger:          logger,
		jsonRPCResponse: jsonrpcResponse,
	}
}

// responseRESTGeneric captures the fields expected in response to any REST request on the CosmosSDK blockchain.
// It is intended to be used when no validation/observation is applicable to the corresponding request's REST endpoint.
// i.e. when there are no unmarshallers/structs matching the endpoint specified by the request.
type responseRESTGeneric struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSON-RPC response parsed from an endpoint's response bytes
	jsonRPCResponse jsonrpc.Response

	// validationError tracks any validation issues with the response
	validationError *qosobservations.CosmosSDKResponseValidationError
}

// GetObservation returns an observation that is NOT used in validating endpoints.
// This allows sharing data with other entities, e.g. a data pipeline.
// Implements the response interface.
// This is a default catchall for any response to any REST requests other than specific validated endpoints.
func (r responseRESTGeneric) GetObservation() qosobservations.CosmosSDKEndpointObservation {
	return qosobservations.CosmosSDKEndpointObservation{
		ResponseObservation: &qosobservations.CosmosSDKEndpointObservation_RestObservation{
			RestObservation: &qosobservations.CosmosSDKEndpointRestObservation{
				ParsedResponse: &qosobservations.CosmosSDKEndpointRestObservation_UnrecognizedResponse{
					UnrecognizedResponse: &qosobservations.CosmosSDKRestUnrecognizedResponse{
						HttpStatusCode: int32(r.GetResponseStatusCode()),
						JsonrpcResponse: &qosobservations.JsonRpcResponse{
							Id: r.jsonRPCResponse.ID.String(),
						},
					},
				},
			},
		},
	}
}

// GetResponsePayload returns the payload for the response to a REST request.
// Implements the response interface.
func (r responseRESTGeneric) GetResponsePayload() []byte {
	return r.getResponsePayload()
}

// getResponsePayload returns the JSON-RPC response as bytes
func (r responseRESTGeneric) getResponsePayload() []byte {
	responseBytes, _ := json.Marshal(r.jsonRPCResponse)
	return responseBytes
}

// GetResponseStatusCode returns an HTTP status code corresponding to the underlying JSON-RPC response code
// DEV_NOTE: This is an opinionated mapping following best practice but not enforced by any specifications or standards
// Implements the response interface
func (r responseRESTGeneric) GetResponseStatusCode() int {
	// If we have a validation error, return 500
	if r.validationError != nil {
		return http.StatusInternalServerError
	}

	// Use JSON-RPC response's recommended status code
	return r.jsonRPCResponse.GetRecommendedHTTPStatusCode()
}

// GetHTTPResponse builds and returns the httpResponse matching the responseRESTGeneric instance
// Implements the response interface
func (r responseRESTGeneric) GetHTTPResponse() httpResponse {
	return httpResponse{
		responsePayload: r.GetResponsePayload(),
		httpStatusCode:  r.GetResponseStatusCode(),
	}
}

// getRESTGenericEmptyErrorResponse creates an error response for empty REST responses
func getRESTGenericEmptyErrorResponse(logger polylog.Logger) responseRESTGeneric {
	errorResp := jsonrpc.GetErrorResponse(
		jsonrpc.IDFromInt(-1), // Use -1 for unknown ID
		errCodeJSONRPCEmptyResponse,
		"the REST endpoint returned an empty response",
		nil,
	)

	validationError := qosobservations.CosmosSDKResponseValidationError_COSMOS_SDK_RESPONSE_VALIDATION_ERROR_EMPTY

	return responseRESTGeneric{
		logger:          logger,
		jsonRPCResponse: errorResp,
		validationError: &validationError,
	}
}

// getRESTGenericUnmarshalErrorResponse creates an error response for REST unmarshaling failures
func getRESTGenericUnmarshalErrorResponse(
	logger polylog.Logger,
	err error,
) responseRESTGeneric {
	errData := map[string]string{
		errDataFieldJSONRPCUnmarshalingErr: err.Error(),
	}

	errorResp := jsonrpc.GetErrorResponse(
		jsonrpc.IDFromInt(-1), // Use -1 for unknown ID
		errCodeJSONRPCUnmarshaling,
		errMsgJSONRPCUnmarshaling,
		errData,
	)

	validationError := qosobservations.CosmosSDKResponseValidationError_COSMOS_SDK_RESPONSE_VALIDATION_ERROR_UNMARSHAL

	return responseRESTGeneric{
		logger:          logger,
		jsonRPCResponse: errorResp,
		validationError: &validationError,
	}
}
