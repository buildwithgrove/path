package shannon

import (
	"errors"
	"strings"
)

var (
	// endpoint configuration error:
	// - TLS certificate verification error.
	// - DNS error on lookup of endpoint URL.
	ErrRelayEndpointConfig = errors.New("endpoint configuration error")

	// endpoint timeout
	ErrRelayEndpointTimeout = errors.New("timeout waiting for endpoint response")

	// PATH manually canceled the context for the request.
	// E.g. Parallel requests were made and one succeeded so the other was canceled.
	ErrContextCancelled = errors.New("context canceled manually")

	// Request context setup errors.
	// Used to build observations:
	// There is no request context to provide observations.
	//
	// Unsupported gateway mode
	errProtocolContextSetupUnsupportedGatewayMode = errors.New("unsupported gateway mode")
	// Centralized gateway mode: Error getting onchain data for app
	errProtocolContextSetupCentralizedAppFetchErr = errors.New("error getting onchain data for app owned by the gateway")
	// Centralized gateway mode app does not delegate to the gateway.
	errProtocolContextSetupCentralizedAppDelegation = errors.New("centralized gateway mode app does not delegate to the gateway")
	// Centralized gateway mode: no active sessions could be retrieved for the service.
	errProtocolContextSetupCentralizedNoSessions = errors.New("no active sessions could be retrieved for the service")
	// Centralized gateway mode: no owned apps found for the service.
	errProtocolContextSetupCentralizedNoAppsForService = errors.New("ZERO owned apps found for service")

	// Delegated gateway mode: could not extract app from HTTP request.
	errProtocolContextSetupGetAppFromHTTPReq = errors.New("error getting the selected app from the HTTP request")
	// Delegated gateway mode: could not fetch session for app from the full node
	errProtocolContextSetupFetchSession = errors.New("error getting a session from the full node for app")
	// Delegated gateway mode: gateway does not have delegation for the app.
	errProtocolContextSetupAppDoesNotDelegate = errors.New("gateway does not have delegation for app")
	// Delegated gateway mode: app is not staked for the service.
	errProtocolContextSetupAppNotStaked = errors.New("app is not staked for the service")
	// No endpoints available for the service.
	// Can be due to one or more of the following:
	// - Any of the gateway mode errors above.
	// - Error fetching a session for an app.
	errProtocolContextSetupNoEndpoints = errors.New("no endpoints found for service: relay request will fail")
	// Selected endpoint is no longer available.
	// Can happen due to:
	// - Bug in endpoint selection logic.
	// - Endpoint sanctioned due to an observation while selection logic was running.
	errRequestContextSetupInvalidEndpointSelected = errors.New("selected endpoint is not available: relay request will fail")
	// Error initializing a signer for the current gateway mode.
	errRequestContextSetupErrSignerSetup = errors.New("error getting the permitted signer: relay request will fail")
)

// extractErrFromRelayError:
// • Analyzes errors returned during relay operations
// • Matches errors to predefined types through:
//   - Primary: Error comparison (with unwrapping)
//   - Fallback: String analysis for unrecognized types
//
// • Centralizes error recognition logic to avoid duplicate string matching
func extractErrFromRelayError(err error) error {
	if isEndpointConfigError(err) {
		return ErrRelayEndpointConfig
	}

	// endpoint timeout
	if strings.Contains(err.Error(), "context deadline exceeded") {
		return ErrRelayEndpointTimeout
	}

	// context canceled manually
	if strings.Contains(err.Error(), "context canceled") {
		return ErrContextCancelled
	}

	// No known patterns matched.
	// return the error as-is.
	return err
}

// returns true if the error indicating an endpoint configuration error.
// Examples:
// - Error verifying endpoint's TLS certificate
// - Error on DNS lookup of endpoint's URL.
func isEndpointConfigError(err error) bool {
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "dial tcp: lookup"):
		return true
	case strings.Contains(errStr, "tls: failed to verify certificate"):
		return true
	default:
		return false
	}
}
