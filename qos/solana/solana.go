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

// QoS struct performs the functionality defined by gateway package's ServiceQoS,
// which consists of:
// A) a QoSRequestParser which builds Solana-specific RequestQoSContext objects,
// by parsing user HTTP requests.
// B) an EndpointSelector, which selects an endpoint for performing a service request.
var _ gateway.QoSService = &QoS{}

// QoS is the ServiceQoS implementations for the Solana blockchain.
// It contains logic specific to Solana, including request parsing,
// response building, and endpoint validation/selection.
type QoS struct {
	*EndpointStore
	ServiceState ServiceState
	Logger       polylog.Logger
}

// ParseHTTPRequest builds a request context from the provided HTTP request.
// It returns an error if the HTTP request cannot be parsed as a JSONRPC request.
//
// This method implements the gateway.QoSService interface.
func (qos *QoS) ParseHTTPRequest(_ context.Context, req *http.Request) (gateway.RequestQoSContext, bool) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return requestContextFromInternalError(err), false
	}

	var jsonrpcReq jsonrpc.Request
	if err := json.Unmarshal(body, &jsonrpcReq); err != nil {
		return requestContextFromUserError(err), false
	}

	// TODO_TECHDEBT: validate the resulting JSONRPC request to block invalid requests from being sent to endpoints.
	// TODO_IMPROVE: method-specific validation of the JSONRPC request.
	return &requestContext{
		JSONRPCReq:    jsonrpcReq,
		EndpointStore: qos.EndpointStore,
		Logger:        qos.Logger,

		isValid: true,
	}, true
}

// ApplyObservations updates the stored endpoints and the "estimated" blockchain state using the supplied observations.
// This method implements the gateway.QoSService interface.
func (q *QoS) ApplyObservations(observations qosobservations.Observations) error {
	solanaObservations := observations.GetSolana()
	if solanaObservations == nil {
		return errors.New("ApplyObservations: received nil Solana observation")
	}

	updatedEndpoints := q.EndpointStore.UpdateEndpointsFromObservations(solanaObservations)

	// update the (estimated) current state of the blockchain.
	return q.ServiceState.UpdateFromEndpoints(updatedEndpoints)
}
