package framework

func (re *requestError) buildObservation() *qosobservations.ValidationError {
	return &qosobservations.ValidationError {
		ErrorType: translateToRequestValidationError(re.errorKind),
		ValidationErrorDetails: re.errorDetails,
		// HTTP status code returned to the client.
		HttpStatusCode: re.jsonrpcErrorResponse.GetRecommendedHTTPStatusCode(),
	}
}
// DEV_NOTE: you MUST update this function when changing the set of request errors.
func translateToRequestError(errKind requestErrorKind) qosobservations.RequestErrorKind {
	switch errKind {
	case requestErrKindInternalReadyHTTPBody:
		return RequestValidationErrorKind_REQUEST_ERROR_INTERNAL_BODY_READ_FAILURE
	case requestErrKindInternalProtocolError:
		return RequestValidationErrorKind_REQUEST_ERROR_INTERNAL_PROTOCOL_ERROR
	case requestErrKindJSONRPCParsingErr:
		return RequestValidationErrorKind_REQUEST_ERROR_VALIDATION_UNMARSHALING_FAILURE
	requestErrKindJSONRPCInvalidVersion
		return RequestValidationErrorKind_REQUEST_ERROR_VALIDATION_INVALID_VERSION
	case requestErrKindJSONRPCMissingMethod:
	 	return RequestValidationErrorKind_REQUEST_ERROR_VALIDATION_MISSING_METHOD
	default:
		return RequestValidationErrorKind_REQUEST_ERROR_VALIDATION_UNSPECIFIED
	}
}
