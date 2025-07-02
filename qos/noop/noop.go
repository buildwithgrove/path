// package noop implements a noop QoS module, enabling a gateway operator to support services
// which do not yet have a QoS implementation.
package noop

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/metrics/devtools"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
)

// TODO_TECHDEBT(@adshmh): support customization of the endpoint response's timeout.
// defaultEndpointResponseTimeoutMillisec is the default timeout for an endpoint to return a response to a service request.
const defaultEndpointResponseTimeoutMillisec = 5_000

var _ gateway.QoSService = NoOpQoS{}

type NoOpQoS struct{}

// ParseHTTPRequest reads the supplied HTTP request's body and passes it on to a new requestContext instance.
// It intentionally avoids performing any validation on the request, as is the designed behavior of the noop QoS.
// Implements the gateway.QoSService interface.
func (NoOpQoS) ParseHTTPRequest(_ context.Context, httpRequest *http.Request) (gateway.RequestQoSContext, bool) {
	bz, err := io.ReadAll(httpRequest.Body)
	if err != nil {
		return requestContextFromError(fmt.Errorf("error reading the HTTP request body: %w", err)), false
	}

	return &requestContext{
		httpRequestBody:                 bz,
		httpRequestMethod:               httpRequest.Method,
		httpRequestPath:                 httpRequest.URL.Path,
		endpointResponseTimeoutMillisec: defaultEndpointResponseTimeoutMillisec,
	}, true
}

// ParseWebsocketRequest builds a request context from the provided WebSocket request.
// This method implements the gateway.QoSService interface.
func (q NoOpQoS) ParseWebsocketRequest(_ context.Context) (gateway.RequestQoSContext, bool) {
	return &requestContext{
		endpointResponseTimeoutMillisec: defaultEndpointResponseTimeoutMillisec,
	}, true
}

// ApplyObservations on noop QoS only fulfills the interface requirements and does not perform any actions.
// Implements the gateway.QoSService interface.
func (NoOpQoS) ApplyObservations(_ *qosobservations.Observations) error {
	return nil
}

// GetRequiredQualityChecks on noop QoS only fulfills the interface requirements and does not perform any actions.
// Implements the gateway.QoSService interface.
func (NoOpQoS) GetRequiredQualityChecks(_ protocol.EndpointAddr) []gateway.RequestQoSContext {
	return nil
}

// requestContextFromError constructs and returns a requestContext instance using the supplied error.
// The returned requestContext will returns a user-facing HTTP request with the supplied error when it GetHTTPResponse method is called.
func requestContextFromError(err error) *requestContext {
	return &requestContext{
		presetFailureResponse: &HTTPResponse{
			httpStatusCode: http.StatusOK,
			payload:        []byte(fmt.Sprintf("Error processing the request: %v", err)),
		},
	}
}

// HydrateDisqualifiedEndpointsResponse is a no-op for the noop QoS.
func (NoOpQoS) HydrateDisqualifiedEndpointsResponse(_ protocol.ServiceID, _ *devtools.DisqualifiedEndpointResponse) {
}
