package evm

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
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
	*EndpointStore
	*ServiceState
}

// ParseHTTPRequest builds a request context from an HTTP request.
// Returns (context, false) if POST request is not valid JSON-RPC.
// Implements gateway.QoSService interface.
func (qos *QoS) ParseHTTPRequest(_ context.Context, req *http.Request) (gateway.RequestQoSContext, bool) {
	// TODO_TECHDEBT(@adshmh): Simplify the qos package by refactoring gateway.QoSContextBuilder.
	// Proposed change: Create a new ServiceRequest type containing raw payload data ([]byte)
	// Benefits: Decouples the qos package from HTTP-specific error handling.
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return requestContextFromInternalError(
			qos.logger,
			err,
			qosobservations.EVMRequestValidationErrorKind_EVM_REQUEST_VALIDATION_ERROR_KIND_HTTP_BODY_READ_FAILURE,
		), false
	}

	// TODO_MVP(@adshmh): Add a request validator for JSONRPC requests:
	// e.g. The request is invalid and the service request flow should be stopped if the specified JSONRPC method is not valid.
	var jsonrpcReq jsonrpc.Request
	if err := json.Unmarshal(body, &jsonrpcReq); err != nil {
		return requestContextFromUserError(
			qos.logger,
			err,
			qosobservations.EVMRequestValidationErrorKind_EVM_REQUEST_VALIDATION_ERROR_KIND_REQUEST_UNMARSHALING_FAILURE,
		), false
	}

	// TODO_IMPROVE(@adshmh): Add JSON-RPC request validation to block invalid requests
	// TODO_IMPROVE(@adshmh): Add method-specific JSONRPC request validation
	return &requestContext{
		logger: qos.logger,

		chainID:       qos.ServiceState.chainID,
		jsonrpcReq:    jsonrpcReq,
		endpointStore: qos.EndpointStore,

		isValid: true,
	}, true
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

	evmObservations := observations.GetEvm()
	if evmObservations == nil {
		return errors.New("ApplyObservations: received nil EVM observation")
	}

	updatedEndpoints := q.EndpointStore.UpdateEndpointsFromObservations(evmObservations)

	return q.ServiceState.UpdateFromEndpoints(updatedEndpoints)
}
