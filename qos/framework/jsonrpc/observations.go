package framework

import (
	"github.com/buildwithgrove/path/observation/qos/jsonrpc"
	observations "github.com/buildwithgrove/path/observation/qos/framework"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// getObservations returns the set of observations for the requestJournal.
// This includes:
// - Successful requests
// - Failed requests: internal error
// - Failed requests: invalid request
// - Failed requests: protcol error, i.e. no endpoint data received.
// requestJournal is the top-level struct in the chain of observation generators.
func (rj *requestJournal) getObservations() qosobservations.Observations {
	// initialize the observations to include:
	// - service name
	// - observations related to the request:
	observations := qosobservations.Observations {
		ServiceName: rj.serviceName,
		// request observations:
		//   - parsed JSONRPC (if successful)
		//   - validation error (if invalid)
		RequestObservation: rj.requestDetails.buildObservation(),
	}

	// No endpoint queries were performed: skip adding endpoint observations.
	// e.g. for invalid requests.
	if len(rj.endpointQueries) == 0 {
		return observations
	}

	endpointObservations := make([]*qosobservations.EndpointObservation, len(rj.endpointQueries))
	for index, endpointQuery := range rj.endpointQueries {
		endpointObservations[index] = endpointQuery.buildObservation()
	}

	observations.EndpointObservations = endpointObservations
	return observations
}

func extractEndpointQueryResults(observations *observations.Observations) []*EndpointQueryResult

