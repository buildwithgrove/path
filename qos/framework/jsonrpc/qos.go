package jsonrpc

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

	serviceState *serviceState

	// The definitoin of QoS behavior, supplied by the custom QoS service.
	qosDefinition QoSDefinition
}

// ParseHTTPRequest handles parsing an HTTP request and validating its content
// It returns a RequestQoSContext and a boolean indicating if processing should continue
func (s *QoSService) ParseHTTPRequest(
	_ context.Context,
	httpReq *http.Request,
) (*requestContext, bool) {
	requestDetails := buildRequestDetailsFromHTTP(s.logger, httpReq)

	// initialize a context for processing the HTTP request.
	requestCtx := &requestContext{
		logger: logger,
		// initialize the request journal to track all data on the request.
		journal: &requestJournal{
			requestDetails: requestDetails,
		},
	}
	
	// check if the request processing flow should continue.
	shouldContinue := requestDetails.getRequestErrorJSONRPCResponse() != nil

	rturn requestCtx, shouldContinue
}

// TODO_IN_THIS_PR: implement this method
// func (qos *QoS) ParseWebsocketRequest(_ context.Context) (gateway.RequestQoSContext, bool)

func (s *QoSService) ApplyObservations(observations *qosobservations.Observations) error
) {
//	-> Framework updates the endpoints + state as part of ApplyObservations
//	-> custom ResultBuilders return the set of attributes for the endpoint.
//	--> + expiryTime to drop endpoint attributes after expiry.
	jsonrpcSvcObservations := observations.GetJsonrpc()
	endpointResults := extractEndpointResultsFromObservations(jsonrpcSvcObservations)
	return s.serviceState.UpdateFromEndpointResults(endpointResults)
}

// buildEndpointQueryResultContext creates a context for processing endpoint queries
// The context provides:
// - Read-only access to current service state
// - Mapping of JSONRPC methods to their corresponding result builders.
func (q *QoS) buildEndpointQueryResultContext() *EndpointQueryResultContext {
	// instantiate a result context to process an endpointQuery.
	return &EndpointQueryResultContext{
		// Service State (read-only)
		// Allows the custom QoS service to base the query results on current state if needed.
		ReadonlyServiceState:        q.serviceState,

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
		*ReadonlyServiceState: rc.serviceState,
		// The endpoint selector logic defined by the custom QoS service defintion.
		customSelector:        q.qosDefinition.EndpointSelector,
	}
}

// TODO_IN_THIS_PR: implement this method.
func (q *QoS) buildServiceStateUpdateContext() *ServiceStateUpdateContext {

}
