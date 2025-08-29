package shannon

import (
	"errors"
	"regexp"
	"strings"

	"github.com/pokt-network/poktroll/pkg/polylog"
	sdk "github.com/pokt-network/shannon-sdk"

	pathhttp "github.com/buildwithgrove/path/network/http"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
)

// classifyRelayError determines the ShannonEndpointErrorType and recommended ShannonSanctionType for a given relay error.
//
// - Uses extractErrFromRelayError to identify the specific error type.
// - Maps known errors to endpoint error types and sanctions.
// - Enhanced error type identification for malformed endpoint payloads.
// - Logs and returns a generic internal error for unknown cases.
func classifyRelayError(logger polylog.Logger, err error) (protocolobservations.ShannonEndpointErrorType, protocolobservations.ShannonSanctionType) {
	// No error: return unspecified.
	if err == nil {
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_UNSPECIFIED,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_UNSPECIFIED
	}

	// Classify shannon-sdk errors
	switch {
	// HTTP relay errors - check first to handle HTTP-specific classifications
	case errors.Is(err, errSendHTTPRelay):
		return classifyHttpError(logger, err)

	// Endpoint payload failed to unmarshal/validate
	case errors.Is(err, errMalformedEndpointPayload):
		// Extract the payload content from the error message
		errorStr := err.Error()
		payloadContent := strings.TrimPrefix(errorStr, "raw_payload: ")
		if idx := strings.LastIndex(payloadContent, ": endpoint returned malformed payload"); idx != -1 {
			payloadContent = payloadContent[:idx]
		}
		return classifyMalformedEndpointPayload(logger, payloadContent)

	// Endpoint payload failed to unmarshal into a RelayResponse struct
	case errors.Is(err, sdk.ErrRelayResponseValidationUnmarshal):
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_PAYLOAD_UNMARSHAL_ERR,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION

	// Endpoint response failed basic validation
	case errors.Is(err, sdk.ErrRelayResponseValidationBasicValidation):
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_RESPONSE_VALIDATION_ERR,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION

	// Could not fetch the public key for supplier address used for the relay.
	case errors.Is(err, sdk.ErrRelayResponseValidationGetPubKey):
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_RESPONSE_GET_PUBKEY_ERR,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION

	// Received nil public key on supplier lookup using its address.
	// This means the supplier account is not properly initialized:
	//
	// In Cosmos SDK (and thus in pocketd) accounts:
	// - Are created when they receive tokens.
	// - Get their public key onchain once they sign their first transaction (e.g. send, delegate, stake, etc.)
	case errors.Is(err, sdk.ErrRelayResponseValidationNilSupplierPubKey):
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_NIL_SUPPLIER_PUBKEY,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION

	// RelayResponse's signature failed validation.
	case errors.Is(err, sdk.ErrRelayResponseValidationSignatureError):
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_RESPONSE_SIGNATURE_VALIDATION_ERR,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION

	// TODO_NEXT(@commoddity): Introduce correct error classification for WebSocket errors.
	//   - Error creating a WebSocket connection.
	//   - Error signing the relay request.
	//   - Error validating the relay response.

	// WebSocket connection failed.
	case errors.Is(err, errCreatingWebSocketConnection):
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_WEBSOCKET_CONNECTION_FAILED,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_UNSPECIFIED

	// Error signing the relay request.
	case errors.Is(err, errRelayRequestWebsocketMessageSigningFailed):
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_WEBSOCKET_REQUEST_SIGNING_FAILED,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_UNSPECIFIED

	// Error validating the relay response in a websocket message.
	case errors.Is(err, errRelayResponseInWebsocketMessageValidationFailed):
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_WEBSOCKET_RELAY_RESPONSE_VALIDATION_FAILED,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_UNSPECIFIED
	}

	// Fallback to error matching using the error string.
	// Extract the specific error type using centralized error matching.
	extractedErr := extractErrFromRelayError(err)

	// Map known errors to endpoint error types and sanctions.
	// TODO_TECHDEBT(@Olshansk): Re-evaluate which errors should be session-based or permanent.
	switch extractedErr {

	// Endpoint Configuration error
	case errRelayEndpointConfig:
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_CONFIG,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION

	// endpoint timeout error
	case errRelayEndpointTimeout:
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_TIMEOUT,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION

	// Endpoint's backend service returned a non 2xx HTTP status code.
	case pathhttp.ErrRelayEndpointHTTPError:
		// TODO_IMPROVE: Make this a sanction that just lasts a few blocks
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_HTTP_BAD_RESPONSE,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_UNSPECIFIED

	case errContextCanceled:
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_REQUEST_CANCELED_BY_PATH,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_DO_NOT_SANCTION

	default:
		// Unknown error: log and return generic internal error.
		// TODO_IMPROVE: Automate tracking and code updates for unrecognized errors.
		// Any logged entry here should result in a code update to handle the new error.
		logger.Error().Err(err).
			Msg("Unrecognized relay error type encountered - code update needed to properly classify this error")

		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_UNKNOWN,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_UNSPECIFIED
	}
}

