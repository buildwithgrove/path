package shannon

import (
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
)

// builds a Shannon endpoint success observation to include:
// - endpoint details: address, url, app
// - endpoint query and response timestamps.
func buildEndpointSuccessObservation(
	logger polylog.Logger,
	endpoint endpoint,
	endpointQueryTimestamp time.Time,
	endpointResponseTimestamp time.Time,
) *protocolobservations.ShannonEndpointObservation {
	// initialize an observation with endpoint details: URL, app, etc.
	endpointObs := buildEndpointObservation(logger, endpoint)

	// Update the observation with endpoint query and response timestamps.
	endpointObs.EndpointQueryTimestamp = timestamppb.New(endpointQueryTimestamp)
	endpointObs.EndpointResponseTimestamp = timestamppb.New(endpointResponseTimestamp)

	return endpointObs
}

// builds a Shannon endpoint error observation to include:
// - endpoint details
// - the encountered error
// - any sanctions resulting from the error.
func buildEndpointErrorObservation(
	logger polylog.Logger,
	endpoint endpoint,
	endpointQueryTimestamp time.Time,
	endpointResponseTimestamp time.Time,
	errorType protocolobservations.ShannonEndpointErrorType,
	errorDetails string,
	sanctionType protocolobservations.ShannonSanctionType,
) *protocolobservations.ShannonEndpointObservation {
	// initialize an observation with endpoint details: URL, app, etc.
	endpointObs := buildEndpointObservation(logger, endpoint)

	// Update the observation with endpoint query/response timestamps.
	endpointObs.EndpointQueryTimestamp = timestamppb.New(endpointQueryTimestamp)
	endpointObs.EndpointResponseTimestamp = timestamppb.New(endpointResponseTimestamp)

	// Update the observation with error details and any resulting sanctions
	endpointObs.ErrorType = &errorType
	endpointObs.ErrorDetails = &errorDetails
	endpointObs.RecommendedSanction = &sanctionType

	return endpointObs
}

// builds a Shannon endpoint observation to include:
// endpoint: supplier, URL
// session: app, service ID, session ID, session start and end heights (using `buildEndpointObservationFromSession`).
func buildEndpointObservation(
	logger polylog.Logger,
	endpoint endpoint,
) *protocolobservations.ShannonEndpointObservation {
	// Add session fields to the observation:
	// app, serviceID, session ID, session start and end heights
	observation := buildEndpointObservationFromSession(logger, endpoint.session)

	// Add endpoint-level details: supplier, URL.
	observation.Supplier = endpoint.supplier
	observation.EndpointUrl = endpoint.url

	return observation
}

// builds an endpoint observation using session's fields, to include:
// session: app, service ID, session ID, session start/end height.
func buildEndpointObservationFromSession(
	logger polylog.Logger,
	session sessiontypes.Session,
) *protocolobservations.ShannonEndpointObservation {
	header := session.Header
	// Nil session: skip.
	if header == nil {
		logger.With("method", "buildEndpointObservationFromSession").Warn().Msg("SHOULD NEVER HAPPEN: received nil session header. Skip session fields.")
		return &protocolobservations.ShannonEndpointObservation{}
	}

	// Build an endpoint observation using session fields.
	return &protocolobservations.ShannonEndpointObservation{
		EndpointAppAddress: header.ApplicationAddress,
		SessionServiceId:   header.ServiceId,
		SessionId:          header.SessionId,
		SessionStartHeight: header.SessionStartBlockHeight,
		SessionEndHeight:   header.SessionEndBlockHeight,
	}
}

// builds a Shannon endpoint from an endpoint observation.
// Used to identify an endpoint for applying sanctions.
func buildEndpointFromObservation(
	observation *protocolobservations.ShannonEndpointObservation,
) *endpoint {
	session := buildSessionFromObservation(observation)
	return &endpoint{
		session:  session,
		supplier: observation.GetSupplier(),
		url:      observation.GetEndpointUrl(),
	}
}

// builds the details of a session from an endpoint observation.
// Used to identify an endpoint for applying sanctions.
func buildSessionFromObservation(
	observation *protocolobservations.ShannonEndpointObservation,
) sessiontypes.Session {
	return sessiontypes.Session{
		// Only Session Header is required for processing observations.
		Header: &sessiontypes.SessionHeader{
			ApplicationAddress:      observation.GetEndpointAppAddress(),
			ServiceId:               observation.GetSessionServiceId(),
			SessionId:               observation.GetSessionId(),
			SessionStartBlockHeight: observation.GetSessionStartHeight(),
			SessionEndBlockHeight:   observation.GetSessionEndHeight(),
		},
	}
}

// builds and returns a request error observation for the supplied internal error.
func buildInternalRequestProcessingErrorObservation(internalErr error) *protocolobservations.ShannonRequestError {
	return &protocolobservations.ShannonRequestError{
		ErrorType: protocolobservations.ShannonRequestErrorType_SHANNON_REQUEST_ERROR_INTERNAL,
		// Use the error message as the request error details.
		ErrorDetails: internalErr.Error(),
	}
}
