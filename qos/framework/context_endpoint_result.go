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

// buildResult uses the supplied method builder to build the EndpointResult for the supplied endpointQuery.
// A default builder is used if no matches were found for the request method.
// Returns the endpointQuery augmented with the endpoint result.
func (ctx *EndpointQueryResultContext) buildEndpointQueryResult() *EndpointQueryResult {
	// Parse the endpoint's payload into JSONRPC response.
	// Stores the parsed JSONRPC response in the endpointQuery.
	shouldContinue := ctx.updateEndpointQueryResultWithParsedResponse()

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
func (ctx *EndpointQueryResultContext) updateEndpointQueryResultWithParsedResponse() bool {
	logger := ctx.getHydratedLogger()

	// Check for empty response
	if len(ctx.EndpointQueryResult.endpointPayload) == 0 {
		ctx.logger.Info().Msg("Received payload with 0 length from the endpoint. Service request will fail.")

		ctx.EndpointQueryResult = buildResultForEmptyResponse(ctx.EndpointQueryResult)
		return false
	}

	// Parse JSONRPC response
	var jsonrpcResp jsonrpc.JsonRpcResponse
	if err := json.Unmarshal(ctx.EndpointQueryResult.endpointPayload, &jsonrpcResp); err != nil {
		// Error parsing the endpoint payload: return generic response to the client.
		ctx.EndpointQueryResult = buildResultForErrorUnmarshalingEndpointReturnedData(ctx.EndpointQueryResult, err)
		return false
	}

	// Validate the JSONRPC response
	if err := jsonrpcResp.Validate(eq.request.ID); err != nil {
		// TODO_IN_THIS_PR: define a separate method for JSONRPC response validation errors.
		ctx.EndpointQueryResult = buildResultForErrorUnmarshalingEndpointPayload(ctx.EndpointQueryResult, err)
		return endpointQuery, false
	}

	// Store the parsed result
	ctx.EndpointQueryResult.ParsedJSONRPCResponse = jsonrpcResp

	// Return true to signal that parsing was successful.
	// Processing will continue to the next step.
	return true
}

// TODO_IN_THIS_PR: implement.
func (ctx *EndpointQueryResultContext) getHydratedLogger() polylog.Logger() {
	// hydrate the logger with endpointQuery fields.

}
