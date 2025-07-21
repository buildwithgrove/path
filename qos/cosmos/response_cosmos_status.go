package cosmos

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseToCosmosStatus provides the functionality required from a response by a requestContext instance.
var _ response = responseToCosmosStatus{}

// responseUnmarshallerCosmosStatus deserializes the provided payload
// into a responseToCosmosStatus struct, adding any encountered errors
// to the returned struct.
func responseUnmarshallerCosmosStatus(
	logger polylog.Logger,
	_ jsonrpc.Response,
	restResponse []byte,
) (response, error) {
	// Then unmarshal the JSON bytes into the node.StatusResponse struct
	var result node.StatusResponse
	if err := json.Unmarshal(restResponse, &result); err != nil {
		return responseToCosmosStatus{
			logger:       logger,
			restResponse: restResponse,
		}, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return responseToCosmosStatus{
		logger:       logger,
		restResponse: restResponse,
		height:       result.Height,
	}, nil
}

// responseToCosmosStatus captures the fields expected in a
// response to a Cosmos SDK status request.
type responseToCosmosStatus struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSON-RPC response parsed from an endpoint's response bytes.
	restResponse []byte

	// height stores the latest block height of a
	// response to a Cosmos SDK status request as a string.
	// Comes from the `height` field in the `/cosmos/base/node/v1beta1/status` response.
	// Reference: https://docs.cosmos.network/main/core/grpc_rest.html#status
	height uint64
}

// GetObservation returns an observation using a Cosmos SDK status request's response.
// Implements the response interface.
func (r responseToCosmosStatus) GetObservation() qosobservations.CosmosSDKEndpointObservation {
	return qosobservations.CosmosSDKEndpointObservation{
		ResponseObservation: &qosobservations.CosmosSDKEndpointObservation_CosmosStatusResponse{
			CosmosStatusResponse: &qosobservations.CosmosSDKStatusResponse{
				LatestBlockHeightResponse: r.height,
			},
		},
	}
}

// GetResponsePayload returns the payload for the response to a `/cosmos/base/node/v1beta1/status` request.
// Implements the response interface.
func (r responseToCosmosStatus) GetResponsePayload() []byte {
	return r.restResponse
}

// returns an HTTP status code corresponding to the underlying JSON-RPC response code.
// DEV_NOTE: This is an opinionated mapping following best practice but not enforced by any specifications or standards.
// Implements the response interface.
func (r responseToCosmosStatus) GetResponseStatusCode() int {
	return http.StatusOK
}

// GetHTTPResponse builds and returns the httpResponse matching the responseToCosmosStatus instance.
// Implements the response interface.
func (r responseToCosmosStatus) GetHTTPResponse() httpResponse {
	return httpResponse{
		responsePayload: r.GetResponsePayload(),
		httpStatusCode:  r.GetResponseStatusCode(),
	}
}
