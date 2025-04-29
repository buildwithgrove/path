// Package jsonrpc provides a framework for implementing Quality of Service (QoS) for JSONRPC-based services.
//
// Key components:
// - Context-based processing for standardizing service interactions
// - Custom endpoint selection based on service state
// - Custom result processing and extraction
// - Service state management with observability
//
// Users implement the QoSDefinition interface to create custom QoS services that
// leverage the framework's request handling, endpoint management, and state tracking.
package framework

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
)

// TODO_MVP(@adshmh): Allow custom QoS services to supply custom request validation logic.
// Example use case: specifying a list of allowed JSONRPC request methods.
// This would require:
// 1. Declaring a public RequestValidator interface.
// 2. Helper functions, e.g. BuildRequestValidatorForAllowedMethods.
//
// TODO_FUTURE(@adshmh): Provide reasonable defaults for components to enable a no-config JSONRPC service QoS.
//
// QoSDefinition contains all custom behavior for a JSONRPC QoS service.
// Implementers must provide all of the customization functions below.
type QoSDefinition struct {
	// Logger for service logs. If nil, a default logger is used
	Logger polylog.Logger

	// ServiceName identifies and describes the service.
	// e.g. "ETH"
	ServiceName string

	// ResultBuilders maps JSONRPC methods to custom result processing logic
	ResultBuilders map[string]EndpointQueryResultBuilder

	// StateUpdater defines how endpoint results affect service state
	StateUpdater StateUpdater

	// EndpointSelector defines custom endpoint selection logic
	EndpointSelector EndpointSelector

	// TODO_MVP(@adshmh): Enable custom service QoS implementations to provide a list of allowed methods which the requestValidator needs to enforce:
	// - Uncomment the following line.
	// - Use the supplied request validator in the framework.
	// RequestValidator RequestValidator

	// TODO_FUTURE(@adshmh): Add additional configuration options:
	// - InitialState: Starting values for service state
	// - AllowedMethods: Restrict which JSONRPC methods can be processed
	// - RequestTimeout: Custom timeout for requests
	// - RetryPolicy: Configuration for request retries
	// - StateExpiryPolicy: Rules for expiring state entries
	// - MetricsCollection: Settings for performance metrics
}

// EndpointQueryResultBuilder processes a response and extracts the relevant result.
// It is implemented by the custom service implementations to extract result(s) from an endpoint query.
// It processes a valid JSONRPC response for a specific method and extracts the relevant data or error information.
// It can potentially mark a JSONRPC response as invalid:
// For example if the result field cannot be parsed into a number in an endpoint's response to an `eth_blockNumber` request.
type EndpointQueryResultBuilder func(ctx *EndpointQueryResultContext) *EndpointQueryResult

// StateUpdater updates service state based on endpoint results
type StateUpdater func(ctx *StateUpdateContext) *StateParameterUpdateSet

// EndpointSelector chooses an endpoint for a request based on service state
type EndpointSelector func(ctx *EndpointSelectionContext) (protocol.EndpointAddr, error)

// NewQoSService creates a new QoS service with the given definition
func NewQoSService(def QoSDefinition) *QoS {
	// TODO_IN_THIS_PR: instantiate the framrwork using the QoSDeinition struct.
}
