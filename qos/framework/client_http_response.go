package framework

import (
	"encoding/json"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/qos/jsonrpc"
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
func buildHTTPResponse(
	logger polylog.Logger,
	jsonrpcResp *jsonrpc.Response,
) gateway.HTTPResponse {
	if jsonrpcResp == nil {
		logger.Error().Msg("Received nil JSONRPC response")
		return buildErrorResponse(logger, jsonrpc.ID{})
	}

	payload, err := json.Marshal(jsonrpcResp)
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
	errPayload, _ := json.Marshal(errResp)
	return &ClientHTTPResponse{
		StatusCode: errResp.GetRecommendedHTTPStatusCode(),
		Headers:    map[string]string{"Content-Type": "application/json"},
		Payload:    errPayload,
	}
}
