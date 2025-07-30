package evm

import (
	"encoding/json"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// defaultServiceRequestTimeoutMillisec is the default timeout when sending a request to an EVM blockchain endpoint.
// TODO_IMPROVE(@adshmh): Support method level specific timeouts and allow the user to configure them.
const defaultServiceRequestTimeoutMillisec = 10_500

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
	jsonrpcReqs []jsonrpc.Request

	// endpointResponses is the set of responses received from one or
	// more endpoints as part of handling this service request.
	// Supports both single and batch JSON-RPC requests.
	endpointResponses []endpointResponse

	// endpointSelectionMetadata contains metadata about the endpoint selection process
	endpointSelectionMetadata EndpointSelectionMetadata
}

// GetServicePayload implements the gateway.RequestQoSContext interface.
// TODO_MVP(@adshmh): Ensure the JSONRPC request struct can handle all valid service requests.
func (rc requestContext) GetServicePayloads() []protocol.Payload {
	payloads := make([]protocol.Payload, len(rc.jsonrpcReqs))

	for _, req := range rc.jsonrpcReqs {
		reqBz, err := json.Marshal(req)
		if err != nil {
			rc.logger.Error().Err(err).Msg("SHOULD RARELY HAPPEN: requestContext.GetServicePayload() should never fail marshaling the JSONRPC request.")
			return []protocol.Payload{protocol.EmptyErrorPayload()}
		}

		payloads = append(payloads, protocol.Payload{
			Data:            string(reqBz),
			Method:          http.MethodPost, // Method is alway POST for EVM-based blockchains.
			Path:            "",              // Path field is not used for EVM-based blockchains.
			Headers:         map[string]string{},
			TimeoutMillisec: defaultServiceRequestTimeoutMillisec,
			RPCType:         sharedtypes.RPCType_JSON_RPC,
		})
	}

	return payloads
}

// UpdateWithResponse is NOT safe for concurrent use
func (rc *requestContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, responseBz []byte) {
	// TODO_IMPROVE: check whether the request was valid, and return an error if it was not.
	// This would be an extra safety measure, as the caller should have checked the returned value
	// indicating the validity of the request when calling on QoS instance's ParseHTTPRequest

	// For batch requests, use the first request for response unmarshalling
	// TODO_FUTURE: Enhance to properly handle batch response processing
	var jsonrpcReq jsonrpc.Request
	if len(rc.jsonrpcReqs) > 0 {
		jsonrpcReq = rc.jsonrpcReqs[0]
	}

	response, err := unmarshalResponse(rc.logger, jsonrpcReq, responseBz, endpointAddr)

	rc.endpointResponses = append(rc.endpointResponses,
		endpointResponse{
			EndpointAddr: endpointAddr,
			response:     response,
			unmarshalErr: err,
		},
	)
}

// GetHTTPResponse builds the HTTP response that should be returned for
// an EVM blockchain service request.
// Implements the gateway.RequestQoSContext interface.
func (rc requestContext) GetHTTPResponse() gateway.HTTPResponse {
	// Use a noResponses struct if no responses were reported by the protocol from any endpoints.
	if len(rc.endpointResponses) == 0 {
		// For batch requests, use the first request for response processing
		var jsonrpcReq jsonrpc.Request
		if len(rc.jsonrpcReqs) > 0 {
			jsonrpcReq = rc.jsonrpcReqs[0]
		}

		responseNoneObj := responseNone{
			logger:     rc.logger,
			jsonrpcReq: jsonrpcReq,
		}

		return responseNoneObj.GetHTTPResponse()
	}

	// Handle batch requests according to JSON-RPC 2.0 specification
	// https://www.jsonrpc.org/specification#batch
	if len(rc.jsonrpcReqs) > 1 {
		return rc.getBatchHTTPResponse()
	}

	// return the last endpoint response reported to the context for single requests.
	return rc.endpointResponses[len(rc.endpointResponses)-1].GetHTTPResponse()
}

