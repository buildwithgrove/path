package solana

import (
	"encoding/json"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseUnmarshallerGetHealth deserializes the provided payload
// into a responseToBlockNumber struct, adding any encountered errors
// to the returned struct.
func responseUnmarshallerGetHealth(logger polylog.Logger, jsonrpcReq jsonrpc.Request, jsonrpcResp jsonrpc.Response) (response, error) {
	// TODO_MVP(@adshmh): validate a `getHealth` request before sending it out to an endpoint.
	// e.g. If the request contains a params field, it is invalid and should not be sent to any endpoints.
	//
	// There are 2 possible valid responses to a `getHealth` request:
	// 1. A JSONRPC response with `result` field set to "ok".
	// 2. A JSONRPC response with an error field indicating unhealthy status for the endpoint.
	//
	// See the following link for more details:
	// https://solana.com/docs/rpc/http/gethealth
	// The endpoint returned an error: no need to do further processing of the response.
	if jsonrpcResp.IsError() {
		return responseToGetHealth{
			Logger: logger,

			Response: jsonrpcResp,
		}, nil
	}

	resultBz, err := jsonrpcResp.GetResultAsBytes()
	if err != nil {
		return responseToGetHealth{
			Logger: logger,

			Response: jsonrpcResp,
		}, err
	}

	var getHealthResult string
	err = json.Unmarshal(resultBz, &getHealthResult)

	return responseToGetHealth{
		Logger: logger,

		Response:     jsonrpcResp,
		HealthResult: getHealthResult,
	}, err
}

// responseToGetHealth captures the fields expected in a
// response to a `getHealth` request.
type responseToGetHealth struct {
	// Response stores the JSONRPC response parsed from an endpoint's response bytes.
	jsonrpc.Response

	// HealthResult stores the result field of a response to a `getHealth` request.
	HealthResult string

	Logger polylog.Logger
}

// GetObservation returns a Solana Endpoint observation based on an endpoint's response to a `getHealth` request.
// Implements the response interface used by the requestContext struct.
func (r responseToGetHealth) GetObservation() qosobservations.SolanaEndpointObservation {
	return qosobservations.SolanaEndpointObservation{
		ResponseObservation: &qosobservations.SolanaEndpointObservation_GetHealthResponse{
			GetHealthResponse: &qosobservations.SolanaGetHealthResponse{
				Result: r.HealthResult,
			},
		},
	}
}

// TODO_MVP(@adshmh): handle the following scenarios:
//  1. An endpoint returned a malformed, i.e. Not in JSONRPC format, response.
//     The user-facing response should include the request's ID.
//  2. An endpoint returns a JSONRPC response indicating a user error:
//     This should be returned to the user as-is.
//  3. An endpoint returns a valid JSONRPC response to a valid user request:
//     This should be returned to the user as-is.
func (r responseToGetHealth) GetResponsePayload() []byte {
	bz, err := json.Marshal(r.Response)
	if err != nil {
		// This should never happen: log an entry but return the response anyway.
		r.Logger.Warn().Err(err).Msg("responseToGetHealth: Marshaling JSONRPC response failed.")
	}
	return bz
}
