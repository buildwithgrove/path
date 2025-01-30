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

type response interface {
	GetObservation() qosobservations.CometBFTEndpointObservation
	GetResponsePayload() []byte
	GetResponseStatusCode() int
}

type endpointResponse struct {
	protocol.EndpointAddr
	response
	unmarshalErr error
}

// requestContext provides the functionality required
// to support QoS for a CometBFT blockchain service.
type requestContext struct {
	logger polylog.Logger

	// httpReq is the HTTP request received from the user.
	httpReq *http.Request

	// CometBFT supports both REST-like and JSON-RPC requests.
	// This field stores the serialized JSON-RPC request as a
	// byte slice, if it is present in a JSON-RPC POST request.
	// Reference: https://docs.cometbft.com/v1.0/spec/rpc/
	jsonrpcRequestBz []byte

	endpointStore *EndpointStore

	// isValid indicates whether the underlying user request
	// for this request context was found to be valid.
	// This field is set by the corresponding QoS instance
	// when creating this request context during the parsing
	// of the user request.
	isValid bool

	// preSelectedEndpointAddr allows overriding the default
	// endpoint selector with a specific endpoint's addresss.
	// This is used when building a request context as a check
	// for a specific endpoint.
	preSelectedEndpointAddr protocol.EndpointAddr

	// endpointResponses is the set of responses received from one or
	// more endpoints as part of handling this service request.
	// NOTE: these are all related to a single JSONRPC request,
	// enhancing to support batch JSONRPC requests will involve the
	// modification of this field's type.
	endpointResponses []endpointResponse
}

func (rc requestContext) GetServicePayload() protocol.Payload {
	payload := protocol.Payload{
		Method:          rc.httpReq.Method,
		TimeoutMillisec: defaultServiceRequestTimeoutMillisec,
	}
	// IF the request is REST-like, set the path.
	if rc.httpReq.URL.Path != "" {
		payload.Path = rc.httpReq.URL.Path
	}
	// If the request is JSON-RPC, set the data from the stored []byte.
	if rc.isJSONRPCRequest() {
		payload.Data = string(rc.jsonrpcRequestBz)
	}
	return payload
}

// UpdateWithResponse is NOT safe for concurrent use
func (rc *requestContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, responseBz []byte) {
	// TODO_IMPROVE: check whether the request was valid, and return an error if it was not.
	// This would be an extra safety measure, as the caller should have checked the returned value
	// indicating the validity of the request when calling on QoS instance's ParseHTTPRequest

	response, err := unmarshalResponse(rc.logger, rc.httpReq.URL.Path, responseBz, rc.isJSONRPCRequest())

	rc.endpointResponses = append(rc.endpointResponses,
		endpointResponse{
			EndpointAddr: endpointAddr,
			response:     response,
			unmarshalErr: err,
		},
	)
}

// GetHTTPResponse builds the HTTP response that should be returned for
// a CometBFT blockchain service request.
// TODO_TECHDEBT(@commoddity): Look into refactoring and reusing specific components
// that play identical roles across QoS packages in order to reduce code duplication.
func (rc requestContext) GetHTTPResponse() gateway.HTTPResponse {
	// By default, return a generic HTTP response if no endpoint responses
	// have been reported to the request context.
	// intentionally ignoring the error here, since unmarshallResponse
	// is being called with an empty endpoint response payload.
	response, _ := unmarshalResponse(rc.logger, rc.httpReq.URL.Path, []byte(""), rc.isJSONRPCRequest())

	if len(rc.endpointResponses) >= 1 {
		// return the last endpoint response reported to the context.
		response = rc.endpointResponses[len(rc.endpointResponses)-1]
	}

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
				// TODO_TECHDEBT(@adshmh): Set JSONRPCRequest field.
				// Requires utility function to convert between:
				// - qos.jsonrpc.Request
				// - observation.qos.JsonRpcRequest
				// Needed for setting JSONRPC fields in any QoS service's observations.
				EndpointObservations: observations,
			},
		},
	}
}

func (rc *requestContext) GetEndpointSelector() protocol.EndpointSelector {
	return rc
}

// Select returns the address of an endpoint using the request context's endpoint store.
// Implements the protocol.EndpointSelector interface.
func (rc *requestContext) Select(allEndpoints []protocol.Endpoint) (protocol.EndpointAddr, error) {
	if rc.preSelectedEndpointAddr != "" {
		return preSelectedEndpoint(rc.preSelectedEndpointAddr, allEndpoints)
	}

	return rc.endpointStore.Select(allEndpoints)
}

// isJSONRPCRequest returns true if the request context contains a JSON-RPC request.
// This is determined by checking whether the request context contains a serialized JSON-RPC request.
// Reference: https://docs.cometbft.com/v1.0/spec/rpc/
func (rc *requestContext) isJSONRPCRequest() bool {
	return len(rc.jsonrpcRequestBz) > 0
}

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
