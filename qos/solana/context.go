package solana

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/message"
	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/buildwithgrove/path/relayer"
)

const (
	// The default timeout when sending a request to
	// a Solana blockchain endpoint.
	defaultServiceRequestTimeoutMillisec = 5000
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
	GetObservation() (observation, bool)
	GetResponsePayload() []byte
	// TODO_TECHDEBT: add method(s) to support retrying a request, e.g. IsUserError(), IsEndpointError().
}

type endpointResponse struct {
	relayer.EndpointAddr
	response
	unmarshalErr error
}

// requestContext provides the functionality required
// to support QoS for a Solana blockchain service.
type requestContext struct {
	JSONRPCReq    jsonrpc.Request
	ServiceState  *ServiceState
	EndpointStore *EndpointStore
	Logger        polylog.Logger

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
	preSelectedEndpointAddr relayer.EndpointAddr

	// endpointResponses is the set of responses received from one or
	// more endpoints as part of handling this service request.
	// NOTE: these are all related to a single JSONRPC request,
	// enhancing to support batch JSONRPC requests will involve the
	// modification of this field's type.
	endpointResponses []endpointResponse
}

// TODO_UPNEXT(@adshmh): Ensure the JSONRPC request struct
// can handle all valid service requests.
func (rc requestContext) GetServicePayload() relayer.Payload {
	reqBz, err := json.Marshal(rc.JSONRPCReq)
	if err != nil {
		// TODO_UPNEXT(@adshmh): find a way to guarantee this never happens,
		// e.g. by storing the serialized form of the JSONRPC request
		// at the time of creating the request context.
		return relayer.Payload{}
	}

	return relayer.Payload{
		Data: string(reqBz),
		// Method is alway POST for Solana.
		Method: http.MethodPost,

		// Path field is not used for Solana.

		// TODO_IMPROVE: adjust the timeout based on the request method:
		// An endpoint may need more time to process certain requests,
		// as indicated by the request's method and/or parameters.
		TimeoutMillisec: defaultServiceRequestTimeoutMillisec,
	}
}

// UpdateWithResponse is NOT safe for concurrent use
func (rc *requestContext) UpdateWithResponse(endpointAddr relayer.EndpointAddr, responseBz []byte) {
	// TODO_IMPROVE: check whether the request was valid, and return an error if it was not.
	// This would be an extra safety measure, as the caller should have checked the returned value
	// indicating the validity of the request when calling on QoS instance's ParseHTTPRequest

	response, err := unmarshalResponse(rc.JSONRPCReq, responseBz, rc.Logger)

	// TODO_UPNEXT(@adshmh): Drop the unmarshalling error: the returned response interface should provide methods to allow the caller to:
	// 1. Check if the response from the endpoint was valid or malformed. This is needed to support retrying with a different endpoint if
	// the originally selected one fails to return a valid response to the user's request.
	// 2. Return a generic but valid JSONRPC response to the user.
	rc.endpointResponses = append(rc.endpointResponses,
		endpointResponse{
			EndpointAddr: endpointAddr,
			response:     response,
			unmarshalErr: err,
		},
	)
}

// GetHTTPResponse builds the HTTP response that should be returned for
// a Solana blockchain service request.
func (rc requestContext) GetHTTPResponse() gateway.HTTPResponse {
	var response response

	if len(rc.endpointResponses) >= 1 {
		// return the last endpoint response reported to the context.
		response = rc.endpointResponses[len(rc.endpointResponses)-1]
	} else {
		// By default, return a generic HTTP response if no endpoint responses
		// have been reported to the request context.
		// intentionally ignoring the error here, since unmarshallResponse
		// is being called with an empty endpoint response payload.
		response, _ = unmarshalResponse(rc.JSONRPCReq, []byte(""), rc.Logger)
	}

	return httpResponse{
		responsePayload: response.GetResponsePayload(),
	}
}

func (rc requestContext) GetObservationSet() message.ObservationSet {
	// No updates needed if the request was invalid
	if !rc.isValid {
		return observationSet{}
	}

	observations := make(map[relayer.EndpointAddr][]observation)
	for _, response := range rc.endpointResponses {
		obs, ok := response.GetObservation()
		if !ok {
			continue
		}

		addr := response.EndpointAddr
		observations[addr] = append(observations[addr], obs)
	}

	return observationSet{
		EndpointStore: rc.EndpointStore,
		ServiceState:  rc.ServiceState,
		Observations:  observations,
	}
}

func (rc *requestContext) GetEndpointSelector() relayer.EndpointSelector {
	return rc
}

func (rc *requestContext) Select(allEndpoints []relayer.Endpoint) (relayer.EndpointAddr, error) {
	if rc.preSelectedEndpointAddr != "" {
		return preSelectedEndpoint(rc.preSelectedEndpointAddr, allEndpoints)
	}

	return rc.EndpointStore.Select(allEndpoints)
}

func preSelectedEndpoint(
	preSelectedEndpointAddr relayer.EndpointAddr,
	allEndpoints []relayer.Endpoint,
) (relayer.EndpointAddr, error) {
	for _, endpoint := range allEndpoints {
		if endpoint.Addr() == preSelectedEndpointAddr {
			return preSelectedEndpointAddr, nil
		}
	}

	return relayer.EndpointAddr(""), fmt.Errorf("singleEndpointSelector: endpoint %s not found in available endpoints", preSelectedEndpointAddr)
}
