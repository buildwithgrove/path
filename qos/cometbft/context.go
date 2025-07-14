package cometbft

import (
	"errors"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const defaultServiceRequestTimeoutMillisec = 10_000

// requestContext provides the support required by the gateway
// package for handling service requests.
var _ gateway.RequestQoSContext = &requestContext{}

// response is an interface that represents the response received from an endpoint.
type response interface {
	GetObservation() qosobservations.CometBFTEndpointObservation

	// TODO_TECHDEBT(@adshmh): Verify that a JSONRPC response covers all supported uses of CometBFT services.
	// Reference:
	// https://docs.cometbft.com/v1.0/rpc
	//
	// GetJSONRPCResponse Returns the JSONRPC response to be sent back to the client.
	GetJSONRPCResponse() jsonrpc.Response
}

// endpointResponse stores the response received from an endpoint.
type endpointResponse struct {
	protocol.EndpointAddr
	response
	unmarshalErr error
}

// requestContext implements QoS functionality for CometBFT blockchain services.
type requestContext struct {
	logger        polylog.Logger
	endpointStore *EndpointStore

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

	// httpReq is the original HTTP request from the user
	httpReq *http.Request

	// CometBFT supports both REST and JSON-RPC formats.
	// For JSON-RPC POST requests, jsonrpcRequestBz stores the serialized request body.
	// See: https://docs.cometbft.com/v1.0/spec/rpc/
	jsonrpcRequestBz []byte

	// isValid indicates if the user request was valid when parsed.
	// Set by QoS instance during request context creation.
	isValid bool

	// endpointResponses contains responses from endpoints handling this service request
	// NOTE: Currently only supports responses associated with a single JSON-RPC request.
	// TODO_FUTURE: Batch support will require modifying the field type.
	endpointResponses []endpointResponse

	// TODO_TECHDEBT(@adshmh): Add endpoint selection metadata, consistent with the `evm` package.
}

// GetServicePayload returns the payload for the service request.
// It accounts for both REST-like and JSON-RPC requests.
// Implements gateway.RequestQoSContext interface.
func (rc requestContext) GetServicePayload() protocol.Payload {
	payload := protocol.Payload{
		Method:          rc.httpReq.Method,
		TimeoutMillisec: defaultServiceRequestTimeoutMillisec,
	}

	// If the request is REST-like, set the path including query parameters.
	if rc.httpReq.URL.Path != "" {
		payload.Path = rc.httpReq.URL.Path

		if rc.httpReq.URL.RawQuery != "" {
			payload.Path += "?" + rc.httpReq.URL.RawQuery
		}
	}

	// If the request is JSON-RPC, set the data from the stored []byte.
	if rc.isJSONRPCRequest() {
		payload.Data = string(rc.jsonrpcRequestBz)
	}

	return payload
}

// UpdateWithResponse stores (appends) the response from an endpoint in the request context.
// CRITICAL: NOT safe for concurrent use.
// Implements gateway.RequestQoSContext interface.
func (rc *requestContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, responseBz []byte) {
	// TODO_IMPROVE: check whether the request was valid, and return an error if it was not.
	// This would be an extra safety measure, as the caller should have checked the returned value
	// indicating the validity of the request when calling on QoS instance's ParseHTTPRequest
	response, err := unmarshalResponse(rc.logger, rc.httpReq.URL.Path, responseBz, rc.isJSONRPCRequest(), endpointAddr)

	// Multiple responses can be associated with a single request for multiple reasons, such as:
	// - Retries from single/multiple endpoints
	// - Collecting a quorum of from different endpoints
	// - Organic vs synthetic responses
	rc.endpointResponses = append(rc.endpointResponses,
		endpointResponse{
			EndpointAddr: endpointAddr,
			response:     response,
			unmarshalErr: err,
		},
	)
}

// GetHTTPResponse builds the HTTP response for a CometBFT blockchain service request.
// Returns the last endpoint response if available, otherwise returns generic response.
// Implements gateway.RequestQoSContext interface.
func (rc requestContext) GetHTTPResponse() gateway.HTTPResponse {
	// No responses received: this is an internal error:
	// e.g. protocol-level errors like endpoint timing out.
	if len(rc.endpointResponses) == 0 {
		// TODO_TECHDEBT(@adshmh): Use request's ID once a request validator is implemented for CometBFT services.
		// Build the JSONRPC response indicating a protocol-level error.
		jsonrpcErrorResponse := jsonrpc.NewErrResponseInternalErr(jsonrpc.ID{}, errors.New("protocol-level error: no endpoint responses received"))
		return qos.BuildHTTPResponseFromJSONRPCResponse(rc.logger, jsonrpcErrorResponse)
	}

	// Use the most recent endpoint response.
	// As of PR #253 there is no retry, meaning there is at most 1 endpoint response.
	selectedResponse := rc.endpointResponses[len(rc.endpointResponses)-1].GetJSONRPCResponse()

	// CometBFT response codes:
	// returns an HTTP status code corresponding to the underlying JSON-RPC response code.
	// DEV_NOTE: This is an opinionated mapping following best practice but not enforced by any specifications or standards.
	return qos.BuildHTTPResponseFromJSONRPCResponse(rc.logger, selectedResponse)
}

// GetObservations returns all endpoint observations from the request context.
// Implements gateway.RequestQoSContext interface.
func (rc requestContext) GetObservations() qosobservations.Observations {
	// Set the observation fields common for all requests: successful or failed.
	observations := &qosobservations.CometBFTRequestObservations{
		ChainId:       rc.chainID,
		ServiceId:     string(rc.serviceID),
		RequestOrigin: rc.requestOrigin,
	}

	// No endpoint responses received.
	// Set request error.
	if len(rc.endpointResponses) == 0 {
		observations.RequestError = qos.GetRequestErrorForProtocolError()

		return qosobservations.Observations{
			ServiceObservations: &qosobservations.Observations_Cometbft{
				Cometbft: observations,
			},
		}
	}

	// Build the endpoint(s) observations.
	endpointObservations := make([]*qosobservations.CometBFTEndpointObservation, len(rc.endpointResponses))
	for idx, endpointResponse := range rc.endpointResponses {
		obs := endpointResponse.GetObservation()
		obs.EndpointAddr = string(endpointResponse.EndpointAddr)
		endpointObservations[idx] = &obs
	}

	// Set the endpoint observations fields.
	observations.EndpointObservations = endpointObservations

	return qosobservations.Observations{
		ServiceObservations: &qosobservations.Observations_Cometbft{
			// TODO_TECHDEBT(@adshmh): Set JSON-RPCRequest field.
			// Requires utility function to convert between:
			// - qos.jsonrpc.Request
			// - observation.qos.JsonRpcRequest
			// Needed for setting JSON-RPC fields in any QoS service's observations.
			Cometbft: observations,
		},
	}
}

// GetEndpointSelector returns the endpoint selector for the request context.
// Implements the gateway.RequestQoSContext interface.
func (rc *requestContext) GetEndpointSelector() protocol.EndpointSelector {
	return rc
}

// Select returns the address of an endpoint using the request context's endpoint store.
// Implements the protocol.EndpointSelector interface.
func (rc *requestContext) Select(allEndpoints protocol.EndpointAddrList) (protocol.EndpointAddr, error) {
	// Select an endpoint from the available endpoints using the endpoint store.
	return rc.endpointStore.Select(allEndpoints)
}

// isJSONRPCRequest checks if the request context contains a serialized JSON-RPC request.
// Reference: https://docs.cometbft.com/v1.0/spec/rpc/
func (rc *requestContext) isJSONRPCRequest() bool {
	return len(rc.jsonrpcRequestBz) > 0
}
