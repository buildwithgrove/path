package framework

import (
	"github.com/buildwithgrove/path/observation/qos/jsonrpc"
	observations "github.com/buildwithgrove/path/observation/qos/framework"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

func (eq *endpointQuery) buildObservations() *observations.Observations {
	return observations.Observations{
		// Service info
		ServiceName:
		ServiceDescription:

		// Observation of the client request
		RequestObservation *RequestObservation
		// Observations from endpoint(s)
		EndpointObservations []*EndpointObservation

	}
}

func (eqr *EndpointQueryResult) buildObservations() *observations.EndpointQueryResult

func extractEndpointQueryResults(observations *observations.Observations) []*EndpointQueryResult

