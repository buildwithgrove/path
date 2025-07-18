package cosmos

import (
	"io"
	"net/http"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/pokt-network/poktroll/pkg/polylog"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
)

// restRequestValidator handles validation of REST API requests
// Determines RPC type from request path and creates appropriate context
type restRequestValidator struct{}

// validateRESTRequest validates a REST request by:
// 1. Validating HTTP method and path
// 2. Determining the specific RPC type from the path
// 3. Checking if the RPC type is supported
// 4. Creating the request context with all necessary information
func (rv *restRequestValidator) validateRESTRequest(
	req *http.Request,
	supportedAPIs map[sharedtypes.RPCType]struct{},
	logger polylog.Logger,
	chainID string,
	serviceID protocol.ServiceID,
) (gateway.RequestQoSContext, bool) {

	logger = logger.With("validator", "REST")

	// Validate HTTP method is appropriate for REST
	if !rv.isValidRESTMethod(req.Method) {
		logger.Warn().Str("method", req.Method).Msg("Invalid HTTP method for REST API")
		return rv.createInvalidMethodContext(req.Method, logger, chainID, serviceID), false
	}

	// Determine the specific RPC type based on path patterns
	rpcType := rv.determineRESTRPCType(req.URL.Path)
	logger = logger.With("detected_rpc_type", rpcType.String())

	// Check if this RPC type is supported by the service
	if _, supported := supportedAPIs[rpcType]; !supported {
		logger.Warn().Msg("Request uses unsupported RPC type")
		return rv.createUnsupportedRPCTypeContext(rpcType, logger, chainID, serviceID), false
	}

	// Validate the path is a recognized REST API pattern
	if !rv.isValidRESTPath(req.URL.Path) {
		logger.Warn().Str("path", req.URL.Path).Msg("Invalid path for REST API")
		return rv.createInvalidPathContext(req.URL.Path, logger, chainID, serviceID), false
	}

	// Read request body if present (for POST/PUT requests)
	body, err := rv.readRequestBody(req)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to read REST request body")
		return rv.createBodyReadErrorContext(err, logger, chainID, serviceID), false
	}

	logger.Debug().
		Int("body_length", len(body)).
		Msg("REST request validation successful")

	// Create request context with detected RPC type
	return &requestContext{
		logger:               logger,
		httpReq:              *req,
		chainID:              chainID,
		serviceID:            serviceID,
		rpcType:              rpcType,
		requestPayloadLength: uint(len(body)),
		requestOrigin:        qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
		restBody:             body,
	}, true
}

// determineRESTRPCType determines the specific RPC type for REST requests
func (rv *restRequestValidator) determineRESTRPCType(path string) sharedtypes.RPCType {
	// CometBFT REST-style endpoints should be tagged as COMET_BFT
	if isCometBftRpc(path) {
		return sharedtypes.RPCType_COMET_BFT
	}

	// Everything else is regular REST
	return sharedtypes.RPCType_REST
}

// isValidRESTMethod checks if the HTTP method is valid for REST API requests
func (rv *restRequestValidator) isValidRESTMethod(method string) bool {
	validMethods := []string{
		http.MethodGet,     // Query data
		http.MethodPost,    // Submit transactions, complex queries
		http.MethodPut,     // Update operations (less common)
		http.MethodDelete,  // Delete operations (less common)
		http.MethodHead,    // Header-only requests
		http.MethodOptions, // CORS preflight
	}

	for _, validMethod := range validMethods {
		if method == validMethod {
			return true
		}
	}
	return false
}

