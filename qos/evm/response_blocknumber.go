package evm

import (
	"encoding/json"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseToBlockNumber provides the functionality required from a response by a requestContext instance.
var _ response = responseToBlockNumber{}

// responseUnmarshallerBlockNumber deserializes the provided payload
// into a responseToBlockNumber struct, adding any encountered errors
// to the returned struct.
func responseUnmarshallerBlockNumber(
	logger polylog.Logger,
	jsonrpcReq jsonrpc.Request,
	jsonrpcResp jsonrpc.Response,
) (response, error) {
	// The endpoint returned an error: no need to do further processing of the response.
	if jsonrpcResp.IsError() {
		// TODO_TECHDEBT(@adshmh): validate the `eth_blockNumber`
		// request that was sent to the endpoint.
		return responseToBlockNumber{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,
		}, nil
	}

	resultBz, err := jsonrpcResp.GetResultAsBytes()
	if err != nil {
		return responseToBlockNumber{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,
		}, err
	}

	var result string
	err = json.Unmarshal(resultBz, &result)

	return responseToBlockNumber{
		logger:          logger,
		jsonRPCResponse: jsonrpcResp,
		result:          result,
	}, err
}

// responseToBlockNumber captures the fields expected in a
// response to an `eth_blockNumber` request.
type responseToBlockNumber struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSONRPC response parsed from an endpoint's response bytes.
	jsonRPCResponse jsonrpc.Response

	// result stores the result field of a response to a `eth_blockNumber` request.
	result string
}

// GetObservation returns an observation using an `eth_blockNumber` request's response.
// Implements the response interface.
// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_blocknumber
func (r responseToBlockNumber) GetObservation() qosobservations.EVMEndpointObservation {
	return qosobservations.EVMEndpointObservation{
		ResponseObservation: &qosobservations.EVMEndpointObservation_BlockNumberResponse{
			BlockNumberResponse: &qosobservations.EVMBlockNumberResponse{
				BlockNumberResponse: r.result,
			},
		},
	}
}

func (r responseToBlockNumber) GetResponsePayload() []byte {
	// TODO_MVP(@adshmh): return a JSONRPC response indicating the error if unmarshaling failed.
	bz, err := json.Marshal(r.jsonRPCResponse)
	if err != nil {
		// This should never happen: log an entry but return the response anyway.
		r.logger.Warn().Err(err).Msg("responseToGetHealth: Marshaling JSON-RPC response failed.")
	}
	return bz
}
