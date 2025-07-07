package shannon

import (
	"errors"

	"github.com/pokt-network/poktroll/pkg/polylog"
	sdk "github.com/pokt-network/shannon-sdk"

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

	switch {
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
	}

	// No known error matched:
	// Fallback to error matching using the error string.
	// Extract the specific error type using centralized error matching.
	extractedErr := extractErrFromRelayError(err)

	// Map known errors to endpoint error types and sanctions.
	// TODO_TECHDEBT(@Olshansk): Re-evaluate which errors should be session-based or permanent.
	switch extractedErr {

	// Endpoint Configuration error
	case ErrRelayEndpointConfig:
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_CONFIG,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION

	// endpoint timeout error
	case ErrRelayEndpointTimeout:
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_TIMEOUT,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION

	case ErrContextCancelled:
		return protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_REQUEST_CANCELED_BY_PATH,
			protocolobservations.ShannonSanctionType_SHANNON_SANCTION_DO_NOT_SANCTION

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
