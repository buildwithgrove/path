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
func responseUnmarshallerBlockNumber(jsonrpcReq jsonrpc.Request, jsonrpcResp jsonrpc.Response, logger polylog.Logger) (response, error) {
	// The endpoint returned an error: no need to do further processing of the response.
	if jsonrpcResp.Error.Code != 0 {
		// Note: this assumes the `eth_blockNumber` request sent to the endpoint was valid.
		return responseToBlockNumber{
			Response: jsonrpcResp,
			Logger:   logger,
		}, nil
	}

	resultBz, err := jsonrpcResp.GetResultAsBytes()
	if err != nil {
		return responseToBlockNumber{
			Response: jsonrpcResp,
			Logger:   logger,
		}, err
	}

	var result string
	err = json.Unmarshal(resultBz, &result)

	return responseToBlockNumber{
		Response: jsonrpcResp,
		Result:   result,
		Logger:   logger,
	}, err
}

// responseToBlockNumber captures the fields expected in a
// response to an `eth_blockNumber` request.
type responseToBlockNumber struct {
	// Response stores the JSONRPC response parsed from an endpoint's response bytes.
	jsonrpc.Response

	// Result stores the result field of a response to a `eth_blockNumber` request.
	Result string

	Logger polylog.Logger
}

// GetObservation returns an observation using an `eth_blockNumber` request's response.
// This method implements the response interface.
func (r responseToBlockNumber) GetObservation() qosobservations.EVMEndpointObservation {
	return qosobservations.EVMEndpointObservation{
		ResponseObservation: &qosobservations.EVMEndpointObservation_BlockNumberResponse{
			BlockNumberResponse: &qosobservations.EVMBlockNumberResponse{
				BlockNumberResponse: r.Result,
			},
		},
	}
}

func (r responseToBlockNumber) GetResponsePayload() []byte {
	// TODO_UPNEXT(@adshmh): return a JSONRPC response indicating the error,
	// if the unmarshalling failed.
	bz, err := json.Marshal(r.Response)
	if err != nil {
		// This should never happen: log an entry but return the response anyway.
		r.Logger.Warn().Err(err).Msg("responseToGetHealth: Marshalling JSONRPC response failed.")
	}
	return bz
}
