package framework

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	observations "github.com/buildwithgrove/path/observation/qos/framework"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)


=======>>>>>>
		// Convert expiry timestamp if available
		if !ee.RecommendedSanction.ExpiryTimestamp.IsZero() {
			// Convert Go time.Duration to proto timestamp
			observationError.Sanction.ExpiryTimestamp = timestampProto(time.Now().Add(ee.RecommendedSanction.Duration))
		}

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

// buildEndpointQueryResultFromObservation builds a single EndpointQueryResult from an observation's EndpointQueryResult
func buildEndpointQueryResultFromObservation(
	logger polylog.Logger,
	observation *observations.EndpointQueryResult,
) *EndpointQueryResult {
	// hydrate the logger
	logger := logger.With("method", "extractEndpointQueryResultFromObservation")

	// Create a new result and populate it from the observation
	result := &EndpointQueryResult{
		// Set the result values to be copied from the observations.
		StringValues: make(map[string]string),
		IntValues:    make(map[string]int),
		ExpiryTime:   timeFromProto(observation.ExpiryTime),
	}

	// Copy string values
	for key, value := range observation.StringValues {
		result.StringValues[key] = value
	}

	// Copy int values
	for key, value := range observation.IntValues {
		result.IntValues[key] = int(value)
	}

	// Convert error information
	if endpointErr := observation.GetEndpointError(); endpointError != nil {
		result.EndpointError = extractEndpointErrorFromObservation(endpointError)
	}

	return result
}
