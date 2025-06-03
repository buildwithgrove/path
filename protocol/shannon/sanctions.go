package shannon

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
)

// classifyRelayError determines the ShannonEndpointErrorType and recommended ShannonSanctionType for a given relay error.
//
// - Uses extractErrFromRelayError to identify the specific error type.
// - Maps known errors to endpoint error types and sanctions.
// - Logs and returns a generic internal error for unknown cases.
func classifyRelayError(logger polylog.Logger, err error) (protocolobservations.ShannonEndpointErrorType, protocolobservations.ShannonSanctionType) {
	if err == nil {
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_UNSPECIFIED,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_UNSPECIFIED
	}

	// Extract the specific error type using centralized error matching.
	extractedErr := extractErrFromRelayError(err)

	// Map known errors to endpoint error types and sanctions.
	// TODO_TECHDEBT(@Olshansk): Re-evaluate which errors should be session-based or permanent.
	switch extractedErr {

	// Endpoint Configuration error
	case RelayErrEndpointConfigError:
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_CONFIG,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION

	// endpoint timeout error
	case RelayErrEndpointTimeout:
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_TIMEOUT,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION

	// endpoint unexpected EOF error
	case RelayErrEndpointUnexpectedEOF:
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_INTERNAL,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION

	// endpoint protocol parsing error
	case RelayErrEndpointProtocolError:
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_INTERNAL,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION

	// endpoint invalid session header error
	case RelayErrEndpointInvalidSession:
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_INTERNAL,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION

	default:
		// Unknown error: log and return generic internal error.
		// TODO_IMPROVE: Automate tracking and code updates for unrecognized errors.
		// Any logged entry here should result in a code update to handle the new error.
		logger.Error().Err(err).
			Msg("Unrecognized relay error type encountered - code update needed to properly classify this error")

		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_INTERNAL,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_UNSPECIFIED
	}
}
