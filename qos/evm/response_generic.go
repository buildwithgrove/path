package evm

import (
	"encoding/json"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const (
	// errCodeUnmarshaling is set as the JSONRPC response's error code if the endpoint returns a malformed response.
	// -32000 Error code will result in returning a 500 HTTP Status Code to the client.
	errCodeUnmarshaling = -32000

	// errMsgUnmarshaling is the generic message returned to the user if the endpoint returns a malformed response.
	errMsgUnmarshaling = "the response returned by the endpoint is not a valid JSONRPC response"

	// errDataFieldRawBytes is the key of the entry in the JSONRPC error response's "data" map which holds the endpoint's original response.
	errDataFieldRawBytes = "endpoint_response"

	// errDataFieldUnmarshalingErr is the key of the entry in the JSONRPC error response's "data" map which holds the unmarshaling error.
	errDataFieldUnmarshalingErr = "unmarshaling_error"
)

// responseGeneric represents the standard response structure for EVM-based blockchain requests.
// Used as a fallback when:
// - No validation/observation is needed for the JSON-RPC method
// - No specific unmarshallers/structs exist for the request method
// responseGeneric captures the fields expected in response to any request on an
// EVM-based blockchain. It is intended to be used when no validation/observation
// is applicable to the corresponding request's JSONRPC method.
// i.e. when there are no unmarshallers/structs matching the method specified by the request.
type responseGeneric struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSONRPC response parsed from an endpoint's response bytes.
	jsonRPCResponse jsonrpc.Response

	// Why the response has failed validation.
	// Only set if the response is invalid.
	// As of PR #152, a response is deemed valid if it can be unmarshaled as a JSONRPC struct
	// regardless of the contents of the response.
	// Used when generating observations.
	validationError *qosobservations.EVMResponseValidationError
}

// GetObservation returns an observation that is used in validating endpoints.
// This generates observations for unrecognized responses that can trigger endpoint
// disqualification when validation errors are present.
// Implements the response interface.
func (r responseGeneric) GetObservation() qosobservations.EVMEndpointObservation {
	return qosobservations.EVMEndpointObservation{
		ResponseObservation: &qosobservations.EVMEndpointObservation_UnrecognizedResponse{
			UnrecognizedResponse: &qosobservations.EVMUnrecognizedResponse{
				// Include JSONRPC response's details in the observation.
				JsonrpcResponse:         r.jsonRPCResponse.GetObservation(),
				ResponseValidationError: r.validationError,
				HttpStatusCode:          int32(r.getHTTPStatusCode()),
			},
		},
	}
}

// GetHTTPResponse builds and returns the httpResponse matching the responseGeneric instance.
// Implements the response interface.
func (r responseGeneric) GetHTTPResponse() httpResponse {
	return httpResponse{
		responsePayload: r.getResponsePayload(),
		// Use the HTTP status code recommended by for the underlying JSONRPC response by the jsonrpc package.
		httpStatusCode: r.getHTTPStatusCode(),
	}
}

// TODO_MVP(@adshmh): handle any unmarshaling errors
// TODO_INCOMPLETE: build a method-specific payload generator.
func (r responseGeneric) getResponsePayload() []byte {
	bz, err := json.Marshal(r.jsonRPCResponse)
	if err != nil {
		// This should never happen: log an entry but return the response anyway.
		r.logger.Warn().Err(err).Msg("responseGeneric: Marshaling JSONRPC response failed.")
	}
	return bz
}

// getHTTPStatusCode returns an HTTP status code corresponding to the underlying JSON-RPC response code.
// DEV_NOTE: This is an opinionated mapping following best practice but not enforced by any specifications or standards.
func (r responseGeneric) getHTTPStatusCode() int {
	return r.jsonRPCResponse.GetRecommendedHTTPStatusCode()
}

// responseUnmarshallerGenericFromResponse processes an already unmarshaled JSON-RPC response into a responseGeneric struct.
// This avoids double unmarshaling when the response has already been parsed.
func responseUnmarshallerGenericFromResponse(logger polylog.Logger, jsonrpcReq jsonrpc.Request, jsonrpcResponse jsonrpc.Response) (response, error) {
	httpStatus := jsonrpcResponse.GetRecommendedHTTPStatusCode()
	logger.With(
		"jsonrpc_response", jsonrpcResponse,
		"jsonrpc_request", jsonrpcReq,
		"http_status", httpStatus,
	).Debug().Msg("Processing EVM generic response")

	// Response successfully parsed into JSONRPC format.
	return responseGeneric{
		logger:          logger,
		jsonRPCResponse: jsonrpcResponse,
		validationError: nil, // Intentionally set to nil to indicate a valid JSONRPC response.
	}, nil
}

// getGenericJSONRPCErrResponse creates a generic response containing:
// - JSON-RPC error with supplied ID
// - Error details
// - Invalid payload in the "data" field
// Sets validation error to UNMARSHAL to trigger endpoint disqualification.
func getGenericJSONRPCErrResponse(_ polylog.Logger, id jsonrpc.ID, malformedResponsePayload []byte, err error) responseGeneric {
	errData := map[string]string{
		errDataFieldRawBytes:        string(malformedResponsePayload),
		errDataFieldUnmarshalingErr: err.Error(),
	}

	validationError := qosobservations.EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_UNMARSHAL
	return responseGeneric{
		jsonRPCResponse: jsonrpc.GetErrorResponse(id, errCodeUnmarshaling, errMsgUnmarshaling, errData),
		validationError: &validationError, // Set validation error to trigger endpoint disqualification
	}
}

// TODO_INCOMPLETE: Handle the string `null`, as it could be returned
// when an object is expected.
// See the following link for more details:
// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_gettransactionbyhash
