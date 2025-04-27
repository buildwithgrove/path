package framework

// TODO_IN_THIS_PR: verify the EmptyResponse and NoResponse scenarios:
// - EmptyResponse is an EndpointQueryResult, because the endpoint did return an empty payload.
// - NoReponse is a requestError: e.g. there may have been ZERO ENDPOINTS available at the PROTOCOL-LEVEL.
//   - It is an INTERNAL error: like failing to read HTTP request's body.

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

	requestDetails *requestDetails	

	// All endpoint interactions that occurred during processing.
	// These are expected to be processed: i.e. have a non-nil result pointer and a client JSONRPC response.
	processedEndpointQueries []*endpointQuery
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

// TODO_FUTURE(@adshmh): A retry mechanism would require support from this struct to determine if the most recent endpoint query has been successful.
//
// getHTTPResponse returns the client's HTTP response:
// - Uses the request error if set
// - Uses the most recent endpoint query if the request has no errors set.
func (rj *requestJournal) getHTTPResponse() gateway.HTTPResponse {
	// For failed requests, return the preset JSONRPC error response.
	// - Invalid request: e.g. malformed payload from client.
	// - Internal error: error reading HTTP request's body
	// - Internal error: Protocol-level error, e.g. selected endpoint timed out.
	if requestErrorJSONRPCResponse := rj.requestDetails.getRequestErrorJSONRPCResponse(); requestErrorJSONRPCResponse != nil {
		return buildHTTPResponse(rj.Logger, requestErrorJSONRPCResponse)
	}

	// TODO_IMPROVE(@adshmh): find a refactor:
	// Goal: guarantee that valid request -> at least 1 endpoint query.
	// Constraint: Such a refactor should keep the requestJournal as a data container.
	//
	// Use the most recently reported endpoint query.
	// There MUST be an entry if the request has no error set.
	selectedQuery := rj.processedEndpointQueries[len(rj.processedEndpointQueries)-1]
	jsonrpcResponse := selectedQuery.result.clientJSONRPCResponse
	return buildHTTPResponse(rj.Logger, jsonrpcResponse)
}
