package evm

import (
	"github.com/buildwithgrove/path/gateway"
)

// requestContext provides the support required by the gateway
// package for handling service requests.
var _ gateway.ServiceRequestContext = &requestContext{}

// TODO_TECHDEBT: Need a Validate() method here to allow
// the caller, e.g. gateway, determine whether the endpoint's
// response was valid, and whether a retry makes sense.
//
// response defines the functionality required from
// a parsed endpoint response.
type response interface {
	GetObservation() (observation, bool)
	GetResponsePayload() []byte
}

type endpointResponse struct {
	relayer.EndpointAddr
	Response response
}

// requestContext provides the functionality required
// to support QoS for an EVM blockchain service.
type requestContext struct {
	// TODO_TECHDEBT: support batch JSONRPC requests
	method method
	id     jsonrpc.ID

	// isValid indicates whether the underlying user request
	// for this request context was found to be valid.
	// This field is set by the corresponding QoS instance
	// when creating this request context during the parsing
	// of the user request.
	isValid bool

	// endpointResponses is the set of responses received from one or
	// more endpoints as part of handling this service request.
	// NOTE: these are all related to a single JSONRPC request,
	// enhancing to support batch JSONRPC requests will involve the
	// modification of this field's type.
	endpointResponses []endpointResponse
}

// TODO_IN_THIS_COMMIT: implement this by adding a request parser.
func (rc requestContext) GetServicePayload() relayer.Payload {
	return relayer.Payload{}
}

// UpdateWithResponse is NOT safe for concurrent use
func (rc requestContext) UpdateWithResponse(endpointAddr relayer.EndpointAddr, endpointResponse []byte) {
	// TODO_IMPROVE: check whether the request was valid, and return an error if it was not.
	// This would be an extra safety measure, as the caller should have checked the returned value
	// indicating the validity of the request when calling on QoS instance's ParseHTTPRequest

	response, err := unmarshalResponse(rc.method, endpointResponse)
	if err != nil {
		// TODO_FUTURE: log the error
	}

	rc.endpointResponses = append(rc.endpointResponses,
		endpointResponse{
			EndpointAddr: endpointAddr,
			Response:     response,
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
func (rc requestContext) GetHTTPResponse() HTTPResponse {
	// By default, return a generic HTTP response if no endpoint responses
	// have been reported to the request context.
	// intentionally ignoring the error here, since unmarshallResponse
	// is being called with an empty endpoint response payload.
	response, _ := unmarshalResponse(rc.method, []byte(""))

	if len(rc.endpointResponses) >= 1 {
		// return the last endpoint response reported to the context.
		response = rc.endpointResponses[len(rc.endpointResponses)-1]
	}

	return httpResponse{
		responsePayload: response.GetResponsePayload(),
	}
}

func (rc requestContext) GetObservationSet() observationSet {
	// No updates needed if the request was invalid
	if !rc.isValid {
		return observationSet{}
	}

	observations := make(map[relayer.EndpointAddr][]observation)
	for _, endpointResponse := range rc.endpointResponses {
		obs, ok := endpointResponse.Response.GetObservation()
		if ok {
			endpointObservations := observations[endpointResponse.EndpointAddr]
			endpointObservations = append(endpointObservations, obs)
			observations[endpointResponse.EndpointAddr] = endpointObservations
		}
	}

	return observationSet{
		EndpointStore: rc.EndpointStore,
		Observations:  observations,
	}
}