// getBatchHTTPResponse handles batch requests by combining individual JSON-RPC responses
// into an array according to the JSON-RPC 2.0 specification.
// https://www.jsonrpc.org/specification#batch
func (rc requestContext) getBatchHTTPResponse() httpResponse {
	// Collect individual response payloads
	var individualResponses []json.RawMessage
	var httpStatusCode int = http.StatusOK // Default to 200 OK for batch responses

	// Process each endpoint response
	for _, endpointResp := range rc.endpointResponses {
		individualHTTPResp := endpointResp.GetHTTPResponse()

		// Extract the JSON payload from each response
		payload := individualHTTPResp.GetPayload()
		if len(payload) > 0 {
			individualResponses = append(individualResponses, json.RawMessage(payload))
		}

		// If any individual response has an error status, use internal server error for the batch
		if individualHTTPResp.GetHTTPStatusCode() >= 400 {
			httpStatusCode = http.StatusInternalServerError
		}
	}

	// According to JSON-RPC spec: "If there are no Response objects contained within the Response array
	// as it is to be sent to the client, the server MUST NOT return an empty Array and should return nothing at all."
	// However, since we're in the context of having endpointResponses, we should return the array.

	// Combine individual responses into a JSON array
	batchResponse, err := json.Marshal(individualResponses)
	if err != nil {
		rc.logger.Error().Err(err).Msg("Failed to marshal batch response")
		// Return an internal server error response
		return httpResponse{
			responsePayload: []byte(`{"jsonrpc":"2.0","error":{"code":-32603,"message":"Internal error"},"id":null}`),
			httpStatusCode:  http.StatusInternalServerError,
		}
	}

	return httpResponse{
		responsePayload: batchResponse,
		httpStatusCode:  httpStatusCode,
	}
}

// GetObservations returns all endpoint observations from the request context.
// Implements gateway.RequestQoSContext interface.
func (rc requestContext) GetObservations() qosobservations.Observations {
	var observations []*qosobservations.EVMEndpointObservation

	// If no (zero) responses were received, create a single observation for the no-response scenario.
	if len(rc.endpointResponses) == 0 {
		// For batch requests, use the first request for response processing
		var jsonrpcReq jsonrpc.Request
		if len(rc.jsonrpcReqs) > 0 {
			jsonrpcReq = rc.jsonrpcReqs[0]
		}

		responseNoneObj := responseNone{
			logger:     rc.logger,
			jsonrpcReq: jsonrpcReq,
		}
		responseNoneObs := responseNoneObj.GetObservation()
		observations = append(observations, &responseNoneObs)
	} else {
		// Otherwise, process all responses as individual observations.
		observations = make([]*qosobservations.EVMEndpointObservation, len(rc.endpointResponses))
		for idx, endpointResponse := range rc.endpointResponses {
			obs := endpointResponse.GetObservation()
			obs.EndpointAddr = string(endpointResponse.EndpointAddr)
			observations[idx] = &obs
		}
	}

	// Convert validation results to proto format
	var validationResults []*qosobservations.EndpointValidationResult
	validationResults = append(validationResults, rc.endpointSelectionMetadata.ValidationResults...)

	// Return the set of observations for the JSONRPC request(s).
	// For batch requests, use the first request for observation generation
	var jsonrpcRequestObs *qosobservations.JsonRpcRequest
	if len(rc.jsonrpcReqs) > 0 {
		jsonrpcRequestObs = rc.jsonrpcReqs[0].GetObservation()
	}

	return qosobservations.Observations{
		ServiceObservations: &qosobservations.Observations_Evm{
			Evm: &qosobservations.EVMRequestObservations{
				JsonrpcRequest:       jsonrpcRequestObs,
				ChainId:              rc.chainID,
				ServiceId:            string(rc.serviceID),
				RequestPayloadLength: uint32(rc.requestPayloadLength),
				RequestOrigin:        rc.requestOrigin,
				EndpointObservations: observations,
				EndpointSelectionMetadata: &qosobservations.EndpointSelectionMetadata{
					RandomEndpointFallback: rc.endpointSelectionMetadata.RandomEndpointFallback,
					ValidationResults:      validationResults,
				},
			},
		},
	}
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
