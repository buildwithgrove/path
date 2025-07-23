package cosmos

import (
	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/pokt-network/poktroll/pkg/polylog"
)

var (
	// CosmosSDK response IDs for different request types:
	// - REST-like responses: -1
	// TODO_NEXT(@adshmh): Use proper JSON-RPC ID response validation that works for all CosmosSDK chains.
	restLikeResponseID = jsonrpc.IDFromInt(-1)
)

// unmarshalRESTRequestEndpointResponse routes REST endpoint responses to appropriate validators.
// Uses the request path to match the response validator.
// It serves as the main entry point for processing REST responses.
// Always returns a valid response interface, never returns an error.
func unmarshalRESTRequestEndpointResponse(
	logger polylog.Logger,
	requestPath string,
	endpointResponseBz []byte,
) response {
	// Route to specific unmarshalers based on endpoint path
	switch requestPath {
	case "/status":
		// TODO_TECHDEBT(@adshmh): Refactor to properly separate response validation functionality shared between REST and JSONRPC.
		//
		// Build a JSONRPC request with the correct method and ID to use JSONRPC validator.
		// `/status` returns a JSONRPC response.
		jsonrpcReq := jsonrpc.Request{
			// Use -1 as Response
			ID:     restLikeResponseID,
			Method: "status",
		}
		return unmarshalJSONRPCRequestEndpointResponse(logger, jsonrpcReq, endpointResponseBz)

	case "/health":
		// TODO_TECHDEBT(@adshmh): Refactor to properly separate response validation functionality shared between REST and JSONRPC.
		//
		// Build a JSONRPC request with the correct method and ID to use JSONRPC validator.
		jsonrpcReq := jsonrpc.Request{
			// Use -1 as Response
			ID:     restLikeResponseID,
			Method: "health",
		}
		return unmarshalJSONRPCRequestEndpointResponse(logger, jsonrpcReq, endpointResponseBz)
	default:
		// For unrecognized endpoints, use the generic unmarshaler
		return responseRESTUnrecognized{
			logger:             logger,
			endpointResponseBz: endpointResponseBz,
		}
	}
}
