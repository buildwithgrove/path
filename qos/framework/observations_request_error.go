package framework

import (
	observations "github.com/buildwithgrove/path/observation/qos/framework"
)

func (re *requestError) buildObservation() *observations.RequestError {
	return &observations.RequestError{
		ErrorKind:    translateToObservationRequestErrorKind(re.errorKind),
		ErrorDetails: re.errorDetails,
		// The JSONRPC response returned to the client.
		JsonRpcResponse: buildJSONRPCResponseObservation(re.jsonrpcResponse),
	}
}

func buildRequestErrorFromObservation(obs *observations.RequestError) *requestError {
	return &requestErro {
		errorKind: translateFromObservationRequestErrorKind(obs.ErrorKind()),
		errorDetails: obs.GetErrorDetails(),
		jsonrpcErrorResponse: buildJSONRPCResponseFromObservation(obs.GetJsonRpcResponse()),
	}
}

// DEV_NOTE: you MUST update this function when changing the set of request errors.
func translateToObservationRequestErrorKind(errKind requestErrorKind) observations.RequestErrorKind {
	switch errKind {
	case requestErrKindInternalReadyHTTPBody:
		return observations.RequestValidationErrorKind_REQUEST_ERROR_INTERNAL_BODY_READ_FAILURE
	case requestErrKindInternalProtocolError:
		return observations.RequestValidationErrorKind_REQUEST_ERROR_INTERNAL_PROTOCOL_ERROR
	case requestErrKindJSONRPCParsingErr:
		return observations.RequestValidationErrorKind_REQUEST_ERROR_VALIDATION_UNMARSHALING_FAILURE
	case requestErrKindJSONRPCInvalidVersion:
		return observations.RequestValidationErrorKind_REQUEST_ERROR_VALIDATION_INVALID_VERSION
	case requestErrKindJSONRPCMissingMethod:
		return observations.RequestValidationErrorKind_REQUEST_ERROR_VALIDATION_MISSING_METHOD
	default:
		return observations.RequestValidationErrorKind_REQUEST_ERROR_VALIDATION_UNSPECIFIED
	}
}
