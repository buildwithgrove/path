package morse

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pokt-foundation/pocket-go/provider"
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

	// ErrRelayRequestTimeout is returned when a relay request from the gateway to the backend service (endpoint) times out
	ErrRelayRequestTimeout = errors.New("relay request to backend service (endpoint) timed out")

	// ErrConnectionFailed is returned when a connection to an endpoint fails
	ErrConnectionFailed = errors.New("connection to endpoint failed")

	// ErrMaxedOut is returned when an endpoint is maxed out and cannot handle any more incoming requests
	ErrMaxedOut = errors.New("endpoint is maxed out and cannot handle any more incoming requests")

	// ErrInvalidResponse is returned when an endpoint returns an invalid response
	ErrInvalidResponse = errors.New("invalid response from endpoint")

	// ErrPocketCore is returned when an endpoint returns an SDK or Pocket-core error.
	ErrPocketCore = errors.New("endpoint returned SDK/Pocket-Core error")

	// ErrSDK4XX is the Morse SDK's 4XX error.
	// https://github.com/pokt-foundation/pocket-go/blob/0cb5a3a2ab762e7af18b3482f864d2d9d211a71f/provider/provider.go#L24
	ErrSDK4XX = provider.Err4xxOnConnection

	// ErrSDK5XX is the Morse SDK's 5XX error.
	// https://github.com/pokt-foundation/pocket-go/blob/0cb5a3a2ab762e7af18b3482f864d2d9d211a71f/provider/provider.go#L26
	ErrSDK5XX = provider.Err5xxOnConnection

	// ErrTLSCertificateVerificationFailed is returned when TLS certificate verification failed
	ErrTLSCertificateVerificationFailed = errors.New("TLS certificate verification failed")

	// ErrNonJSONResponse is returned when an endpoint returns a non-JSON response
	ErrNonJSONResponse = errors.New("non JSON response received from endpoint")

	// ErrHTTPContentLengthIncorrect is returned when an endpoint returns an HTTP response with a mismatch between:
	// - The ContentLength HTTP header
	// - Actual body length
	ErrHTTPContentLengthIncorrect = errors.New("endpoint returned HTTP response with ContentLength mismatching the actual length")

	// ErrExecutingHTTPRequest is returned when an endpoint returned an error indicating it encountered an error executing the HTTP request.
	ErrExecutingHTTPRequest = errors.New("endpoint indicated error executing the HTTP request")
)

// NewNoEndpointsError creates a formatted error for when no endpoints are available
// for a specific service ID
func NewNoEndpointsError(serviceID string) error {
	return fmt.Errorf("%w for service %s", ErrNoEndpointsAvailable, serviceID)
}

// NewEndpointSelectionError creates a formatted error for endpoint selection issues
// that includes the service ID and underlying error
func NewEndpointSelectionError(serviceID string, err error) error {
	return fmt.Errorf("SelectEndpoint: %w for service %s: %w", ErrEndpointSelectionFailed, serviceID, err)
}

// NewEndpointNotFoundError creates a formatted error for when a selected endpoint is not found
// in the available endpoints for a specific service
func NewEndpointNotFoundError(endpointAddr, serviceID string) error {
	return fmt.Errorf("%w: endpoint address %q does not match any available endpoints on service %s", ErrEndpointNotFound, endpointAddr, serviceID)
}

// NewNullRelayResponseError creates a formatted error for null relay responses with details
// about what specific part of the response was null or invalid
func NewNullRelayResponseError(detail string) error {
	return fmt.Errorf("%w: %s", ErrNullRelayResponse, detail)
}

// extractErrFromRelayError:
// • Analyzes errors returned during relay operations
// • Matches errors to predefined types through:
//   - Primary: Error comparison (with unwrapping)
//   - Fallback: String analysis for unrecognized types
//
// • Centralizes error recognition logic to avoid duplicate string matching
func extractErrFromRelayError(err error) error {
	if err == nil {
		return nil
	}

	// Check for known predefined errors through unwrapping
	if errors.Is(err, ErrRelayRequestTimeout) ||
		errors.Is(err, ErrConnectionFailed) ||
		errors.Is(err, ErrInvalidResponse) ||
		errors.Is(err, ErrNullRelayResponse) ||
		errors.Is(err, ErrEndpointNotInSession) ||
		errors.Is(err, ErrEndpointSelectionFailed) ||
		errors.Is(err, ErrNoEndpointsAvailable) ||
		errors.Is(err, ErrEndpointNotFound) ||
		errors.Is(err, ErrMaxedOut) ||
		errors.Is(err, ErrSDK4XX) ||
		errors.Is(err, ErrSDK5XX) {
		return err
	}

	// Fall back to string matching for errors not using our predefined errors
	errStr := strings.ToLower(err.Error())

	// Check for endpoint maxed out errors
	if isEndpointMaxedOutError(errStr) {
		return ErrMaxedOut
	}

	// Check for endpoint SDK/Pocket-Core errors.
	if isEndpointPocketCoreError(errStr) {
		return ErrPocketCore
	}

	if isEndpointHTTPContentLengthMismatchErr(errStr) {
		return ErrHTTPContentLengthIncorrect
	}

	if isEndpointErrorExecutingHTTPRequest(errStr) {
		return ErrExecutingHTTPRequest
	}

	// Check for TLS certificate verification errors
	if strings.Contains(errStr, "tls: failed to verify certificate") {
		return ErrTLSCertificateVerificationFailed
	}

	// Check for non-JSON response errors
	if strings.Contains(errStr, "non json response") {
		return ErrNonJSONResponse
	}

	// Check for timeouts using strings
	if strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline") {
		return ErrRelayRequestTimeout
	}

	// Check for connection errors using strings
	if strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "dial") ||
		strings.Contains(errStr, "connect") {
		return ErrConnectionFailed
	}

	// If no specific error type matched, return the original error
	return err
}

// isEndpointMaxedOutError determines if the error message indicates the endpoint has reached its capacity.
// Returns true when the endpoint cannot accept additional relays for the current (app,session) pair.
func isEndpointMaxedOutError(errStr string) bool {
	return matchesAllSubstrings(
		errStr,
		[]string{
			"code: 90",
			"codespace: pocketcore",
			"the evidence is sealed",
		},
	)
}

// isEndpointHTTPContentLengthMismatchErr checks if the error indicates a mismatch in the endpoint's HTTP response between:
// - The `ContentLength` header, and
// - Actual body length.
func isEndpointHTTPContentLengthMismatchErr(errStr string) bool {
	return matchesAllSubstrings(
		errStr,
		[]string{
			"post", // the supplied error string is lower case.
			"v1/client/relay",
			"http: contentlength=",
			"with body length",
		},
	)
}

// isEndpointPocketCoreError checks if the error string indicates any SDK/Pocket-core errors.
// The error message contains specific sections which are checked to identify this error type.
func isEndpointPocketCoreError(errStr string) bool {
	return matchesAllSubstrings(
		errStr,
		[]string{
			"codespace: sdk",
			"code: 1",
		},
	)
}

// isEndpointErrorExecutingHTTPRequest checks if the error string indicates an endpoint-reported error on executing the HTTP request.
func isEndpointErrorExecutingHTTPRequest(errStr string) bool {
	return strings.Contains(errStr, "error executing the http request: blockchain request for chain")
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
