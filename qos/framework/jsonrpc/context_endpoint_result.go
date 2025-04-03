package jsonrpc

// ResultBuilder is implemented by custom service implementations to build an EndpointResult from an endpointCall.
// It processes a valid JSONRPC response for a specific method and extracts the relevant data or error information.
// It can potentially makr a JSONRPC response as invalid:
// For example if the result field cannot be parsed into a number in a response to an `eth_blockNumber` request.
type EndpointQueryResultBuilder func(ctx *EndpointQueryResultContext) *EndpointQueryResult

// TODO_FUTURE(@adshmh): Support overriding the JSONRPC response through the EndpointQueryResultContext, IFF there is a use case for it.
//
// EndpointQueryResultContext provides context for processing a result with the service state.
// Provides a fluent API for custom service implementations to create endpoint query results without directly constructing types.
type EndpointQueryResultContext struct {
	// Allows direct Get calls on the current service state.
	// It is read only: this is not the context for updating service state.
	*ReadonlyServiceState

	// The endpoint query for which results are being built.
	endpointQuery *endpointQuery

	// Custom result builders, supplied by the QoS Definition.
	jsonrpcMethodResultBuilders map[string]EndpointQueryResultBuilder

	// Response is the parsed JSON-RPC response received from the endpoint.
	response *jsonrpc.JsonRpcResponse

	// The result data that will be returned to the caller (requestContext)
	result *EndpointQueryResult
}

// buildResult uses the supplied method builder to build and return the EndpointResult.
// A default builder is used if no matches were found for the request method.
func (ctx *EndpointResultContext) buildResult() *EndpointQueryResult {
	if parsingFailureResult, shouldContinue := ctx.parseEndpointPayload(); !shouldContinue {
		return parsingFailureResult
	}

	builder, found := ctx.jsonrpcMethodResultBuilders[ctx.endpointQuery.request.Method]

	if !found {
		// Use default processor for unrecognized methods
		builder = defaultResultBuilder
	}

	// Process the result using service-specific processor with context
	ctx.result.ResultData = builder(ctx)
	return ctx.result
}

// TODO_IN_THIS_PR: define/allow customization of sanctions for endpoint errors: e.g. malformed response.
//
// parseEndpointPayload parses the payload from an endpoint and handles empty responses and parse errors.
// It returns the result and a boolean indicating whether processing should continue (true) or stop (false).
func (ctx *EndpointResultContext) parseEndpointPayload() (*EndpointQueryResult, bool) {
	// Check for empty response
	if len(call.ReceivedData) == 0 {
		result := buildResultForEmptyResponse(eq)
		return result, false
	}

	// Parse JSONRPC response
	var jsonrpcResp jsonrpc.JsonRpcResponse
	if err := json.Unmarshal(eq.receivedData, &jsonrpcResp); err != nil {
		result := buildResultForErrorUnmarshalingEndpointReturnedData(call, err)
		return result, false
	}

	// Validate the JSONRPC response
	if err := jsonrpcResp.Validate(eq.request.ID); err != nil {
		// TODO_IN_THIS_PR: define a separate method for JSONRPC response validation errors.
		result := buildResultForErrorUnmarshalingEndpointReturnedData(eq, err)
		return result, false
	}

	// Store the parsed result
	ctx.parsedResponse = jsonrpcResp

	// Return true to signal that parsing was successful.
	// Processing will continue to the next step.
	return nil, true
}

// Success creates a success result with the given value.
func (ctx *ResultBuilderContext) Success(value string) *ResultData {
	valuePtr := &value
	return &ResultData{
		Type:  ctx.Method,
		Value: valuePtr,
	}
}

// ErrorResult creates an error result with the given message and no sanction.
func (ctx *ResultBuilderContext) ErrorResult(description string) *ResultData {
	return &ResultData{
		Type: ctx.Method,
		Error: &ResultError{
			Description: description,
			kind:        EndpointDataErrorKindInvalidResult,
		},
		CreatedTime: time.Now(),
	}
}

// SanctionEndpoint creates an error result with a temporary sanction.
func (ctx *ResultBuilderContext) SanctionEndpoint(description, reason string, duration time.Duration) *ResultData {
	return &ResultData{
		Type: ctx.Method,
		Error: &ResultError{
			Description: description,
			RecommendedSanction: &SanctionRecommendation{
				Sanction: Sanction{
					Type:        SanctionTypeTemporary,
					Reason:      reason,
					ExpiryTime:  time.Now().Add(duration),
					CreatedTime: time.Now(),
				},
				SourceDataType: ctx.Method,
				TriggerDetails: description,
			},
			kind: EndpointDataErrorKindInvalidResult,
		},
		CreatedTime: time.Now(),
	}
}

// PermanentSanction creates an error result with a permanent sanction.
func (ctx *ResultBuilderContext) PermanentSanction(description, reason string) *ResultData {
	return &ResultData{
		Type: ctx.Method,
		Error: &ResultError{
			Description: description,
			RecommendedSanction: &SanctionRecommendation{
				Sanction: Sanction{
					Type:        SanctionTypePermanent,
					Reason:      reason,
					CreatedTime: time.Now(),
				},
				SourceDataType: ctx.Method,
				TriggerDetails: description,
			},
			kind: EndpointDataErrorKindInvalidResult,
		},
		CreatedTime: time.Now(),
	}
}
