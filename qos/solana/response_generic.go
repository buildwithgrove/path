package solana

import (
	"encoding/json"

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

// responseUnmarshallerGeneric processes raw response data into a responseGeneric struct.
// It extracts and stores any data needed for generating a response payload.
func responseUnmarshallerGeneric(
	logger polylog.Logger,
	jsonrpcReq jsonrpc.Request,
	data []byte,
) (response, error) {
	var response jsonrpc.Response
	if err := json.Unmarshal(data, &response); err != nil {
		return getGenericJSONRPCErrResponse(logger, jsonrpcReq.ID, data, err), nil
	}

	return responseGeneric{
		Logger:   logger,
		Response: response,
	}, nil
}

// TODO_MVP(@adshmh): implement the generic jsonrpc response
// (with the scope limited to the Solana blockchain)
// responseGeneric captures the fields expected in response to any request on the Solana blockchain.
// It is intended to be used when no validation/observation is applicable to the corresponding request's JSON-RPC method.
// i.e. when there are no unmarshallers/structs matching the method specified by the request.
type responseGeneric struct {
	Logger polylog.Logger
	jsonrpc.Response
}

// GetObservation returns an observation that is NOT used in validating endpoints.
// This allows sharing data with other entities, e.g. a data pipeline.
// Implements the response interface.
// As of PR 372, this is a default catchall for any response to any requests other than `getHealth` and `getEpochInfo`.
func (r responseGeneric) GetObservation() qosobservations.SolanaEndpointObservation {
	return qosobservations.SolanaEndpointObservation{
		// TODO_TECHDEBT(@adshmh): set additional JSON-RPC response fields, specifically the `error` object, on the observation.
		// This needs a utility function to convert a `qos.jsonrpc.Response` to an `observation.qos.JsonRpcResponse.
		ResponseObservation: &qosobservations.SolanaEndpointObservation_UnrecognizedResponse{
			UnrecognizedResponse: &qosobservations.SolanaUnrecognizedResponse{
				JsonrpcResponse: &qosobservations.JsonRpcResponse{
					Id: r.Response.ID.String(),
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
	bz, err := json.Marshal(r.Response)
	if err != nil {
		// This should never happen: log an entry but return the response anyway.
		r.Logger.Warn().Err(err).Msg("responseGeneric: Marshaling JSON-RPC response failed.")
	}
	return bz
}

// getGenericJSONRPCErrResponse returns a generic response wrapped around a JSON-RPC error response with the supplied ID, error, and the invalid payload in the "data" field.
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

	return responseGeneric{
		Response: jsonrpc.GetErrorResponse(id, errCodeUnmarshaling, errMsgUnmarshaling, errData),
	}
}
