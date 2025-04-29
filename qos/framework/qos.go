package framework

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_TECHDEBT(@adshmh): Simplify the qos package by refactoring gateway.QoSContextBuilder.
// Proposed change: Create a new ServiceRequest type containing raw payload data ([]byte)
// Benefits: Decouples the qos package from HTTP-specific error handling.
//
// QoS represents a service that processes JSONRPC requests and applies QoS policies based on data returned by endpoints.
type QoS struct {
	// Logger for diagnostics
	logger polylog.Logger

	serviceState *ServiceState

	// The definitoin of QoS behavior, supplied by the custom QoS service.
	qosDefinition QoSDefinition
}

// ParseHTTPRequest handles parsing an HTTP request and validating its content
// It returns a RequestQoSContext and a boolean indicating if processing should continue
func (s *QoS) ParseHTTPRequest(
	_ context.Context,
	httpReq *http.Request,
) (*requestContext, bool) {
	// Context for processing the HTTP request.
	requestCtx := &requestContext{
		logger: s.logger,
	}

	// Initialize the request context from the HTTP request.
	shouldContinue := requestCtx.initFromHTTP(httpReq)

	return requestCtx, shouldContinue
}

// TODO_IN_THIS_PR: implement this method
// func (qos *QoS) ParseWebsocketRequest(_ context.Context) (gateway.RequestQoSContext, bool)

func (q *QoS) ApplyObservations(observations *qosobservations.Observations) error {
	serviceRequestObservations := observations.GetJsonrpc()

	// Validate the Service Name
	if serviceRequestObservations.ServiceName != q.qosDefinition.ServiceName {
		return fmt.Errorf("Reported observations mismatch: service name %q, expected %q", serviceRequestObservations.ServiceName, q.qosDefinitions.ServiceName)
	}

	// reconstruct the request journal matching the observations.
	requestJournal, err := buildRequestJournalFromObservations(q.logger, serviceRequestObservations)
	if err != nil {
		q.logger.Error().Err(err).Msg("Error building the request journal from observations: skipping the application of observations.")a
		return err
	}

	// update the stored endpoints
	updatedEndpoints := s.serviceState.updateStoredEndpoints(requestJournal.endpointQueryResults)

	// instantiate a state update context.
	stateUpdateCtx := s.buildServiceStateUpdateContext()

	// update the service state through the context, using stored endpoints.
	return stateUpdateCtx.updateFromEndpoints(updatedEndpoints)
}

// Implements gateway.QoSEndpointCheckGenerator interface
func (q *QoS) GetRequiredQualityChecks(endpointAddr protocol.EndpointAddr) []RequestQoSContext {
	endpointChecksCtx := q.buildEndpointChecksContext(endpointAddr)
	return endpointChecksCtx.BuildRequests()
}

// buildEndpointQueryResultContext creates a context for processing endpoint queries
// The context provides:
// - Read-only access to current service state
// - Mapping of JSONRPC methods to their corresponding result builders.
func (q *QoS) buildEndpointQueryResultContext(endpointQueryResult *EndpointQueryResult) *EndpointQueryResultContext {
	// instantiate a result context to process an endpointQuery.
	return &EndpointQueryResultContext{
		// Service State (read-only)
		// Allows the custom QoS service to base the query results on current state if needed.
		ServiceState:        q.serviceState,

		// Tracks the result of the endpoint query.
		EndpointQueryResult: endpointQueryResult,
		// Map of JSONRPC request method to the corresponding query result builders.
		jsonrpcMethodResultBuilders:  q.qosDefinition.ResultBuilders,
	}
}

// buildEndpointSelectionContext creates a context for endpoint validation and selection 
// The context provides:
// - Read-only access to current service state and endpoint store
// - Custom endpoint selector logic from QoS service definition
func (q *QoS) buildEndpointSelectionContext() *EndpointSelectionContext {
	return &EndpointSelectionContext{
		// Service State (read-only)
		// Allows the custom QoS service to base the validation/selection of endpoints on current state.
		// Includes the endpoint store in read-only mode.
		*ReadonlyServiceState: q.serviceState,
		// The endpoint selector logic defined by the custom QoS service defintion.
		customSelector:        q.qosDefinition.EndpointSelector,
	}
}

// TODO_IN_THIS_PR: implement this method.
func (q *QoS) buildEndpointChecksContext(endpointAddr protocol.EndpointAddr) *EndpointChecksContext {
	// Ignore the second return value: an empty endpoint is a valid value when determining the required endpoint checks.
	endpoint, _ := q.serviceState.GetEndpoint(endpointAddr)

	return &EndpointChecksContext{
		// Service State (read-only)
		// Allows the custom QoS service to base the endpoint checks on current state.
		// Includes the endpoint store in read-only mode.
		ReadonlyServiceState: q.serviceState,

		// Endpoint loaded from the endpoint store.
		Endpoint: endpoint,

		// Custom service's Endpoint Checks function
		endpointChecksBuilder: q.qosDefinition.EndpointChecksBuilder,
	}
}

// TODO_IN_THIS_PR: implement this method.
func (q *QoS) buildServiceStateUpdateContext() *ServiceStateUpdateContext {
	return &ServiceStateUpdateContext {
		ServiceState: q.ServiceState,
		// the custom service's State Updater function.
		stateUpdater: q.qosDefinition.StateUpdater,
	}

}
