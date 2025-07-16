package gateway

import (
	"context"
	"net/http"

	"github.com/buildwithgrove/path/metrics/devtools"
	"github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
)

// RequestQoSContext
//
// Represents interactions between the gateway and the QoS instance for a given service request.
//
// Construction methods:
// - Parse an organic request from an end-user.
// - Rebuild from a shared context deserialized from another PATH instance.
type RequestQoSContext interface {
	// TODO_TECHDEBT: Should eventually return []Payload
	// - Allows mapping a single RelayRequest into multiple ServiceRequests.
	// - Example: A batch relay request on JSONRPC should decompose into multiple independent requests.
	GetServicePayload() protocol.Payload

	// TODO_FUTURE:
	// - Add retry-related return values to UpdateWithResponse,
	//   or add retry-related methods (e.g., Failed(), ShouldRetry()).
	//
	// UpdateWithResponse:
	// - Informs the request QoS context of the payload returned by a specific endpoint.
	// - Response is for the service payload produced by GetServicePayload.
	UpdateWithResponse(endpointAddr protocol.EndpointAddr, endpointSerializedResponse []byte)

	// GetHTTPResponse:
	// - Returns the user-facing HTTP response.
	// - Response depends on the current state of the service request context.
	// - State is set at context creation and updated via UpdateWithResponse.
	// - If never updated, may return 404 HTTP status.
	GetHTTPResponse() HTTPResponse

	// GetObservations:
	// - Returns QoS-level observations in the context.
	//
	// Example:
	//   Context:
	//     - Service: Solana
	//     - SelectedEndpoint: `endpoint_101`
	//     - Request: `getHealth`
	//     - Endpoint response: error
	//   Observation:
	//     - `endpoint_101` is unhealthy.
	GetObservations() qos.Observations

	// GetEndpointSelector:
	// - Enables specialized endpoint selection (e.g., method-based selection for EVM requests).
	GetEndpointSelector() protocol.EndpointSelector
}

// QoSContextBuilder
//
// Builds the QoS context required for all steps of a service request.
// Example: Generate a user-facing HTTP response from an endpoint's response.
type QoSContextBuilder interface {
	// ParseHTTPRequest:
	// - Ensures the HTTP request is valid for the target service.
	ParseHTTPRequest(context.Context, *http.Request) (RequestQoSContext, bool)

	// ParseWebsocketRequest:
	// - Ensures a WebSocket request is valid for the target service.
	// - WebSocket connection requests have no body, so no parsing needed.
	// - If service supports WebSocket, returns a valid RequestQoSContext.
	ParseWebsocketRequest(context.Context) (RequestQoSContext, bool)
}

// QoSEndpointCheckGenerator
//
// Returns one or more service request contexts that:
// - Provide data on endpoint quality by sending payloads and parsing responses.
// - Checks are service-specific; the QoS instance decides what checks to run.
type QoSEndpointCheckGenerator interface {
	// TODO_FUTURE:
	// - Add GetOptionalQualityChecks() to collect additional QoS data (e.g., endpoint latency).
	//
	// GetRequiredQualityChecks:
	// - Returns required quality checks for a QoS instance to assess endpoint validity.
	// - Example: EVM QoS may skip block height check if chain ID check already failed.
	GetRequiredQualityChecks(protocol.EndpointAddr) []RequestQoSContext
}

// TODO_IMPLEMENT:
// - Add a QoS instance per service supported by the gateway (e.g., Ethereum, Solana, RESTful).
//
// QoSService:
// - Represents the embedded definition of a service (e.g., JSONRPC blockchain).
// - Responsibilities:
//  1. QoSRequestParser: Translates service requests (currently only HTTP) into service request contexts.
//  2. EndpointSelector: Chooses the best endpoint for a specific service request.
type QoSService interface {
	QoSContextBuilder
	QoSEndpointCheckGenerator

	// ApplyObservations:
	// - Applies QoS-related observations to the local QoS instance.
	// - TODO_FUTURE: Observations can be:
	//   - "local": from requests sent to an endpoint by THIS PATH instance.
	//   - "shared": from QoS observations shared by OTHER PATH instances.
	ApplyObservations(*qos.Observations) error

	// HydrateDisqualifiedEndpointsResponse:
	// - Fills the disqualified endpoint response with QoS-specific data.
	HydrateDisqualifiedEndpointsResponse(protocol.ServiceID, *devtools.DisqualifiedEndpointResponse)
}
