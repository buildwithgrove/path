package cosmos

import (
	"net/http"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
	"github.com/pokt-network/poktroll/pkg/polylog"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
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
// 1. Determining the validation strategy (REST vs JSON-RPC)
// 2. Delegating to the appropriate validator
// 3. The validator determines RPC type and builds context with necessary info
func (crv *cosmosSDKRequestValidator) validateHTTPRequest(req *http.Request) (gateway.RequestQoSContext, bool) {
	crv.logger = crv.logger.With(
		"qos", "CosmosSDK",
		"http_method", req.Method,
		"path", req.URL.Path,
	)

	// 1. Determine validation strategy based on request characteristics
	strategy := crv.determineRPCValidationStrategy(req)
	crv.logger = crv.logger.With("validation_strategy", string(strategy))

	// 2. Delegate to appropriate validator
	// The validator will determine RPC type and check if supported
	if strategy == rpcValidationStrategyREST {
		return crv.restValidator.validateRESTRequest(req, crv.supportedAPIs, crv.logger, crv.chainID, crv.serviceID)
	} else {
		return crv.jsonrpcValidator.validateJSONRPCRequest(req, crv.supportedAPIs, crv.logger, crv.chainID, crv.serviceID)
	}
}
