package solana

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
		Logger: logger,

		Response: response,
	}, nil
}

// TODO_MVP(@adshmh): implement the generic jsonrpc response
// (with the scope limited to the Solana blockchain)
// responseGeneric captures the fields expected in response to any request on the Solana blockchain.
// It is intended to be used when no validation/observation is applicable to the corresponding request's JSONRPC method.
// i.e. when there are no unmarshallers/structs matching the method specified by the request.
type responseGeneric struct {
	Logger polylog.Logger
	jsonrpc.Response
}

// GetObservation on a generic response returns an observation not utilized for any endpoint validations.
// As of PR 372, this is a default catchall for any response to any requests other than `getHealth` and `getEpochInfo`.
// GetObservation implements the response interface used by the requestContext struct.
func (r responseGeneric) GetObservation() qosobservations.SolanaEndpointObservation {
	return qosobservations.SolanaEndpointObservation{
		// TODO_TECHDEBT(@adshmh): set additional JSONRPC response fields, specifically the `error` object, on the observation.
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

func (r responseGeneric) GetResponsePayload() []byte {
	bz, err := json.Marshal(r.Response)
	if err != nil {
		// This should never happen: log an entry but return the response anyway.
		r.Logger.Warn().Err(err).Msg("responseGeneric: Marshalling JSONRPC response failed.")
	}
	return bz
}

// getGenericJSONRPCErrResponse returns a generic response wrapped around a JSONRPC error response with the supplied ID, error, and the invalid payload in the "data" field.
func getGenericJSONRPCErrResponse(logger polylog.Logger, id jsonrpc.ID, malformedResponsePayload []byte, err error) responseGeneric {
	errData := map[string]string{
		errDataFieldRawBytes:         string(malformedResponsePayload),
		errDataFieldUnmarshallingErr: err.Error(),
	}

	return responseGeneric{
		Response: jsonrpc.GetErrorResponse(id, errCodeUnmarshalling, errMsgUnmarshalling, errData),
	}
}
