package cometbft

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
)

// QoS struct performs the functionality defined by gateway package's ServiceQoS,
// which consists of:
// A) a QoSRequestParser which builds CometBFT-specific RequestQoSContext objects,
// by parsing user HTTP requests.
// B) an EndpointSelector, which selects an endpoint for performing a service request.
var _ gateway.QoSService = &QoS{}

// QoS is the ServiceQoS implementations for CometBFT-based chains.
// It contains logic specific to CometBFT-based chains, including request parsing,
// response building, and endpoint validation/selection.
type QoS struct {
	*EndpointStore
	*ServiceState
	Logger polylog.Logger
}

// ParseHTTPRequest builds a request context from an HTTP request.
// Returns (context, false) if request cannot be parsed as JSONRPC.
// Implements gateway.QoSService interface.
func (qos *QoS) ParseHTTPRequest(_ context.Context, req *http.Request) (gateway.RequestQoSContext, bool) {
	if req.Method != http.MethodGet {
		err := fmt.Errorf("ParseHTTPRequest: received non-GET request")
		return requestContextFromInternalError(err), false
	}

	return &requestContext{
		logger:        qos.Logger,
		httpReq:       req,
		endpointStore: qos.EndpointStore,
		isValid:       true,
	}, true
}

// ParseWebsocketRequest builds a request context from the provided WebSocket request.
// WebSocket connection requests do not have a body, so we don't need to parse it.
//
// This method implements the gateway.QoSService interface.
// TODO_HACK(@commoddity, #143): Utilize this method once the Shannon protocol supports websocket connections.
func (qos *QoS) ParseWebsocketRequest(_ context.Context) (gateway.RequestQoSContext, bool) {
	return &requestContext{
		logger:        qos.Logger,
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
