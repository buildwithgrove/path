package morse

import (
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"google.golang.org/protobuf/types/known/timestamppb"

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
	case ErrRelayRequestTimeout:
		return protocolobservations.MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_TIMEOUT,
			protocolobservations.MorseSanctionType_MORSE_SANCTION_SESSION

	case ErrConnectionFailed:
		return protocolobservations.MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_CONNECTION_FAILED,
			protocolobservations.MorseSanctionType_MORSE_SANCTION_SESSION

	case ErrInvalidResponse:
		return protocolobservations.MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_INVALID_RESPONSE,
			protocolobservations.MorseSanctionType_MORSE_SANCTION_SESSION

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

	case ErrTLSCertificateVerificationFailed:
		return protocolobservations.MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_TLS_CERTIFICATE_VERIFICATION_FAILED,
			protocolobservations.MorseSanctionType_MORSE_SANCTION_SESSION

	case ErrNonJSONResponse:
		return protocolobservations.MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_NON_JSON_RESPONSE,
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

// builds a Morse endpoint success observation to include:
// - endpoint details: address, url, app
// - endpoint query and response timestamps.
func buildEndpointSuccessObservation(
	endpoint endpoint,
	endpointQueryTimestamp time.Time,
	endpointResponseTimestamp time.Time,
) *protocolobservations.MorseEndpointObservation {
	// initialize an observation with endpoint details: URL, app, etc.
	endpointObs := buildEndpointObservation(endpoint)

	// Update the observation with endpoint query and response timestamps.
	endpointObs.EndpointQueryTimestamp = timestamppb.New(endpointQueryTimestamp)
	endpointObs.EndpointResponseTimestamp = timestamppb.New(endpointResponseTimestamp)

	return endpointObs
}

// builds a Morse endpoint error observation to include:
// - endpoint details: address, url, app, query/response timestamps.
// - the encountered error
// - any sanctions resulting from the error.
func buildEndpointErrorObservation(
	endpoint endpoint,
	endpointQueryTimestamp time.Time,
	endpointResponseTimestamp time.Time,
	errorType protocolobservations.MorseEndpointErrorType,
	errorDetails string,
	sanctionType protocolobservations.MorseSanctionType,
) *protocolobservations.MorseEndpointObservation {
	// initialize an observation with endpoint details: URL, app, etc.
	endpointObs := buildEndpointObservation(endpoint)

	// Update the observation with endpoint query/response timestamps.
	endpointObs.EndpointQueryTimestamp = timestamppb.New(endpointQueryTimestamp)
	endpointObs.EndpointResponseTimestamp = timestamppb.New(endpointResponseTimestamp)

	// Update the observation with error details and any resulting sanctions
	endpointObs.ErrorType = &errorType
	endpointObs.ErrorDetails = &errorDetails
	endpointObs.RecommendedSanction = &sanctionType

	return endpointObs
}

// builds a Morse endpoint observation to include:
// - endpoint details: address, url, app
// - session details: key, height
func buildEndpointObservation(
	endpoint endpoint,
) *protocolobservations.MorseEndpointObservation {
	return &protocolobservations.MorseEndpointObservation{
		AppAddress:       endpoint.app.Addr(),
		AppPublicKey:     endpoint.app.publicKey,
		SessionKey:       endpoint.session.Key,
		SessionServiceId: endpoint.session.Header.Chain,
		SessionHeight:    int32(endpoint.session.Header.SessionHeight),
		EndpointAddr:     endpoint.address,
		EndpointUrl:      endpoint.url,
	}
}

// builds and returns a request error observation for the supplied internal error.
func buildInternalRequestProcessingErrorObservation(internalErr error) *protocolobservations.MorseRequestError {
	return &protocolobservations.MorseRequestError{
		ErrorType: protocolobservations.MorseRequestErrorType_MORSE_REQUEST_ERROR_INTERNAL,
		// Use the error message as the request error details.
		ErrorDetails: internalErr.Error(),
	}
}
