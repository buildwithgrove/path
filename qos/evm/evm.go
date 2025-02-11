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

// QoS struct performs the functionality defined by gateway package's ServiceQoS,
// which consists of:
// A) a QoSRequestParser which builds EVM-specific RequestQoSContext objects,
// by parsing user HTTP requests.
// B) an EndpointSelector, which selects an endpoint for performing a service request.
var _ gateway.QoSService = &QoS{}

// QoS is the ServiceQoS implementations for EVM-based chains.
// It contains logic specific to EVM-based chains, including request parsing,
// response building, and endpoint validation/selection.
type QoS struct {
	*EndpointStore
	*ServiceState
	logger polylog.Logger
}

// ParseHTTPRequest builds a request context from an HTTP request.
// Returns (context, false) if request cannot be parsed as JSONRPC.
// Implements gateway.QoSService interface.
func (qos *QoS) ParseHTTPRequest(_ context.Context, req *http.Request) (gateway.RequestQoSContext, bool) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return requestContextFromInternalError(err), false
	}

	var jsonrpcReq jsonrpc.Request
	if err := json.Unmarshal(body, &jsonrpcReq); err != nil {
		return requestContextFromUserError(err), false
	}

	// TODO_TECHDEBT(@adshmh): Add JSONRPC request validation to block invalid requests
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
