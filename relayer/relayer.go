// relayer package defines the requirements and steps
// of sending relays from the perspective of:
// a) protocols, i.e. Morse and Shannon protocols, which provide:
// - a list of endpoints available for a service.
// - a function for sending a relay to a specific endpoint.
// b) gateways, which are required to provide a function for
// selecting an endpoint to which the relay is to be sent.
package relayer

import (
	"context"
	"fmt"

	"github.com/buildwithgrove/path/health"
)

// ServiceID represents a unique onchain ID for a service.
// It is defined in the `relayer` package and not the `service` package
// because `service` is intended to handle off-chain details, while
// `relayer` handles onchain details. See discussion here for more:
// https://github.com/buildwithgrove/path/pull/767#discussion_r1722001685
type ServiceID string

// AppAddr is used as the unique identifier on an onchain application.
// Both Morse and Shannon use the Application's Address for this purpose.
type AppAddr string

// App represents an onchain application on a supported protocol.
type App interface {
	Addr() AppAddr
}

// EndpointAddr is used as the unique identifier for an endpoint.
type EndpointAddr string

// Endpoint represents an entity which serves relay requests.
type Endpoint interface {
	// Addr is used to uniquely identify an endpoint
	Addr() EndpointAddr
	// PublicURL is the URL to which relay requests can be sent.
	PublicURL() string
}

// TODO_TECHDEBT: use an interace here that returns the serialized form the request:
// Payload should return the serialized form of the request to be delivered to the backend service,
// i.e. the service to which the protocol endpoint proxies relay requests.
//
// Payload currently only supports HTTP requests to an EVM blockchain (through its Data/Method/Path fields)
// TODO_DOCUMENT: add more examples, e.g. for RESTful services, as support for more types of services
// is added.
type Payload struct {
	Data            string
	Method          string
	Path            string
	TimeoutMillisec int
}

type Request struct {
	ServiceID
	AppAddr
	EndpointAddr
	Payload
}

// Response is a general purpose struct for capturing the response
// to a relay, received from an endpoint.
// TODO_FUTURE: It only supports HTTP responses for now.
type Response struct {
	// Bytes is the response to a relay received from an endpoint.
	// This can be a response to any type of RPC(GRPC, HTTP, etc.)
	Bytes []byte
	// HTTPStatusCode is the HTTP status returned by an endpoint
	// in response to a relay request.
	HTTPStatusCode int
}

// Protocol defines the core functionality of a protocol,
// from the perspective of a gateway.
// It expects a protocol to provide functions to:
// 1) List the endpoins available for sending relays for a specific service.
// 2) Send a relay to a specific endpoint and return its response.
// There are two implementations of this interface:
// - Morse: in the relayer/morse package, and
// - Shannon: in the relayer/shannon package.
type Protocol interface {
	Endpoints(ServiceID) (map[AppAddr][]Endpoint, error)
	SendRelay(Request) (Response, error)
	// All components that report their ready status to /healthz must implement the health.Check interface.
	health.Check
}

// EndpointSelector defines the functionality that the user of a relayer needs to provide,
// i.e. selecting an endpoint, from the list of available ones, to which the relay is to be sent.
type EndpointSelector interface {
	Select(map[AppAddr][]Endpoint) (AppAddr, EndpointAddr, error)
}

// Relayer defines the components and their interactions
// required for sending relays through a single entry point,
// i.e. the SendRelay method.
type Relayer struct {
	Protocol
}

// TODO_INCOMPLETE: use the supplied context to store any details, including protocol-specific details,
// which are not an immediate concern for the caller of SendRelay.
// e.g. EndpointLatency should be attached to the relay request's context, rather than being stored
// in the Response struct.
//
// SendRelay is sending a relay from the perspective of a gateway.
// It is responsible for calling Protocol.SendRelay to a specific endpoint
// for a specific application.
// It does so by calling the correct sequence of functions on
// the Relayer and the EndpointSelector.
//
// SendRelay is written as a template method to allow the customization of key steps,
// e.g. endpoint selection and protocol-specific details of sending a relay.
// See the following link for more details:
// https://en.wikipedia.org/wiki/Template_method_pattern
func (r Relayer) SendRelay(
	ctx context.Context,
	serviceID ServiceID,
	payload Payload,
	endpointSelector EndpointSelector,
) (Response, error) {
	endpoints, err := r.Protocol.Endpoints(serviceID)
	if err != nil {
		return Response{}, fmt.Errorf("Relay: error getting available endpoints for service %s: %w", serviceID, err)
	}

	appAddr, endpointAddr, err := endpointSelector.Select(endpoints)
	if err != nil {
		return Response{}, fmt.Errorf("Serve: error selecting an endpoint for service %s: %w", serviceID, err)
	}

	return r.Protocol.SendRelay(Request{
		ServiceID:    serviceID,
		AppAddr:      appAddr,
		EndpointAddr: endpointAddr,
		Payload:      payload,
	})
}
