package evm

import (
	"encoding/json"
	"errors"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

var (
	// Error recording that endpoint selection was attempted but failed due to an invalid request
	errInvalidSelectorUsage = errors.New("endpoint selection attempted on failed request")
)

// errorContext provides the support required by the gateway package for handling service requests.
var _ gateway.RequestQoSContext = &errorContext{}

// requestContextFromInternalError creates an errorContext when encountering internal system errors.
// For example, failures while reading an HTTP request body.
func requestContextFromInternalError(logger polylog.Logger, err error, internalErrReason qosobservations.EVMRequestValidationErrorKind) errorContext {
	return errorContext{
		logger:                     logger,
		response:                   newErrResponseInternalErr(err),
		requestValidationErrorKind: &internalErrReason,
	}
}

// requestContextFromUserError creates an errorContext for client-side errors.
// FOr example, malformed JSON-RPC requests that fail to deserialize.
func requestContextFromUserError(logger polylog.Logger, err error, userErrReason qosobservations.EVMRequestValidationErrorKind) errorContext {
	return errorContext{
		logger:                     logger,
		response:                   newErrResponseInvalidRequest(err, jsonrpc.ID{}),
		requestValidationErrorKind: &userErrReason,
	}
}

// errorContext terminates EVM request processing on errors (internal failures or invalid requests).
// Provides:
//  1. Detailed error response to the user
//  2. Observation: feed into Metrics and data pipeline.
//
// Implements gateway.RequestQoSContext
type errorContext struct {
	logger polylog.Logger

	// chainID is the chain identifier for EVM QoS implementation.
	// Expected as the `Result` field in eth_chainId responses.
	chainID string

	// The response to be returned to the user.
	response jsonrpc.Response

	// Indicates why the request processing failed.
	requestValidationErrorKind *qosobservations.EVMRequestValidationErrorKind
}

// GetHTTPResponse formats the stored JSONRPC error as an HTTP response
// Implements the gateway.RequestQoSContext interface.
func (ec errorContext) GetHTTPResponse() gateway.HTTPResponse {
	bz, err := json.Marshal(ec.response)
	if err != nil {
		// TODO_IMPROVE(@adshmh): Standardize logger labels across packages
		// 1. Create shared label schema for the evm package
		// 2. Extend schema to other QoS packages
		ec.logger.With(
			"qos", "evm",
			"component", "errorContext",
			"method", "GetHTTPResponse",
		).Warn().Err(err).Msg("Failed to serialize client response.")
	}

	return httpResponse{
		responsePayload: bz,
	}
}

// GetObservation returns a QoS observation explaining why the request failed based on its error context.
// Implements the gateway.RequestQoSContext interface.
func (ec errorContext) GetObservations() qosobservations.Observations {
	return qosobservations.Observations{
		ServiceObservations: &qosobservations.Observations_Evm{
			Evm: &qosobservations.EVMRequestObservations{
				ChainId:                    ec.chainID,
				RequestValidationErrorKind: ec.requestValidationErrorKind,
			},
		},
	}
}

// GetServicePayload should never be called.
// It logs a warning and returns nil.
// Implements the gateway.RequestQoSContext interface.
func (ec errorContext) GetServicePayload() protocol.Payload {
	ec.logger.Warn().Msg("Invalid usage: errorContext.GetServicePayload() should never be called.")
	return protocol.Payload{}
}

// UpdateWithResponse logs a warning - should never be called.
// Implements the gateway.RequestQoSContext interface.
func (ec errorContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, endpointSerializedResponse []byte) {
	ec.logger.With(
		"endpoint_addr", endpointAddr,
		"endpoint_response_len", len(endpointSerializedResponse),
	).Warn().Msg("Invalid usage: errorContext.UpdateWithResponse() should never be called.")
}

// UpdateWithResponse logs a warning and returns an errorTrackingSelector - should never be called.
// Implements the gateway.RequestQoSContext interface.
func (ec errorContext) GetEndpointSelector() protocol.EndpointSelector {
	ec.logger.Warn().Msg("Invalid usage: errorContext.GetEndpointSelector() should never be called.")

	return errorTrackingSelector{
		logger: ec.logger,
	}
}

// errorTrackingSelector prevents panics in request handling goroutines by:
// - Intentionally failing all endpoint selection attempts 
// - Logging diagnostic information when endpoint selection is incorrectly attempted on failed requests
// Acts as a failsafe mechanism for request handling.
type errorTrackingSelector struct {
	logger polylog.Logger
}

// Select method of an errorTrackingSelector always returns an invalid usage error.
// Implements the protocol.EndpointSelector interface.
func (ets errorTrackingSelector) Select(endpoints []protocol.Endpoint) (protocol.EndpointAddr, error) {
	ets.logger.With(
		"num_endpoints", len(endpoints),
	).Warn().Msg("Invalid usage: errorTrackingSelector.Select() should never be called.")

	return protocol.EndpointAddr(""), errInvalidSelectorUsage
}
