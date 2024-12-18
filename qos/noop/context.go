package noop

import (
	"net/http"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/message"
	"github.com/buildwithgrove/path/protocol"
)

// requestContext implements all the functionality required by gateway.RequestQoSContext interface.
var _ gateway.RequestQoSContext = &requestContext{}

// requestContext provides the functionality required to fulfill the role of a Noop QoS service,
// i.e. no validation of requests or responses, and no data is kept on endpoints to guide
// the endpoint selection process.
type requestContext struct {
	// httpRequestBody contains the body of the HTTP request for which this instance of
	// requestContext was constructed.
	httpRequestBody []byte

	// endpointResponseTimeoutMillisec specifies the timeout for receiving a response from an endpoint serving
	// the request represented by this requestContext instance.
	endpointResponseTimeoutMillisec int

	// receivedResponses maintains response(s) received from one or more endpoints, for the
	// request represented by this instance of requestContext.
	receivedResponses []endpointResponse

	// presetFailureResponse, if set, is used to return a preconstructed response to the user.
	// This is used by the consutrctor of the requestContext instance, e.g. if reading the HTTP request's body fails.
	presetFailureResponse *HTTPResponse
}

// GetServicePayload returns the payload to be sent to a service endpoint.
// This method implements the gateway.RequestQoSContext interface.
func (rc *requestContext) GetServicePayload() protocol.Payload {
	return protocol.Payload{
		Method:          http.MethodPost,
		Data:            string(rc.httpRequestBody),
		TimeoutMillisec: rc.endpointResponseTimeoutMillisec,
	}
}

// UpdateWithResponse is used to inform the requestContext of the response to its underlying service request, returned from an endpoint.
// UpdateWithResponse is NOT safe for concurrent use
// This method implements the gateway.RequestQoSContext interface.
func (rc *requestContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, endpointSerializedResponse []byte) {
	rc.receivedResponses = append(rc.receivedResponses, endpointResponse{EndpointAddr: endpointAddr, ResponseBytes: endpointSerializedResponse})
}

// GetHTTPResponse returns a user-facing response that fulfills the gateway.HTTPResponse interface.
// Any preset failure responses, e.g. set during the construction of the requestContext instance, take priority.
// After that, this method simply returns an HTTP response based on the most recently reported endpoint response.
// This method implements the gateway.RequestQoSContext interface.
func (rc *requestContext) GetHTTPResponse() gateway.HTTPResponse {
	if rc.presetFailureResponse != nil {
		return rc.presetFailureResponse
	}

	if len(rc.receivedResponses) == 0 {
		return &HTTPResponse{
			httpStatusCode: http.StatusOK,
			payload:        []byte("No responses have been received from any endpoints"),
		}
	}

	return &HTTPResponse{
		httpStatusCode: http.StatusOK,
		payload:        rc.receivedResponses[len(rc.receivedResponses)].ResponseBytes,
	}
}

// GetObservationSet returns an empty struct that fulfill the required interface, since the noop QoS does not make or use
// any endpoint observations to improve endpoint selection.
// This method implements the gateway.RequestQoSContext interface.
func (rc *requestContext) GetObservationSet() message.ObservationSet {
	return observationSet{}
}

// GetEndpointSelector returns an endpoint selector which simply makes a random selection among available endpoints.
// This method implements the gateway.RequestQoSContext interface.
func (rc *requestContext) GetEndpointSelector() protocol.EndpointSelector {
	return RandomEndpointSelector{}
}
