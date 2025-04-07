package jsonrpc

const (
	// TODO_MVP(@adshmh): Support individual configuration of timeout for every service that uses EVM QoS.
	// The default timeout when sending a request to an EVM blockchain endpoint.
	defaultServiceRequestTimeoutMillisec = 10000
)

// requestJournal holds the data for a complete JSONRPC request lifecycle.
type requestJournal struct {
	logger polylog.Logger

	// Service identification
	serviceName string

	// The client's JSONRPC request
	request *jsonrpc.Request

	// Error response to return if a request parsing error occurred:
	// - error reading HTTP request's body.
	// - error parsing the request's payload into a jsonrpc.Request struct.
	errorResponse *jsonrpc.Response

	// All endpoint interactions that occurred during processing.
	endpointQueries []*endpointQuery
}

func (rj *requestJournal) buildEndpointQuery(endpointAddr protocol.EndpointAddr, receivedData []byte) *endpointQuery {
	return &endpointQuery{
		request:      rj.request,
		endpointAddr: endpointAddr,
		receivedData: receivedData,
	}
}

func (rj *requestJournal) reportProcessedEndpointQuery(processedEndpointQuery endpointQuery) {
	rj.endpointQueries = append(rj.endpointQueries, processedEndpointQuery)
}

func (rj *requestJournal) getServicePayload() protocol.Payload {
	// TODO_IN_THIS_PR: update this code
	reqBz, err := json.Marshal(*rc.Request)
	if err != nil {
		// TODO_MVP(@adshmh): find a way to guarantee this never happens,
		// e.g. by storing the serialized form of the JSONRPC request
		// at the time of creating the request context.
		return protocol.Payload{}
	}

	return protocol.Payload{
		Data: string(reqBz),
		// Method is alway POST for EVM-based blockchains.
		Method: http.MethodPost,

		// Path field is not used for JSONRPC services.

		// TODO_IMPROVE: adjust the timeout based on the request method:
		// An endpoint may need more time to process certain requests,
		// as indicated by the request's method and/or parameters.
		TimeoutMillisec: defaultServiceRequestTimeoutMillisec,
	}
}

func (rj *requestJournal) getHTTPResponse() gateway.HTTPResponse {
	if rj.JSONRPCErrorResponse != nil {
		return buildHTTPResponse(rj.Logger, rj.JSONRPCErrorResponse)
	}

	return buildHTTPResponse(rj.Logger, rj.getJSONRPCResponse())
}

func (rj *requestJournal) getObservations() qosobservations.Observations {
	/*
		// Service identification
		ServiceName:
		ServiceDescription:

		// Observation of the client request
		RequestObservation *RequestObservation
		// Observations from endpoint(s)
		EndpointObservations []*EndpointObservation

		return qosobservations.Observations {
			RequestObservations: rc. resut???? .GetObservation(),
			EndpointObservations: rc.EndpointCallsProcessor.GetObservations(),
		// TODO_IN_THIS_PR: Implement this method in observations.go.
		// Return basic observations for now
			ServiceId:          p.ServiceID,
			ServiceDescription: p.ServiceDescription,
			RequestObservation: p.RequestObservation,
		}
	*/
}

// TODO_FUTURE(@adshmh): A retry mechanism would require support from this struct to determine if the most recent endpoint query has been successful.
//
// getJSONRPCResponse simply returns the result associated with the most recently reported endpointQuery.
func (rj *requestJournal) getJSONRPCResponse() *jsonrpc.Response {
	// Check if we received any endpoint results
	if len(rc.processedResults) == 0 {
		// If no results were processed, handle it as a protocol error
		return buildResultForNoResponse(rc.Request)
	}

	// Return the latest result.
	return rc.processedResults[len(rc.processedResults)-1]
}
