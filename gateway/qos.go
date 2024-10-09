package gateway

import (
	"context"
	"net/http"

	"github.com/buildwithgrove/path/message"
	"github.com/buildwithgrove/path/relayer"
)

// RequestQoSContext represents the interactions of
// the gateway with the QoS instance corresponding
// to the service specified by a service request.
//
// A RequestQoSContext can be built in various ways such as:
//   - 1. Building a new context by parsing an organic request from an end-user
//   - 2. Building a new context based on a desired endpoint check, e.g. an `eth_chainId` request on an EVM blockchain.
//   - 3. Rebuilding an existing context by deserializing a shared context from another PATH instance
type RequestQoSContext interface {
	// TODO_TECHDEBT: This should eventually return a []relayer.Payload
	// to allow mapping a single RelayRequest into multiple ServiceRequests,
	// e.g. A single batch relay request on a JSONRPC blockchain should be decomposable into
	// multiple independent requests.
	GetServicePayload() relayer.Payload

	// TODO_FUTURE: add retry-related return values to UpdateWithResponse,
	// or add retry-related methods to the interface, e.g. Failed(), ShouldRetry().
	// UpdateWithResponse is used to inform the request QoS context of the
	// payload returned by a specific endpoint in response to the service
	// payload produced (through the `GetServicePayload` method) by the
	// request QoS context instance
	UpdateWithResponse(endpointAddr relayer.EndpointAddr, endpointSerializedResponse []byte)

	// GetHTTPResponse returns the user-facing HTTP response.
	// The received response will depend on the state of the service request context,
	// which is set at the time of establishing the context,
	// and updated using the UpdateWithResponse method above.
	// e.g. Calling this on a ServiceRequestContext instance which has
	// never been updated with a response could return an HTTP response
	// with a 404 HTTP status code.
	GetHTTPResponse() HTTPResponse

	// GetObservationSet returns the list of observations resulting from
	// the response(s) received from one or more endpoints as part of fulfilling
	// the request underlying the RequestQoSContext instance.
	GetObservationSet() message.ObservationSet

	// GetEndpointSelector is part of this interface to enable more specialized endpoint
	// selection, e.g. method-based endpoint selection for an EVM blockchain service request.
	GetEndpointSelector() relayer.EndpointSelector
}

// QoSContextBuilder builds the QoS context required for handling
// all steps of a service request, e.g. generating a user-facing
// HTTP response from an endpoint's response.
// TODO_FUTURE: It only supports HTTP service requests at this point.
type QoSContextBuilder interface {
	// ParseHTTPRequest ensures that an HTTP request represents a valid request on the target service.
	ParseHTTPRequest(context.Context, *http.Request) (RequestQoSContext, bool)
}

// QoSEndpointCheckGenerator returns one or more service request contexts
// that can provide data on the quality of an enpoint by sending it the
// corresponding payloads and parsing its response.
// These checks are service-specific, i.e. the QoS instance for a
// service decides what checks should be done against an endpoint.
type QoSEndpointCheckGenerator interface {
	// TODO_FUTURE: add a GetOptionalQualityChecks() method, e.g. to enable
	// a higher level of quality of service by collecting endpoints' latency
	// in responding to certain requests.
	//
	// GetRequiredQualityChecks returns the set of quality checks required by
	// the a QoS instance to assess the validity of an endpoint.
	// The endpoint address is passed here because it allows the QoS instance to
	// make a decision based on the specific endpoint.
	// e.g. An EVM-based blockchain service QoS may decide to skip quering an endpoint on
	// its current block height if it has already failed the chain ID check.
	GetRequiredQualityChecks(relayer.EndpointAddr) []RequestQoSContext
}

// QoSPublisher is used to publish a message package's ObservationSet.
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
	QoSContextBuilder
}
