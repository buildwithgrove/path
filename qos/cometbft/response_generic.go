package cometbft

import (
	"encoding/json"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const (
	// errCodeUnmarshalling is set as the JSONRPC response's error code if the endpoint returns a malformed response
	errCodeUnmarshalling = -32600

	// errMsgUnmarshalling is the generic message returned to the user if the endpoint returns a malformed response.
	errMsgUnmarshalling = "the response returned by the endpoint is not a valid JSONRPC response"

	// errDataFieldRawBytes is the key of the entry in the JSONRPC error response's "data" map which holds the endpoint's original response.
	errDataFieldRawBytes = "endpoint_response"

	// errDataFieldUnmarshallingErr is the key of the entry in the JSONRPC error response's "data" map which holds the unmarshalling error.
	errDataFieldUnmarshallingErr = "unmarshalling_error"
)

// responseGeneric captures the fields expected in response to any request on an
// CometBFT-based blockchain. It is intended to be used when no validation/observation
// is applicable to the corresponding request's JSONRPC method.
// i.e. when there are no unmarshallers/structs matching the method specified by the request.
type responseGeneric struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSONRPC response parsed from an endpoint's response bytes.
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

// TODO_MVP(@adshmh): handle any unmarshalling errors
// TODO_INCOMPLETE: build a method-specific payload generator.
func (r responseGeneric) GetResponsePayload() []byte {
	bz, err := json.Marshal(r.jsonRPCResponse)
	if err != nil {
		// This should never happen: log an entry but return the response anyway.
		r.logger.Warn().Err(err).Msg("responseGeneric: Marshalling JSONRPC response failed.")
	}
	return bz
}

// CometBFT always returns either a 500 (on error) or 200 (on success).
// Reference: https://docs.cometbft.com/v0.38/rpc/
func (r responseGeneric) GetResponseStatusCode() int {
	if r.jsonRPCResponse.IsError() {
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

// responseUnmarshallerGeneric unmarshal the provided byte slice
// into a responseGeneric struct and saves any data that may be
// needed for producing a response payload into the struct.
func responseUnmarshallerGeneric(logger polylog.Logger, _ jsonrpc.Response, data []byte) (response, error) {
	var response jsonrpc.Response
	if err := json.Unmarshal(data, &response); err != nil {
		return getGenericJSONRPCErrResponse(logger, response, data, err), nil
	}

	return responseGeneric{
		logger:          logger,
		jsonRPCResponse: response,
	}, nil
}

// getGenericJSONRPCErrResponse returns a generic response wrapped around a JSONRPC error response with the supplied ID, error, and the invalid payload in the "data" field.
func getGenericJSONRPCErrResponse(_ polylog.Logger, response jsonrpc.Response, malformedResponsePayload []byte, err error) responseGeneric {
	errData := map[string]string{
		errDataFieldRawBytes:         string(malformedResponsePayload),
		errDataFieldUnmarshallingErr: err.Error(),
	}

	// CometBFT always returns a "1" ID for error responses.
	if response.ID.IsEmpty() {
		response.ID = cometBFTErrResponseID
	}

	return responseGeneric{
		jsonRPCResponse: jsonrpc.GetErrorResponse(response.ID, errCodeUnmarshalling, errMsgUnmarshalling, errData),
	}
}
