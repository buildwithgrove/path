package cosmos

import (
	"encoding/json"
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseToStatus provides the functionality required from a response by a requestContext instance.
var _ response = responseToStatus{}

// TODO_IMPROVE(@commoddity): The actual `coretypes.ResultStatus` struct causes
// an unmarshalling error due to type mismatch in a number of fields:
//   - Node returns string values for the following required field:
//   - `sync_info.latest_block_height`
//   - The `coretypes.ResultStatus` struct expects this field to be int64.
//   - Many other non-required fields are also of the wrong type and will
//     cause an unmarshalling error if the `coretypes.ResultStatus` struct is used.
//
// Update to use the CometBFT `coretypes.ResultStatus` struct once the issue is fixed.
//
// The following structs are a workaround to fix the unmarshalling error.
//
// These structs represent the subset of the JSON data from the CometBFT `ResultStatus` struct
// needed to satisfy the `/status` endpoint checks.
//
// Reference: https://github.com/cometbft/cometbft/blob/4226b0ea6ab4725ef807a16b86d6d24835bb45d4/rpc/core/types/responses.go#L100
type (
	// Node Status
	ResultStatus struct {
		NodeInfo DefaultNodeInfo `json:"node_info"`
		SyncInfo SyncInfo        `json:"sync_info"`
	}

	// Info about the node's syncing state
	SyncInfo struct {
		LatestBlockHeight string `json:"latest_block_height"`
		CatchingUp        bool   `json:"catching_up"`
	}

	// DefaultNodeInfo is the basic node information exchanged
	// between two peers during the CometBFT P2P handshake.
	DefaultNodeInfo struct {
		Network string `json:"network"` // network/chain ID
	}
)

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
	var result ResultStatus
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		return responseToStatus{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,
		}, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	here := responseToStatus{
		logger:            logger,
		jsonRPCResponse:   jsonrpcResp,
		chainID:           result.NodeInfo.Network,
		catchingUp:        result.SyncInfo.CatchingUp,
		latestBlockHeight: result.SyncInfo.LatestBlockHeight,
	}

	return here, nil
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
func (r responseToStatus) GetObservation() qosobservations.CosmosSDKEndpointObservation {
	return qosobservations.CosmosSDKEndpointObservation{
		ResponseObservation: &qosobservations.CosmosSDKEndpointObservation_StatusResponse{
			StatusResponse: &qosobservations.CosmosSDKStatusResponse{
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

// GetHTTPResponse builds and returns the httpResponse matching the responseToStatus instance.
// Implements the response interface.
func (r responseToStatus) GetHTTPResponse() httpResponse {
	return httpResponse{
		responsePayload: r.GetResponsePayload(),
		httpStatusCode:  r.GetResponseStatusCode(),
	}
}
