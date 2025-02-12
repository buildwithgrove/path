package solana

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
//  1. QoSRequestParser - Builds Solana-specific RequestQoSContext objects from HTTP requests
//  2. EndpointSelector - Selects endpoints for service requests
var _ gateway.QoSService = &QoS{}

// QoS implements ServiceQoS for Solana-based chains.
// It handles chain-specific:
//   - Request parsing
//   - Response building
//   - Endpoint validation and selection
type QoS struct {
	logger polylog.Logger
	*EndpointStore
	*ServiceState
}

// ParseHTTPRequest builds a request context from the provided HTTP request.
// It returns an error if the HTTP request cannot be parsed as a JSONRPC request.
//
// Implements the gateway.QoSService interface.
func (qos *QoS) ParseHTTPRequest(_ context.Context, req *http.Request) (gateway.RequestQoSContext, bool) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return requestContextFromInternalError(err), false
	}

	var jsonrpcReq jsonrpc.Request
	if err := json.Unmarshal(body, &jsonrpcReq); err != nil {
		return requestContextFromUserError(err), false
	}

	// TODO_TECHDEBT(@adshmh): validate the JSONRPC request to block invalid requests from being sent to endpoints.
	// TODO_IMPROVE(@adshmh): perform method-specific validation of the JSONRPC request.
	// e.g. for a `getTokenAccountBalance` request, ensure there is a single account public key is specified as the `params` object.
	// https://solana.com/docs/rpc/http/gettokenaccountbalance
	return &requestContext{
		logger: qos.logger,

		JSONRPCReq:    jsonrpcReq,
		endpointStore: qos.EndpointStore,

		// set isValid to true to signal to the requestContext that the request is considered valid.
		// The requestContext can be enhanced (see the above TODOs) to e.g. skip sending an invalid request to any endpoints,
		// and directly return an error response to the user instead.
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
		logger: qos.logger,

		endpointStore: qos.EndpointStore,

		isValid: true,
	}, true
}

// ApplyObservations updates the stored endpoints and the perceived blockchain state using the supplied observations.
// Implements the gateway.QoSService interface.
func (q *QoS) ApplyObservations(observations *qosobservations.Observations) error {
	if observations == nil {
		return errors.New("ApplyObservations: received nil observations")
	}

	solanaObservations := observations.GetSolana()
	if solanaObservations == nil {
		return errors.New("ApplyObservations: received nil Solana observation")
	}

	updatedEndpoints := q.EndpointStore.UpdateEndpointsFromObservations(solanaObservations)

	// update the perceived current state of the blockchain.
	return q.ServiceState.UpdateFromEndpoints(updatedEndpoints)
}
