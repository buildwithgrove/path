package cometbft

import (
	"encoding/json"
	"net/http"

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

	return responseToHealth{
		logger:          logger,
		jsonRPCResponse: jsonrpcResp,
		healthy:         true,
	}, nil
}

// responseToHealth captures the fields expected in a
// response to an `eth_blockNumber` request.
type responseToHealth struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSONRPC response parsed from an endpoint's response bytes.
	jsonRPCResponse jsonrpc.Response

	// statusCode stores the status code of a response to a `/health` request.
	healthy bool
}

// GetObservation returns an observation using an `eth_blockNumber` request's response.
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

func (r responseToHealth) GetResponsePayload() []byte {
	// TODO_MVP(@adshmh): return a JSONRPC response indicating the error if unmarshalling failed.
	bz, err := json.Marshal(r.jsonRPCResponse)
	if err != nil {
		// This should never happen: log an entry but return the response anyway.
		r.logger.Warn().Err(err).Msg("responseToGetHealth: Marshaling JSONRPC response failed.")
	}
	return bz
}

// CometBFT always returns either a 500 (on error) or 200 (on success).
// Reference: https://docs.cometbft.com/v0.38/rpc/
func (r responseToHealth) GetResponseStatusCode() int {
	if r.jsonRPCResponse.IsError() {
		return http.StatusInternalServerError
	}
	return http.StatusOK
}
