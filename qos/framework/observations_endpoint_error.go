package framework

import (
	"time"

	observations "github.com/buildwithgrove/path/observation/qos/framework"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// buildObservation converts an EndpointError to an observations.EndpointError
func (ee *EndpointError) buildObservation() *observations.EndpointError {
	if ee == nil {
		return nil
	}

	observationError := &observations.EndpointError{
		ErrorKind:    translateToObservationRequestErrorKind(re.errorKind),
		ErrorDetails: re.errorDetails,
		// The JSONRPC response returned to the client.
		JsonRpcResponse: buildJSONRPCResponseObservation(re.jsonrpcResponse),
	}

	// Include sanction information if available
	if ee.RecommendedSanction != nil {
		observationError.Sanction = ee.RecommendedSanction.buildObservation()

	}

	return observationError
}

// extractEndpointErrorFromObservation extracts an EndpointError from an observations.EndpointError
func extractEndpointErrorFromObservation(obsError *observations.EndpointError) *EndpointError {
	if obsError == nil {
		return nil
	}

	err := &EndpointError{
		Description: obsError.Description,
		ErrorKind:   translateFromObservationErrorKind(obsError.ErrorKind),
	}

	// Include sanction information if available
	if obsError.Sanction != nil {
		err.RecommendedSanction = &Sanction{}

		// Convert sanction expiry timestamp to Duration
		if obsError.Sanction.ExpiryTimestamp != nil {
			sanctionExpiry := timeFromProto(obsError.Sanction.ExpiryTimestamp)
			err.RecommendedSanction.Duration = sanctionExpiry.Sub(time.Now())
		}
	}

	return err
}

// TODO_IN_THIS_PR: verify errorKind conversion to/from proto.
//
// DEV_NOTE: you MUST update this function when changing the set of endpoint error kinds.
func translateToObservationErrorKind(errKind EndpointErrorKind) observations.EndpointErrorKind {
	switch errKind {
	case EndpointErrKindEmptyPayload:
		return observations.EndpointErrorKind_ENDPOINT_ERROR_KIND_EMPTY_PAYLOAD
	case EndpointErrKindParseErr:
		return observations.EndpointErrorKind_ENDPOINT_ERROR_KIND_UNMARSHALING
	case EndpointErrKindValidationErr:
		return observations.EndpointErrorKind_ENDPOINT_ERROR_KIND_VALIDATION_ERR
	case EndpointErrKindInvalidResult:
		return observations.EndpointErrorKind_ENDPOINT_ERROR_KIND_INVALID_RESULT
	default:
		return observations.EndpointErrorKind_ENDPOINT_ERROR_KIND_UNSPECIFIED
	}
}

// DEV_NOTE: you MUST update this function when changing the set of endpoint error kinds.
func translateFromObservationErrorKind(errKind observations.EndpointErrorKind) EndpointErrorKind {
	switch errKind {
	case observations.EndpointErrorKind_ENDPOINT_ERROR_KIND_EMPTY_PAYLOAD:
		return EndpointErrKindEmptyPayload
	case observations.EndpointErrorKind_ENDPOINT_ERROR_KIND_UNMARSHALING:
		return EndpointErrKindParseErr
	case observations.EndpointErrorKind_ENDPOINT_ERROR_KIND_VALIDATION_ERR:
		return EndpointErrKindValidationErr
	case observations.EndpointErrorKind_ENDPOINT_ERROR_KIND_INVALID_RESULT:
		return EndpointErrKindInvalidResult
	default:
		return EndpointErrorKindUnspecified
	}
}


