package framework

import (
	observations "github.com/buildwithgrove/path/observation/qos/framework"
)

// buildObservation converts an EndpointError to an observations.EndpointError
func (ee *EndpointError) buildObservation() *observations.EndpointError {
	endpointErrorObs := &observations.EndpointError{
		ErrorKind:   translateToObservationEndpointErrorKind(ee.ErrorKind),
		Description: ee.Description,
	}

	// Include sanction information if available
	if ee.RecommendedSanction != nil {
		endpointErrorObs.RecommendedSanction = ee.RecommendedSanction.buildObservation()
	}

	return endpointErrorObs
}

// buildEndpointErrorFromObservation extracts an EndpointError from an observations.EndpointError
func buildEndpointErrorFromObservation(endpointErrorObs *observations.EndpointError) *EndpointError {
	endpointErr := &EndpointError{
		ErrorKind:   translateFromObservationEndpointErrorKind(endpointErrorObs.GetErrorKind()),
		Description: endpointErrorObs.Description,
	}

	recommendedSanctionObs := endpointErrorObs.GetRecommendedSanction()
	// No sanctions: skip the rest of the processing.
	if recommendedSanctionObs == nil {
		return endpointErr
	}

	endpointErr.RecommendedSanction = buildSanctionFromObservation(recommendedSanctionObs)

	return endpointErr
}

// TODO_IN_THIS_PR: verify errorKind conversion to/from proto.
//
// DEV_NOTE: you MUST update this function when changing the set of endpoint error kinds.
func translateToObservationEndpointErrorKind(errKind EndpointErrorKind) observations.EndpointErrorKind {
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
func translateFromObservationEndpointErrorKind(errKind observations.EndpointErrorKind) EndpointErrorKind {
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
		return EndpointErrKindUnspecified
	}
}
