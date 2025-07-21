package cosmos

import (
	"encoding/json"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseToCometbftHealth provides the functionality required from a response by a requestContext instance.
var _ response = responseToCometbftHealth{}

// responseUnmarshallerCometbftHealth deserializes the provided payload
// into a responseToCometbftHealth struct, adding any encountered errors
// to the returned struct.
func responseUnmarshallerCometbftHealth(
	logger polylog.Logger,
	jsonrpcResp jsonrpc.Response,
	_ []byte,
) (response, error) {
	// The endpoint returned an error: no need to do further processing of the response.
	if jsonrpcResp.IsError() {
		return responseToCometbftHealth{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,
			healthy:         false,
		}, nil
	}

	// `/health` endpoint returns an empty response on success,
	// so any non-error response is considered healthy.
	return responseToCometbftHealth{
		logger:          logger,
		jsonRPCResponse: jsonrpcResp,
		healthy:         true,
	}, nil
}

// responseToCometbftHealth captures a CometBFT-based blockchain's /health endpoint response.
// Reference: https://docs.cometbft.com/v1.0/spec/rpc/#health
type responseToCometbftHealth struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSON-RPC response parsed from an endpoint's response bytes.
	jsonRPCResponse jsonrpc.Response

	// statusCode stores the status code of a response to a `/health` request.
	healthy bool
}

// GetObservation returns a CosmosSDK-based /health observation
// Implements the response interface.
func (r responseToCometbftHealth) GetObservation() qosobservations.CosmosSDKEndpointObservation {
	return qosobservations.CosmosSDKEndpointObservation{
		ResponseObservation: &qosobservations.CosmosSDKEndpointObservation_CometbftHealthResponse{
			CometbftHealthResponse: &qosobservations.CometBFTHealthResponse{
				HealthStatusResponse: r.healthy,
			},
		},
	}
}

// GetResponsePayload returns the payload for the response to a `/health` request.
// Implements the response interface.
func (r responseToCometbftHealth) GetResponsePayload() []byte {
	// TODO_MVP(@adshmh): return a JSON-RPC response indicating the error if unmarshaling failed.
	bz, err := json.Marshal(r.jsonRPCResponse)
	if err != nil {
		// This should never happen: log an entry but return the response anyway.
		r.logger.Warn().Err(err).Msg("responseToGetHealth: Marshaling JSON-RPC response failed.")
	}
	return bz
}

// returns an HTTP status code corresponding to the underlying JSON-RPC response code.
// DEV_NOTE: This is an opinionated mapping following best practice but not enforced by any specifications or standards.
// Implements the response interface.
func (r responseToCometbftHealth) GetResponseStatusCode() int {
	return r.jsonRPCResponse.GetRecommendedHTTPStatusCode()
}

// GetHTTPResponse builds and returns the httpResponse matching the responseToCometbftHealth instance.
// Implements the response interface.
func (r responseToCometbftHealth) GetHTTPResponse() httpResponse {
	return httpResponse{
		responsePayload: r.GetResponsePayload(),
		httpStatusCode:  r.GetResponseStatusCode(),
	}
}
