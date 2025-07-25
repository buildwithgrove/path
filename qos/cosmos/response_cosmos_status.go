package cosmos

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
)

// TODO_IMPROVE(@commoddity): The actual Cosmos SDK `node.StatusResponse` struct
// causes an unmarshalling error due to type mismatch in all fields.
//
// The following struct is a workaround to fix the unmarshalling error.
//
// The node returns all fields as strings, but the `node.StatusResponse` struct
// expects all fields to be of type int64.
//
// Update to use the Cosmos SDK `node.StatusResponse` struct once the issue is fixed.
// Reference: https://github.com/cosmos/cosmos-sdk/blob/main/client/grpc/node/query.pb.go#L180
//
// The `cosmosStatusResponse` struct is a workaround to fix the unmarshalling error.

// cosmosStatusResponse is the expected response from the /cosmos/base/node/v1beta1/status endpoint.
// Only the `height` field is needed to satisfy the `/status` endpoint checks.
//
// Reference: https://docs.cosmos.network/api#tag/Service/operation/Status
type cosmosStatusResponse struct {
	Height string `json:"height"`
}

// responseCosmosStatus provides the functionality required from a response by a requestContext instance.
var _ response = responseCosmosStatus{}

// responseUnmarshallerCosmosStatus deserializes the provided payload
// into a responseCosmosStatus struct, adding any encountered errors
// to the returned struct.
func responseValidatorCosmosStatus(
	logger polylog.Logger,
	restResponse []byte,
) (response, error) {
	// Then unmarshal the JSON bytes into the node.StatusResponse struct
	var result cosmosStatusResponse
	if err := json.Unmarshal(restResponse, &result); err != nil {
		return responseCosmosStatus{
			logger:       logger,
			restResponse: restResponse,
		}, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	heightInt, err := strconv.Atoi(result.Height)
	if err != nil {
		return responseCosmosStatus{
			logger:       logger,
			restResponse: restResponse,
		}, fmt.Errorf("failed to convert height to int: %w", err)
	}

	return responseCosmosStatus{
		logger:       logger,
		restResponse: restResponse,
		height:       uint64(heightInt),
	}, nil
}

// responseCosmosStatus captures the fields expected in a
// response to a Cosmos SDK status request.
type responseCosmosStatus struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSON-RPC response parsed from an endpoint's response bytes.
	restResponse []byte

	// height stores the latest block height of a
	// response to a Cosmos SDK status request as a string.
	// Comes from the `height` field in the `/cosmos/base/node/v1beta1/status` response.
	// Reference: https://docs.cosmos.network/main/core/grpc_rest.html#status
	height uint64
}

// GetObservation returns an observation using a /status request's response
// Implements the response interface
func (r responseCosmosStatus) GetObservation() qosobservations.CosmosEndpointObservation {
	return qosobservations.CosmosEndpointObservation{
		EndpointResponseValidationResult: &qosobservations.CosmosEndpointResponseValidationResult{
			ResponseValidationType: qosobservations.CosmosResponseValidationType_COSMOS_RESPONSE_VALIDATION_TYPE_JSON,
			HttpStatusCode:         int32(r.GetResponseStatusCode()),
			ValidationError:        nil, // No validation error for successfully processed responses
			ParsedResponse: &qosobservations.CosmosEndpointResponseValidationResult_ResponseCosmosSdkStatus{
				ResponseCosmosSdkStatus: &qosobservations.CosmosResponseCosmosSDKStatus{
					LatestBlockHeight: r.height,
				},
			},
		},
	}
}

// GetResponsePayload returns the payload for the response to a `/cosmos/base/node/v1beta1/status` request.
// Implements the response interface.
func (r responseCosmosStatus) GetResponsePayload() []byte {
	return r.restResponse
}

// returns an HTTP status code corresponding to the underlying JSON-RPC response code.
// DEV_NOTE: This is an opinionated mapping following best practice but not enforced by any specifications or standards.
// Implements the response interface.
func (r responseCosmosStatus) GetResponseStatusCode() int {
	return http.StatusOK
}

// GetHTTPResponse builds and returns the httpResponse matching the responseCosmosStatus instance.
// Implements the response interface.
func (r responseCosmosStatus) GetHTTPResponse() gateway.HTTPResponse {
	return httpResponse{
		responsePayload: r.GetResponsePayload(),
		httpStatusCode:  r.GetResponseStatusCode(),
	}
}
