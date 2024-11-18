package solana

import (
	"encoding/json"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseUnmarshallerGetHealth deserializes the provided payload
// into a responseToBlockNumber struct, adding any encountered errors
// to the returned struct.
func responseUnmarshallerGetHealth(jsonrpcReq jsonrpc.Request, jsonrpcResp jsonrpc.Response, logger polylog.Logger) (response, error) {
	// TODO_UPNEXT(@adshmh): validate a `getHealth` request before sending it out to an endpoint.
	// e.g. If the request contains a params field, it is invalid and should not be sent to any endpoints.
	//
	// There are 2 possible valid responses to a `getHealth` request:
	// 1. A JSONRPC response with `result` field set to "ok".
	// 2. A JSONRPC response with an error field indicating unhealthy status for the endpoint.
	//
	// See the following link for more details:
	// https://solana.com/docs/rpc/http/gethealth
	if jsonrpcResp.Error.Code != 0 { // The endpoint returned an error: no need to do further processing of the response.
		// Note: this assumes the `getHealth` request sent to the endpoint was valid.
		return responseToGetHealth{
			Response: jsonrpcResp,
			Logger:   logger,
		}, nil
	}

	resultBz, err := jsonrpcResp.GetResultAsBytes()
	if err != nil {
		return responseToGetHealth{
			Response: jsonrpcResp,
			Logger:   logger,
		}, err
	}

	var getHealthResult string
	err = json.Unmarshal(resultBz, &getHealthResult)

	return responseToGetHealth{
		Response:     jsonrpcResp,
		HealthResult: getHealthResult,
		Logger:       logger,
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
// This method implements the response interface used by the requestContext struct. 
func (r responseToGetHealth) GetObservation() observation.qos.SolanaEndpointDetails {
	return observation.qos.SolanaEndpointDetails{
		HealthResult: &r.HealthResult,
	}
}

// TODO_UPNEXT(@adshmh): handle the following scenarios:
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
		r.Logger.Warn().Err(err).Msg("responseToGetHealth: Marshalling JSONRPC response failed.")
	}
	return bz
}
