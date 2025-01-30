package cometbft

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
// Returns (context, false) if POST request cannot be parsed as JSONRPC.
// Implements gateway.QoSService interface.
func (qos *QoS) ParseHTTPRequest(_ context.Context, req *http.Request) (gateway.RequestQoSContext, bool) {
	requestContext := &requestContext{
		logger:        qos.Logger,
		httpReq:       req,
		endpointStore: qos.EndpointStore,
		isValid:       true,
	}

	// CometBFT supports both REST-like and JSON-RPC requests.
	// If the request is a JSON-RPC POST request, read the JSON-RPC
	// request body and store it on the request context as a []byte.
	// Reference: https://docs.cometbft.com/v1.0/spec/rpc/
	if req.Method == http.MethodPost {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return requestContextFromInternalError(err), false
		}

		// Validate the JSON-RPC request body.
		if err := json.Unmarshal(body, &jsonrpc.Request{}); err != nil {
			return requestContextFromUserError(err), false
		}

		// Store the serialized JSONRPC request as a byte slice
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
