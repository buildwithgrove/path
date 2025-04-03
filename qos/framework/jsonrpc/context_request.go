package jsonrpc

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// TODO_REFACTOR: Improve naming clarity by distinguishing between interfaces and adapters
// in the metrics/qos/evm and qos/evm packages, and elsewhere names like `response` are used.
// Consider renaming:
//   - metrics/qos/evm: response → EVMMetricsResponse
//   - qos/evm: response → EVMQoSResponse
//   - observation/evm: observation -> EVMObservation
//
// TODO_TECHDEBT: Need to add a Validate() method here to allow the caller (e.g. gateway)
// determine whether the endpoint's response was valid, and whether a retry makes sense.
//

const (
	// TODO_MVP(@adshmh): Support individual configuration of timeout for every service that uses EVM QoS.
	// The default timeout when sending a request to an EVM blockchain endpoint.
	defaultServiceRequestTimeoutMillisec = 10000
)

// requestContext provides the support required by the gateway
// package for handling service requests.
var _ gateway.RequestQoSContext = &requestContext{}

// TODO_IN_THIS_PR: change the errorKind to private + find the correct file for it.

// TODO_FUTURE(@adshmh): implement custom, typed result extractors that are commonly used by custom QoS implementations.
// Example:
// ResultContext.
// endpointErrorKind identifies different kinds of endpoint data errors.
type endpointErrorKind int

const (
	EndpointDataErrorKindUnspecified   endpointErrorKind = iota
	EndpointDataErrorKindNoInteraction                   // No endpoint interaction occurred or no payload received
	EndpointDataErrorKindEmptyPayload                    // Empty payload from endpoint
	EndpointDataErrorKindUnmarshaling                    // Could not parse endpoint payload
	EndpointDataErrorKindInvalidResult                   // Payload result doesn't match expected format
)

// TODO_IN_THIS_PR: sort out the scope of fields and methods: private/public on private structs.
//
// requestQoSContext holds the context for a request through its lifecycl.
// It contains all the state needed to process the request, build responses, and generate observations.
type requestQoSContext struct {
	Logger polylog.Logger

	// Service Identification fields
	ServiceID ServiceID

	// Request is the JSONRPC request that was sent
	Request *jsonrpc.Request

	// Error response to return if validation failed
	JSONRPCErrorResponse *jsonrpc.Response

	// Read-only form of the service state.
	// Used to instantiate EndpointQueryResultContext and EndpointSelectionContext.
	serviceState *ServiceStateReadOnly

	// Used to instantiate the EndpointSelectionContext, to select an endpoint for serving the client's request.
	endpointStore *endpointStore

	// Used to instantiate the EndpointSelectionContext
	customEndpointSelector EndpointSelector

	// Used to instantiate the EndpointResultContext, to build an endpoint result from an endpoint query.
	jsonrpcMethodResultBuilders map[string]EndpointResultBuilder

	// Tracks results processed in the current request context.
	processedResults []*EndpointQueryResult
}

// TODO_MVP(@adshmh): Ensure the JSONRPC request struct can handle all valid service requests.
func (rc requestQoSContext) GetServicePayload() protocol.Payload {
	reqBz, err := json.Marshal(*rc.Request)
	if err != nil {
		// TODO_MVP(@adshmh): find a way to guarantee this never happens,
		// e.g. by storing the serialized form of the JSONRPC request
		// at the time of creating the request context.
		return protocol.Payload{}
	}

	return protocol.Payload{
		Data: string(reqBz),
		// Method is alway POST for EVM-based blockchains.
		Method: http.MethodPost,

		// Path field is not used for JSONRPC services.

		// TODO_IMPROVE: adjust the timeout based on the request method:
		// An endpoint may need more time to process certain requests,
		// as indicated by the request's method and/or parameters.
		TimeoutMillisec: defaultServiceRequestTimeoutMillisec,
	}
}

// UpdateWithResponse is NOT safe for concurrent use
func (rc *requestQoSContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, receivedData []byte) {
	// TODO_IMPROVE(@adshmh): check whether the request was valid, and return an error if it was not.
	// This would be an extra safety measure, as the caller should have checked the returned value
	// indicating the validity of the request when calling on QoS instance's ParseHTTPRequest
	//
	// Instantiate an endpointQuery to capture the interaction with the service endpoint.
	endpointQuery := &endpointQuery{
		serviceID:    rc.ServiceID,
		request:      rc.Request,
		endpointAddr: endpointAddr,
		receivedData: receivedData,
	}

	// instantiate a result context to process the endpointQuery.
	resultCtx := &EndpointQueryResultContext{
		ReadonlyServiceState:        rc.readonlyServiceState,
		jsonrpcMethodResultBuilders: rc.jsonrpcMethodResultBuilders,
		endpointQuery:               endpointQuery,
	}

	// Process the endpointQuery using the correct context.
	endpointQueryResult := resultCtx.buildEndpointResult()

	// Track the processed result
	p.processedResults = append(p.processedResults, endpointQueryResult)
}

// TODO_TECHDEBT: support batch JSONRPC requests by breaking them into single JSONRPC requests and tracking endpoints' response(s) to each.
// This would also require combining the responses into a single, valid response to the batch JSONRPC request.
// See the following link for more details:
// https://www.jsonrpc.org/specification#batch
//
// GetHTTPResponse builds the HTTP response that should be returned for a JSONRPC service request.
// Implements the gateway.RequestQoSContext interface.
func (rc requestQoSContext) GetHTTPResponse() gateway.HTTPResponse {
	if rc.JSONRPCErrorResponse != nil {
		return buildHTTPResponse(rc.Logger, rc.JSONRPCErrorResponse)
	}

	return buildHTTPResponse(rc.Logger, rc.getJSONRPCResponse())
}

// GetObservations returns QoS observations from all processed results.
// GetObservations returns all endpoint observations from the request context.
// Implements gateway.RequestQoSContext interface.
func (rc requestContext) GetObservations() qosobservations.Observations {
	/*
		return qosobservations.Observations {
			RequestObservations: rc. resut???? .GetObservation(),
			EndpointObservations: rc.EndpointCallsProcessor.GetObservations(),
		// TODO_IN_THIS_PR: Implement this method in observations.go.
		// Return basic observations for now
			ServiceId:          p.ServiceID,
			ServiceDescription: p.ServiceDescription,
			RequestObservation: p.RequestObservation,
		}
	*/
}

// Build and returns an instance EndpointSelectionContext to perform endpoint selection for the client request.
// Implements the gateway.RequestQoSContext
func (rc *requestQoSContext) GetEndpointSelector() protocol.EndpointSelector {
	return &EndpointSelectionContext{
		*ReadonlyServiceState: rc.serviceState,
		Request:               rc.Request,
		endpointStore:         rc.endpointStore,
		customSelector:        rc.customEndpointSelector,
	}
}

// TODO_FUTURE(@adshmh): A retry mechanism would require support from this struct to determine if the most recent endpoint call has been successful.
//
// getJSONRPCResponse simply returns the result associated with the most recently reported EndpointCall.
func (rc requestContext) getJSONRPCResponse() *jsonrpc.Response {
	// Check if we received any endpoint results
	if len(rc.processedResults) == 0 {
		// If no results were processed, handle it as a protocol error
		return buildResultForNoResponse(rc.Request)
	}

	// Return the latest result.
	return rc.processedResults[len(rc.processedResults)-1]
}
