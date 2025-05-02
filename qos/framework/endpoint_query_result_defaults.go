package framework

import (
	"fmt"
)

// TODO_IN_THIS_PR: reword/rename the method and the comment.
//
// defaultResultBuilder is applied by the endpointCallProcessor on endpoint responses not matching any of the JSONRPC methods specified by the custom service QoS.
// It builds an EndpointQueryResult to track JSONRPC requests/responses not utilized by the custom QoS service for updating the service state or endpoint selection.
// Used in generating observations for:
// - Metrics
// - Data Pipeline
func defaultResultBuilder(ctx *EndpointQueryResultContext) *EndpointQueryResult {
	//TODO_IN_THIS_PR: implement this function:
	/*
		JsonrpcResponse: &qosobservations.JsonRpcResponse{
			Id: r.jsonRPCResponse.ID.String(),
		},
		ResponseValidationError: r.validationError,
		HttpStatusCode:          int32(r.getHTTPStatusCode()),
	*/
	return nil
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
func buildResultForEmptyResponse(endpointQueryResult *EndpointQueryResult) *EndpointQueryResult {
	endpointError := &EndpointError{
		ErrorKind:   EndpointErrKindEmptyPayload,
		Description: "endpoint returned an empty response",
		// Set the recommended sanction based on the error
		RecommendedSanction: getRecommendedSanction(EndpointErrKindEmptyPayload, nil),
	}

	// Set a generic response.
	endpointQueryResult.parsedJSONRPCResponse = newErrResponseEmptyEndpointResponse(endpointQueryResult.getJSONRPCRequestID())
	// Set the endpoint error
	endpointQueryResult.EndpointError = endpointError

	return endpointQueryResult
}

// buildResultForErrorUnmarshalingEndpointReturnedData handles the case when parsing the endpoint's returned data failed.
func buildResultForErrorUnmarshalingEndpointReturnedData(
	endpointQueryResult *EndpointQueryResult,
	parseError error,
) *EndpointQueryResult {
	endpointError := &EndpointError{
		ErrorKind:           EndpointErrKindParseErr,
		Description:         fmt.Sprintf("endpoint payload failed to unmarshal: %q", parseError.Error()),
		RecommendedSanction: getRecommendedSanction(EndpointErrKindParseErr, parseError),
	}

	// Set a generic response.
	endpointQueryResult.parsedJSONRPCResponse = newErrResponseParseError(endpointQueryResult.getJSONRPCRequestID(), parseError)
	// Set the endpoint error
	endpointQueryResult.EndpointError = endpointError

	return endpointQueryResult
}

// buildResultForErrorValidatingEndpointResponse handles the case when validating the unmarshaled endpoint's JSONRPC response has failed.
func buildResultForErrorValidatingEndpointResponse(
	endpointQueryResult *EndpointQueryResult,
	parseError error,
) *EndpointQueryResult {
	endpointError := &EndpointError{
		ErrorKind:           EndpointErrKindValidationErr,
		Description:         fmt.Sprintf("endpoint payload failed to unmarshal: %q", parseError.Error()),
		RecommendedSanction: getRecommendedSanction(EndpointErrKindValidationErr, parseError),
	}

	// Set a generic response.
	endpointQueryResult.parsedJSONRPCResponse = newErrResponseParseError(endpointQueryResult.getJSONRPCRequestID(), parseError)
	// Set the endpoint error
	endpointQueryResult.EndpointError = endpointError

	return endpointQueryResult
}
