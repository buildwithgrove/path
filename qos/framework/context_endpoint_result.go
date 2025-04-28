package framework

import (
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_IN_THIS_PR: add hydratedLoggers.
//
// TODO_FUTURE(@adshmh): Support overriding the JSONRPC response through the EndpointQueryResultContext, IFF there is a use case for it.
//
// EndpointQueryResultContext provides context for processing a result with the service state.
// Provides a fluent API for custom service implementations to create endpoint query results without directly constructing types.
type EndpointQueryResultContext struct {
	// Allows direct Get calls on the current service state.
	// ServiceState's public methods provide read only access: this is not the context for updating service state.
	*ServiceState

	// Tracks the result of the endpoint query.
	// Declared public to expose EndpointQueryResult's setter/getter methods.
	*EndpointQueryResult

	// Custom result builders, supplied by the QoS Definition.
	jsonrpcMethodResultBuilders map[jsonrpc.Method]EndpointQueryResultBuilder
}

// ===> TODO_IN_THIS_PR: find a better name to signal to the client they should call this when done with updating the EndpointQueryResult.
func (ctx *EndpointQueryResultContext) Success() *EndpointQueryResult {
	return &ctx.endpointQueryResult
}



// buildResult uses the supplied method builder to build the EndpointResult for the supplied endpointQuery.
// A default builder is used if no matches were found for the request method.
// Returns the endpointQuery augmented with the endpoint result.
func (ctx *EndpointQueryResultContext) buildEndpointQueryResult() *EndpointQueryResult {
	// Parse the endpoint's payload into JSONRPC response.
	// Stores the parsed JSONRPC response in the endpointQuery.
	shouldContinue := ctx.updateEndpointQueryWithParsedResponse()

	// Parsing failed: skip the rest of the processing.
	if !shouldContinue {
		// parsing the request failed: stop the request processing flow.
		// Return a failure result for building the client's response and observations.
		return ctx.EndpointQueryResult
	}

	// Use the custom endpoint query result builder, if one is found matching the JSONRPC request's method.
	builder, found := ctx.jsonrpcMethodResultBuilders[parsedEndpointQuery.request.Method]
	if !found {
		// Use default processor for methods not specified by the custom QoS service definition.
		builder = defaultResultBuilder
	}

	// Process the result using custom service's result processor.
	// Pass the context to the builder to provide helper methods.
	queryResult := builder(ctx)

	// Return the endpoint query result.
	return queryResult
}

// TODO_IN_THIS_PR: define/allow customization of sanctions for endpoint errors: e.g. malformed response.
//
// parseEndpointQuery parses the payload from an endpoint and handles empty responses and parse errors.
// It returns a boolean indicating whether processing should continue (true) or stop (false).
func (ctx *EndpointQueryResultContext) updateEndpointQueryWithParsedResponse() bool {
	logger := ctx.getHydratedLogger()

	endpointQuery := ctx.EndpointQueryResult.endpointQuery

	// Check for empty response
	if len(endpointQuery.receivedData) == 0 {
		ctx.logger.Info()
		endpointQuery.result = buildResultForEmptyResponse(endpointQuery)
		return endpointQuery, false
	}

	// Parse JSONRPC response
	var jsonrpcResp jsonrpc.JsonRpcResponse
	if err := json.Unmarshal(endpointQuery.receivedData, &jsonrpcResp); err != nil {
		endpointQuery.result = buildResultForErrorUnmarshalingEndpointReturnedData(endpointQuery, err)
		return endpointQuery, false
	}

	// Validate the JSONRPC response
	if err := jsonrpcResp.Validate(eq.request.ID); err != nil {
		// TODO_IN_THIS_PR: define a separate method for JSONRPC response validation errors.
		endpointQuery.result = buildResultForErrorUnmarshalingEndpointReturnedData(endpointQuery, err)
		return endpointQuery, false
	}

	// Store the parsed result
	endpointQuery.parsedResponse = jsonrpcResp

	// Return true to signal that parsing was successful.
	// Processing will continue to the next step.
	return endpointQuery, true
}

// TODO_IN_THIS_PR: implement.
func (ctx *EndpointQueryResultContext) getHydratedLogger() polylog.Logger() {
	// hydrate the logger with endpointQuery fields.

}
