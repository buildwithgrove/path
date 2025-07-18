package cosmos

import (
	"net/http"
	"strings"
)

// rpcValidationStrategy represents which RPC validation strategy to use
type rpcValidationStrategy string

const (
	rpcValidationStrategyREST    rpcValidationStrategy = "REST"
	rpcValidationStrategyJSONRPC rpcValidationStrategy = "JSONRPC"
)

// determineRPCValidationStrategy determines which RPC validation strategy to use (REST vs JSONRPC)
// Uses HTTP method + path + content-type following the multi-service API router strategy
func (crv *cosmosSDKRequestValidator) determineRPCValidationStrategy(req *http.Request) rpcValidationStrategy {
	// REST validation: GET/PUT/DELETE requests OR non-JSON content-type
	if req.Method != http.MethodPost {
		return rpcValidationStrategyREST
	}

	// POST with non-JSON content-type uses REST validation
	contentType := req.Header.Get("Content-Type")
	if contentType != "" && !strings.Contains(contentType, "application/json") {
		return rpcValidationStrategyREST
	}

	// POST with JSON content-type uses JSONRPC validation
	return rpcValidationStrategyJSONRPC
}
