package framework

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	observations "github.com/buildwithgrove/path/observation/qos/framework"
)

// buildObservation converts an EndpointQueryResult to observations.EndpointQueryResult
// Used for reporting metrics.
func (eqr *EndpointQueryResult) buildObservation(logger polylog.Logger) *observations.EndpointQueryResult {
	logger = logger.With("endpoint_addr", eqr.endpointAddr)

	// Create the observation result
	obs := &observations.EndpointQueryResult{
		// Store the endpoint address
		EndpointAddr: string(eqr.endpointAddr),
	}

	// This should never happen.
	// The parsed JSONRPC response is set by the framework, as either:
	// - Parsed from the payload returned by the service endpoint.
	// - A generic JSONRPC response if the above failed to parse into a JSONRPC response.
	if eqr.parsedJSONRPCResponse == nil {
		logger.Warn().Msg("Should never happen: EndpointQueryResult has no JSONRPC response set.")
	}

	// Set JSONRPC response
	if eqr.parsedJSONRPCResponse != nil {
		obs.JsonrpcResponse = buildObservationFromJSONRPCResponse(eqr.parsedJSONRPCResponse)
	}

	// Copy string values
	if len(eqr.StrValues) > 0 {
		obs.StringValues = make(map[string]string)
	}
	for key, value := range eqr.StrValues {
		obs.StringValues[key] = value
	}

	// Copy int values
	if len(eqr.IntValues) > 0 {
		obs.IntValues = make(map[string]int64)
	}
	for key, value := range eqr.IntValues {
		obs.IntValues[key] = int64(value)
	}

	// Convert error information if available
	if eqr.EndpointError != nil {
		obs.EndpointError = eqr.EndpointError.buildObservation()
	}

	// Set expiry time
	if !eqr.ExpiryTime.IsZero() {
		obs.ExpiryTime = timestampProto(eqr.ExpiryTime)
	}

	return obs
}

// buildEndpointQueryResultFromObservation builds a single EndpointQueryResult from an observation's EndpointQueryResult
func buildEndpointQueryResultFromObservation(
	logger polylog.Logger,
	observation *observations.EndpointQueryResult,
) *EndpointQueryResult {
	// hydrate the logger
	logger = logger.With("method", "extractEndpointQueryResultFromObservation")

	// Create a new result and populate it from the observation
	result := &EndpointQueryResult{
		// Set the result values to be copied from the observations.
		ExpiryTime: timeFromProto(observation.ExpiryTime),
	}

	// Copy string values
	if len(observation.StringValues) > 0 {
		result.StrValues = make(map[string]string)
	}
	for key, value := range observation.StringValues {
		result.StrValues[key] = value
	}

	// Copy int values
	if len(observation.IntValues) > 0 {
		result.IntValues = make(map[string]int)
	}
	for key, value := range observation.IntValues {
		result.IntValues[key] = int(value)
	}

	// Convert error information
	if endpointErr := observation.GetEndpointError(); endpointErr != nil {
		result.EndpointError = buildEndpointErrorFromObservation(endpointErr)
	}

	return result
}
