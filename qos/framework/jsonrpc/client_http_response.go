package framework

import (
	"github.com/buildwithgrove/path/gateway"
)

// ClientHTTPResponse implements the gateway.HTTPResponse interface
// and provides a standardized way to return HTTP responses to clients.
type ClientHTTPResponse struct {
	StatusCode int
	Headers    map[string]string
	Payload    []byte
}

// GetPayload returns the response body payload.
func (r *ClientHTTPResponse) GetPayload() []byte {
	return r.Payload
}

// GetHTTPStatusCode returns the HTTP status code.
func (r *ClientHTTPResponse) GetHTTPStatusCode() int {
	return r.StatusCode
}

// GetHTTPHeaders returns the HTTP response headers.
func (r *ClientHTTPResponse) GetHTTPHeaders() map[string]string {
	return r.Headers
}

// newHTTPResponse creates a new HTTP response with the given status code and payload.
func newHTTPResponse(statusCode int, payload []byte) *ClientHTTPResponse {
	return &ClientHTTPResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Payload:    payload,
	}
}

// buildHTTPResponse creates an HTTP response from a JSONRPC response.
// It performs logging only if errors occur during the process.
func buildHTTPResponse(logger polylog.Logger, jsonrpcResp *jsonrpc.Response) gateway.HTTPResponse {
	if jsonrpcResp == nil {
		logger.Error().Msg("Received nil JSONRPC response")
		return buildErrorResponse(logger, jsonrpc.ID{})
	}

	payload, err := jsonrpcResp.MarshalJSON()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to marshal JSONRPC response")
		return buildErrorResponse(logger, jsonrpcResp.ID)
	}

	return &ClientHTTPResponse{
		StatusCode: jsonrpcResp.GetRecommendedHTTPStatusCode(),
		Headers:    map[string]string{"Content-Type": "application/json"},
		Payload:    payload,
	}
}

// buildErrorResponse creates an internal error HTTP response with the given ID.
func buildErrorResponse(logger polylog.Logger, id jsonrpc.ID) gateway.HTTPResponse {
	errResp := newErrResponseInternalError(id)
	errPayload, _ := errResp.MarshalJSON()
	return &ClientHTTPResponse{
		StatusCode: errResp.GetRecommendedHTTPStatusCode(),
		Headers:    map[string]string{"Content-Type": "application/json"},
		Payload:    errPayload,
	}
}

// ====================> DROP/REFACTOR these methods

// BuildInternalErrorResponse creates a generic internal error HTTP response.
func BuildInternalErrorResponse() gateway.HTTPResponse {
	errResp := newErrResponseInternalError(jsonrpc.ID{})
	errPayload, _ := MarshalErrorResponse(nil, errResp)
	return NewHTTPResponse(errResp.GetRecommendedHTTPStatusCode(), errPayload)
}

// BuildInternalErrorHTTPResponse creates an internal error HTTP response with logging.
func BuildInternalErrorHTTPResponse(logger polylog.Logger) gateway.HTTPResponse {
	errResp := newErrResponseInternalError(jsonrpc.ID{})
	errPayload, err := MarshalErrorResponse(logger, errResp)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to marshal internal error response")
	}
	return NewHTTPResponse(errResp.GetRecommendedHTTPStatusCode(), errPayload)
}

// BuildErrorHTTPResponse creates an HTTP response for an error response.
func BuildErrorHTTPResponse(logger polylog.Logger, errResp *jsonrpc.ErrorResponse) gateway.HTTPResponse {
	payload, err := MarshalErrorResponse(logger, errResp)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to marshal error response")
	}
	return NewHTTPResponse(errResp.GetRecommendedHTTPStatusCode(), payload)
}

// BuildSuccessHTTPResponse creates an HTTP response for a successful JSONRPC response.
func BuildSuccessHTTPResponse(logger polylog.Logger, jsonrpcResp *jsonrpc.JsonRpcResponse) gateway.HTTPResponse {
	response := jsonrpc.Response{
		ID:      jsonrpcResp.Id,
		JSONRPC: jsonrpc.Version2,
		Result:  jsonrpcResp.Result,
	}
	payload, err := response.MarshalJSON()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to marshal JSONRPC response")
		errResp := newErrResponseMarshalError(response.ID, err)
		errPayload, _ := MarshalErrorResponse(logger, errResp)
		return NewHTTPResponse(errResp.GetRecommendedHTTPStatusCode(), errPayload)
	}
	return NewHTTPResponse(response.GetRecommendedHTTPStatusCode(), payload)
}

// MarshalErrorResponse marshals an error response to JSON.
func MarshalErrorResponse(logger polylog.Logger, errResp *jsonrpc.ErrorResponse) ([]byte, error) {
	payload, err := errResp.MarshalJSON()
	if err != nil && logger != nil {
		logger.Error().Err(err).Msg("Failed to marshal error response")
	}
	return payload, err
}

// NewHTTPResponse creates a new HTTP response with the given status code and payload.
func NewHTTPResponse(statusCode int, payload []byte) gateway.HTTPResponse {
	return gateway.HTTPResponse{
		StatusCode: statusCode,
		Body:       payload,
	}
}
