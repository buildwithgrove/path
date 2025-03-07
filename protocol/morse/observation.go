package morse

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
)

// classifyRelayError analyzes an error returned during relay and classifies it
// into one of the MorseEndpointErrorType categories along with a recommended sanction type.
// It relies on the extractErrFromRelayError function to identify the specific error type.
func classifyRelayError(logger polylog.Logger, err error) (protocolobservations.MorseEndpointErrorType, protocolobservations.MorseSanctionType) {
	if err == nil {
		return protocolobservations.MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_UNSPECIFIED,
			protocolobservations.MorseSanctionType_MORSE_SANCTION_UNSPECIFIED
	}

	// Extract the specific error type using our centralized error matching function
	extractedErr := extractErrFromRelayError(err)

	// Check for known predefined errors and map them to appropriate endpoint error types and sanctions
	// TODO_TECHDEBT(@Olshansk): Re-evaluate which errors should be session based or permanent.
	switch extractedErr {
	case ErrRelayTimeout:
		return protocolobservations.MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_TIMEOUT,
			protocolobservations.MorseSanctionType_MORSE_SANCTION_SESSION

	case ErrConnectionFailed:
		return protocolobservations.MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_CONNECTION_FAILED,
			protocolobservations.MorseSanctionType_MORSE_SANCTION_SESSION

	case ErrInvalidResponse:
		return protocolobservations.MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_INVALID_RESPONSE,
			protocolobservations.MorseSanctionType_MORSE_SANCTION_SESSION

	case ErrValidationFailed:
		return protocolobservations.MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_VALIDATION_FAILED,
			protocolobservations.MorseSanctionType_MORSE_SANCTION_PERMANENT

	case ErrNullRelayResponse:
		return protocolobservations.MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_INVALID_RESPONSE,
			protocolobservations.MorseSanctionType_MORSE_SANCTION_SESSION

	case ErrEndpointNotInSession, ErrEndpointSelectionFailed, ErrNoEndpointsAvailable, ErrEndpointNotFound:
		return protocolobservations.MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_INTERNAL,
			protocolobservations.MorseSanctionType_MORSE_SANCTION_UNSPECIFIED

	case ErrMaxedOut:
		return protocolobservations.MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_MAXED_OUT,
			protocolobservations.MorseSanctionType_MORSE_SANCTION_SESSION

	case ErrMisconfigured:
		return protocolobservations.MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_MISCONFIGURED,
			protocolobservations.MorseSanctionType_MORSE_SANCTION_SESSION
	}

	// If the error doesn't match any of our defined errors, log it and return a generic internal error.
	// TODO_IMPROVE: Find a way to make tracking these during deployment part of the (automated?) process.
	// This should never happen because any logged entry should result in code updates to handle the newly encountered error.
	logger.Error().Err(err).
		Msg("Unrecognized relay error type encountered - code update needed to properly classify this error")

	return protocolobservations.MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_INTERNAL,
		protocolobservations.MorseSanctionType_MORSE_SANCTION_UNSPECIFIED
}
