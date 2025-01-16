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
	Logger polylog.Logger
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

	// TODO_TECHDEBT(@adshmh): validate the resulting JSONRPC request to block invalid requests from being sent to endpoints.
	// TODO_IMPROVE(@adshmh): method-specific validation of the JSONRPC request.
	return &requestContext{
		jsonrpcReq:    jsonrpcReq,
		endpointStore: qos.EndpointStore,
		logger:        qos.Logger,

		isValid: true,
	}, true
}

// ApplyObservations updates the stored endpoints and the perceived blockchain state using the supplied observations.
// This method implements the gateway.QoSService interface.
func (q *QoS) ApplyObservations(observations *qosobservations.Observations) error {
	if observations == nil {
		return errors.New("ApplyObservations: received nil")
	}

	evmObservations := observations.GetEVM()
	if evmObservations == nil {
		return errors.New("ApplyObservations: received nil EVM observation")
	}

	updatedEndpoints := q.EndpointStore.UpdateEndpointsFromObservations(evmObservations)

	// update the perceived state of the blockchain.
	return q.ServiceState.UpdateFromEndpoints(updatedEndpoints)
}
