package cometbft

import (
	"fmt"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
)

const defaultServiceRequestTimeoutMillisec = 10_000

// requestContext provides the support required by the gateway
// package for handling service requests.
var _ gateway.RequestQoSContext = &requestContext{}

// response is an interface that represents the response received from an endpoint.
type response interface {
	GetObservation() qosobservations.CometBFTEndpointObservation
	GetResponsePayload() []byte
	GetResponseStatusCode() int
}

// endpointResponse stores the response received from an endpoint.
type endpointResponse struct {
	protocol.EndpointAddr
	response
	unmarshalErr error
}

// requestContext implements QoS functionality for CometBFT blockchain services.
type requestContext struct {
	logger        polylog.Logger
	endpointStore *EndpointStore

	// httpReq is the original HTTP request from the user
	httpReq *http.Request

	// CometBFT supports both REST and JSON-RPC formats.
	// For JSON-RPC POST requests, jsonrpcRequestBz stores the serialized request body.
	// See: https://docs.cometbft.com/v1.0/spec/rpc/
	jsonrpcRequestBz []byte

	// isValid indicates if the user request was valid when parsed.
	// Set by QoS instance during request context creation.
	isValid bool

	// preSelectedEndpointAddr overrides default endpoint selection with specific address.
	// Used when building request context to check specific endpoint.
	preSelectedEndpointAddr protocol.EndpointAddr

	// endpointResponses contains responses from endpoints handling this service request
	// NOTE: Currently only supports responses associated with a single JSON-RPC request.
	// TODO_FUTURE: Batch support will require modifying the field type.
	endpointResponses []endpointResponse
}

// GetServicePayload returns the payload for the service request.
// It accounts for both REST-like and JSON-RPC requests.
// Implements gateway.RequestQoSContext interface.
func (rc requestContext) GetServicePayload() protocol.Payload {
	payload := protocol.Payload{
		Method:          rc.httpReq.Method,
		TimeoutMillisec: defaultServiceRequestTimeoutMillisec,
	}

	// If the request is REST-like, set the path including query parameters.
	if rc.httpReq.URL.Path != "" {
		payload.Path = rc.httpReq.URL.Path

		if rc.httpReq.URL.RawQuery != "" {
			payload.Path += "?" + rc.httpReq.URL.RawQuery
		}
	}

	// If the request is JSON-RPC, set the data from the stored []byte.
	if rc.isJSONRPCRequest() {
		payload.Data = string(rc.jsonrpcRequestBz)
	}

	return payload
}

// UpdateWithResponse stores (appends) the response from an endpoint in the request context.
// CRITICAL: NOT safe for concurrent use.
// Implements gateway.RequestQoSContext interface.
func (rc *requestContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, responseBz []byte) {
	// TODO_IMPROVE: check whether the request was valid, and return an error if it was not.
	// This would be an extra safety measure, as the caller should have checked the returned value
	// indicating the validity of the request when calling on QoS instance's ParseHTTPRequest
	response, err := unmarshalResponse(rc.logger, rc.httpReq.URL.Path, responseBz, rc.isJSONRPCRequest())

	// Multiple responses can be associated with a single request for multiple reasons, such as:
	// - Retries from single/multiple endpoints
	// - Collecting a quorum of from different endpoints
	// - Organic vs synthetic responses
	rc.endpointResponses = append(rc.endpointResponses,
		endpointResponse{
			EndpointAddr: endpointAddr,
			response:     response,
			unmarshalErr: err,
		},
	)
}

// SetPreSelectedEndpointAddr assigns the endpoint address to be used for hydrator checks.
// It is called to override the endpoint selection process with a specific endpoint.
// Is used to enforce performing quality checks on a specific endpoint.
func (rc *requestContext) SetPreSelectedEndpointAddr(endpointAddr protocol.EndpointAddr) {
	rc.preSelectedEndpointAddr = endpointAddr
}

// GetHTTPResponse builds the HTTP response for a CometBFT blockchain service request.
// Returns the last endpoint response if available, otherwise returns generic response.
// Implements gateway.RequestQoSContext interface.
func (rc requestContext) GetHTTPResponse() gateway.HTTPResponse {
	// Ignore unmarshaling errors since the payload is empty for REST-like requests.
	// By default, return a generic HTTP response if no endpoint responses
	// have been reported to the request context.
	response, _ := unmarshalResponse(rc.logger, rc.httpReq.URL.Path, []byte(""), rc.isJSONRPCRequest())

	// If at least one endpoint response exists, return the last one
	if len(rc.endpointResponses) >= 1 {
		response = rc.endpointResponses[len(rc.endpointResponses)-1]
	}

	// Default to generic response if no endpoint responses exist
	return httpResponse{
		responsePayload: response.GetResponsePayload(),
	}
}

// GetObservations returns all endpoint observations from the request context.
// Implements gateway.RequestQoSContext interface.
func (rc requestContext) GetObservations() qosobservations.Observations {
	observations := make([]*qosobservations.CometBFTEndpointObservation, len(rc.endpointResponses))
	for idx, endpointResponse := range rc.endpointResponses {
		obs := endpointResponse.response.GetObservation()
		obs.EndpointAddr = string(endpointResponse.EndpointAddr)
		observations[idx] = &obs
	}

	return qosobservations.Observations{
		ServiceObservations: &qosobservations.Observations_Cometbft{
			Cometbft: &qosobservations.CometBFTRequestObservations{
				// TODO_TECHDEBT(@adshmh): Set JSON-RPCRequest field.
				// Requires utility function to convert between:
				// - qos.jsonrpc.Request
				// - observation.qos.JsonRpcRequest
				// Needed for setting JSON-RPC fields in any QoS service's observations.
				EndpointObservations: observations,
			},
		},
	}
}

// GetEndpointSelector returns the endpoint selector for the request context.
// Implements the gateway.RequestQoSContext interface.
func (rc *requestContext) GetEndpointSelector() protocol.EndpointSelector {
	return rc
}

// Select returns the address of an endpoint using the request context's endpoint store.
// Implements the protocol.EndpointSelector interface.
func (rc *requestContext) Select(allEndpoints []protocol.Endpoint) (protocol.EndpointAddr, error) {
	// If set, override the selection with the pre-selected endpoint address.
	if rc.preSelectedEndpointAddr != "" {
		return preSelectedEndpoint(rc.preSelectedEndpointAddr, allEndpoints)
	}

	// Select an endpoint from the available endpoints using the endpoint store.
	return rc.endpointStore.Select(allEndpoints)
}

// isJSONRPCRequest checks if the request context contains a serialized JSON-RPC request.
// Reference: https://docs.cometbft.com/v1.0/spec/rpc/
func (rc *requestContext) isJSONRPCRequest() bool {
	return len(rc.jsonrpcRequestBz) > 0
}

// preSelectedEndpoint returns the pre-selected endpoint address if it exists in the list of available endpoints.
// It is used to override the default endpoint selection with a specific address.
func preSelectedEndpoint(
	preSelectedEndpointAddr protocol.EndpointAddr,
	allEndpoints []protocol.Endpoint,
) (protocol.EndpointAddr, error) {
	for _, endpoint := range allEndpoints {
		if endpoint.Addr() == preSelectedEndpointAddr {
			return preSelectedEndpointAddr, nil
		}
	}

	return protocol.EndpointAddr(""), fmt.Errorf("singleEndpointSelector: endpoint %s not found in available endpoints", preSelectedEndpointAddr)
}
