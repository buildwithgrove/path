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
	observations := qosobservations.RequestJournal {
		ServiceName: rj.serviceName,
	}

	// observation for parsed JSONRPC (if parsed)
	if rj.jsonrpcRequest != nil {
		observations.JsonRpcRequest = buildJSONRPCRequestObservation(rj.jsonrpcRequest)
	}

	// observation for request error (if set)
	if rj.requestErr != nil {
		observations.RequestError = buildRequestErrorObservations(rj.requestErr)
	}

	// No endpoint query results.
	// e.g. for invalid requests.
	// Skip adding endpoint observations.
	if len(rj.endpointQueryResults) == 0 {
		return observations
	}

	endpointObservations := make([]*qosobservations.EndpointQueryResultObservation, len(rj.endpointQueryResults))
	for index, endpointQueryResult := range rj.endpointQueryResults {
		endpointObservation[index] = endpointQueryResult.buildObservations()
	}

	observations.EndpointQueryResultObservations = endpointObservations
	return observations
}


func buildRequestJournalFromObservations(
	logger polylog.Logger,
	observations *qosobservations.Observations,
) (*requestJournal, error) {
	// hydrate the logger
	logger := logger.With("method", "buildRequestJournalFromObservations")

	// sanity check the observations.
	if observations == nil {
		errMsg := "Received nil observation: skip the processing."
		logger.Warn().Msg(errMsg)
		return nil, errors.New(errMsg)
	}

	reqObs := observations.GetRequestObservation()
	// No request observation present: skip the processing.
	if reqObs == nil {
		errMsg := "Received nil request observation: skip the processing."
		logger.Warn().Msg(errMsg)
		return nil, errors.New(errMsg)
	}

	// construct the request and any errors from the observations.
	jsonrpcRequest := buildJSONRPCRequestFromObservation(reqObs)
	requestErr := buildRequestErrorFromObservations(reqObs)

	// Instantiate the request journal.
	requestJournal := &requestJournal{
		logger: logger,
		jsonrpcRequest: jsonrpcRequest,
		requestErr: requestErr,
	}

	// request had an error: internal, parsing, validation, etc.
	// no further processing required.
	if requestErr != nil {
		logger.With("num_endpoint_observations", len(observations.GetEndpointQueryResultObservatios()).
			Info().Msg("Request had an error: no endpoint observations expected.")

		return requestJournal
	}

	// reconstruct endpoint query results.
	endpointsObs := observations.GetEndpointQueyResultObservations()
	// No endpoint observation present: skip the processing.
	if endpointsObs == nil || len(endpointsObs) == 0 {
		logger.Warn().Msg("Received nil/empty endpoint observation: skip the processing.")
		return nil
	}

	// Initialize the endpoint query results of the request journal.
	requestJournal.endpointQueryResults = make([]*EndpointQueryResult, len(endpointsObs))

	// add one endpoint query result per endpoint observation.
	for index, endpointObs := range endpointsObs {
		// Construct the query result from the endpoint observation.
		endpointQueryResult := extractEndpointQueryResultsFromObservations(endpointObs)

		// add a reference to the request journal: e.g. for retrieving the JSONRPC request method.
		endpointQueryResult.requestJournal = requestJournal

		// add the endpoint query result to the request journal.
		requestJournal.endpointQueryResults[index] = endpointQueryResult
	}

	return requestJournal, nil
}
