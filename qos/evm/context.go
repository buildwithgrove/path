package evm

import (
	"encoding/json"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"

	"github.com/buildwithgrove/path/gateway"
	pathhttp "github.com/buildwithgrove/path/network/http"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// requestContext provides the support required by the gateway
// package for handling service requests.
var _ gateway.RequestQoSContext = &requestContext{}

// requestContext provides the endpoint selection capability required
// by the protocol package for handling a service request.
var _ protocol.EndpointSelector = &requestContext{}

// TODO_REFACTOR: Improve naming clarity by distinguishing between interfaces and adapters
// in the metrics/qos/evm and qos/evm packages, and elsewhere names like `response` are used.
// Consider renaming:
//   - metrics/qos/evm: response → EVMMetricsResponse
//   - qos/evm: response → EVMQoSResponse
//   - observation/evm: observation -> EVMObservation
//
// TODO_TECHDEBT: Need to add a Validate() method here to allow the caller (e.g. gateway)
// determine whether the endpoint's response was valid, and whether a retry makes sense.
//
// response defines the functionality required from a parsed endpoint response, which all response types must implement.
// It provides methods to:
//  1. Generate observations for endpoint quality tracking
//  2. Format HTTP responses to send back to clients
type response interface {
	// GetObservation returns an observation of the endpoint's response
	// for quality metrics tracking, including HTTP status code.
	GetObservation() qosobservations.EVMEndpointObservation

	// GetHTTPResponse returns the HTTP response to be sent back to the client.
	GetHTTPResponse() httpResponse

	GetJSONRPCID() jsonrpc.ID
}

var _ response = &endpointResponse{}

type endpointResponse struct {
	protocol.EndpointAddr
	response
	unmarshalErr error
}

// requestContext implements the functionality for EVM-based blockchain services.
type requestContext struct {
	logger polylog.Logger

	// chainID is the chain identifier for EVM QoS implementation.
	// Expected as the `Result` field in eth_chainId responses.
	chainID string

	// service_id is the identifier for the evm QoS implementation.
	// It is the "alias" or human readable interpratation of the chain_id.
	// Used in generating observations.
	serviceID protocol.ServiceID

	// The origin of the request handled by the context.
	// Either:
	// - Organic: user requests
	// - Synthetic: requests built by the QoS service to get additional data points on endpoints.
	requestOrigin qosobservations.RequestOrigin

	// The length of the request payload in bytes.
	requestPayloadLength uint

	serviceState *serviceState

	// JSON-RPC requests - supports both single and batch requests per JSON-RPC 2.0 spec
	jsonrpcReqs map[string]jsonrpc.Request

	// endpointResponses is the set of responses received from one or
	// more endpoints as part of handling this service request.
	// Supports both single and batch JSON-RPC requests.
	endpointResponses []endpointResponse

	// endpointSelectionMetadata contains metadata about the endpoint selection process
	endpointSelectionMetadata EndpointSelectionMetadata
}

// GetServicePayloads returns the service payloads for the JSON-RPC requests in the request context.
//
// jsonrpcReqs is a map of JSON-RPC request IDs to JSON-RPC requests.
// It is assigned to the request context in the `evmRequestValidator.validateHTTPRequest` method.
//
// TODO_MVP(@adshmh): Ensure the JSONRPC request struct can handle all valid service requests.
func (rc requestContext) GetServicePayloads() []protocol.Payload {
	var payloads []protocol.Payload

	for _, req := range rc.jsonrpcReqs {
		reqBz, err := json.Marshal(req)
		if err != nil {
			rc.logger.Error().Err(err).Msg("SHOULD RARELY HAPPEN: requestContext.GetServicePayload() should never fail marshaling the JSONRPC request.")
			return []protocol.Payload{protocol.EmptyErrorPayload()}
		}

		payloads = append(payloads, protocol.Payload{
			Data:    string(reqBz),
			Method:  http.MethodPost, // Method is always POST for EVM-based blockchains.
			Headers: map[string]string{},
			RPCType: sharedtypes.RPCType_JSON_RPC,
		})
	}

	return payloads
}

// UpdateWithResponse is NOT safe for concurrent use
func (rc *requestContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, responseBz []byte) {
	// TODO_IMPROVE: check whether the request was valid, and return an error if it was not.
	// This would be an extra safety measure, as the caller should have checked the returned value
	// indicating the validity of the request when calling on QoS instance's ParseHTTPRequest

	response, err := unmarshalResponse(rc.logger, rc.jsonrpcReqs, responseBz, endpointAddr)

	rc.endpointResponses = append(rc.endpointResponses, endpointResponse{
		EndpointAddr: endpointAddr,
		response:     response,
		unmarshalErr: err,
	})
}

// GetHTTPResponse builds the HTTP response that should be returned for
// an EVM blockchain service request.
// Implements the gateway.RequestQoSContext interface.
func (rc requestContext) GetHTTPResponse() pathhttp.HTTPResponse {
	// Use a noResponses struct if no responses were reported by the protocol from any endpoints.
	if len(rc.endpointResponses) == 0 {
		responseNoneObj := responseNone{
			logger:      rc.logger,
			jsonrpcReqs: rc.jsonrpcReqs,
		}

		return responseNoneObj.GetHTTPResponse()
	}

	if len(rc.jsonrpcReqs) == 1 {
		// return the only endpoint response reported to the context for single requests.
		return rc.endpointResponses[0].GetHTTPResponse()
	}

	// Handle batch requests according to JSON-RPC 2.0 specification
	// https://www.jsonrpc.org/specification#batch
	return rc.getBatchHTTPResponse()
}

// getBatchHTTPResponse handles batch requests by combining individual JSON-RPC responses
// into an array according to the JSON-RPC 2.0 specification.
// https://www.jsonrpc.org/specification#batch
func (rc requestContext) getBatchHTTPResponse() pathhttp.HTTPResponse {
	// Collect individual response payloads
	var individualResponses []json.RawMessage

	// Process each endpoint response
	for _, endpointResp := range rc.endpointResponses {
		individualHTTPResp := endpointResp.GetHTTPResponse()

		// Extract the JSON payload from each response
		payload := individualHTTPResp.GetPayload()
		if len(payload) > 0 {
			individualResponses = append(individualResponses, json.RawMessage(payload))
		}
	}

	// According to JSON-RPC spec: "If there are no Response objects contained within the Response array
	// as it is to be sent to the client, the server MUST NOT return an empty Array and should return nothing at all."
	// This can happen when all requests in the batch are notifications (which don't get responses)
	// or when all individual responses are empty/invalid.
	if len(individualResponses) == 0 {
		emptyBatchResponse := getGenericResponseBatchEmpty(rc.logger)
		return emptyBatchResponse.GetHTTPResponse()
	}

	// Combine individual responses into a JSON array
	batchResponse, err := json.Marshal(individualResponses)
	if err != nil {
		// Create a responseGeneric for batch marshaling failure and return its HTTP response
		errorResponse := getGenericJSONRPCErrResponseBatchMarshalFailure(rc.logger, err)
		return errorResponse.GetHTTPResponse()
	}

	return httpResponse{
		responsePayload: batchResponse,
		// According to the JSON-RPC 2.0 specification, even if individual responses
		// in a batch contain errors, the entire batch should still return HTTP 200 OK.
		httpStatusCode: http.StatusOK,
	}
}

// GetObservations returns all endpoint observations from the request context.
// Implements gateway.RequestQoSContext interface.
func (rc requestContext) GetObservations() qosobservations.Observations {
	// Create observations for each JSON-RPC request in the batch (or single request)
	requestObservations := rc.createRequestObservations()

	// Convert endpoint selection validation results to proto format
	validationResults := rc.convertValidationResults()

	return qosobservations.Observations{
		ServiceObservations: &qosobservations.Observations_Evm{
			Evm: &qosobservations.EVMRequestObservations{
				ChainId:              rc.chainID,
				ServiceId:            string(rc.serviceID),
				RequestPayloadLength: uint32(rc.requestPayloadLength),
				RequestOrigin:        rc.requestOrigin,
				RequestObservations:  requestObservations,
				EndpointSelectionMetadata: &qosobservations.EndpointSelectionMetadata{
					RandomEndpointFallback: rc.endpointSelectionMetadata.RandomEndpointFallback,
					ValidationResults:      validationResults,
				},
			},
		},
	}
}

// createRequestObservations creates observations for all JSON-RPC requests in the batch.
// For batch requests, jsonrpcReqs contains multiple requests keyed by their ID strings.
// For single requests, jsonrpcReqs contains one request.
// Each observation correlates a JSON-RPC request with its corresponding endpoint response(s).
func (rc requestContext) createRequestObservations() []*qosobservations.EVMRequestObservation {
	// Handle the special case where no endpoint responses were received
	if len(rc.endpointResponses) == 0 {
		return rc.createNoResponseObservations()
	}

	// Create observations by correlating endpoint responses with their corresponding JSON-RPC requests
	return rc.createResponseObservations()
}

// createNoResponseObservations creates a single observation when no endpoint responses were received.
// This can happen when all endpoints are unreachable or fail to respond.
// The observation includes all JSON-RPC requests from the batch but no endpoint observations.
func (rc requestContext) createNoResponseObservations() []*qosobservations.EVMRequestObservation {
	responseNoneObj := responseNone{
		logger:      rc.logger,
		jsonrpcReqs: rc.jsonrpcReqs,
	}
	responseNoneObs := responseNoneObj.GetObservation()

	return []*qosobservations.EVMRequestObservation{
		{EndpointObservations: []*qosobservations.EVMEndpointObservation{
			&responseNoneObs,
		}},
	}
}

// createResponseObservations creates observations by correlating endpoint responses with their
// corresponding JSON-RPC requests from the batch. Each endpoint response contains a JSON-RPC ID
// that is used to look up the original request in the jsonrpcReqs map.
//
// For batch requests: multiple responses are correlated with multiple requests
// For single requests: one response is correlated with one request
func (rc requestContext) createResponseObservations() []*qosobservations.EVMRequestObservation {
	var observations []*qosobservations.EVMRequestObservation

	for _, endpointResp := range rc.endpointResponses {
		responseIDStr := endpointResp.GetJSONRPCID().String()

		// Look up the original JSON-RPC request using the response ID
		// This correlation is critical for batch requests where multiple requests/responses
		// need to be properly matched
		jsonrpcReq, ok := rc.jsonrpcReqs[responseIDStr]
		if !ok {
			rc.logger.Error().Msgf("SHOULD RARELY HAPPEN: requestContext.createResponseObservations() should never fail to find the JSONRPC request for response ID: %s", responseIDStr)
			continue
		}

		// Create observations for both the request and its corresponding endpoint response
		endpointObs := endpointResp.GetObservation()

		observations = append(observations, &qosobservations.EVMRequestObservation{
			JsonrpcRequest: jsonrpcReq.GetObservation(),
			EndpointObservations: []*qosobservations.EVMEndpointObservation{
				&endpointObs,
			},
		})
	}

	return observations
}

// convertValidationResults converts endpoint selection validation results to proto format.
// These results contain information about which endpoints were considered during selection
// and why they were accepted or rejected.
func (rc requestContext) convertValidationResults() []*qosobservations.EndpointValidationResult {
	var validationResults []*qosobservations.EndpointValidationResult
	validationResults = append(validationResults, rc.endpointSelectionMetadata.ValidationResults...)
	return validationResults
}

func (rc *requestContext) GetEndpointSelector() protocol.EndpointSelector {
	return rc
}

// Select returns endpoint address using request context's endpoint store.
// Implements protocol.EndpointSelector interface.
// Tracks random selection when all endpoints fail validation.
func (rc *requestContext) Select(allEndpoints protocol.EndpointAddrList) (protocol.EndpointAddr, error) {
	// TODO_FUTURE(@adshmh): Enhance the endpoint selection meta data to track, e.g.:
	// * Endpoint Selection Latency
	// * Number of available endpoints
	selectionResult, err := rc.serviceState.SelectWithMetadata(allEndpoints)
	if err != nil {
		return protocol.EndpointAddr(""), err
	}

	// Store selection metadata for observation tracking
	rc.endpointSelectionMetadata = selectionResult.Metadata

	return selectionResult.SelectedEndpoint, nil
}

// SelectMultiple returns multiple endpoint addresses using the request context's endpoint store.
// Implements the protocol.EndpointSelector interface.
func (rc *requestContext) SelectMultiple(allEndpoints protocol.EndpointAddrList, numEndpoints uint) (protocol.EndpointAddrList, error) {
	return rc.serviceState.SelectMultiple(allEndpoints, numEndpoints)
}
