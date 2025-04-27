package framework

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

func (re *requestError) getJSONRPCResponse() *jsonrpc.Response {
	return &re.jsonrpcErrorResponse
}

func buildRequestErrorForInternalErrHTTPRead(err error) *requestError {
	return &requestError {
		errorKind: requestErrKindInternalErrReadyHTTPBody,
		errorDetails: fmt.Sprintf("error reading HTTP request body: %v", err),
		// Create JSONRPC error response for read failure
		jsonrpcErrorResponse: newJSONRPCErrResponseInternalReadError(err),
	}
}

// TODO_TECHDEBT(@adshmh): Report the protocol-level error to the QoS system to use here.
// Use these steps:
// - Update gateway.RequestQoSContext interface: add a ReportProtocolError(error) method.
// - Update requestContext: add ReportProtocolError to pass the error to the requestCtx.journal.request object.
//
// Protocol-level error: e.g. endpoint timeout has occurred.
// This is an internal error, causing a valid request to fail.
// The exact error is not known here: see the TODO_TECHDEBT above.
func buildRequestErrorForInternalErrProtocolErr(requestID jsonrpc.ID) *requestError {
	return &requestError {
		errorKind: requestErrKindInternalErrProtocolError,
		errorDetails: "error handling the request due to protocol-level error.",
		// Create JSONRPC error response for protocol error.
		jsonrpcErrorResponse: newJSONRPCErrResponseInternalProtocolError(requestID),
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
