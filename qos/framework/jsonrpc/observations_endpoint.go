package jsonrpc

import (
	observations "github.com/buildwithgrove/path/observation/qos/framework"
)

func (eq *endpointQuery) buildObservations() *qosobservations.EndpointObservation {
	return &qosobservations.EndpointObservation{
		EndpointAddr: string(eq.endpointAddr),
		EndpointQueryResult: result.buildObservation(), 
	}
}

func extractEndpointQueryFromObservation(observation *qosobservations.EndpointObservation) *endpointQuery {
	return &endpointQuery {
		endpointAddr: observation.GetEndpointAddr(),
		// Single result item extracted from this endpoint query.
		result: extractEndpointQueryResultFromObservation(observation.GetResult()),
	}
}
