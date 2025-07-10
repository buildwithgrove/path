package cometbft

import (
	"encoding/json"
	"fmt"
	"strconv"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseToStatus provides the functionality required from a response by a requestContext instance.
var _ response = responseToStatus{}

// responseUnmarshallerStatus deserializes the provided payloadxz
// into a responseToStatus struct, adding any encountered errors
// to the returned struct.
func responseUnmarshallerStatus(
	logger polylog.Logger,
	jsonrpcResp jsonrpc.Response,
) (response, error) {
	// The endpoint returned an error: no need to do further processing of the response.
	if jsonrpcResp.IsError() {
		return responseToStatus{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,
		}, nil
	}

	resultBytes, err := json.Marshal(jsonrpcResp.Result)
	if err != nil {
		return responseToStatus{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,
		}, fmt.Errorf("failed to marshal result: %w", err)
	}

	// Then unmarshal the JSON bytes into the ResultStatus struct
	// from the CometBFT's `coretypes` package.
	// Reference: https://github.com/cometbft/cometbft/blob/4226b0ea6ab4725ef807a16b86d6d24835bb45d4/rpc/core/types/responses.go#L100
	var result coretypes.ResultStatus
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		return responseToStatus{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,
		}, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return responseToStatus{
		logger:            logger,
		jsonRPCResponse:   jsonrpcResp,
		chainID:           result.NodeInfo.Network,
		catchingUp:        result.SyncInfo.CatchingUp,
		latestBlockHeight: strconv.FormatInt(result.SyncInfo.LatestBlockHeight, 10),
	}, nil
}

// responseToStatus captures the fields expected in a
// response to a block height request.
type responseToStatus struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSON-RPC response parsed from an endpoint's response bytes.
	jsonRPCResponse jsonrpc.Response

	// chainID stores the chain ID of the endpoint.
	// Comes from the `NodeInfo.Network` field in the `/status` response.
	// Reference: https://docs.cometbft.com/v1.0/spec/rpc/#status
	chainID string

	// catchingUp indicates if the endpoint is catching up to the network.
	// Comes from the `SyncInfo.CatchingUp` field in the `/status` response.
	// Reference: https://docs.cometbft.com/v1.0/spec/rpc/#status
	catchingUp bool

	// latestBlockHeight stores the latest block height of a
	// response to a block height request as a string.
	// Comes from the `SyncInfo.LatestBlockHeight` field in the `/status` response.
	// Reference: https://docs.cometbft.com/v1.0/spec/rpc/#status
	latestBlockHeight string
}

// GetObservation returns an observation using a block height request's response.
// Implements the response interface.
func (r responseToStatus) GetObservation() qosobservations.CometBFTEndpointObservation {
	return qosobservations.CometBFTEndpointObservation{
		ResponseObservation: &qosobservations.CometBFTEndpointObservation_StatusResponse{
			StatusResponse: &qosobservations.CometBFTStatusResponse{
				ChainIdResponse:           r.chainID,
				CatchingUpResponse:        r.catchingUp,
				LatestBlockHeightResponse: r.latestBlockHeight,
			},
		},
	}
}

// GetResponsePayload returns the payload for the response to a `/status` request.
// Implements the response interface.
func (r responseToStatus) GetResponsePayload() []byte {
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
func (r responseToStatus) GetResponseStatusCode() int {
	return r.jsonRPCResponse.GetRecommendedHTTPStatusCode()
}
