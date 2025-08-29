package noop

import (
	"net/http"

	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"

	"github.com/buildwithgrove/path/gateway"
	pathhttp "github.com/buildwithgrove/path/network/http"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
)

// clientRespMsgNoProtocolEndpoints is the error message sent to clients when
// the underlying protocol fails to register any endpoint responses with the NoOp QoS service.
// This can occur due to:
//   - User error: invalid service ID in the request's HTTP header
//   - Protocol error: selected endpoint failed to provide a valid response
//   - System timeout: no endpoints responded within the allowed time window
const clientRespMsgNoProtocolEndpoints = "NoOp QoS service error: No responses received from any service endpoints. Please verify your service ID and retry."

// requestContext implements all the functionality required by gateway.RequestQoSContext interface.
var _ gateway.RequestQoSContext = &requestContext{}

// requestContext provides the functionality required to fulfill the role of a Noop QoS service,
// i.e. no validation of requests or responses, and no data is kept on endpoints to guide
// the endpoint selection process.
type requestContext struct {
	// httpRequestBody contains the body of the HTTP request for which this instance of
	// requestContext was constructed.
	httpRequestBody []byte

	// httpRequestMethod contains the HTTP method (GET, POST, PUT, etc.) of the request for
	// which this instance of requestContext was constructed.
	// For more details, see https://pkg.go.dev/net/http#Request
	httpRequestMethod string

	// httpRequestPath contains the path of the HTTP request for which this instance of
	// requestContext was constructed.
	httpRequestPath string

	// receivedResponses maintains response(s) received from one or more endpoints, for the
	// request represented by this instance of requestContext.
	receivedResponses []endpointResponse

	// presetFailureResponse, if set, is used to return a preconstructed response to the user.
	// This is used by the conductor of the requestContext instance, e.g. if reading the HTTP request's body fails.
	presetFailureResponse *HTTPResponse
}

// GetServicePayload returns the payload to be sent to a service endpoint.
// Implements the gateway.RequestQoSContext interface.
func (rc *requestContext) GetServicePayloads() []protocol.Payload {
	payload := protocol.Payload{
		Data:    string(rc.httpRequestBody),
		Method:  rc.httpRequestMethod,
		Path:    "", // set below
		Headers: map[string]string{},
		RPCType: sharedtypes.RPCType_UNKNOWN_RPC,
	}
	if rc.httpRequestPath != "" {
		payload.Path = rc.httpRequestPath
	}
	return []protocol.Payload{payload}
}

// UpdateWithResponse is used to inform the requestContext of the response to its underlying service request, returned from an endpoint.
// UpdateWithResponse is NOT safe for concurrent use
// Implements the gateway.RequestQoSContext interface.
func (rc *requestContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, endpointSerializedResponse []byte) {
	rc.receivedResponses = append(rc.receivedResponses, endpointResponse{EndpointAddr: endpointAddr, ResponseBytes: endpointSerializedResponse})
}

// GetHTTPResponse returns a user-facing response that fulfills the pathhttp.HTTPResponse interface.
// Any preset failure responses, e.g. set during the construction of the requestContext instance, take priority.
// After that, this method simply returns an HTTP response based on the most recently reported endpoint response.
// Implements the gateway.RequestQoSContext interface.
func (rc *requestContext) GetHTTPResponse() pathhttp.HTTPResponse {
	if rc.presetFailureResponse != nil {
		return rc.presetFailureResponse
	}

	if len(rc.receivedResponses) == 0 {
		return &HTTPResponse{
			httpStatusCode: http.StatusOK,
			payload:        []byte(clientRespMsgNoProtocolEndpoints),
		}
	}

	return &HTTPResponse{
		httpStatusCode: http.StatusOK,
		payload:        rc.receivedResponses[len(rc.receivedResponses)-1].ResponseBytes,
	}
}

// GetObservations returns an empty struct that fulfill the required interface, since the noop QoS does not make or use
// any endpoint observations to improve endpoint selection.
// Implements the gateway.RequestQoSContext interface.
func (rc *requestContext) GetObservations() qosobservations.Observations {
	return qosobservations.Observations{}
}

// GetEndpointSelector returns an endpoint selector which simply makes a random selection among available endpoints.
// Implements the gateway.RequestQoSContext interface.
func (rc *requestContext) GetEndpointSelector() protocol.EndpointSelector {
	return RandomEndpointSelector{}
}
