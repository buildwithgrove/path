package framework

import (
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

type requestErrorKind int

const (
	_ requestErrorKind = iota // skip the 0 value: it matches the "UNSPECIFIED" enum value in proto definitions.
	requestErrKindInternalErrReadyHTTPBody
	requestErrKindInternalProtocolErr
	requestErrKindJSONRPCParsingErr
	requestErrKindJSONRPCValidationErr
)

// TODO_FUTURE(@adshmh): Consider making requestError public.
// This would allow custom QoS to reject valid JSONRPC requests.
// e.g. reject a JSONRPC request with an unsupported method.
// 
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
	return &requestError{
		errorKind:    requestErrKindInternalErrReadyHTTPBody,
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
// No endpoint responses are reported to the QoS.
// This is an internal error, causing a valid request to fail.
// The exact error is not known here: see the TODO_TECHDEBT above.
func buildRequestErrorForInternalErrProtocolErr(requestID jsonrpc.ID) *requestError {
	return &requestError{
		errorKind:    requestErrKindInternalErrProtocolError,
		errorDetails: "error handling the request due to protocol-level error.",
		// Create JSONRPC error response for protocol error.
		jsonrpcErrorResponse: newJSONRPCErrResponseInternalProtocolError(requestID),
	}
}

func buildRequestErrorForParseError(err error) *requestError {
	return &requestError{
		errorKind:    requestErrKindJSONRPCParsingErr,
		errorDetails: fmt.Sprintf("error parsing HTTP request into JSONRPC: %v", err),
		// Create JSONRPC error response for parse failure
		jsonrpcErrorResponse: newJSONRPCErrResponseParseError(err),
	}
}

func buildRequestErrorJSONRPCValidationError(requestID jsonrpc.ID, validationErr error) *requestError {
	return &requestError{
		errorKind:    requestErrKindJSONRPCValidationErr,
		errorDetails: fmt.Sprintf("JSONRPC request failed validation: %s", validationErr.Error()),
		// Create JSONRPC error response for parse failure
		jsonrpcErrorResponse: newJSONRPCErrResponseJSONRPCRequestValidationError(requestID, validationErr),
	}
}
