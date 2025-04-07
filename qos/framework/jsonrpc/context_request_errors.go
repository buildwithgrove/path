
////          THIS MAY NOT BE NEEDED with the ServiceRequestJournal
///
package jsonrpc

import (
	"encoding/json"
	"errors"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos/framework"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

var (
	// Error recording that endpoint selection was attempted but failed due to an invalid request
	errInvalidSelectorUsage = errors.New("endpoint selection attempted on failed request")
)

// errorContext provides the support required by the gateway package for handling service requests.
var _ gateway.RequestQoSContext = &errorContext{}

func newRequestErrorContextInternalError(logger polylog.Logger, err error) *errorContext {
	return &errorContext{
		logger: logger,

		// Create error response for read failure
		errResp := newErrResponseInternalReadError(err)

		requestCtx.JSONRPCErrorResponse = &errResp 
		rb.context = requestCtx
		return rb
}


// errorContext terminates EVM request processing on errors (internal failures or invalid requests).
// Provides:
//  1. Detailed error response to the user
//  2. Observation: feed into Metrics and data pipeline.
//
// Implements gateway.RequestQoSContext
type errorContext struct {
	logger polylog.Logger

	// The observation to return, to be processed by the metrics and data pipeline components.
	observations *qosobservations.Observations

	// The response to be returned to the user.
	response jsonrpc.Response
}

// GetHTTPResponse formats the stored JSONRPC error as an HTTP response
// Implements the gateway.RequestQoSContext interface.
func (ec *errorContext) GetHTTPResponse() gateway.HTTPResponse {
	return buildHTTPResponse(ec.Logger, ec.errorResponse)
}

// GetObservation returns the QoS observation set for the error context.
// Implements the gateway.RequestQoSContext interface.
func (ec *errorContext) GetObservations() qosobservations.Observations {
	return qosobservations.Observations{
		ServiceObservations: ec.evmObservations,
	}
}

// GetServicePayload should never be called.
// It logs a warning and returns nil.
// Implements the gateway.RequestQoSContext interface.
func (ec *errorContext) GetServicePayload() protocol.Payload {
	ec.logger.Warn().Msg("Invalid usage: errorContext.GetServicePayload() should never be called.")
	return protocol.Payload{}
}

// UpdateWithResponse should never be called.
// Only logs a warning.
// Implements the gateway.RequestQoSContext interface.
func (ec *errorContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, endpointSerializedResponse []byte) {
	ec.logger.With(
		"endpoint_addr", endpointAddr,
		"endpoint_response_len", len(endpointSerializedResponse),
	).Warn().Msg("Invalid usage: errorContext.UpdateWithResponse() should never be called.")
}

// UpdateWithResponse should never be called.
// It logs a warning and returns a failing selector that logs a warning on all selection attempts.
// Implements the gateway.RequestQoSContext interface.
func (ec *errorContext) GetEndpointSelector() protocol.EndpointSelector {
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

// Select method of an errorTrackingSelector should never be called.
// It logs a warning and returns an invalid usage error.
// Implements the protocol.EndpointSelector interface.
func (ets errorTrackingSelector) Select(endpoints []protocol.Endpoint) (protocol.EndpointAddr, error) {
	ets.logger.With(
		"num_endpoints", len(endpoints),
	).Warn().Msg("Invalid usage: errorTrackingSelector.Select() should never be called.")

	return protocol.EndpointAddr(""), errInvalidSelectorUsage
}
