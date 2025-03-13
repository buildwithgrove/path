package morse

// TODO_FUTURE(#agent): Ensure that the helpers in this files are used everywhere appropriate.
// For existing code and new code incoming from PRs, find opportunities to provide additional context
// or refactor existing code that should be using them instead of reimplementing similar logic.
import (
	"fmt"

	"github.com/pokt-foundation/pocket-go/provider"
	"github.com/pokt-network/poktroll/pkg/polylog"

	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/protocol"
)

// loggerWithEndpointObservation enhances a logger with detailed information from a
// MorseEndpointObservation. This creates contextually rich logs that include all relevant
// fields from the observation for better debugging and traceability.
//
// If the provided endpointObservation is nil, the original logger is returned unchanged.
//
// Parameters:
//   - logger: The base logger to enhance
//   - endpointObservation: The observation containing fields to add to the logger
//
// Returns:
//   - An enhanced logger with all relevant observation fields attached
func loggerWithEndpointObservation(
	logger polylog.Logger,
	endpointObservation *protocolobservations.MorseEndpointObservation,
) polylog.Logger {

	if endpointObservation == nil {
		return logger
	}

	loggerWithSession := loggerWithSessionFields(
		logger,
		endpointObservation.GetAppAddress(),
		endpointObservation.GetSessionKey(),
		endpointObservation.GetSessionServiceId(),
		int(endpointObservation.GetSessionHeight()),
	)

	return loggerWithSession.With(
		"endpoint_addr", endpointObservation.GetEndpointAddr(),
		"error_type", endpointObservation.GetErrorType().String(),
		"error_details", endpointObservation.GetErrorDetails(),
		"recommended_sanction", endpointObservation.GetRecommendedSanction().String(),
	)
}

// loggerWithEndpoint creates an enhanced logger with contextual information about
// an endpoint and its sanction status. This is typically used when filtering endpoints
// or when applying sanctions to provide a clear context for log entries.
//
// Parameters:
//   - logger: The base logger to enhance
//   - appAddr: Address of the application
//   - sessionKey: Key of the current session
//   - endpointAddr: Address of the endpoint being processed
//   - reason: Explanation for the action (e.g., reason for sanction)
//
// Returns:
//   - An enhanced logger with endpoint information and context attached
func loggerWithEndpoint(
	logger polylog.Logger,
	appAddr string,
	sessionKey string,
	endpointAddr protocol.EndpointAddr,
	reason string,
) polylog.Logger {
	return logger.With(
		"app_addr", appAddr,
		"session_key", sessionKey,
		"endpoint_addr", string(endpointAddr),
		"reason", reason,
	)
}

// loggerWithSession enhances a logger with session information from a provider.Session object.
// This is a convenience wrapper around loggerWithSessionFields, extracting the
// necessary session fields from the provider.Session structure.
//
// Parameters:
//   - logger: The base logger to enhance
//   - appAddr: Address of the application associated with the session
//   - session: The Session object containing session details
//
// Returns:
//   - A logger enhanced with session context information
func loggerWithSession(
	logger polylog.Logger,
	appAddr string,
	session provider.Session,
) polylog.Logger {
	return loggerWithSessionFields(
		logger,
		appAddr,
		session.Key,
		session.Header.Chain,
		session.Header.SessionHeight,
	)
}

// loggerWithSessionFields adds session-specific fields to a logger.
// This function is used internally by other logging helpers to maintain consistent
// session-related fields across different log contexts.
//
// Parameters:
//   - logger: The base logger to enhance
//   - appAddr: Address of the application
//   - sessionKey: Unique identifier for the session
//   - sessionServiceID: Service/chain ID for which the session was created
//   - sessionHeight: Blockchain height at which the session was established
//
// Returns:
//   - A logger with session details attached for context
func loggerWithSessionFields(
	logger polylog.Logger,
	appAddr string,
	sessionKey string,
	sessionServiceID string,
	sessionHeight int,
) polylog.Logger {
	return logger.With(
		"app_addr", appAddr,
		"session_key", sessionKey,
		"session_service_id", sessionServiceID,
		"session_height", fmt.Sprintf("%d", sessionHeight),
	)
}
