package cosmos

import (
	"encoding/json"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_MVP(@adshmh): Support individual configuration of timeout for every service that uses CosmosSDK QoS.
// The default timeout when sending a request to a CosmosSDK blockchain endpoint.
const defaultServiceRequestTimeoutMillisec = 10_000

// requestContext provides the support required by the gateway
// package for handling service requests.
var _ gateway.RequestQoSContext = &requestContext{}

// endpointResponse implements the response interface for CosmosSDK-based blockchain services.
var _ response = &endpointResponse{}

// TODO_REFACTOR: Improve naming clarity by distinguishing between interfaces and adapters
// in the metrics/qos/cosmos and qos/cosmos packages, and elsewhere names like `response` are used.
// Consider renaming:
//   - metrics/qos/cosmos: response → CosmosSDKMetricsResponse
//   - qos/cosmos: response → CosmosSDKQoSResponse
//   - observation/cosmos: observation -> CosmosSDKObservation
//
// TODO_TECHDEBT: Need to add a Validate() method here to allow the caller (e.g. gateway)
// determine whether the endpoint's response was valid, and whether a retry makes sense.
//
// response defines the functionality required from a parsed endpoint response, which all response types must implement.
// It provides methods to:
//  1. Generate observations for endpoint quality tracking
//  2. Format HTTP responses to send back to clients
type response interface {
	// GetObservation returns an observation of the endpoint's response
	// for quality metrics tracking, including HTTP status code.
	GetObservation() qosobservations.CosmosSDKEndpointObservation

	// GetHTTPResponse returns the HTTP response to be sent back to the client.
	GetHTTPResponse() httpResponse
}

// endpointResponse stores the response received from an endpoint.
type endpointResponse struct {
	protocol.EndpointAddr
	response
	unmarshalErr error
}

// requestContext implements the functionality for CosmosSDK-based blockchain services.
type requestContext struct {
	logger polylog.Logger

	// httpReq is the original HTTP request from the user
	httpReq http.Request

	// chainID is the chain identifier for CosmosSDK QoS implementation.
	chainID string

	// service_id is the identifier for the CosmosSDK QoS implementation.
	// It is the "alias" or human readable interpretation of the chain_id.
	// Used in generating observations.
	serviceID protocol.ServiceID

	// The origin of the request handled by the context.
	// Either:
	// - Organic: user requests
	// - Synthetic: requests built by the QoS service to get additional data points on endpoints.
	requestOrigin qosobservations.RequestOrigin

	// The length of the request payload in bytes.
	requestPayloadLength uint

	serviceState *serviceState

	// For JSON-RPC POST requests (when applicable)
	jsonrpcReq *jsonrpc.Request

	// endpointResponses is the set of responses received from one or
	// more endpoints as part of handling this service request.
	endpointResponses []endpointResponse
}

// GetServicePayload returns the payload for the service request.
// It accounts for both REST-like and JSON-RPC requests.
// Implements gateway.RequestQoSContext interface.
func (rc requestContext) GetServicePayload() protocol.Payload {
	payload := protocol.Payload{
		Method:          rc.httpReq.Method,
		TimeoutMillisec: defaultServiceRequestTimeoutMillisec,
		Headers:         getRPCTypeHeaders(rc.httpReq.URL.Path, rc.jsonrpcReq),
	}

	// If the request is REST-like set the path including query parameters.
	if rc.isRESTLikeRequest() {
		rc.setPathWithQueryParams(&payload)
	}

	// Determine if request is a JSON-RPC request by checking if:
	//  - The request method is POST
	//  - The JSON-RPC request is not empty.
	if rc.isJsonRpcRequest() {
		rc.setJSONRPCRequest(&payload)
	}

	return payload
}

// isRESTLikeRequest checks if the request is a REST-like request.
func (rc *requestContext) isRESTLikeRequest() bool {
	return rc.httpReq.URL.Path != ""
}

// setPathWithQueryParams sets the path of the payload with the query parameters from the request.
func (rc *requestContext) setPathWithQueryParams(payload *protocol.Payload) {
	payload.Path = rc.httpReq.URL.Path
	if rc.httpReq.URL.RawQuery != "" {
		payload.Path += "?" + rc.httpReq.URL.RawQuery
	}
}

// isEmptyJSONRPCRequest checks if the JSON-RPC request is empty/uninitialized.
func (rc requestContext) isJsonRpcRequest() bool {
	return rc.httpReq.Method == http.MethodPost && rc.jsonrpcReq != nil
}

// setJSONRPCRequest sets the JSON-RPC request in the payload.
func (rc *requestContext) setJSONRPCRequest(payload *protocol.Payload) {
	reqBz, err := json.Marshal(rc.jsonrpcReq)
	if err != nil {
		rc.logger.Error().Err(err).Msg("failed to marshal JSON-RPC request")
	}
	payload.Data = string(reqBz)
}

// UpdateWithResponse stores (appends) the response from an endpoint in the request context.
// CRITICAL: NOT safe for concurrent use.
// Implements gateway.RequestQoSContext interface.
func (rc *requestContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, responseBz []byte) {
	// TODO_IMPROVE: check whether the request was valid, and return an error if it was not.
	// This would be an extra safety measure, as the caller should have checked the returned value
	// indicating the validity of the request when calling on QoS instance's ParseHTTPRequest
	response, err := unmarshalResponse(
		rc.logger,
		rc.httpReq.URL.Path,
		responseBz,
		rc.isJsonRpcRequest(),
		endpointAddr,
	)

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

// TODO_TECHDEBT: support batch JSONRPC requests by breaking them into single JSONRPC requests and tracking endpoints' response(s) to each.
// This would also require combining the responses into a single, valid response to the batch JSONRPC request.
// See the following link for more details:
// https://www.jsonrpc.org/specification#batch
//
// GetHTTPResponse builds the HTTP response that should be returned for
// a CosmosSDK blockchain service request.
// Implements the gateway.RequestQoSContext interface.
func (rc requestContext) GetHTTPResponse() gateway.HTTPResponse {
	// Use a noResponses struct if no responses were reported by the protocol from any endpoints.
	if len(rc.endpointResponses) == 0 {
		responseNoneObj := responseNone{
			logger:     rc.logger,
			httpReq:    rc.httpReq,
			jsonrpcReq: rc.jsonrpcReq,
		}

		return responseNoneObj.GetHTTPResponse()
	}

	// return the last endpoint response reported to the context.
	return rc.endpointResponses[len(rc.endpointResponses)-1].GetHTTPResponse()
}

// GetObservations returns all endpoint observations from the request context.
// Implements gateway.RequestQoSContext interface.
func (rc requestContext) GetObservations() qosobservations.Observations {
	// Set the observation fields common for all requests: successful or failed.
	observations := &qosobservations.CosmosSDKRequestObservations{
		ChainId:       rc.chainID,
		ServiceId:     string(rc.serviceID),
		RequestOrigin: rc.requestOrigin,
	}

	// No endpoint responses received.
	// Set request error.
	if len(rc.endpointResponses) == 0 {
		observations.RequestError = qos.GetRequestErrorForProtocolError()

		return qosobservations.Observations{
			ServiceObservations: &qosobservations.Observations_Cosmos{
				Cosmos: observations,
			},
		}
	}

	// Build the endpoint(s) observations.
	endpointObservations := make([]*qosobservations.CosmosSDKEndpointObservation, len(rc.endpointResponses))
	for idx, endpointResponse := range rc.endpointResponses {
		obs := endpointResponse.GetObservation()
		obs.EndpointAddr = string(endpointResponse.EndpointAddr)
		endpointObservations[idx] = &obs
	}

	// Set the endpoint observations fields.
	observations.EndpointObservations = endpointObservations

	return qosobservations.Observations{
		ServiceObservations: &qosobservations.Observations_Cosmos{
			// TODO_TECHDEBT(@adshmh): Set JSON-RPCRequest field.
			// Requires utility function to convert between:
			// 		- qos.jsonrpc.Request
			// 		- observation.qos.JsonRpcRequest
			// Needed for setting JSON-RPC fields in any QoS service's observations.
			Cosmos: observations,
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
func (rc *requestContext) Select(allEndpoints protocol.EndpointAddrList) (protocol.EndpointAddr, error) {
	return rc.serviceState.Select(allEndpoints)
}
