package evm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const (
	// TODO_MVP(@adshmh): Support individual configuration of timeout for every service that uses EVM QoS.
	// The default timeout when sending a request to
	// an EVM blockchain endpoint.
	defaultServiceRequestTimeoutMillisec = 10000
)

// requestContext provides the support required by the gateway
// package for handling service requests.
var _ gateway.RequestQoSContext = &requestContext{}

// TODO_TECHDEBT: Need a Validate() method here to allow
// the caller, e.g. gateway, determine whether the endpoint's
// response was valid, and whether a retry makes sense.
//
// response defines the functionality required from
// a parsed endpoint response.
type response interface {
	GetObservation() qosobservations.EVMEndpointObservation
	GetResponsePayload() []byte
}

type endpointResponse struct {
	protocol.EndpointAddr
	response
	unmarshalErr error
}

// requestContext implements the functionality for EVM-based blockchain services.
type requestContext struct {
	logger polylog.Logger

	// chainID is the chain identifier for EVM QoS implementation.
	// Expected as the `Result` field in eth_chainId responses.
	chainID string

	endpointStore *EndpointStore

	// TODO_TECHDEBT(@adshmh): support batch JSONRPC requests
	jsonrpcReq jsonrpc.Request

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

// TODO_MVP(@adshmh): Ensure the JSONRPC request struct
// can handle all valid service requests.
func (rc requestContext) GetServicePayload() protocol.Payload {
	reqBz, err := json.Marshal(rc.jsonrpcReq)
	if err != nil {
		// TODO_MVP(@adshmh): find a way to guarantee this never happens,
		// e.g. by storing the serialized form of the JSONRPC request
		// at the time of creating the request context.
		return protocol.Payload{}
	}

	return protocol.Payload{
		Data: string(reqBz),
		// Method is alway POST for EVM-based blockchains.
		Method: http.MethodPost,

		// Path field is not used for EVM-based blockchains.

		// TODO_IMPROVE: adjust the timeout based on the request method:
		// An endpoint may need more time to process certain requests,
		// as indicated by the request's method and/or parameters.
		TimeoutMillisec: defaultServiceRequestTimeoutMillisec,
	}
}

// UpdateWithResponse is NOT safe for concurrent use
func (rc *requestContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, responseBz []byte) {
	// TODO_IMPROVE: check whether the request was valid, and return an error if it was not.
	// This would be an extra safety measure, as the caller should have checked the returned value
	// indicating the validity of the request when calling on QoS instance's ParseHTTPRequest

	response, err := unmarshalResponse(rc.logger, rc.jsonrpcReq, responseBz)

	rc.endpointResponses = append(rc.endpointResponses,
		endpointResponse{
			EndpointAddr: endpointAddr,
			response:     response,
			unmarshalErr: err,
		},
	)
}

// TODO_TECHDEBT: support batch JSONRPC requests by breaking them into
// single JSONRPC requests and tracking endpoints' response(s) to each.
// This would also require combining the responses into a single, valid
// response to the batch JSONRPC request.
// See the following link for more details:
// https://www.jsonrpc.org/specification#batch
//
// GetHTTPResponse builds the HTTP response that should be returned for
// an EVM blockchain service request.
func (rc requestContext) GetHTTPResponse() gateway.HTTPResponse {
	// TODO_MVP(@adshmh): Add `responseNone` type to handle cases where no endpoint response was received
	// (e.g., protocol-level failures) and update both user response and metrics.
	//
	// By default, return a generic HTTP response if no endpoint responses
	// have been reported to the request context.
	// intentionally ignoring the error here, since unmarshallResponse
	// is being called with an empty endpoint response payload.
	response, _ := unmarshalResponse(rc.logger, rc.jsonrpcReq, []byte(""))

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
	// TODO_MVP(@adshmh): Add responseNone type to track requests that fail without receiving endpoint responses
	// (e.g., protocol failures) via metrics.
	observations := make([]*qosobservations.EVMEndpointObservation, len(rc.endpointResponses))
	for idx, endpointResponse := range rc.endpointResponses {
		obs := endpointResponse.response.GetObservation()
		obs.EndpointAddr = string(endpointResponse.EndpointAddr)
		observations[idx] = &obs
	}

	return qosobservations.Observations{
		ServiceObservations: &qosobservations.Observations_Evm{
			Evm: &qosobservations.EVMRequestObservations{
				JsonrpcRequest:       rc.jsonrpcReq.GetObservation(),
				ChainId:              rc.chainID,
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
