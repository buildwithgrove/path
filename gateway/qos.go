package gateway

import (
	"context"
	"net/http"

	"github.com/buildwithgrove/path/relayer"
)

// TODO_IMPLEMENT: Add one QoS instance per service that is to be supported by the gateway, implementing the QoSService interface below.
// e.g. a QoSService implementation for Ethereum, another for Solana, and third one for a RESTful service.
//
// QoSService represents the embedded definition of a service, e.g. a JSONRPC blockchain.
// It is broken into 3 pieces to clarify its responsibilities:
// 1. QoSRequestParser: Translates a service request from a supported format (currently only HTTP) into a service payload.
// 2. QoSResponseBuilder: Builds HTTP responses from service endpoints' responses.
// 3. EndpointSelector: chooses the best endpoint for performing a particular service request.
type QoSService interface {
	QoSRequestParser
	QoSResponseBuilder
	relayer.EndpointSelector
}

// QoSRequestParser can build the payload to be delivered to a service endpoint.
// It only supports HTTP service requests at this point.
type QoSRequestParser interface {
	// TODO_TECHDEBT: This should eventually return a []ServiceRequest
	// to allow mapping a single RelayRequest into multiple ServiceRequests,
	// e.g. a batch relay request on a JSONRPC blockchain.
	//
	// ParseHTTPRequest ensures that an HTTP request represents a valid request on the target service.
	ParseHTTPRequest(context.Context, *http.Request) (relayer.Payload, error)
}

// QoSResponseBuilder builds HTTP responses from service endpoints' responses and/or errors.
type QoSResponseBuilder interface {
	// GetHTTPResponse validates the response received from a service endpoint.
	GetHTTPResponse(context.Context, relayer.Response) (HTTPResponse, error)

	// GetHTTPErrorResponse returns a service-specific error response as an HTTP response.
	// e.g. a JSONRPC blockchain service could return an HTTP response with JSONRPC-formatted payload.
	GetHTTPErrorResponse(context.Context, error) HTTPResponse
}