// classifyHttpError classifies HTTP-related errors.
// It returns the appropriate endpoint error type and sanction to be applied to the endpoint.
// Analyzes the raw error from sendHttpRelay and maps it to defined error types
func classifyHttpError(logger polylog.Logger, err error) (protocolobservations.ShannonEndpointErrorType, protocolobservations.ShannonSanctionType) {
	logger = logger.With("error_message", err.Error())

	// Endpoint returned non 2xx HTTP Status code
	if errors.Is(err, pathhttp.ErrRelayEndpointHTTPError) {
		// TODO_IMPROVE: Make this a sanction that just lasts a few blocks
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_HTTP_NON_2XX_STATUS,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_UNSPECIFIED
	}

	errStr := err.Error()

	// Connection establishment failures
	switch {
	case strings.Contains(errStr, "connection refused"):
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_HTTP_CONNECTION_REFUSED,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	case strings.Contains(errStr, "connection reset"):
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_HTTP_CONNECTION_RESET,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	case strings.Contains(errStr, "no route to host"):
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_HTTP_NO_ROUTE_TO_HOST,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	case strings.Contains(errStr, "network is unreachable"):
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_HTTP_NETWORK_UNREACHABLE,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	}

	// Transport layer errors
	switch {
	case strings.Contains(errStr, "broken pipe"):
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_HTTP_BROKEN_PIPE,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	case strings.Contains(errStr, "i/o timeout"):
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_HTTP_IO_TIMEOUT,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	case strings.Contains(errStr, "context deadline exceeded"):
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_HTTP_CONTEXT_DEADLINE_EXCEEDED,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	}

	// Connection timeout (separate from i/o timeout)
	if strings.Contains(errStr, "dial tcp") && strings.Contains(errStr, "timeout") {
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_HTTP_CONNECTION_TIMEOUT,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	}

	// HTTP protocol errors
	switch {
	case strings.Contains(errStr, "malformed HTTP"):
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_HTTP_BAD_RESPONSE,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	case strings.Contains(errStr, "invalid status"):
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_HTTP_INVALID_STATUS,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	}

	// Generic transport errors (catch-all for other transport issues)
	if strings.Contains(errStr, "transport") {
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_HTTP_TRANSPORT_ERROR,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	}

	// If we can't classify the HTTP error, it's an internal error
	logger.With(
		"err_preview", errStr[:min(100, len(errStr))],
	).Warn().Msg("Unable to classify HTTP error - defaulting to internal error")

	// SHANNON_ENDPOINT_ERROR_HTTP_UNKNOWN is the default if we have no details
	return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_HTTP_UNKNOWN,
		protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
}

// classifyMalformedEndpointPayload classifies errors found in the malformed endpoint response payload
// This is where the original error analysis data gets processed
func classifyMalformedEndpointPayload(logger polylog.Logger, payloadContent string) (protocolobservations.ShannonEndpointErrorType, protocolobservations.ShannonSanctionType) {
	logger = logger.With("payload_content_preview", payloadContent[:min(len(payloadContent), 200)])

	// Connection refused errors - most common pattern in the data (~52% of errors)
	if strings.Contains(payloadContent, "connection refused") {
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_RAW_PAYLOAD_CONNECTION_REFUSED, protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	}

	// Service not configured - second most common pattern (~17% of errors)
	if strings.Contains(payloadContent, "service endpoint not handled by relayer proxy") ||
		regexp.MustCompile(`service "[^"]+" not configured`).MatchString(payloadContent) {
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_RAW_PAYLOAD_SERVICE_NOT_CONFIGURED, protocolobservations.ShannonSanctionType_SHANNON_SANCTION_PERMANENT
	}

	// Protocol parsing errors
	if regexp.MustCompile(`proto: illegal wireType \d+`).MatchString(payloadContent) {
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_RAW_PAYLOAD_PROTOCOL_WIRE_TYPE, protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	}

	if strings.Contains(payloadContent, "proto: RelayRequest: wiretype end group") {
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_RAW_PAYLOAD_PROTOCOL_RELAY_REQUEST, protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	}

	// Unexpected EOF
	if strings.Contains(payloadContent, "unexpected EOF") {
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_RAW_PAYLOAD_UNEXPECTED_EOF, protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	}

	// Backend service errors
	if regexp.MustCompile(`backend service returned an error with status code \d+`).MatchString(payloadContent) {
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_RAW_PAYLOAD_BACKEND_SERVICE, protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	}

	// Suppliers not reachable
	if strings.Contains(payloadContent, "supplier(s) not reachable") {
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_RAW_PAYLOAD_SUPPLIERS_NOT_REACHABLE, protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	}

	// Response size exceeded
	if strings.Contains(payloadContent, "body size exceeds maximum allowed") ||
		strings.Contains(payloadContent, "response limit exceed") {
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_RAW_PAYLOAD_RESPONSE_SIZE_EXCEEDED, protocolobservations.ShannonSanctionType_SHANNON_SANCTION_UNSPECIFIED
	}

	// Server closed connection
	if strings.Contains(payloadContent, "server closed idle connection") {
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_RAW_PAYLOAD_SERVER_CLOSED_CONNECTION, protocolobservations.ShannonSanctionType_SHANNON_SANCTION_UNSPECIFIED
	}

	// TCP connection errors
	if strings.Contains(payloadContent, "write tcp") && strings.Contains(payloadContent, "connection") {
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_RAW_PAYLOAD_TCP_CONNECTION, protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	}

	// DNS resolution errors
	if strings.Contains(payloadContent, "no such host") {
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_RAW_PAYLOAD_DNS_RESOLUTION, protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	}

	// TLS handshake errors
	if strings.Contains(payloadContent, "tls") {
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_RAW_PAYLOAD_TLS_HANDSHAKE, protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	}

	// General HTTP transport errors
	if strings.Contains(payloadContent, "http:") || strings.Contains(payloadContent, "HTTP") {
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_RAW_PAYLOAD_HTTP_TRANSPORT, protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	}

	// If we can't classify the malformed payload, it's an internal error
	logger.With(
		"endpoint_payload_preview", payloadContent[:min(100, len(payloadContent))],
	).Warn().Msg("Unable to classify malformed endpoint payload - defaulting to internal error")
	return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_RAW_PAYLOAD_UNKNOWN, protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
}
