package evm

import (
	"context"
	"errors"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
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
	*serviceState
	*evmRequestValidator
}

// NewQoSInstance builds and returns an instance of the EVM QoS service.
func NewQoSInstance(logger polylog.Logger, config EVMServiceQoSConfig) *QoS {
	evmChainID := config.getEVMChainID()
	serviceId := config.GetServiceID()

	logger = logger.With(
		"qos_instance", "evm",
		"service_id", serviceId,
		"evm_chain_id", evmChainID,
	)

	store := &endpointStore{
		logger: logger,
		// Initialize the endpoint store with an empty map.
		endpoints: make(map[protocol.EndpointAddr]endpoint),
	}

	serviceState := &serviceState{
		logger:        logger,
		serviceConfig: config,
		endpointStore: store,
	}

	// TODO_CONSIDERATION(@olshansk): Archival checks are currently optional to enable iteration
	// and optionality. In the future, evaluate whether it should be mandatory for all EVM services.
	if config.archivalCheckEnabled() {
		serviceState.archivalState = archivalState{
			logger:              logger.With("state", "archival"),
			archivalCheckConfig: config.getEVMArchivalCheckConfig(),
			// Initialize the balance consensus map.
			// It keeps track and maps a balance (at the configured address and contract)
			// to the number of occurrences seen across all endpoints.
			balanceConsensus: make(map[string]int),
		}
	}

	evmRequestValidator := &evmRequestValidator{
		logger:       logger,
		serviceID:    serviceId,
		chainID:      evmChainID,
		serviceState: serviceState,
	}

	return &QoS{
		logger:              logger,
		serviceState:        serviceState,
		evmRequestValidator: evmRequestValidator,
	}
}

// ParseHTTPRequest builds a request context from an HTTP request.
// Returns (requestContext, true) if the request is valid JSONRPC
// Returns (errorContext, false) if the request is not valid JSONRPC.
//
// Implements gateway.QoSService interface.
func (qos *QoS) ParseHTTPRequest(_ context.Context, req *http.Request) (gateway.RequestQoSContext, bool) {
	return qos.evmRequestValidator.validateHTTPRequest(req)
}

// ParseWebsocketRequest builds a request context from the provided WebSocket request.
// WebSocket connection requests do not have a body, so we don't need to parse it.
//
// Implements gateway.QoSService interface.
func (qos *QoS) ParseWebsocketRequest(_ context.Context) (gateway.RequestQoSContext, bool) {
	return &requestContext{
		logger:       qos.logger,
		serviceState: qos.serviceState,
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

	updatedEndpoints := q.endpointStore.updateEndpointsFromObservations(
		evmObservations,
		q.serviceState.archivalState.blockNumberHex,
	)

	return q.serviceState.updateFromEndpoints(updatedEndpoints)
}
