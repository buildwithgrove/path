package qos

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

// RequestErrorContext provides the support required by the gateway package for handling service requests.
var _ gateway.RequestQoSContext = &RequestErrorContext{}

// RequestErrorContext terminates the processing of a JSONRPC-service request on errors (internal failures or invalid requests).
// Provides:
//  1. Detailed error response to the user.
//  2. Log entries to warn on potential incorrect usage.
//
// Implements gateway.RequestQoSContext
type RequestErrorContext struct {
	Logger polylog.Logger

	// The response to be returned to the user.
	Response jsonrpc.Response

	// The observations to use for the error.
	Observations *qosobservations.Observations
}

// GetHTTPResponse formats the stored JSONRPC error as an HTTP response
// Implements the gateway.RequestQoSContext interface.
func (rec *RequestErrorContext) GetHTTPResponse() gateway.HTTPResponse {
	bz, err := json.Marshal(rec.Response)
	if err != nil {
		rec.Logger.With(
			"component", "RequestErrorContext",
			"method", "GetHTTPResponse",
		).Warn().Err(err).Msg("Failed to serialize client response.")
	}

	httpStatusCode := rec.Response.GetRecommendedHTTPStatusCode()

	return HTTPResponse{
		responsePayload: bz,
		httpStatusCode:  httpStatusCode,
	}
}

// TODO_MVP(@adshmh): Generate observations for the error context.
// GetObservation returns the QoS observation set for the error context.
// Implements the gateway.RequestQoSContext interface.
func (rec *RequestErrorContext) GetObservations() qosobservations.Observations {
	return qosobservations.Observations{
		ServiceObservations: rec.Observations.ServiceObservations,
	}
}

// GetServicePayload should never be called.
// It logs a warning and returns nil.
// Implements the gateway.RequestQoSContext interface.
func (rec *RequestErrorContext) GetServicePayload() protocol.Payload {
	rec.Logger.Warn().Msg("Invalid usage: RequestErrorContext.GetServicePayload() should never be called.")
	// POTENTIAL NULL ISSUE: Returning empty Payload{} here could lead to "null" being sent
	// as request body if this error context is used in request processing pipeline.
	// This may be related to RelayMiner receiving "null" (base64: bnVsbA==) as request body.
	return protocol.Payload{}
}

// UpdateWithResponse should never be called.
// Only logs a warning.
// Implements the gateway.RequestQoSContext interface.
func (rec *RequestErrorContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, endpointSerializedResponse []byte) {
	rec.Logger.With(
		"endpoint_addr", endpointAddr,
		"endpoint_response_len", len(endpointSerializedResponse),
	).Warn().Msg("Invalid usage: RequestErrorContext.UpdateWithResponse() should never be called.")
}

// UpdateWithResponse should never be called.
// It logs a warning and returns a failing selector that logs a warning on all selection attempts.
// Implements the gateway.RequestQoSContext interface.
func (rec *RequestErrorContext) GetEndpointSelector() protocol.EndpointSelector {
	rec.Logger.Warn().Msg("Invalid usage: RequestErrorContext.GetEndpointSelector() should never be called.")

	return errorTrackingSelector{
		logger: rec.Logger,
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
func (ets errorTrackingSelector) Select(endpoints protocol.EndpointAddrList) (protocol.EndpointAddr, error) {
	ets.logger.With(
		"num_endpoints", len(endpoints),
	).Warn().Msg("Invalid usage: errorTrackingSelector.Select() should never be called.")

	return protocol.EndpointAddr(""), errInvalidSelectorUsage
}

// SelectMultiple method of an errorTrackingSelector should never be called.
// It logs a warning and returns an invalid usage error.
// Implements the protocol.EndpointSelector interface.
func (ets errorTrackingSelector) SelectMultiple(endpoints protocol.EndpointAddrList, numEndpoints uint) (protocol.EndpointAddrList, error) {
	ets.logger.Warn().Msg("Invalid usage: errorTrackingSelector.SelectMultiple() should never be called.")

	return nil, errInvalidSelectorUsage
}
