package cosmos

import (
	"encoding/json"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const (
	// Error codes for JSONRPC response validation failures
	errCodeJSONRPCUnmarshaling  = -32000
	errCodeJSONRPCEmptyResponse = -32000

	// Error messages for JSONRPC response validation failures
	errMsgJSONRPCUnmarshaling  = "the JSON-RPC response returned by the endpoint is not valid"
	errMsgJSONRPCEmptyResponse = "the endpoint returned an empty JSON-RPC response"

	// Error data field keys for JSONRPC responses
	errDataFieldJSONRPCRawBytes        = "endpoint_response"
	errDataFieldJSONRPCUnmarshalingErr = "unmarshaling_error"
)

// responseJSONRPC represents a JSON-RPC response from a Cosmos endpoint
// Used for all JSON-RPC requests regardless of method
type responseJSONRPC struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSON-RPC response parsed from an endpoint's response bytes
	jsonRPCResponse jsonrpc.Response

	// validationError tracks any validation issues with the response
	validationError *qosobservations.CosmosSDKResponseValidationError
}

// GetObservation returns an observation for JSON-RPC responses
// Implements the response interface
func (r responseJSONRPC) GetObservation() qosobservations.CosmosSDKEndpointObservation {
	return qosobservations.CosmosSDKEndpointObservation{
		ResponseObservation: &qosobservations.CosmosSDKEndpointObservation_JsonrpcObservation{
			JsonrpcObservation: &qosobservations.CosmosSDKEndpointJsonRpcObservation{
				JsonrpcResponse: &qosobservations.JsonRpcResponse{
					Id:     r.jsonRPCResponse.ID.String(),
					Result: r.getResultAsString(),
					Err:    r.getErrorObservation(),
				},
				ParsedResponse: &qosobservations.CosmosSDKEndpointJsonRpcObservation_UnrecognizedResponse{
					UnrecognizedResponse: &qosobservations.CosmosSDKJSONRPCUnrecognizedResponse{
						JsonrpcResponse: &qosobservations.JsonRpcResponse{
							Id:     r.jsonRPCResponse.ID.String(),
							Result: r.getResultAsString(),
							Err:    r.getErrorObservation(),
						},
					},
				},
			},
		},
	}
}

// GetResponsePayload returns the payload for the JSON-RPC response
// Implements the response interface
func (r responseJSONRPC) GetResponsePayload() []byte {
	bz, err := json.Marshal(r.jsonRPCResponse)
	if err != nil {
		// This should never happen: log an entry but return the response anyway
		r.logger.Warn().Err(err).Msg("responseJSONRPC: Marshaling JSON-RPC response failed")
	}
	return bz
}

// GetResponseStatusCode returns the appropriate HTTP status code for JSON-RPC responses
// Follows CometBFT response codes: 200 for success, 500 for errors
// Reference: https://docs.cometbft.com/v1.0/spec/rpc/
// Implements the response interface
func (r responseJSONRPC) GetResponseStatusCode() int {
	if r.jsonRPCResponse.IsError() {
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

// GetHTTPResponse builds and returns the httpResponse matching the responseJSONRPC instance
// Implements the response interface
func (r responseJSONRPC) GetHTTPResponse() httpResponse {
	return httpResponse{
		responsePayload: r.GetResponsePayload(),
		httpStatusCode:  r.GetResponseStatusCode(),
	}
}

// getResultAsString safely extracts the result field as a string
func (r responseJSONRPC) getResultAsString() string {
	if r.jsonRPCResponse.Result == nil {
		return ""
	}

	resultBz, err := json.Marshal(r.jsonRPCResponse.Result)
	if err != nil {
		r.logger.Warn().Err(err).Msg("Failed to marshal JSON-RPC result for observation")
		return ""
	}

	return string(resultBz)
}

// getErrorObservation extracts error details for observations
func (r responseJSONRPC) getErrorObservation() *qosobservations.JsonRpcResponseError {
	if r.jsonRPCResponse.Error == nil {
		return nil
	}

	return &qosobservations.JsonRpcResponseError{
		Code:    r.jsonRPCResponse.Error.Code,
		Message: r.jsonRPCResponse.Error.Message,
	}
}

// responseUnmarshalerJSONRPC processes raw response data into a responseJSONRPC struct
func responseUnmarshalerJSONRPC(
	logger polylog.Logger,
	data []byte,
) (response, error) {
	// Handle empty responses
	if len(data) == 0 {
		logger.Error().Msg("Received empty JSON-RPC response from endpoint")
		return getEmptyJSONRPCErrorResponse(logger), nil
	}

	// Unmarshal the raw response payload into a JSON-RPC response
	var jsonrpcResponse jsonrpc.Response
	if err := json.Unmarshal(data, &jsonrpcResponse); err != nil {
		logger.Error().
			Err(err).
			Str("raw_payload", string(data)).
			Msg("Failed to unmarshal JSON-RPC response from endpoint")

		return getJSONRPCUnmarshalErrorResponse(logger, data, err), nil
	}

	// Create response with successful unmarshaling
	return responseJSONRPC{
		logger:          logger,
		jsonRPCResponse: jsonrpcResponse,
		validationError: nil, // No validation error for successfully unmarshaled responses
	}, nil
}

// getEmptyJSONRPCErrorResponse creates an error response for empty JSON-RPC responses
func getEmptyJSONRPCErrorResponse(logger polylog.Logger) responseJSONRPC {
	errorResp := jsonrpc.GetErrorResponse(
		jsonrpc.IDFromInt(-1), // Use -1 for unknown ID
		errCodeJSONRPCEmptyResponse,
		errMsgJSONRPCEmptyResponse,
		nil,
	)

	validationError := qosobservations.CosmosSDKResponseValidationError_COSMOS_SDK_RESPONSE_VALIDATION_ERROR_EMPTY

	return responseJSONRPC{
		logger:          logger,
		jsonRPCResponse: errorResp,
		validationError: &validationError,
	}
}

// getJSONRPCUnmarshalErrorResponse creates an error response for JSON-RPC unmarshaling failures
func getJSONRPCUnmarshalErrorResponse(
	logger polylog.Logger,
	malformedResponsePayload []byte,
	err error,
) responseJSONRPC {
	errData := map[string]string{
		errDataFieldJSONRPCRawBytes:        string(malformedResponsePayload),
		errDataFieldJSONRPCUnmarshalingErr: err.Error(),
	}

	errorResp := jsonrpc.GetErrorResponse(
		jsonrpc.IDFromInt(-1), // Use -1 for unknown ID
		errCodeJSONRPCUnmarshaling,
		errMsgJSONRPCUnmarshaling,
		errData,
	)

	validationError := qosobservations.CosmosSDKResponseValidationError_COSMOS_SDK_RESPONSE_VALIDATION_ERROR_UNMARSHAL

	return responseJSONRPC{
		logger:          logger,
		jsonRPCResponse: errorResp,
		validationError: &validationError,
	}
}
