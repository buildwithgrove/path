package cosmos

import (
	"net/http"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
	"github.com/pokt-network/poktroll/pkg/polylog"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
)

// requestValidator handles validation for all Cosmos service requests
// Coordinates between different protocol validators (JSONRPC, REST)
type requestValidator struct {
	logger        polylog.Logger
	chainID       string
	serviceID     protocol.ServiceID
	supportedAPIs map[sharedtypes.RPCType]struct{}
	serviceState  protocol.EndpointSelector
}

// validateHTTPRequest validates an HTTP request and routes to appropriate sub-validator
// Returns (context, true) on success or (errorContext, false) on failure
func (rv *requestValidator) validateHTTPRequest(req *http.Request) (gateway.RequestQoSContext, bool) {
	logger := rv.logger.With(
		"qos", "Cosmos",
		"method", "validateHTTPRequest",
		"path", req.URL.Path,
		"http_method", req.Method,
	)

	// Determine request type and route to appropriate validator
	if rv.isJSONRPCRequest(req) {
		logger.Debug().Msg("Routing to JSONRPC validator")

		// Initialize JSONRPC validator with shared fields and validate
		jsonrpcValidator := jsonrpcRequestValidator{}
		return jsonrpcValidator.validateJSONRPCRequest(
			req,
			rv.supportedAPIs,
			logger,
			rv.chainID,
			rv.serviceID,
			rv.serviceState,
		)
	} else {
		logger.Debug().Msg("Routing to REST validator")

		// Initialize REST validator with shared fields and validate
		restValidator := restRequestValidator{}
		return restValidator.validateRESTRequest(
			req,
			rv.supportedAPIs,
			logger,
			rv.chainID,
			rv.serviceID,
			rv.serviceState,
		)
	}
}

// isJSONRPCRequest determines if the incoming HTTP request is a JSONRPC request
// Uses simple heuristics: POST method and specific content types or paths
func (rv *requestValidator) isJSONRPCRequest(req *http.Request) bool {
	// JSONRPC requests are typically POST with specific patterns
	if req.Method != http.MethodPost {
		return false
	}

	// Check content type if present
	contentType := req.Header.Get("Content-Type")
	if contentType == "application/json" || contentType == "application/json-rpc" {
		return true
	}

	// Check for common JSONRPC paths (many services use root path)
	path := req.URL.Path
	jsonrpcPaths := []string{
		"/",
		"/rpc",
		"/jsonrpc",
		"/v1",
		"/api/v1",
	}

	for _, jsonrpcPath := range jsonrpcPaths {
		if path == jsonrpcPath {
			return true
		}
	}

	// Default to JSONRPC for POST requests without clear REST patterns
	return true
}
