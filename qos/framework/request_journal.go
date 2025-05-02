package framework

import (
	"encoding/json"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_IN_THIS_PR: verify the EmptyResponse and NoResponse scenarios:
// - EmptyResponse is an EndpointQueryResult, because the endpoint did return an empty payload.
// - NoReponse is a requestError: e.g. there may have been ZERO ENDPOINTS available at the PROTOCOL-LEVEL.
//   - It is an INTERNAL error: like failing to read HTTP request's body.

const (
	// TODO_MVP(@adshmh): Support individual configuration of timeout for every service that uses EVM QoS.
	// The default timeout when sending a request to an EVM blockchain endpoint.
	defaultServiceRequestTimeoutMillisec = 10000

	// maximum length of the error message stored in request validation failure observations and logs.
	// This is used to prevent overly verbose error messages from being stored in logs and metrics leading to excessive memory usage and cost.
	maxErrMessageLen = 1000
)

// requestJournal holds the data for a complete JSONRPC request lifecycle.
type requestJournal struct {
	logger polylog.Logger

	// Service identification
	serviceName string

	// The client's JSONRPC request
	// Only set if the request was successfully parsed.
	jsonrpcRequest *jsonrpc.Request

	// Request error, if any.
	requestError *requestError

	// All endpoint interactions that occurred during processing.
	endpointQueryResults []*EndpointQueryResult
}

func (rj *requestJournal) setProtocolLevelError() {
	// request already marked as failed.
	// skip setting an error.
	if rj.requestError != nil {
		return
	}

	// set the request as failed with protocol-level error.
	rj.requestError = buildRequestErrorForInternalErrProtocolErr(rj.jsonrpcRequest.ID)
}

func (rj *requestJournal) buildEndpointQueryResult(endpointAddr protocol.EndpointAddr, receivedData []byte) *EndpointQueryResult {
	return &EndpointQueryResult{
		// request journal reference.
		// Used to access the request, e.g. JSONRPC request method.
		requestJournal: rj,

		// Address of the queried endpoint.
		endpointAddr: endpointAddr,
		// Data received from the endpoint.
		endpointPayload: receivedData,

		// Initialize attribute maps
		IntValues: make(map[string]int),
		StrValues: make(map[string]string),
	}
}

func (rj *requestJournal) reportEndpointQueryResult(endpointQueryResult *EndpointQueryResult) {
	rj.endpointQueryResults = append(rj.endpointQueryResults, endpointQueryResult)
}

// Example: setting protocol-level error, i.e. no endpoint responses were received.
func (rj *requestJournal) setRequestError(requestErr *requestError) {
	rj.requestError = requestErr
}

func (rj *requestJournal) getServicePayload() protocol.Payload {
	// Sanity check the request fields.
	// A non-nil requestErr indicates the request failed to parse/validate.
	if rj.requestError != nil {
		rj.logger.With("request_error", rj.requestError).Error().Msg("Should never happen: getServicePayload() called for invalid/failed request.")
		return protocol.Payload{}
	}

	// JSONRPC request not set on the journal: skip the processing.
	if rj.jsonrpcRequest == nil {
		rj.logger.Error().Msg("Should never happen: getServicePayload() called with nil JSONRPC request.")
		return protocol.Payload{}
	}

	reqBz, err := json.Marshal(*rj.jsonrpcRequest)
	if err != nil {
		// TODO_MVP(@adshmh): find a way to guarantee this never happens,
		// e.g. by storing the serialized form of the JSONRPC request
		// at the time of creating the request context.
		rj.logger.With("marshal_err", err).Error().Msg("Should never happen: getServicePayload() failed to marshal the JSONRPC request.")
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
	if requestErr := rj.requestError; requestErr != nil {
		return buildHTTPResponse(rj.logger, requestErr.jsonrpcErrorResponse)
	}

	// Use the most recently reported endpoint query.
	// There MUST be an entry if the request has no error set.
	selectedEndpointQueryResult := rj.endpointQueryResults[len(rj.endpointQueryResults)-1]
	return buildHTTPResponse(rj.logger, selectedEndpointQueryResult.parsedJSONRPCResponse)
}

func (rj *requestJournal) getJSONRPCRequestMethod() jsonrpc.Method {
	request := rj.jsonrpcRequest
	if request == nil {
		return jsonrpc.Method("")
	}

	return request.Method
}

func (rj *requestJournal) getJSONRPCRequestID() jsonrpc.ID {
	request := rj.jsonrpcRequest
	if request == nil {
		return jsonrpc.ID{}
	}

	return request.ID
}
