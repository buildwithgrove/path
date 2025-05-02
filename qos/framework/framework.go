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
	"github.com/buildwithgrove/path/qos/jsonrpc"
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

	// Constructs JSONRPC requests to assess endpoint eligibility to handle service requests.
	EndpointQualityChecksBuilder

	// ResultBuilders maps JSONRPC methods to custom result processing logic
	ResultBuilders map[jsonrpc.Method]EndpointQueryResultBuilder

	// StateUpdater defines how endpoint results affect service state
	StateUpdater

	// EndpointSelector defines custom endpoint selection logic
	EndpointSelector

	// TODO_MVP(@adshmh): Enable custom service QoS implementations to provide a list of allowed methods which the requestValidator needs to enforce:
	// - Uncomment the following line.
	// - Use the supplied request validator in the framework.
	// RequestValidator RequestValidator

	// TODO_FUTURE(@adshmh): Add additional configuration options:
	// - AllowedMethods: Restrict which JSONRPC methods can be processed
	// - RequestTimeout: Custom timeout for requests
	// - RetryPolicy: Configuration for request retries
	// - StateExpiryPolicy: Rules for expiring state entries
}

// NewQoSService creates a new QoS service with the given definition
func (qd *QoSDefinition) NewQoSService() *QoS {
	return &QoS{
		logger: qd.Logger,
		// set the definitions required for building different contexts.
		qosDefinition: qd,
		// initialize the service state and endpoint store.
		serviceState: &ServiceState{
			// hydrate the logger with component name: service state.
			logger: qd.Logger.With("component", "serviceState"),
			// initialize the endpoint store
			endpointStore: &endpointStore{
				logger: qd.Logger.With("component", "endpointStore"),
			},
		},
	}
}

// EndpointQueryResultBuilder processes a response and extracts the relevant result.
// It is implemented by the custom service implementations to extract result(s) from an endpoint query.
// It processes a valid JSONRPC response for a specific method and extracts the relevant data or error information.
// It can potentially mark a JSONRPC response as invalid:
// For example if the result field cannot be parsed into a number in an endpoint's response to an `eth_blockNumber` request.
type EndpointQueryResultBuilder func(*EndpointQueryResultContext) *EndpointQueryResult

// StateUpdater updates service state based on endpoint results
type StateUpdater func(*StateUpdateContext) *StateParameterUpdateSet

// EndpointSelector chooses an endpoint for a request based on service state
type EndpointSelector func(*EndpointSelectionContext) (protocol.EndpointAddr, error)

// EndpointQualityChecksBuilder constructs JSONRPC requests.
// Used to assess endpoint eligibility to handle service requests.
// Custom QoS service implements this.
// Determines what data points are needed on an endpoint, considering:
// - The existing observations on the endpoint
// - Current service state.
type EndpointQualityChecksBuilder func(*EndpointQualityChecksContext) []*jsonrpc.Request
