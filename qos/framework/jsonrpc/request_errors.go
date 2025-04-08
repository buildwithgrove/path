package jsonrpc

type requestErrorKind int
const (
	_ requestErrorKind = iota // skip the 0 value: it matches the "UNSPECIFIED" enum value in proto definitions.
	requestErrKindInternalErrReadyHTTPBody
	requestErrKindJSONRPCParsingErr
	requestErrKindJSONRPCInvalidVersion
	requestErrKindJSONRPCMissingMethod
)

type requestError struct {
	// Captures the kind of error the request encountered.
	// e.g. error parsing HTTP payload into a JSONRPC request.
	errorKind requestErrorKind

	// Stores a description of the request error.
	errorDetails string

	// Error response to return if a request parsing error occurred:
	// - error reading HTTP request's body.
	// - error parsing the request's payload into a jsonrpc.Request struct.
	jsonrpcErrorResponse jsonrpc.Response
}

func buildRequestErrorForInternalErrHTTPRead(err error) *requestError {
	return &requestError {
		errorKind: requestErrKindInternalErrReadyHTTPBody,
		errorDetails: fmt.Sprintf("error reading HTTP request body: %v", err),
		// Create JSONRPC error response for read failure
		jsonrpcErrorResponse: newJSONRPCErrResponseInternalReadError(err),
	}
}

func buildRequestErrorForParseError(err error) *requestError {
	return &requestError {
		errorKind: requestErrKindJSONRPCParsingErr,
		errorDetails: fmt.Sprintf("error parsing HTTP request into JSONRPC: %v", err),
		// Create JSONRPC error response for parse failure
		jsonrpcErrorResponse: newJSONRPCErrResponseParseError(err),
	}
}

func buildRequestErrorJSONRPCErrInvalidVersion(requestID jsonrpc.ID, version jsonrpc.Version) *requestError {
	err := fmt.Errorf("invalid version in JSONRPC request: %s", version)

	return &requestError {
		errorKind: requestErrKindJSONRPCInvalidVersion,
		errorDetails: err.Error(),
		// Create JSONRPC error response for parse failure
		jsonrpcErrorResponse: newJSONRPCErrResponseInvalidVersion(err, requestID),
	}
}

func buildRequestErrorJSONRPCErrMissingMethod(requestID jsonrpc.Request) *requestError {
	return &requestError {
		errorKind: requestErrKindJSONRPCMissingMethod,
		errorDetails: "No method specified by the JSONRPC request",
		// Create JSONRPC error response for parse failure
		jsonrpcErrorResponse: newJSONRPCErrResponseMissingMethod(requestID),
	}
}
