package framework

import (
	observations "github.com/buildwithgrove/path/observation/qos/framework"
)

// getObservations returns the set of observations for the requestJournal.
// This includes:
// - Successful requests
// - Failed requests, due to:
//    - internal error:
//      - error reading HTTP request's body.
//      - any protocol-level error: e.g. endpoint timed out.
//    - invalid request
//
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

	// Add one endpoint observation entry per processed enpoint query stored in the journal.
	endpointObservations := make([]*qosobservations.EndpointObservation, len(rj.processedEndpointQueries))
	for index, endpointQuery := range rj.processedEndpointQueries {
		endpointObservations[index] = endpointQuery.buildObservation()
	}

	observations.EndpointObservations = endpointObservations
	return observations
}

// TODO_IN_THIS_PR: check the observations have the correct service name.
func (rj *requestJournal) extractEndpointQueriesFromObservations(observations *observations.Observations) []*endpointQuery {
	// fetch endpoint observations
	endpointObservations := observations.GetEndpointObservations()

	endpointQueries := make(*endpointQuery, len(endpointObservations))
	for index, endpointObservation := range endpointObservations {
		endpointQueries[index] = extractEndpointQueryFromObservation(endpointObservation)
	}

	return endpointQueries
}
