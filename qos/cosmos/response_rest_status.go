package cosmos

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseRESTStatus provides the functionality required from a response by a requestContext instance
var _ response = responseRESTStatus{}

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

// responseUnmarshalerRESTStatus deserializes the provided payload
// into a responseRESTStatus struct, handling JSON-RPC responses from /status endpoint
// Always returns a valid response interface, never returns an error.
func responseUnmarshalerRESTStatus(
	logger polylog.Logger,
	data []byte,
) response {
	logger = logger.With("response_processor", "status")

	// Handle empty responses
	if len(data) == 0 {
		logger.Error().
			Str("endpoint", "/status").
			Msg("Received empty JSON-RPC response from /status endpoint")

		return getRESTStatusEmptyErrorResponse(logger)
	}

	// Unmarshal as JSON-RPC response since /status returns JSON-RPC
	var jsonrpcResponse jsonrpc.Response
	if err := json.Unmarshal(data, &jsonrpcResponse); err != nil {
		logger.Error().
			Err(err).
			Str("raw_payload", string(data)).
			Msg("Failed to unmarshal /status response as JSON-RPC")

		return getRESTStatusUnmarshalErrorResponse(logger, err)
	}

	// The endpoint returned an error: no need to do further processing of the response
	if jsonrpcResponse.IsError() {
		logger.Warn().
			Str("jsonrpc_error", jsonrpcResponse.Error.Message).
			Int("jsonrpc_error_code", jsonrpcResponse.Error.Code).
			Msg("Endpoint returned JSON-RPC error for /status request")

		return responseRESTStatus{
			logger:          logger,
			jsonRPCResponse: jsonrpcResponse,
		}
	}

	// Marshal the result to parse it as ResultStatus
	resultBytes, err := json.Marshal(jsonrpcResponse.Result)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to marshal JSON-RPC result for /status")

		return getRESTStatusUnmarshalErrorResponse(logger, err)
	}

	// Then unmarshal the JSON bytes into the ResultStatus struct
	var result ResultStatus
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		logger.Error().
			Err(err).
			Str("result_data", string(resultBytes)).
			Msg("Failed to unmarshal JSON-RPC result into ResultStatus structure")

		return getRESTStatusUnmarshalErrorResponse(logger, err)
	}

	logger.Debug().
		Str("chain_id", result.NodeInfo.Network).
		Bool("catching_up", result.SyncInfo.CatchingUp).
		Str("latest_block_height", result.SyncInfo.LatestBlockHeight).
		Msg("Successfully parsed /status response")

	return responseRESTStatus{
		logger:            logger,
		jsonRPCResponse:   jsonrpcResponse,
		chainID:           result.NodeInfo.Network,
		catchingUp:        result.SyncInfo.CatchingUp,
		latestBlockHeight: result.SyncInfo.LatestBlockHeight,
		validationError:   nil, // No validation error for successfully unmarshaled responses
	}
}

// responseRESTStatus captures the fields expected in a
// response to a /status request (which returns JSON-RPC)
type responseRESTStatus struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSON-RPC response parsed from an endpoint's response bytes
	jsonRPCResponse jsonrpc.Response

	// chainID stores the chain ID of the endpoint
	// Comes from the `NodeInfo.Network` field in the `/status` response
	// Reference: https://docs.cometbft.com/v1.0/spec/rpc/#status
	chainID string

	// catchingUp indicates if the endpoint is catching up to the network
	// Comes from the `SyncInfo.CatchingUp` field in the `/status` response
	// Reference: https://docs.cometbft.com/v1.0/spec/rpc/#status
	catchingUp bool

	// latestBlockHeight stores the latest block height of a
	// response to a block height request as a string
	// Comes from the `SyncInfo.LatestBlockHeight` field in the `/status` response
	// Reference: https://docs.cometbft.com/v1.0/spec/rpc/#status
	latestBlockHeight string

	// validationError tracks any validation issues with the response
	validationError *qosobservations.CosmosSDKResponseValidationError
}

// GetObservation returns an observation using a /status request's response
// Implements the response interface
func (r responseRESTStatus) GetObservation() qosobservations.CosmosSDKEndpointObservation {
	return qosobservations.CosmosSDKEndpointObservation{
		ResponseObservation: &qosobservations.CosmosSDKEndpointObservation_RestObservation{
			RestObservation: &qosobservations.CosmosSDKEndpointRestObservation{
				ParsedResponse: &qosobservations.CosmosSDKEndpointRestObservation_StatusResponse{
					StatusResponse: &qosobservations.CosmosSDKRESTStatusResponse{
						HttpStatusCode:            int32(r.GetResponseStatusCode()),
						ChainIdResponse:           r.chainID,
						CatchingUpResponse:        r.catchingUp,
						LatestBlockHeightResponse: r.latestBlockHeight,
					},
				},
			},
		},
	}
}

// GetResponsePayload returns the payload for the response to a `/status` request
// Implements the response interface
func (r responseRESTStatus) GetResponsePayload() []byte {
	return r.getResponsePayload()
}

// getResponsePayload returns the JSON-RPC response as bytes
func (r responseRESTStatus) getResponsePayload() []byte {
	responseBytes, _ := json.Marshal(r.jsonRPCResponse)
	return responseBytes
}

// GetResponseStatusCode returns an HTTP status code corresponding to the underlying JSON-RPC response code
// DEV_NOTE: This is an opinionated mapping following best practice but not enforced by any specifications or standards
// Implements the response interface
func (r responseRESTStatus) GetResponseStatusCode() int {
	// If we have a validation error, return 500
	if r.validationError != nil {
		return http.StatusInternalServerError
	}

	// Use JSON-RPC response's recommended status code
	return r.jsonRPCResponse.GetRecommendedHTTPStatusCode()
}

// GetHTTPResponse builds and returns the httpResponse matching the responseRESTStatus instance
// Implements the response interface
func (r responseRESTStatus) GetHTTPResponse() httpResponse {
	return httpResponse{
		responsePayload: r.GetResponsePayload(),
		httpStatusCode:  r.GetResponseStatusCode(),
	}
}

// getRESTStatusEmptyErrorResponse creates an error response for empty /status responses
func getRESTStatusEmptyErrorResponse(logger polylog.Logger) responseRESTStatus {
	errorResp := jsonrpc.GetErrorResponse(
		jsonrpc.IDFromInt(-1), // Use -1 for unknown ID
		errCodeJSONRPCEmptyResponse,
		"the /status endpoint returned an empty response",
		nil,
	)

	validationError := qosobservations.CosmosSDKResponseValidationError_COSMOS_SDK_RESPONSE_VALIDATION_ERROR_EMPTY

	return responseRESTStatus{
		logger:          logger,
		jsonRPCResponse: errorResp,
		validationError: &validationError,
	}
}

// getRESTStatusUnmarshalErrorResponse creates an error response for /status unmarshaling failures
func getRESTStatusUnmarshalErrorResponse(
	logger polylog.Logger,
	err error,
) responseRESTStatus {
	errData := map[string]string{
		errDataFieldJSONRPCUnmarshalingErr: err.Error(),
	}

	errorResp := jsonrpc.GetErrorResponse(
		jsonrpc.IDFromInt(-1), // Use -1 for unknown ID
		errCodeJSONRPCUnmarshaling,
		"Failed to parse /status response as JSON-RPC",
		errData,
	)

	validationError := qosobservations.CosmosSDKResponseValidationError_COSMOS_SDK_RESPONSE_VALIDATION_ERROR_UNMARSHAL

	return responseRESTStatus{
		logger:          logger,
		jsonRPCResponse: errorResp,
		validationError: &validationError,
	}
}
