package cometbft

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseToHealth provides the functionality required from a response by a requestContext instance.
var _ response = responseToHealth{}

// responseUnmarshallerHealth deserializes the provided payload
// into a responseToHealth struct, adding any encountered errors
// to the returned struct.
func responseUnmarshallerHealth(
	logger polylog.Logger,
	jsonrpcResp jsonrpc.Response,
) (response, error) {
	// The endpoint returned an error: no need to do further processing of the response.
	if jsonrpcResp.IsError() {
		return responseToHealth{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,
			healthy:         false,
		}, nil
	}

	// `/health` endpoint returns an empty response on success,
	// so any non-error response is considered healthy.
	return responseToHealth{
		logger:          logger,
		jsonRPCResponse: jsonrpcResp,
		healthy:         true,
	}, nil
}

// responseToHealth captures a CometBFT-based blockchain's /health endpoint response.
// Reference: https://docs.cometbft.com/v0.38/rpc/#/Info/health
type responseToHealth struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSON-RPC response parsed from an endpoint's response bytes.
	jsonRPCResponse jsonrpc.Response

	// statusCode stores the status code of a response to a `/health` request.
	healthy bool
}

// GetObservation returns a CometBFT-based /health observation
// Implements the response interface.
func (r responseToHealth) GetObservation() qosobservations.CometBFTEndpointObservation {
	return qosobservations.CometBFTEndpointObservation{
		ResponseObservation: &qosobservations.CometBFTEndpointObservation_HealthResponse{
			HealthResponse: &qosobservations.CometBFTHealthResponse{
				HealthStatusResponse: r.healthy,
			},
		},
	}
}

// Returns the parsed JSONRPC response.
// Implements the response interface.
func (r responseToHealth) GetJSONRPCResponse() jsonrpc.Response {
	return r.jsonRPCResponse
}
