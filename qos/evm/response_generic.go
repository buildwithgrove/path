package evm

import (
	"encoding/json"

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

// TODO_MVP(@adshmh): implement the generic jsonrpc response
// (with the scope limited to an EVM-based blockchain)
// responseGeneric captures the fields expected in response to any request on an
// EVM-based blockchain. It is intended to be used when no validation/observation
// is applicable to the corresponding request's JSONRPC method.
// i.e. when there are no unmarshallers/structs matching the method specified by the request.
type responseGeneric struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSONRPC response parsed from an endpoint's response bytes.
	jsonRPCResponse jsonrpc.Response

	// valid is set to true if the parsed response is deemed valid.
	// As of PR #152, a response is deemed valid if it can be unmarshaled as a JSONRPC struct
	// regardless of the contents of the response.
	valid bool
}

// GetObservation returns an observation that is NOT used in validating endpoints.
// This allows sharing data with other entities, e.g. a data pipeline.
// Implements the response interface.
func (r responseGeneric) GetObservation() qosobservations.EVMEndpointObservation {
	return qosobservations.EVMEndpointObservation{
		ResponseObservation: &qosobservations.EVMEndpointObservation_UnrecognizedResponse{
			UnrecognizedResponse: &qosobservations.EVMUnrecognizedResponse{
				JsonrpcResponse: &qosobservations.JsonRpcResponse{
					Id: r.jsonRPCResponse.ID.String(),
				},
				Valid: r.valid,
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

// responseUnmarshallerGeneric unmarshal the provided byte slice
// into a responseGeneric struct and saves any data that may be
// needed for producing a response payload into the struct.
func responseUnmarshallerGeneric(logger polylog.Logger, jsonrpcReq jsonrpc.Request, data []byte) (response, error) {
	var response jsonrpc.Response
	err := json.Unmarshal(data, &response)
	if err != nil {
		return getGenericJSONRPCErrResponse(logger, jsonrpcReq.ID, data, err), nil
	}

	return responseGeneric{
		logger: logger,

		jsonRPCResponse: response,

		// The response is assumed valid if it can be successfully unmarshaled into a JSONRPC response struct.
		valid: true,
	}, nil
}

// getGenericJSONRPCErrResponse returns a generic response wrapped around a JSONRPC error response with the supplied ID, error, and the invalid payload in the "data" field.
func getGenericJSONRPCErrResponse(_ polylog.Logger, id jsonrpc.ID, malformedResponsePayload []byte, err error) responseGeneric {
	errData := map[string]string{
		errDataFieldRawBytes:         string(malformedResponsePayload),
		errDataFieldUnmarshallingErr: err.Error(),
	}

	return responseGeneric{
		jsonRPCResponse: jsonrpc.GetErrorResponse(id, errCodeUnmarshalling, errMsgUnmarshalling, errData),
	}
}

// TODO_INCOMPLETE: Handle the string `null`, as it could be returned
// when an object is expected.
// See the following link for more details:
// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_gettransactionbyhash
