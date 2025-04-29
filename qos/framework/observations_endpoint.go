package framework

import (
	observations "github.com/buildwithgrove/path/observation/qos/framework"
)

func extractEndpointQueryFromObservation(observation *qosobservations.Observations) *endpointQuery {
	return &endpointQuery{
		// Extract the JSONRPC request corresponding to the observation.
		request: extractJSONRPCRequestFromObservation(observation.GetRequestObservation()),
	}
}
