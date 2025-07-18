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
		return createInvalidMethodError(req.Method, logger, chainID, serviceID, "REST"), false
	}

	// Determine the specific RPC type based on path patterns - delegate to specialized detection
	rpcType := determineRESTRPCType(req.URL.Path)
	logger = logger.With("detected_rpc_type", rpcType.String())

	// Check if this RPC type is supported by the service
	if _, supported := supportedAPIs[rpcType]; !supported {
		logger.Warn().Msg("Request uses unsupported RPC type")
		return createUnsupportedRPCTypeError(rpcType, logger, chainID, serviceID), false
	}

	// Validate the path is a recognized REST API pattern - delegate to specialized validation
	if !isValidRESTPath(req.URL.Path) {
		logger.Warn().Str("path", req.URL.Path).Msg("Invalid path for REST API")
		return createInvalidPathError(req.URL.Path, logger, chainID, serviceID), false
	}

	// Read request body if present (for POST/PUT requests)
	body, err := rv.readRequestBody(req)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to read REST request body")
		return createBodyReadError(err, logger, chainID, serviceID), false
	}

	logger.Debug().
		Int("body_length", len(body)).
		Msg("REST request validation successful")

	// Create specialized REST context
	return &restContext{
		logger:               logger,
		chainID:              chainID,
		serviceID:            serviceID,
		httpMethod:           req.Method,
		urlPath:              req.URL.Path,
		requestBody:          body,
		rpcType:              rpcType,
		requestPayloadLength: uint(len(body)),
		requestOrigin:        qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
		headers:              extractRelevantHeaders(req),
	}, true
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

// extractRelevantHeaders extracts headers that should be forwarded to endpoints
func extractRelevantHeaders(req *http.Request) map[string]string {
	headers := make(map[string]string)

	// Forward common headers that might be relevant for REST API calls
	relevantHeaders := []string{
		"Accept",
		"Accept-Encoding",
		"Authorization",
		"User-Agent",
		"X-Forwarded-For",
		"X-Real-IP",
	}

	for _, headerName := range relevantHeaders {
		if value := req.Header.Get(headerName); value != "" {
			headers[headerName] = value
		}
	}

	return headers
}
