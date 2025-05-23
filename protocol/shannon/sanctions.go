package shannon

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
)

// classifyRelayError analyzes an error returned during relay and classifies it
// into one of the ShannonEndpointErrorType categories along with a recommended sanction type.
// It relies on the extractErrFromRelayError function to identify the specific error type.
func classifyRelayError(logger polylog.Logger, err error) (protocolobservations.ShannonEndpointErrorType, protocolobservations.ShannonSanctionType) {
	if err == nil {
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_UNSPECIFIED,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_UNSPECIFIED
	}

	// Extract the specific error type using our centralized error matching function
	extractedErr := extractErrFromRelayError(err)

	// Check for known predefined errors and map them to appropriate endpoint error types and sanctions
	// TODO_TECHDEBT(@Olshansk): Re-evaluate which errors should be session based or permanent.
	switch extractedErr {
	// Endpoint Configuration error
	case RelayErrEndpointConfigError:
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_CONFIG,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION

	// endpoint timeout error
	case RelayErrEndpointTimeout:
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_TIMEOUT,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION
	}

	// If the error doesn't match any of our defined errors, log it and return a generic internal error.
	// TODO_IMPROVE: Find a way to make tracking these during deployment part of the (automated?) process.
	// This should never happen because any logged entry should result in code updates to handle the newly encountered error.
	logger.Error().Err(err).
		Msg("Unrecognized relay error type encountered - code update needed to properly classify this error")

	return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_INTERNAL,
		protocolobservations.ShannonSanctionType_SHANNON_SANCTION_UNSPECIFIED
}
