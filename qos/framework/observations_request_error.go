package framework

import (
	observations "github.com/buildwithgrove/path/observation/qos/framework"
)

func (re *requestError) buildObservation() *observations.RequestError {
	return &observations.RequestError{
		ErrorKind:    translateToObservationRequestErrorKind(re.errorKind),
		ErrorDetails: re.errorDetails,
		// The JSONRPC response returned to the client.
		JsonRpcResponse: buildObservationFromJSONRPCResponse(re.jsonrpcErrorResponse),
	}
}

func buildRequestErrorFromObservation(obs *observations.RequestError) *requestError {
	return &requestError{
		errorKind:            translateFromObservationRequestErrorKind(obs.GetErrorKind()),
		errorDetails:         obs.GetErrorDetails(),
		jsonrpcErrorResponse: buildJSONRPCResponseFromObservation(obs.GetJsonRpcResponse()),
	}
}

// DEV_NOTE: you MUST update this function when changing the set of request errors.
func translateToObservationRequestErrorKind(errKind requestErrorKind) observations.RequestErrorKind {
	switch errKind {
	case requestErrKindInternalErrReadyHTTPBody:
		return observations.RequestErrorKind_REQUEST_ERROR_INTERNAL_BODY_READ_FAILURE
	case requestErrKindInternalProtocolError:
		return observations.RequestErrorKind_REQUEST_ERROR_INTERNAL_PROTOCOL_ERROR
	case requestErrKindJSONRPCParsingError:
		return observations.RequestErrorKind_REQUEST_ERROR_UNMARSHALING_ERROR
	case requestErrKindJSONRPCValidationError:
		return observations.RequestErrorKind_REQUEST_ERROR_JSONRPC_VALIDATION_ERROR
	default:
		return observations.RequestErrorKind_REQUEST_ERROR_UNSPECIFIED
	}
}

// translateFromObservationRequestErrorKind converts proto enum to Go enum:
// - Maps proto validation error kinds to their local equivalents
// - Handles unknown values with unspecified default
func translateFromObservationRequestErrorKind(errKind observations.RequestErrorKind) requestErrorKind {
	switch errKind {
	case observations.RequestErrorKind_REQUEST_ERROR_INTERNAL_BODY_READ_FAILURE:
		return requestErrKindInternalErrReadyHTTPBody
	case observations.RequestErrorKind_REQUEST_ERROR_INTERNAL_PROTOCOL_ERROR:
		return requestErrKindInternalProtocolError
	case observations.RequestErrorKind_REQUEST_ERROR_UNMARSHALING_ERROR:
		return requestErrKindJSONRPCParsingError
	case observations.RequestErrorKind_REQUEST_ERROR_JSONRPC_VALIDATION_ERROR:
		return requestErrKindJSONRPCValidationError
	default:
		return requestErrKindUnspecified
	}
}
