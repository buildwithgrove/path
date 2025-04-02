package gateway

import (
	"context"
	"net/http"

	"github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
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
	// TODO_TECHDEBT: This should eventually return a []Payload
	// to allow mapping a single RelayRequest into multiple ServiceRequests,
	// e.g. A single batch relay request on a JSONRPC blockchain should be decomposable into
	// multiple independent requests.
	GetServicePayload() protocol.Payload

	// TODO_FUTURE: add retry-related return values to UpdateWithResponse,
	// or add retry-related methods to the interface, e.g. Failed(), ShouldRetry().
	// UpdateWithResponse is used to inform the request QoS context of the
	// payload returned by a specific endpoint in response to the service
	// payload produced (through the `GetServicePayload` method) by the
	// request QoS context instance
	UpdateWithResponse(endpointAddr protocol.EndpointAddr, endpointSerializedResponse []byte)

	// GetHTTPResponse returns the user-facing HTTP response.
	// The received response will depend on the state of the service request context,
	// which is set at the time of establishing the context,
	// and updated using the UpdateWithResponse method above.
	// e.g. Calling this on a ServiceRequestContext instance which has
	// never been updated with a response could return an HTTP response
	// with a 404 HTTP status code.
	GetHTTPResponse() HTTPResponse

	// GetObservations returns the set of QoS-level observations contained in the context.
	//
	// Hypothetical illustrative example.
	//
	// If the context is:
	// 	- Service: Solana
	// 	- SelectedEndpoint: `endpoint_101`
	// 	- Request: `getHealth`
	// 	- Endpoint response: an error
	//
	// Then the observation can be:
	// 	- `endpoint_101` is unhealthy.
	GetObservations() qos.Observations

	// GetEndpointSelector is part of this interface to enable more specialized endpoint
	// selection, e.g. method-based endpoint selection for an EVM blockchain service request.
	GetEndpointSelector() protocol.EndpointSelector
}

// QoSContextBuilder builds the QoS context required for handling
// all steps of a service request, e.g. generating a user-facing
// HTTP response from an endpoint's response.
type QoSContextBuilder interface {
	// ParseHTTPRequest ensures that an HTTP request represents a valid request on the target service.
	ParseHTTPRequest(context.Context, *http.Request) (RequestQoSContext, bool)

	// ParseWebsocketRequest ensures that a WebSocket request represents a valid request on the target service.
	// WebSocket connection requests do not have a body so there is no need to parse anything.
	// As long as the service supports WebSocket connections, this method should return a valid RequestQoSContext.
	ParseWebsocketRequest(context.Context) (RequestQoSContext, bool)
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
	// e.g. An EVM-based blockchain service QoS may decide to skip querying an endpoint on
	// its current block height if it has already failed the chain ID check.
	GetRequiredQualityChecks(protocol.EndpointAddr) []RequestQoSContext
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
	QoSEndpointCheckGenerator

	// ApplyObservations is used to apply QoS-related observations to the local QoS instance.
	// The observations can be either of:
	// 	- "local": from requests sent to an endpoint by **THIS** PATH instance
	// 	- "shared": from QoS observations shared by **OTHER** PATH instances.
	ApplyObservations(*qos.Observations) error
}
