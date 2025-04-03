package framework

import (
	"fmt"
	"time"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_IN_THIS_PR: reword/rename the method and the comment.
// defaultResultBuilder is applied by the endpointCallProcessor on EndpointCalls not matching any of the JSONRPC methods specified by the custom service QoS.
// It builds an EndpointResult to track JSONRPC requests/responses which are not utilized by the custom QoS service for updating the service state or endpoint selection.
func defaultResultBuilder(ctx *EndpointQueryResultContext) *EndpointQueryResult {
	//TODO_IN_THIS_PR: implement this function:
	/*
				JsonrpcResponse: &qosobservations.JsonRpcResponse{
					Id: r.jsonRPCResponse.ID.String(),
				},
				ResponseValidationError: r.validationError,
				HttpStatusCode:          int32(r.getHTTPStatusCode()),
	*/
}


// TODO_IN_THIS_PR: clarify that the following happens for a NoResponse:
// - NoResponse's underlying getHTTPStatusCode always returns a 500 Internal error.
// - NoResponse is always an invalid response.
//
// buildResultForNoResponse handles the case when no endpoint response was received.
// This can occur due to protocol-level failures or when no endpoint was selected.
// This is not the same as empty responses (where an endpoint responded with empty data).
func buildResultForNoResponse(request *jsonrpc.JsonRpcRequest) *EndpointQueryResult {
	// Create a synthetic result for protocol error
	result := &EndpointResult{
		Call: &EndpointCall{
			Request:    request,
			ReceivedAt: time.Now(),
		},
		parseResult: &parseResult{
			parseError: fmt.Errorf("no endpoint responses received"),
		},
	}
	
	// Create result context for creating the result
	ctx := &ResultContext{
		Request: request,
	}
	
	// Apply default sanction for no response
	result.ResultData = applySanctionForNoResponse(ctx)
	
	// Set error kind
	result.ResultData.Error.kind = EndpointDataErrorKindNoInteraction
	
	// Set error response
	result.ErrorResponse = newErrResponseNoEndpointResponses(request.Id)
	
	return result
}

// TODO_MVP(@adshmh): Implement request retry support:
//  1. Add ShouldRetry() method to gateway.RequestQoSContext
//  2. Integrate ShouldRetry() into gateway request handler
//  3. Extend evm.response interface with ShouldRetry()
//  4. Add ShouldRetry() to evm.requestContext to evaluate retry eligibility based on responses
//
// TODO_IN_THIS_PR: update comments to show the following for Empty response:
// EmptyResponse always returns a 500 Internal error HTTP status code.
// An empty response is always invalid: e.g. EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_EMPTY
//
// buildResultForEmptyResponse handles the case when an endpoint returned an empty response.
func buildResultForEmptyResponse(call *EndpointCall) *EndpointQueryResult {
	// Create a new result with the call
	result := &EndpointResult{
		Call: call,
		parseResult: &parseResult{
			isEmpty:    true,
			parseError: fmt.Errorf("empty response from endpoint"),
		},
	}
	
	// Create result context for creating the result
	ctx := &ResultContext{
		EndpointAddr: call.EndpointAddr,
		Request:      call.Request,
	}
	
	// Apply default sanction for empty response
	result.ResultData = applySanctionForEmptyResponse(ctx)
	
	// Set error kind
	result.ResultData.Error.kind = EndpointDataErrorKindEmptyPayload
	
	// Set error response
	result.ErrorResponse = newErrResponseEmptyResponse(call.Request.Id)
	
	return result
}

// buildResultForErrorUnmarshalingEndpointReturnedData handles the case when parsing the endpoint's returned data failed.
func buildResultForErrorUnmarshalingEndpointReturnedData(
	call *EndpointCall,
	parseError error,
) *EndpointQueryResult {
	// Create a new result with the call
	result := &EndpointResult{
		Call: call,
		parseResult: &parseResult{
			parseError: parseError,
		},
	}
	
	// Create result context for creating the result
	ctx := &ResultContext{
		EndpointAddr: call.EndpointAddr,
		Request:      call.Request,
	}
	
	// Apply default sanction for parse error
	result.ResultData = applySanctionForUnmarshalingError(ctx, parseError)
	
	// Set error kind and raw payload
	result.ResultData.Error.kind = EndpointDataErrorKindUnmarshaling
	result.ResultData.Error.rawPayload = call.ReceivedData
	
	// Set error response
	result.ErrorResponse = newErrResponseParseError(call.Request.Id, parseError)
	
	return result
}
