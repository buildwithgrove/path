package framework

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

	// Tracks all data related to the current request context:
	// - client's request
	// - endpoint query result(s)
	journal *requestJournal

	// QoS service will be used to build the required contexts:
	// - EndpointSelectionContext
	// - EndpointQueryResultContext
	contextBuilder *QoS
}

// TODO_MVP(@adshmh): Ensure the JSONRPC request struct can handle all valid service requests.
func (rc requestQoSContext) GetServicePayload() protocol.Payload {
	return rc.journal.getServicePayload()
}

// UpdateWithResponse is NOT safe for concurrent use
func (rc *requestQoSContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, receivedData []byte) {
	// TODO_IMPROVE(@adshmh): check whether the request was valid, and return an error if it was not.
	// This would be an extra safety measure, as the caller should have checked the returned value
	// indicating the validity of the request when calling on QoS instance's ParseHTTPRequest
	//
	// Instantiate an endpointQuery to capture the interaction with the service endpoint.
	endpointQuery := rc.journal.buildEndpointQuery(endpointAddr, receivedData)

	resultCtx := rc.contextBuilder.buildEndpointQueryResultContext()

	// Process the endpointQuery using the correct context.
	processedEndpointQuery := resultCtx.buildEndpointQueryResult(endpointQuery)

	// Track the processed result in the request journal
	rc.journal.reportProcessedEndpointQuery(processedEndpointQuery)
}

// TODO_TECHDEBT: support batch JSONRPC requests by breaking them into single JSONRPC requests and tracking endpoints' response(s) to each.
// This would also require combining the responses into a single, valid response to the batch JSONRPC request.
// See the following link for more details:
// https://www.jsonrpc.org/specification#batch
//
// GetHTTPResponse builds the HTTP response that should be returned for a JSONRPC service request.
// Implements the gateway.RequestQoSContext interface.
func (rc requestQoSContext) GetHTTPResponse() gateway.HTTPResponse {
	// check if a protocol-level error has occurred.
	rc.checkForProtocolLevelError()

	// use the request journal to build the client's HTTP response.
	return rc.journal.getHTTPResponse()
}

// GetObservations uses the request's journal to build and return all observations.
// Implements gateway.RequestQoSContext interface.
func (rc requestContext) GetObservations() qosobservations.Observations {
	// check if a protocol-level error has occurred.
	rc.checkForProtocolLevelError()

	// Use the request journal to generate the observations.
	return rc.journal.getObservations()
}

// Build and returns an instance EndpointSelectionContext to perform endpoint selection for the client request.
// Implements the gateway.RequestQoSContext
func (rc *requestQoSContext) GetEndpointSelector() protocol.EndpointSelector {
	selectorCtx := rc.contextBuilder.buildEndpointSelectionContext()
	return selectorCtx.buildSelectorForRequest(rc.journal)
}

// Declares the request as failed with protocol-level error if no data from any endpoints has been reported to the request context.
func (rc *requestContext) checkForProtocolLevelError() {
	// TODO_IMPROVE(@adshmh): consider using the journal directly for setting protocol failure error.
	//
	// Assume protocol-level error if no endpoint responses have been received yet.
	if len(rc.journal.processedEndpointQueries) == 0 {
		rc.journal.requestDetails.setProtocolLevelError()
	}
}
