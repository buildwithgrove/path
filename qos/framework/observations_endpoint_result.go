package framework

import (
	observations "github.com/buildwithgrove/path/observation/qos/framework"
	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/buildwithgrove/path/protocol"
)

// buildObservation converts an EndpointQueryResult to observations.EndpointQueryResult
// Used for reporting metrics.
func (eqr *EndpointQueryResult) buildObservation() *observations.EndpointQueryResult {
	if eqr == nil {
		return nil
	}

	// Create the observation result structure
	observationResult := &observations.EndpointQueryResult{
		StringValues: make(map[string]string),
		IntValues:    make(map[string]int64),
	}

	// Copy string values
	for key, value := range eqr.StringValues {
		observationResult.StringValues[key] = value
	}

	// Copy int values
	for key, value := range eqr.IntValues {
		observationResult.IntValues[key] = int64(value)
	}

	// Convert error information if available
	if eqr.Error != nil {
		observationResult.Error = eqr.Error.buildObservation()
	}

	// Set expiry time
	if !eqr.ExpiryTime.IsZero() {
		observationResult.ExpiryTime = timestampProto(eqr.ExpiryTime)
	}

	// Set HTTP response code if available from client response
	if eqr.clientResponse != nil && eqr.clientResponse.HTTPCode != 0 {
		observationResult.ClientHttpResponse = int32(eqr.clientResponse.HTTPCode)
	}

	return observationResult
}

// extractEndpointQueryResultFromObservation extracts a single EndpointQueryResult from an observation's EndpointQueryResult
// Ignores the HTTP stauts code: it is only required when responding to the client.
func extractEndpointQueryResultFromObservation(
	endpointQuery *endpointQuery,
	obsResult *observations.EndpointQueryResult,
) *EndpointQueryResult {
	if obsResult == nil {
		return nil
	}
	
	// Create a new result and populate it from the observation
	result := &EndpointQueryResult{
		// Set the endpointQuery underlying the observations.
		endpointQuery: endpointQuery,

		// Set the result values to be copied from the observations.
		StringValues: make(map[string]string),
		IntValues:    make(map[string]int),
		ExpiryTime:   timeFromProto(obsResult.ExpiryTime),
	}

	// Copy string values
	for key, value := range obsResult.StringValues {
		result.StringValues[key] = value
	}

	// Copy int values
	for key, value := range obsResult.IntValues {
		result.IntValues[key] = int(value)
	}

	// Convert error information
	if obsResult.Error != nil {
		result.Error = extractEndpointErrorFromObservation(obsResult.Error)
	}

	return result
}
