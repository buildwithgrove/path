package cosmos

import (
	"net/http"
	"strings"

	"github.com/pokt-network/poktroll/pkg/polylog"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
)

// cosmosSDKRequestValidator handles request validation for CosmosSDK chains by:
// 1. Detecting RPC type from request path
// 2. Delegating to RPC type-specific validators
// 3. Creating unified contexts with RPC type information
type cosmosSDKRequestValidator struct {
	logger        polylog.Logger
	chainID       string
	serviceID     protocol.ServiceID
	serviceState  *serviceState
	supportedAPIs map[sharedtypes.RPCType]struct{}

	// RPC type-specific validators - focused on RPC type, not domain
	restValidator    restRequestValidator
	jsonrpcValidator jsonrpcRequestValidator
}

// validateHTTPRequest validates an HTTP request by:
// 1. Detecting the RPC type from the request path/method
// 2. Checking if the detected RPC type is supported
// 3. Delegating to the appropriate RPC type-specific validator
func (crv *cosmosSDKRequestValidator) validateHTTPRequest(req *http.Request) (gateway.RequestQoSContext, bool) {
	crv.logger = crv.logger.With(
		"qos", "CosmosSDK",
		"http_method", req.Method,
		"path", req.URL.Path,
	)

	// 1. Detect RPC type from request characteristics
	rpcType := crv.detectRPCType(req)

	crv.logger = crv.logger.With("detected_rpc_type", rpcType.String())

	// 2. Check if this RPC type is supported by the service
	if _, supported := crv.supportedAPIs[rpcType]; !supported {
		crv.logger.Warn().Msg("Request uses unsupported RPC type")
		return crv.createUnsupportedRPCTypeContext(rpcType), false
	}

	// 3. Delegate to RPC type-specific validator
	switch rpcType {
	case sharedtypes.RPCType_REST:
		return crv.restValidator.validateRESTRequest(req)
	case sharedtypes.RPCType_JSONRPC:
		// All JSON-RPC requests (EVM and CometBFT) handled by same validator
		return crv.jsonrpcValidator.validateJSONRPCRequest(req, rpcType)
	default:
		crv.logger.Error().Msg("Unknown RPC type detected")
		return crv.createUnknownRPCTypeContext(rpcType), false
	}
}

// detectRPCType determines the RPC type based on request path and HTTP method
// Uses path-focused detection with explicit rules for different endpoints
func (crv *cosmosSDKRequestValidator) detectRPCType(req *http.Request) sharedtypes.RPCType {
	path := req.URL.Path
	method := req.Method

	// Priority 1: CosmosSDK REST API patterns (always REST)
	// Examples: /cosmos/*, /ibc/*, /staking/*, /bank/*, etc.
	if isCosmosRestAPI(path) {
		return sharedtypes.RPCType_REST
	}

	// Priority 2: CometBFT endpoints (can be REST or JSON-RPC based on HTTP method)
	// Examples: /health, /status, /genesis, /validators, /block, etc.
	if isCometBftRpc(path) {
		if method == http.MethodPost {
			// POST to CometBFT endpoints can contain JSON-RPC payloads
			return sharedtypes.RPCType_JSONRPC
		}
		// GET/other methods to CometBFT endpoints are REST-style
		return sharedtypes.RPCType_REST
	}

	// Priority 3: Root path - determine based on HTTP method
	if path == "/" || path == "" {
		if method == http.MethodPost {
			// POST to root is almost always JSON-RPC (EVM or CometBFT)
			return sharedtypes.RPCType_JSONRPC
		}
		// GET to root is REST-style
		return sharedtypes.RPCType_REST
	}

	// Default: treat unknown paths as REST
	return sharedtypes.RPCType_REST
}

// Specific path mappings can be added here if needed:
//
// var cometBFTJSONRPCPaths = []string{
//     // Paths that should always be JSON-RPC even with GET
// }
//
// var cometBFTRESTPaths = []string{
//     "/health", "/status", "/genesis", "/validators",
//     // Paths that should always be REST even with POST
// }

// createUnsupportedRPCTypeContext creates an error context for unsupported RPC types
func (crv *cosmosSDKRequestValidator) createUnsupportedRPCTypeContext(rpcType sharedtypes.RPCType) gateway.RequestQoSContext {
	observations := &qosobservations.Observations_Cosmos{
		Cosmos: &qosobservations.CosmosSDKRequestObservations{
			ChainId:       crv.chainID,
			ServiceId:     string(crv.serviceID),
			RequestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
			RpcType:       convertToProtoRPCType(rpcType),
			RequestError: &qosobservations.RequestError{
				ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_USER_ERROR_JSONRPC_PARSE_ERROR,
				ErrorDetails:   "RPC type not supported by this service: " + rpcType.String(),
				HttpStatusCode: httpStatusRequestValidationFailureUnsupportedRPCType,
			},
		},
	}

	return &errorContext{
		logger:                 crv.logger,
		responseHTTPStatusCode: httpStatusRequestValidationFailureUnsupportedRPCType,
		cosmosSDKObservations:  observations,
	}
}

