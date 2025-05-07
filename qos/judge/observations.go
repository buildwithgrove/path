package judge

import (
	"errors"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	observations "github.com/buildwithgrove/path/observation/qos/framework"
)

// getObservations returns the set of observations for the requestJournal.
// This includes:
// - Successful requests
// - Failed requests, due to:
//   - internal error:
//   - error reading HTTP request's body.
//   - any protocol-level error: e.g. endpoint timed out.
//   - invalid request
//
// requestJournal is the top-level struct in the chain of observation generators.
func (rj *requestJournal) getObservations() qosobservations.Observations {
	// initialize the observations to include:
	// - service name
	// - observations related to the request:
	journalObservations := observations.RequestJournalObservations{
		ServiceName: rj.serviceName,
	}

	// observation for parsed JSONRPC (if parsed)
	if rj.jsonrpcRequest != nil {
		journalObservations.JsonrpcRequest = buildJSONRPCRequestObservation(rj.jsonrpcRequest)
	}

	// observation for request error (if set)
	if rj.requestError != nil {
		journalObservations.RequestError = rj.requestError.buildObservation()
	}

	// No endpoint query results.
	// e.g. for invalid requests.
	// Skip adding endpoint observations.
	if len(rj.endpointQueryResults) == 0 {
		return qosobservations.Observations{
			ServiceObservations: &qosobservations.Observations_RequestJournalObservations{
				RequestJournalObservations: &journalObservations,
			},
		}
	}

	endpointObservations := make([]*observations.EndpointQueryResult, len(rj.endpointQueryResults))
	for index, endpointQueryResult := range rj.endpointQueryResults {
		endpointObservations[index] = endpointQueryResult.buildObservation(rj.logger)
	}

	journalObservations.EndpointQueryResultObservations = endpointObservations
	return qosobservations.Observations{
		ServiceObservations: &qosobservations.Observations_RequestJournalObservations{
			RequestJournalObservations: &journalObservations,
		},
	}
}

func buildRequestJournalFromObservations(
	logger polylog.Logger,
	journalObs *observations.RequestJournalObservations,
) (*requestJournal, error) {
	// construct the request and any errors from the observations.
	reqObs := journalObs.GetJsonrpcRequest()
	// nil request observation: no further processing can be done.
	if reqObs == nil {
		errMsg := "Should happen very rarely: received nil JSONRPC request observation: skip the processing."
		logger.Warn().Msg(errMsg)
		return nil, errors.New(errMsg)
	}

	// Construct the JSONRPC request from the observation.
	// Only the JSONRPC request method is required: to build endpoint query result.
	jsonrpcRequest := buildJSONRPCRequestFromObservation(reqObs)

	// Instantiate the request journal.
	requestJournal := &requestJournal{
		logger:         logger,
		jsonrpcRequest: jsonrpcRequest,
	}

	// hydrate the logger with endpoint observations count.
	numEndpointObservations := len(journalObs.GetEndpointQueryResultObservations())
	logger = logger.With("num_endpoint_observations", numEndpointObservations)

	requestErrObs := journalObs.GetRequestError()
	// request had an error: internal, parsing, validation, etc.
	// no further processing required.
	if requestErrObs != nil {
		requestJournal.requestError = buildRequestErrorFromObservation(requestErrObs)

		// hydrate the logger with request error kind.
		logger := logger.With("request_error_kind", requestJournal.requestError.errorKind)

		// Request with an error had one or more endpoint observations: this should not happen.
		if numEndpointObservations > 0 {
			errMsg := "Should happen very rarely: received request with both an error and non-zero observations: skip the processing."
			logger.Warn().Msg(errMsg)
			return nil, errors.New(errMsg)
		}

		logger.Debug().Msg("Successfully parsed the request journal from observations.")
		return requestJournal, nil
	}

	// reconstruct endpoint query results.
	endpointsObs := journalObs.GetEndpointQueryResultObservations()

	// No endpoint observation present: skip the processing.
	if endpointsObs == nil || len(endpointsObs) == 0 {
		errMsg := "Should happen very rarely: received nil endpoint observations, but the request has no error set: skip the processing."
		logger.Warn().Msg(errMsg)
		return nil, errors.New(errMsg)
	}

	// Initialize the endpoint query results of the request journal.
	requestJournal.endpointQueryResults = make([]*EndpointQueryResult, len(endpointsObs))

	// add one endpoint query result per endpoint observation.
	for index, endpointObs := range endpointsObs {
		// Construct the query result from the endpoint observation.
		endpointQueryResult := buildEndpointQueryResultFromObservation(logger, endpointObs)

		// add a reference to the request journal: e.g. for retrieving the JSONRPC request method.
		endpointQueryResult.requestJournal = requestJournal

		// add the endpoint query result to the request journal.
		requestJournal.endpointQueryResults[index] = endpointQueryResult
	}

	return requestJournal, nil
}
