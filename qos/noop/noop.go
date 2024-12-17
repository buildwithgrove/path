// package noop implements a noop QoS module, enabling a gateway operator to support services 
// which do not yet have a QoS implementation.
package noop

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/buildwithgrove/path/gateway"
)

var _ gateway.QoSService = NoOpQoS{}

type NoOpQoS struct{}

func (NoOpQoS) ParseHTTPRequest(_ context.Context, httpRequest *http.Request) (gateway.RequestQoSContext, bool) {
	bz, err := io.ReadAll(httpRequest.Body)
	if err != nil {
		return requestContextFromError(fmt.Errorf("Error reading the HTTP request body: %w", err)), false
	}

	return &requestContext{
		httpRequestBody: bz,
	}, true
}

func requestContextFromError(err error) *requestContext {
	return &requestContext{
		presetFailureResponse: &HTTPResponse{
			httpStatusCode: http.StatusOK,
			payload:        []byte(fmt.Sprintf("Error processing the request: %v", err)),
		},
	}
}
