package evm

import (
	"encoding/json"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseToChainID provides the functionality required from a response by a requestContext instance.
var _ response = responseToChainID{}

// TODO_TECHDEBT(@adshmh): consider refactoring all unmarshallers to remove any duplicated logic.
// responseUnmarshallerChainID deserializes the provided byte slice into a responseToChainID struct,
// adding any encountered errors to the returned struct for constructing a response payload.
func responseUnmarshallerChainID(
	logger polylog.Logger,
	jsonrpcReq jsonrpc.Request,
	jsonrpcResp jsonrpc.Response,
) (response, error) {
	// The endpoint returned an error: no need to do further processing of the response.
	if jsonrpcResp.Error.Code != 0 {
		// TODO_TECHDEBT(@adshmh): validate the `eth_chainId` request sent to the endpoint.
		return responseToChainID{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,
		}, nil
	}

	resultBz, err := jsonrpcResp.GetResultAsBytes()
	if err != nil {
		return responseToChainID{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,
		}, err
	}

	var result string
	err = json.Unmarshal(resultBz, &result)

	return &responseToChainID{
		logger:          logger,
		jsonRPCResponse: jsonrpcResp,
		result:          result,
	}, err
}

// responseToChainID captures the fields expected in a
// response to an `eth_chainId` request.
type responseToChainID struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSONRPC response parsed from an endpoint's response bytes.
	jsonRPCResponse jsonrpc.Response
	// result captures the `result` field of a JSONRPC resonse to an `eth_chainId` request.
	result string
}

// GetObservation returns an observation using an `eth_chainId` request's response.
// This method implements the response interface.
func (r responseToChainID) GetObservation() qosobservations.EVMEndpointObservation {
	return qosobservations.EVMEndpointObservation{
		ResponseObservation: &qosobservations.EVMEndpointObservation_ChainIdResponse{
			ChainIdResponse: &qosobservations.EVMChainIDResponse{
				ChainIdResponse: r.result,
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
	bz, err := json.Marshal(r.jsonRPCResponse)
	if err != nil {
		// This should never happen: log an entry but return the response anyway.
		r.logger.Warn().Err(err).Msg("responseToChainID: Marshalling JSONRPC response failed.")
	}
	return bz
}