// createUnknownRPCTypeContext creates an error context for unknown RPC types
func (crv *cosmosSDKRequestValidator) createUnknownRPCTypeContext(rpcType sharedtypes.RPCType) gateway.RequestQoSContext {
	observations := &qosobservations.Observations_Cosmos{
		Cosmos: &qosobservations.CosmosSDKRequestObservations{
			ChainId:       crv.chainID,
			ServiceId:     string(crv.serviceID),
			RequestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
			RpcType:       convertToProtoRPCType(rpcType),
			RequestError: &qosobservations.RequestError{
				ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_INTERNAL_PROTOCOL_ERROR,
				ErrorDetails:   "Unknown RPC type detected: " + rpcType.String(),
				HttpStatusCode: httpStatusInternalError,
			},
		},
	}

	return &errorContext{
		logger:                 crv.logger,
		responseHTTPStatusCode: httpStatusInternalError,
		cosmosSDKObservations:  observations,
	}
}

// convertToProtoRPCType converts sharedtypes.RPCType to proto RPCType
func convertToProtoRPCType(rpcType sharedtypes.RPCType) qosobservations.RPCType {
	switch rpcType {
	case sharedtypes.RPCType_REST:
		return qosobservations.RPCType_RPC_TYPE_REST
	case sharedtypes.RPCType_JSONRPC:
		return qosobservations.RPCType_RPC_TYPE_JSONRPC
	default:
		return qosobservations.RPCType_RPC_TYPE_UNSPECIFIED
	}
}

// HTTP status codes for validation failures
const (
	httpStatusRequestValidationFailureUnsupportedRPCType = 400
	httpStatusInternalError                              = 500
)

// ------------------------------------------------------------------------------------------------
// Path Detection Functions
// ------------------------------------------------------------------------------------------------

// isCosmosRestAPI checks if the URL path corresponds to Cosmos SDK REST API endpoints
// These typically run on port :1317 and include paths like:
func isCosmosRestAPI(urlPath string) bool {
	cosmosRestPrefixes := []string{
		"/cosmos/",       // - /cosmos/* (Cosmos SDK modules)
		"/ibc/",          // - /ibc/* (IBC protocol)
		"/staking/",      // - /staking/* (staking module)
		"/auth/",         // - /auth/* (auth module)
		"/bank/",         // - /bank/* (bank module)
		"/txs/",          // - /txs/* (transaction endpoints)
		"/gov/",          // - /gov/* (governance module)
		"/distribution/", // - /distribution/* (distribution module)
		"/slashing/",     // - /slashing/* (slashing module)
		"/mint/",         // - /mint/* (mint module)
		"/upgrade/",      // - /upgrade/* (upgrade module)
		"/evidence/",     // - /evidence/* (evidence module)
	}

	for _, prefix := range cosmosRestPrefixes {
		if strings.HasPrefix(urlPath, prefix) {
			return true
		}
	}
	return false
}

// isCometBftRpc checks if the URL path corresponds to CometBFT RPC endpoints
// These typically run on port :26657
func isCometBftRpc(urlPath string) bool {
	cometBftPaths := []string{
		"/status",               // - /status (node status)
		"/broadcast_tx_sync",    // - /broadcast_tx_sync (transaction broadcast)
		"/broadcast_tx_async",   // - /broadcast_tx_async (transaction broadcast)
		"/broadcast_tx_commit",  // - /broadcast_tx_commit (transaction broadcast)
		"/block",                // - /block (block queries)
		"/commit",               // - /commit (commit queries)
		"/validators",           // - /validators (validator set queries)
		"/genesis",              // - /genesis (genesis document)
		"/health",               // - /health (health check)
		"/abci_info",            // - /abci_info (ABCI info)
		"/abci_query",           // - /abci_query (ABCI query)
		"/consensus_state",      // - /consensus_state (consensus state)
		"/dump_consensus_state", // - /dump_consensus_state (dump consensus state)
		"/net_info",             // - /net_info (network information)
		"/num_unconfirmed_txs",  // - /num_unconfirmed_txs (number of unconfirmed transactions)
		"/tx",                   // - /tx (transaction information)
		"/tx_search",            // - /tx_search (transaction search)
		"/block_search",         // - /block_search (block search)
		"/consensus_params",     // - /consensus_params (consensus parameters)
	}

	for _, path := range cometBftPaths {
		if strings.HasPrefix(urlPath, path) {
			return true
		}
	}
	return false
}
