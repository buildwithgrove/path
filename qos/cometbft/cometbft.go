package cometbft

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
)

// QoS implements gateway.QoSService by providing:
//  1. QoSRequestParser - Builds CometBFT-specific RequestQoSContext objects from HTTP requests
//  2. EndpointSelector - Selects endpoints for service requests
var _ gateway.QoSService = &QoS{}

// QoS implements ServiceQoS for CometBFT-based chains.
// It handles chain-specific:
//   - Request parsing
//   - Response building
//   - Endpoint validation and selection
type QoS struct {
	logger polylog.Logger
	*EndpointStore
	*ServiceState
}

// ParseHTTPRequest builds a request context from an HTTP request.
// Returns (context, false) if POST request is not valid JSON-RPC.
// Implements gateway.QoSService interface.
func (qos *QoS) ParseHTTPRequest(_ context.Context, req *http.Request) (gateway.RequestQoSContext, bool) {
	requestContext := &requestContext{
		logger:        qos.logger,
		httpReq:       req,
		endpointStore: qos.EndpointStore,
		isValid:       true,
	}

	// Parse both REST and JSON-RPC requests (CometBFT supports both).
	// For JSON-RPC POST requests, read and store the request body as []byte.
	// See: https://docs.cometbft.com/v1.0/spec/rpc/
	if req.Method == http.MethodPost {
		// TODO_IMPROVE(@commoddity): implement JSON-RPC request validation.
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return requestContextFromInternalError(err), false
		}

		// Store the serialized JSON-RPC request as a byte slice
		requestContext.jsonrpcRequestBz = body
	}

	return requestContext, true
}

// ParseWebsocketRequest builds a request context from the provided WebSocket request.
// WebSocket connection requests do not have a body, so we don't need to parse it.
//
// This method implements the gateway.QoSService interface.
// TODO_HACK(@commoddity, #143): Utilize this method once the Shannon protocol supports websocket connections.
func (qos *QoS) ParseWebsocketRequest(_ context.Context) (gateway.RequestQoSContext, bool) {
	return &requestContext{
		logger:        qos.logger,
		endpointStore: qos.EndpointStore,
		isValid:       true,
	}, true
}

// ApplyObservations updates endpoint storage and blockchain state from observations.
// Implements gateway.QoSService interface.
func (q *QoS) ApplyObservations(observations *qosobservations.Observations) error {
	if observations == nil {
		return errors.New("ApplyObservations: received nil")
	}

	cometbftObservations := observations.GetCometbft()
	if cometbftObservations == nil {
		return errors.New("ApplyObservations: received nil CometBFT observation")
	}

	updatedEndpoints := q.EndpointStore.UpdateEndpointsFromObservations(cometbftObservations)

	return q.ServiceState.UpdateFromEndpoints(updatedEndpoints)
}
