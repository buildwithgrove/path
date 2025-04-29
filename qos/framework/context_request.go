package framework

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos/framework"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
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
	endpointQueryResult := rc.journal.buildEndpointQueryResult(endpointAddr, receivedData)

	// Instantiate a result context using the endpointQuery.
	resultCtx := rc.contextBuilder.buildEndpointQueryResultContext(endpointQueryResult)

	// Build an endpoint query result using the context.
	endpointQueryResult := resultCtx.buildEndpointQueryResult()

	// Track the result in the request journal.
	rc.journal.reportEndpointQueryResult(endpointQueryResult)
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
	return rc.journal.buildObservations()
}

// Build and returns an instance EndpointSelectionContext to perform endpoint selection for the client request.
// Implements the gateway.RequestQoSContext
func (rc *requestQoSContext) GetEndpointSelector() protocol.EndpointSelector {
	selectorCtx := rc.contextBuilder.buildEndpointSelectionContext()
	return selectorCtx.buildSelectorForRequest(rc.journal)
}

// Declares the request as failed with protocol-level error if no data from any endpoints has been reported to the request context.
func (rc *requestContext) checkForProtocolLevelError() {
	// Assume protocol-level error if no endpoint responses have been received yet.
	if len(rc.journal.endpointQueryResults) == 0 {
		rc.journal.setProtocolLevelError()
	}
}


func (ctx *requestContext) initFromHTTP(httpReq *http.Request) bool {
	jsonrpcReq, reqErr := parseHTTPRequest(ctx.logger, httpReq)

	// initialize the request journal to track all request data and events.
	journal: &requestJournal{
		jsonrpcRequest: jsonrpcReq,
		requestErr: reqErr,
	},

	// Only proceed with next steps if there were no errors parsing the HTTP request into a JSONRPC request.
	return reqErr == nil 
}

// parseHTTPRequest builds and returns a context for processing the HTTP request:
// - Reads and processes the HTTP request
// - Parses a JSONRPC request from the HTTP request's payload.
// - Validates the resulting JSONRPC request.
func parseHTTPRequest(
	logger polylog.Logger,
	httpReq *http.Request,
) (*jsonrpc.Request, *requestError) {
	// Read the HTTP request body
	body, err := io.ReadAll(httpReq.Body)
	defer httpReq.Body.Close()

	// TODO_IMPROVE(@adshmh): Propagate a request ID parameter on internal errors that occur after successful request parsing.
	// There are no such cases as of PR #186.
	if err != nil {
		// Handle read error (internal server error)
		logger.Error().Err(err).Msg("Failed to read request body")

		// return the error details to be stored in the request journal.
		return nil, buildRequestErrorForInternalErrHTTPRead(err)
	}

	// Parse the JSON-RPC request
	var jsonrpcReq jsonrpc.JsonRpcRequest
	if err := json.Unmarshal(body, &jsonrpcReq); err != nil {
		// TODO_IN_THIS_PR: log the first 1K bytes of the body.
		// Handle parse error (client error)
		logger.Error().Err(err).Msg("Failed to parse JSON-RPC request")

		return nil, buildRequestErrorForParseError(err)
	}

	// Validate the request
	if validationErr := jsonrpcReq.Validate(); validationErr != nil {
		// Request failed basic JSONRPC request validation.
		logger.Info().Err(validationErr).
			Str("method", jsonrpcReq.Method).
			Msg("JSONRPC Request validation failed")

		return jsonrpcReq, buildRequestErrorJSONRPCValidationError(jsonrpcReq.ID, validationErr)
	}

	// Request is valid
	logger.Debug().
		Str("method", jsonrpcReq.Method).
		Msg("Request validation successful")

	return jsonrpcReq, nil
}