// isValidRESTPath validates that the path matches REST API patterns
func (rv *restRequestValidator) isValidRESTPath(path string) bool {
	// CosmosSDK REST API patterns
	if isCosmosRestAPI(path) {
		return true
	}

	// CometBFT REST-style endpoints
	if isCometBftRpc(path) {
		return true
	}

	// Allow some additional common REST patterns
	additionalValidPaths := []string{
		"/txs",          // Legacy transaction endpoint
		"/tx",           // Transaction endpoints
		"/node_info",    // Node information
		"/syncing",      // Sync status
		"/latest_block", // Latest block info
	}

	for _, validPath := range additionalValidPaths {
		if path == validPath || (len(path) > len(validPath) && path[:len(validPath)+1] == validPath+"/") {
			return true
		}
	}

	return false
}

// readRequestBody safely reads the HTTP request body
func (rv *restRequestValidator) readRequestBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// Error context creation methods

func (rv *restRequestValidator) createInvalidMethodContext(method string, logger polylog.Logger, chainID string, serviceID protocol.ServiceID) gateway.RequestQoSContext {
	observations := &qosobservations.Observations_Cosmos{
		Cosmos: &qosobservations.CosmosSDKRequestObservations{
			ChainId:       chainID,
			ServiceId:     string(serviceID),
			RequestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
			RpcType:       qosobservations.RPCType_RPC_TYPE_REST,
			RequestError: &qosobservations.RequestError{
				ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_USER_ERROR_JSONRPC_PARSE_ERROR,
				ErrorDetails:   "Invalid HTTP method for REST API: " + method,
				HttpStatusCode: http.StatusMethodNotAllowed,
			},
		},
	}

	return &errorContext{
		logger:                 logger,
		responseHTTPStatusCode: http.StatusMethodNotAllowed,
		cosmosSDKObservations:  observations,
	}
}

func (rv *restRequestValidator) createUnsupportedRPCTypeContext(rpcType sharedtypes.RPCType, logger polylog.Logger, chainID string, serviceID protocol.ServiceID) gateway.RequestQoSContext {
	observations := &qosobservations.Observations_Cosmos{
		Cosmos: &qosobservations.CosmosSDKRequestObservations{
			ChainId:       chainID,
			ServiceId:     string(serviceID),
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
		logger:                 logger,
		responseHTTPStatusCode: httpStatusRequestValidationFailureUnsupportedRPCType,
		cosmosSDKObservations:  observations,
	}
}

func (rv *restRequestValidator) createInvalidPathContext(path string, logger polylog.Logger, chainID string, serviceID protocol.ServiceID) gateway.RequestQoSContext {
	observations := &qosobservations.Observations_Cosmos{
		Cosmos: &qosobservations.CosmosSDKRequestObservations{
			ChainId:       chainID,
			ServiceId:     string(serviceID),
			RequestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
			RpcType:       qosobservations.RPCType_RPC_TYPE_REST,
			RequestError: &qosobservations.RequestError{
				ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_USER_ERROR_JSONRPC_PARSE_ERROR,
				ErrorDetails:   "Invalid path for REST API: " + path,
				HttpStatusCode: http.StatusNotFound,
			},
		},
	}

	return &errorContext{
		logger:                 logger,
		responseHTTPStatusCode: http.StatusNotFound,
		cosmosSDKObservations:  observations,
	}
}

func (rv *restRequestValidator) createBodyReadErrorContext(err error, logger polylog.Logger, chainID string, serviceID protocol.ServiceID) gateway.RequestQoSContext {
	observations := &qosobservations.Observations_Cosmos{
		Cosmos: &qosobservations.CosmosSDKRequestObservations{
			ChainId:       chainID,
			ServiceId:     string(serviceID),
			RequestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
			RpcType:       qosobservations.RPCType_RPC_TYPE_REST,
			RequestError: &qosobservations.RequestError{
				ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_INTERNAL_READ_HTTP_ERROR,
				ErrorDetails:   "Failed to read REST request body: " + err.Error(),
				HttpStatusCode: http.StatusInternalServerError,
			},
		},
	}

	return &errorContext{
		logger:                 logger,
		responseHTTPStatusCode: http.StatusInternalServerError,
		cosmosSDKObservations:  observations,
	}
}
