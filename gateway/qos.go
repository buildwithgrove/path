package gateway

import (
	"context"
	"net/http"

	"github.com/buildwithgrove/path/relayer"
)

// TODO_UPNEXT(@adshmh): first implementation of ServiceRequestContext in the qos/evm package.
//
// ServiceRequestContext represent the interactions of the gateway with
// a service's QoS instance, in the context of a single service request.
//
// A ServiceRequestContext can be built by:
// A) Parsing an organic, i.e. originating from a user, HTTP service request, or
// B) Using an embedded endpoint data augmenting service request, e.g. an `eth_chainId` request on an EVM blockchain, or
// C) Deserializing the serialized format of a ServiceRequestContext, e.g. one shared by another PATH instance.
type ServiceRequestContext interface {
	// TODO_TECHDEBT: This should eventually return a []relayer.Payload
	// to allow mapping a single RelayRequest into multiple ServiceRequests,
	// e.g. a batch relay request on a JSONRPC blockchain.
	GetPayload() relayer.Payload

	// TODO_FUTURE: add retry-related return values to UpdateWithResponse,
	// or add retry-related methods to the interface, e.g. Failed(), ShouldRetry().
	UpdateWithResponse(relayer.EndpointAddr, []byte)

	// GetHTTPResponse returns the user-facing HTTP response.
	// The received response will depend on the state of the service request context,
	// which is set at the time of establishing the context,
	// and updated using the UpdateWithResponse method above.
	// e.g. Calling this on a ServiceRequestContext instance which has
	// never been updated with a response could return an HTTP response
	// with a 404 HTTP status code.
	GetHTTPResponse() HTTPResponse

	GetObservationSet() message.ObservationSet
}

// QoSRequestParser can build the payload to be delivered to a service endpoint.
// It only supports HTTP service requests at this point.
type QoSRequestParser interface {
	// ParseHTTPRequest ensures that an HTTP request represents a valid request on the target service.
	ParseHTTPRequest(context.Context, *http.Request) (ServiceRequestContext, bool)
}

// QoSEndpointCheckGenerator returns one or more service request contexts
// that can provide data on the quality of an enpoint by sending it the
// corresponding payloads and parsing its response.
// These checks are service-specific, i.e. the QoS instance for a
// service decides what checks should be done against an endpoint.
type QoSEndpointCheckGenerator interface {
	GetRequiredQualityChecks(relayer.EndpointAddr) []ServiceRequestContext
}

// QoSPublisher is used to publish a ServiceRequestContext.
// This is used to share QoS data between PATH instances.
type QoSPublisher interface {
	Publish(message.ObservationSet) error
}

// TODO_IMPLEMENT: Add one QoS instance per service that is to be supported by the gateway, implementing the QoSService interface below.
// e.g. a QoSService implementation for Ethereum, another for Solana, and third one for a RESTful service.
//
// QoSService represents the embedded definition of a service, e.g. a JSONRPC blockchain.
// It is broken into several pieces to clarify its responsibilities:
// 1. QoSRequestParser: Translates a service request from a supported format (currently only HTTP) into a service request context.
// 2. EndpointSelector: chooses the best endpoint for performing a particular service request.
type QoSService interface {
	QoSRequestParser
	relayer.EndpointSelector
}
