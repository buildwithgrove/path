package jsonrpc

func (re *requestError) buildObservation() *qosobservations.ValidationError {
	return &qosobservations.ValidationError {
		ErrorType: translateToRequestValidationError(re.errorKind),
		ValidationErrorDetails: re.errorDetails,
		// HTTP status code returned to the client.
		HttpStatusCode: re.jsonrpcErrorResponse.GetRecommendedHTTPStatusCode(),
	}
}

func translateToRequestValidationError(errKind requestErrorKind) qosobservations.RequestValidationErrorKind {
	switch errKind {
	case requestErrKindInternalErrReadyHTTPBody:
		return RequestValidationErrorKind_REQUEST_VALIDATION_ERROR_BODY_READ_FAILURE
	case requestErrKindJSONRPCParsingErr:
		return RequestValidationErrorKind_REQUEST_VALIDATION_ERROR_UNMARSHALING_FAILURE
	requestErrKindJSONRPCInvalidVersion
		return RequestValidationErrorKind_REQUEST_VALIDATION_ERROR_INVALID_VERSION
	case requestErrKindJSONRPCMissingMethod:
	 	return RequestValidationErrorKind_REQUEST_VALIDATION_ERROR_MISSING_METHOD
	default:
		return RequestValidationErrorKind_REQUEST_VALIDATION_ERROR_UNSPECIFIED
	}
}
