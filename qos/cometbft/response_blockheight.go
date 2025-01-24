package cometbft

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// The Result struct is the expected response from the `/status` endpoint.
// We are only interested in the `latest_block_height` field.
// Reference: https://docs.cometbft.com/v1.0/spec/rpc/#response-1
type (
	Result struct {
		SyncInfo SyncInfo `json:"sync_info"`
	}
	SyncInfo struct {
		LatestBlockHeight string `json:"latest_block_height"`
	}
)

// responseToStatus provides the functionality required from a response by a requestContext instance.
var _ response = responseToStatus{}

// responseUnmarshallerBlockHeight deserializes the provided payloadxz
// into a responseToStatus struct, adding any encountered errors
// to the returned struct.
func responseUnmarshallerBlockHeight(
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
	// the JSONRPC response, so first convert it from any to bytes.
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

	// jsonRPCResponse stores the JSONRPC response parsed from an endpoint's response bytes.
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

func (r responseToStatus) GetResponsePayload() []byte {
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
func (r responseToStatus) GetResponseStatusCode() int {
	if r.jsonRPCResponse.IsError() {
		return http.StatusInternalServerError
	}
	return http.StatusOK
}
