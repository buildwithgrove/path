package evm

import (
	"context"
	"errors"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
)

// QoS implements gateway.QoSService by providing:
//  1. QoSRequestParser - Builds EVM-specific RequestQoSContext objects from HTTP requests
//  2. EndpointSelector - Selects endpoints for service requests
var _ gateway.QoSService = &QoS{}

// QoS implements ServiceQoS for EVM-based chains.
// It handles chain-specific:
//   - Request parsing
//   - Response building
//   - Endpoint validation and selection
type QoS struct {
	logger polylog.Logger
	*endpointStore
	*serviceState
	*evmRequestValidator
}

// ParseHTTPRequest builds a request context from an HTTP request.
// Returns (requestContext, true) if the request is valid JSONRPC
// Returns (errorContext, false) if the request is not valid JSONRPC.
// Implements gateway.QoSService interface.
func (qos *QoS) ParseHTTPRequest(_ context.Context, req *http.Request) (gateway.RequestQoSContext, bool) {
	return qos.evmRequestValidator.validateHTTPRequest(req)
}

// ParseWebsocketRequest builds a request context from the provided WebSocket request.
// WebSocket connection requests do not have a body, so we don't need to parse it.
//
// This method implements the gateway.QoSService interface.
func (qos *QoS) ParseWebsocketRequest(_ context.Context) (gateway.RequestQoSContext, bool) {
	return &requestContext{
		logger:        qos.logger,
		endpointStore: qos.endpointStore,
	}, true
}

// ApplyObservations updates endpoint storage and blockchain state from observations.
// Implements gateway.QoSService interface.
func (q *QoS) ApplyObservations(observations *qosobservations.Observations) error {
	if observations == nil {
		return errors.New("ApplyObservations: received nil")
	}

	evmObservations := observations.GetEvm()
	if evmObservations == nil {
		return errors.New("ApplyObservations: received nil EVM observation")
	}

	updatedEndpoints := q.endpointStore.updateEndpointsFromObservations(evmObservations)

	return q.serviceState.updateFromEndpoints(updatedEndpoints)
}
