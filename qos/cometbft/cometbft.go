package cometbft

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos"
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
	Logger polylog.Logger
	*qos.EndpointStore
	*ServiceState
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
		// TODO_IMPROVE(@commoddity): implement JSON-RPC request validation.
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return requestContextFromInternalError(err), false
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

// ApplyObservations updates endpoint store and CometBFT service state from observations.
// Implements gateway.QoSService interface.
func (q *QoS) ApplyObservations(observations *qosobservations.Observations) error {
	if observations == nil {
		return errors.New("ApplyObservations: received nil")
	}

	// Get the CometBFT observations from the observations object.
	cometbftObservations := observations.GetCometbft()
	if cometbftObservations == nil {
		return errors.New("ApplyObservations: received nil CometBFT observation")
	}

	// Apply the CometBFT observations to the endpoints.
	updatedEndpoints := q.applyCometBFTObservations(cometbftObservations.GetEndpointObservations())

	// Update the endpoint store with the new endpoints.
	q.EndpointStore.UpdateEndpointsFromObservations(updatedEndpoints)

	// Update the CometBFT service state with the new endpoints.
	return q.ServiceState.UpdateFromEndpoints(updatedEndpoints)
}

// applyCometBFTObservations applies observations to the endpoints and returns the updated endpoints.
// This method is used to initialize the endpoint store and service state when first starting the PATH hydrator.
func (q *QoS) applyCometBFTObservations(endpointObservations []*qosobservations.CometBFTEndpointObservation) map[protocol.EndpointAddr]qos.Endpoint {
	logger := q.Logger.With(
		"qos_instance", "cometbft",
		"method", "applyCometBFTObservations",
	)
	logger.Info().Msg(
		fmt.Sprintf("About to update endpoints from %d observations.", len(endpointObservations)),
	)

	storedEndpoints := q.EndpointStore.GetEndpoints()
	updatedEndpoints := make(map[protocol.EndpointAddr]qos.Endpoint)

	for _, observation := range endpointObservations {
		if observation == nil {
			logger.Info().Msg("CometBFT EndpointStore received a nil observation. Skipping...")
			continue
		}

		endpointAddr := protocol.EndpointAddr(observation.GetEndpointAddr())

		logger := logger.With("endpoint", endpointAddr)
		logger.Info().Msg("processing observation for endpoint.")

		// Initialize the CometBFT endpoint as zero-value.
		var cometbftEndpoint endpoint

		// If the endpoint is already stored, use it to initialize the endpoint.
		if storedEndpoint, found := storedEndpoints[endpointAddr]; found {
			storedCometbftEndpoint, ok := storedEndpoint.(endpoint)
			if !ok {
				logger.Warn().Msg("endpoint was not of type cometbft.endpoint. Skipping...")
				continue
			}

			cometbftEndpoint = storedCometbftEndpoint
		}

		// Apply the observation to the endpoint, whether it is already stored or not.
		if isMutated := cometbftEndpoint.ApplyObservation(observation); !isMutated {
			// If the observation did not mutate the endpoint, don't update the stored endpoint entry.
			logger.Warn().Msg("endpoint was not mutated by observations. Skipping...")
			continue
		}

		// If the observation mutated the endpoint, update the stored endpoint entry.
		// A zero-value endpoint will always be mutated by an observation and so will be stored.
		updatedEndpoints[endpointAddr] = cometbftEndpoint
	}

	return updatedEndpoints
}
