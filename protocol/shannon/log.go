package shannon

import (
	"github.com/pokt-network/poktroll/pkg/polylog"
	//	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
)

// hydrateLoggerWithEndpoint enhances a logger with a Shannon endpoint details.
// Creates contextually rich logs.
//
// Parameters:
//   - logger: The base logger to enhance
//   - endpoint: The Shannon endpoint
//
// Returns:
//   - An enhanced logger with all relevant endpoint fields attached
func hydrateLoggerWithEndpoint(
	logger polylog.Logger,
	endpoint *endpoint,
) polylog.Logger {
	hydratedLogger := logger.With(
		"endpoint_supplier", endpoint.Supplier(),
		"endpoint_url", endpoint.PublicURL(),
	)

	sessionHeader := endpoint.session.Header
	// nil session header: skip the processing.
	if sessionHeader == nil {
		return hydratedLogger
	}

	return hydratedLogger.With(
		// Hydrate with session fields.
		"endpoint_app_addr", sessionHeader.ApplicationAddress,
		"endpoint_session_service_id", sessionHeader.ServiceId,
		"endpoint_session_id", sessionHeader.SessionId,
		"endpoint_session_start_height", sessionHeader.SessionStartBlockHeight,
		"endpoint_session_end_height", sessionHeader.GetSessionEndBlockHeight,
	)
}
