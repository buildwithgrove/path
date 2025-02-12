package cometbft

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

type (
	// Result struct is the expected response from the `/status` endpoint.
	// Reference: https://docs.cometbft.com/v1.0/spec/rpc/#response-1
	Result struct {
		SyncInfo SyncInfo `json:"sync_info"`
	}
	// TODO_FUTURE: Extract more than just the `latest_block_height` field.
	SyncInfo struct {
		LatestBlockHeight string `json:"latest_block_height"`
	}
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

	// We only care about the SyncInfo.LatestBlockHeight field of
	// the JSON-RPC response, so first convert it from any to bytes.
	resultBytes, err := json.Marshal(jsonrpcResp.Result)
	if err != nil {
		return responseToStatus{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,
		}, fmt.Errorf("failed to marshal result: %w", err)
	}

	// Then unmarshal the JSON bytes into the Result struct.
	var result Result
	err = json.Unmarshal(resultBytes, &result)
	if err != nil {
		return responseToStatus{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,
		}, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return responseToStatus{
		logger:            logger,
		jsonRPCResponse:   jsonrpcResp,
		latestBlockHeight: result.SyncInfo.LatestBlockHeight,
	}, nil
}

// responseToStatus captures the fields expected in a
// response to a block height request.
type responseToStatus struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSON-RPC response parsed from an endpoint's response bytes.
	jsonRPCResponse jsonrpc.Response

	// latestBlockHeight stores the latest block height of a
	// response to a block height request as a string.
	latestBlockHeight string
}

// GetObservation returns an observation using a block height request's response.
// Implements the response interface.
func (r responseToStatus) GetObservation() qosobservations.CometBFTEndpointObservation {
	return qosobservations.CometBFTEndpointObservation{
		ResponseObservation: &qosobservations.CometBFTEndpointObservation_LatestBlockHeightResponse{
			LatestBlockHeightResponse: &qosobservations.CometBFTLatestBlockHeightResponse{
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

// CometBFT response codes:
// - 200: Success
// - 500: Error
// Reference: https://docs.cometbft.com/v0.38/rpc/
// Implements the response interface.
func (r responseToStatus) GetResponseStatusCode() int {
	if r.jsonRPCResponse.IsError() {
		return http.StatusInternalServerError
	}
	return http.StatusOK
}
