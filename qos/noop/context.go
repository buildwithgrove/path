package noop

import (
	"net/http"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/message"
	"github.com/buildwithgrove/path/protocol"
)

var _ gateway.RequestQoSContext = &requestContext{}

type requestContext struct {
	httpRequestBody []byte

	receivedResponses []endpointResponse

	presetFailureResponse *HTTPResponse
}

func (rc *requestContext) GetServicePayload() protocol.Payload {
	// TODO_TECHDEBT: log a
	return protocol.Payload{
		Method: http.MethodPost,
		Data:   string(rc.httpRequestBody),
		// TODO_TECHDEBT(@adshmh): support customization of the service request's timeout.
		// TimeoutMillisec: time.Duration(60) * time.Second
	}
}

func (rc *requestContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, endpointSerializedResponse []byte) {
	rc.receivedResponses = append(rc.receivedResponses, endpointResponse{EndpointAddr: endpointAddr, ResponseBytes: endpointSerializedResponse})
}

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

func (rc *requestContext) GetObservationSet() message.ObservationSet {
	return observationSet{}
}

func (rc *requestContext) GetEndpointSelector() protocol.EndpointSelector {
	return RandomEndpointSelector{}
}
