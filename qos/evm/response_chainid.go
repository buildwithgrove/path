package evm

import (
	"encoding/json"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseToChainID provides the functionality required from a response by a requestContext instance.
var _ response = responseToChainID{}

// TODO_IMPROVE(@adshmh): consider refactoring all unmarshallers to remove any duplicated logic.
// responseUnmarshallerChainID deserializes the provided byte slice into a responseToChainID struct,
// adding any encountered errors to the returned struct for constructing a response payload.
func responseUnmarshallerChainID(jsonrpcReq jsonrpc.Request, jsonrpcResp jsonrpc.Response, logger polylog.Logger) (response, error) {
	if jsonrpcResp.Error.Code != 0 { // The endpoint returned an error: no need to do further processing of the response.
		// Note: this assumes the `eth_chainId` request sent to the endpoint was valid.
		return responseToChainID{
			Response: jsonrpcResp,
			Logger:   logger,
		}, nil
	}

	resultBz, err := jsonrpcResp.GetResultAsBytes()
	if err != nil {
		return responseToChainID{
			Response: jsonrpcResp,
			Logger:   logger,
		}, err
	}

	var result string
	err = json.Unmarshal(resultBz, &result)

	return &responseToChainID{
		Response: jsonrpcResp,
		Result:   result,
	}, err
}

// responseToChainID captures the fields expected in a
// response to an `eth_chainId` request.
type responseToChainID struct {
	// Response stores the JSONRPC response parsed from an endpoint's response bytes.
	Response jsonrpc.Response
	// Result captures the `result` field of a JSONRPC resonse to an `eth_chainId` request.
	Result string
	Logger polylog.Logger
}

// GetObservation returns an observation using an `eth_chainId` request's response.
// This method implements the response interface.
func (r responseToChainID) GetObservation() qosobservations.EVMEndpointObservation {
	return qosobservations.EVMEndpointObservation{
		ResponseObservation: &qosobservations.EVMEndpointObservation_ChainIdResponse{
			ChainIdResponse: &qosobservations.EVMChainIDResponse{
				ChainIdResponse: r.Result,
			},
		},
	}
}

// TODO_UPNEXT(@adshmh): handle the following scenarios:
//  1. An endpoint returned a malformed, i.e. Not in JSONRPC format, response.
//     The user-facing response should include the request's ID.
//  2. An endpoint returns a JSONRPC response indicating a user error:
//     This should be returned to the user as-is.
//  3. An endpoint returns a valid JSONRPC response to a valid user request:
//     This should be returned to the user as-is.
//
// GetResponsePayload returns the raw byte slice payload to be returned as the response to the JSONRPC request.
// It implements the response interface.
func (r responseToChainID) GetResponsePayload() []byte {
	bz, err := json.Marshal(r.Response)
	if err != nil {
		// This should never happen: log an entry but return the response anyway.
		r.Logger.Warn().Err(err).Msg("responseToChainID: Marshalling JSONRPC response failed.")
	}
	return bz
}
