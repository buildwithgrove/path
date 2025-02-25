package morse

import (
	"errors"
	"fmt"
	"strings"
)

// Error definitions for the Morse protocol
var (
	// ErrNoEndpointsAvailable is returned when attempting to select an endpoint but none are available
	ErrNoEndpointsAvailable = errors.New("no endpoints available for selection")

	// ErrEndpointSelectionFailed is returned when endpoint selection failed
	ErrEndpointSelectionFailed = errors.New("endpoint selection failed")

	// ErrEndpointNotFound is returned when the selected endpoint is not found in the available endpoints
	ErrEndpointNotFound = errors.New("selected endpoint not found in available endpoints")

	// ErrEndpointNotInSession is returned when attempting to get an endpoint from a session but it's not present
	ErrEndpointNotInSession = errors.New("endpoint not found in session")

	// ErrNullRelayResponse is returned when a relay response is null or incomplete
	ErrNullRelayResponse = errors.New("null or incomplete relay response received")

	// ErrRelayTimeout is returned when a relay times out
	ErrRelayTimeout = errors.New("relay request timed out")

	// ErrConnectionFailed is returned when a connection to an endpoint fails
	ErrConnectionFailed = errors.New("connection to endpoint failed")

	// ErrMaxedOut is returned when an endpoint is maxed out
	ErrMaxedOut = errors.New("endpoint is maxed out")

	// ErrInvalidResponse is returned when an endpoint returns an invalid response
	ErrInvalidResponse = errors.New("invalid response from endpoint")

	// ErrValidationFailed is returned when response validation fails
	ErrValidationFailed = errors.New("response validation failed")

	// ErrMisconfigured is returned when an endpoint is misconfigured
	ErrMisconfigured = errors.New("endpoint is misconfigured")
)

// NewNoEndpointsError creates a formatted error for when no endpoints are available
// for a specific service ID
func NewNoEndpointsError(serviceID string) error {
	return fmt.Errorf("%w for service %s", ErrNoEndpointsAvailable, serviceID)
}

// NewEndpointSelectionError creates a formatted error for endpoint selection issues
// that includes the service ID and underlying error
func NewEndpointSelectionError(serviceID string, err error) error {
	return fmt.Errorf("SelectEndpoint: %w for service %s: %v", ErrEndpointSelectionFailed, serviceID, err)
}

// NewEndpointNotFoundError creates a formatted error for when a selected endpoint is not found
// in the available endpoints for a specific service
func NewEndpointNotFoundError(endpointAddr, serviceID string) error {
	return fmt.Errorf("%w: endpoint address %q does not match any available endpoints on service %s", ErrEndpointNotFound, endpointAddr, serviceID)
}

// NewEndpointNotInSessionError creates a formatted error for when an endpoint is not in a session
func NewEndpointNotInSessionError(endpointAddr string) error {
	return fmt.Errorf("%w: %s", ErrEndpointNotInSession, endpointAddr)
}

// NewNullRelayResponseError creates a formatted error for null relay responses with details
// about what specific part of the response was null or invalid
func NewNullRelayResponseError(detail string) error {
	return fmt.Errorf("%w: %s", ErrNullRelayResponse, detail)
}

// extractErrFromRelayError analyzes an error returned during relay and attempts to match it
// to one of our predefined error types. It first checks by error comparison (unwrapping),
// then falls back to string analysis if the error type isn't recognized.
// This function centralizes error recognition logic to avoid replicating string matching
// throughout the codebase.
func extractErrFromRelayError(err error) error {
	if err == nil {
		return nil
	}

	// Check for known predefined errors through unwrapping
	if errors.Is(err, ErrRelayTimeout) ||
		errors.Is(err, ErrConnectionFailed) ||
		errors.Is(err, ErrInvalidResponse) ||
		errors.Is(err, ErrValidationFailed) ||
		errors.Is(err, ErrNullRelayResponse) ||
		errors.Is(err, ErrEndpointNotInSession) ||
		errors.Is(err, ErrEndpointSelectionFailed) ||
		errors.Is(err, ErrNoEndpointsAvailable) ||
		errors.Is(err, ErrEndpointNotFound) ||
		errors.Is(err, ErrMaxedOut) ||
		errors.Is(err, ErrMisconfigured) {
		return err
	}

	// Fall back to string matching for errors not using our predefined errors
	errStr := strings.ToLower(err.Error())

	// Check for endpoint maxed out errors
	if isEndpointMaxedOutError(errStr) {
		return ErrMaxedOut
	}

	// Check for endpoint misconfiguration errors
	if isEndpointRejectingAValidChain(errStr) {
		return ErrMisconfigured
	}

	// Check for timeouts using strings
	if strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline") {
		return ErrRelayTimeout
	}

	// Check for connection errors using strings
	if strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "dial") ||
		strings.Contains(errStr, "connect") {
		return ErrConnectionFailed
	}

	// Check for max quota errors
	if strings.Contains(errStr, "max") ||
		strings.Contains(errStr, "quota") ||
		strings.Contains(errStr, "limit") {
		return ErrMaxedOut
	}

	// Check for validation failures
	if strings.Contains(errStr, "validation") ||
		strings.Contains(errStr, "invalid") {
		return ErrValidationFailed
	}

	// If no specific error type matched, return the original error
	return err
}

// isEndpointMaxedOutError checks if the error string indicates an endpoint is maxed out
func isEndpointMaxedOutError(errStr string) bool {
	return matchesAllSubstrings(
		errStr,
		[]string{
			"codespace: sdk",
			"code: 1",
			"codespace: pocketcore",
			"code: 90",
			"the evidence is sealed",
			"max relays reached",
		},
	)
}

// isEndpointRejectingAValidChain checks if the error string indicates the endpoint rejected a valid chain
// The error message contains specific sections which are checked to identify this error type
func isEndpointRejectingAValidChain(errStr string) bool {
	return matchesAllSubstrings(
		errStr,
		[]string{
			"codespace: sdk",
			"code: 1",
			"codespace: pocketcore",
			"code: 26",
			"blockchain in the relay request is not supported on this node",
		},
	)
}

// matchesAllSubstrings checks if all the specified substrings are present in the given string
func matchesAllSubstrings(str string, piecesToMatch []string) bool {
	for _, pieceToMatch := range piecesToMatch {
		if !strings.Contains(str, pieceToMatch) {
			return false
		}
	}

	return true
}
