package solana

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos"
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
	Logger polylog.Logger
	*qos.EndpointStore
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
		logger: qos.Logger,

		jsonrpcReq:    jsonrpcReq,
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
		logger:        qos.Logger,
		endpointStore: qos.EndpointStore,
		isValid:       true,
	}, true
}

// ApplyObservations updates the stored endpoints and the perceived blockchain state using the supplied observations.
// Implements the gateway.QoSService interface.
func (q *QoS) ApplyObservations(observations *qosobservations.Observations) error {
	if observations == nil {
		return errors.New("ApplyObservations: received nil observations")
	}

	// Get the Solana observations from the observations object.
	solanaObservations := observations.GetSolana()
	if solanaObservations == nil {
		return errors.New("ApplyObservations: received nil Solana observation")
	}

	// Apply the Solana observations to the endpoints.
	updatedEndpoints := q.applySolanaObservations(solanaObservations.GetEndpointObservations())

	// Update the endpoint store with the new endpoints.
	q.EndpointStore.UpdateEndpointsFromObservations(updatedEndpoints)

	// Update the service state with the new endpoints.
	return q.ServiceState.UpdateFromEndpoints(updatedEndpoints)
}

// applySolanaObservations applies observations to the endpoints and returns the updated endpoints.
// This method is used to initialize the endpoint store and service state when first starting the PATH hydrator.
func (q *QoS) applySolanaObservations(solanaObservations []*qosobservations.SolanaEndpointObservation) map[protocol.EndpointAddr]qos.Endpoint {
	logger := q.Logger.With(
		"qos_instance", "solana",
		"method", "applySolanaObservations",
	)
	logger.Info().Msg(fmt.Sprintf("About to update endpoints from %d observations.", len(solanaObservations)))

	storedEndpoints := q.EndpointStore.GetEndpoints()
	updatedEndpoints := make(map[protocol.EndpointAddr]qos.Endpoint)

	for _, observation := range solanaObservations {
		if observation == nil {
			logger.Info().Msg("Solana EndpointStore received a nil observation. Skipping...")
			continue
		}

		endpointAddr := protocol.EndpointAddr(observation.EndpointAddr)

		logger := logger.With("endpoint", endpointAddr)
		logger.Info().Msg("processing observation for endpoint.")

		// Initialize the Solana endpoint as zero-value.
		var solanaEndpoint endpoint

		// If the endpoint is already stored, use it to initialize the endpoint.
		if storedEndpoint, found := storedEndpoints[endpointAddr]; found {
			storedEndpointCast, ok := storedEndpoint.(endpoint)
			if !ok {
				logger.Warn().Msg("endpoint was not of type solana.endpoint. Skipping...")
				continue
			}

			solanaEndpoint = storedEndpointCast
		}

		// Apply the observation to the endpoint, whether it is already stored or not.
		if isMutated := solanaEndpoint.ApplyObservation(observation); !isMutated {
			// If the observation did not mutate the endpoint, don't update the stored endpoint entry.
			logger.Warn().Msg("endpoint was not mutated by observations. Skipping...")
			continue
		}

		// If the observation mutated the endpoint, update the stored endpoint entry.
		// A zero-value endpoint will always be mutated by an observation and so will be stored.
		updatedEndpoints[endpointAddr] = solanaEndpoint
	}

	return updatedEndpoints
}
