package cosmos

import (
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// responseUnmarshalerREST is a multiplexer that routes REST endpoint responses to appropriate unmarshalers
// based on the endpoint path. It serves as the main entry point for processing REST responses.
// Always returns a valid response interface, never returns an error.
func responseUnmarshalerREST(
	logger polylog.Logger,
	endpointPath string,
	data []byte,
) response {
	logger = logger.With("endpoint_path", endpointPath)

	// Route to specific unmarshalers based on endpoint path
	switch endpointPath {
	case "/status":
		return responseUnmarshalerRESTStatus(logger, data)
	case "/health":
		return responseUnmarshalerRESTHealth(logger, data)
	default:
		// For unrecognized endpoints, use the generic unmarshaler
		return responseUnmarshalerRESTGeneric(logger, data)
	}
}
