package solana

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const (
	// The default timeout when sending a request to
	// a Solana blockchain endpoint.
	defaultServiceRequestTimeoutMillisec = 15_000
)

// requestContext provides the support required by the gateway
// package for handling service requests.
var _ gateway.RequestQoSContext = &requestContext{}

// TODO_TECHDEBT: Need a Validate() method here to allow
// the caller, e.g. gateway, determine whether the endpoint's
// response was valid, and whether a retry makes sense.
//
// response defines the functionality required from a parsed endpoint response.
type response interface {
	GetObservation() qosobservations.SolanaEndpointObservation
	GetJSONRPCResponse() jsonrpc.Response

	// TODO_TECHDEBT: add method(s) to support retrying a request, e.g. IsUserError(), IsEndpointError().
}

type endpointResponse struct {
	protocol.EndpointAddr
	response
}

// requestContext provides the functionality required
// to support QoS for a Solana blockchain service.
type requestContext struct {
	logger polylog.Logger

	// chainID is the chain identifier for the Solana QoS implementation.
	chainID string

	// service_id is the identifier for the Solana QoS implementation.
	// It is the "alias" or human readable interpretation of the chain_id.
	// Used in generating observations.
	serviceID protocol.ServiceID

	// The length of the request payload in bytes.
	requestPayloadLength uint

	endpointStore *EndpointStore

	JSONRPCReq jsonrpc.Request

	// The origin of the request handled by the context.
	// Either:
	// - User: user requests
	// - QoS: requests built by the QoS service to get additional data points on endpoints.
	requestOrigin qosobservations.RequestOrigin

	// endpointResponses is the set of responses received from one or
	// more endpoints as part of handling this service request.
	// NOTE: these are all related to a single JSONRPC request,
	// enhancing to support batch JSONRPC requests will involve the
	// modification of this field's type.
	endpointResponses []endpointResponse
}

// TODO_MVP(@adshmh): Ensure the JSONRPC request struct
// can handle all valid service requests.
func (rc requestContext) GetServicePayload() protocol.Payload {
	reqBz, err := json.Marshal(rc.JSONRPCReq)
	if err != nil {
		// TODO_MVP(@adshmh): find a way to guarantee this never happens,
		// e.g. by storing the serialized form of the JSONRPC request
		// at the time of creating the request context.
		return protocol.Payload{}
	}

	return protocol.Payload{
		Data: string(reqBz),
		// Method is alway POST for Solana.
		Method: http.MethodPost,

		// Path field is not used for Solana.

		// TODO_IMPROVE: adjust the timeout based on the request method:
		// An endpoint may need more time to process certain requests,
		// as indicated by the request's method and/or parameters.
		TimeoutMillisec: defaultServiceRequestTimeoutMillisec,
	}
}

// UpdateWithResponse is NOT safe for concurrent use
func (rc *requestContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, responseBz []byte) {
	// TODO_IMPROVE: check whether the request was valid, and return an error if it was not.
	// This would be an extra safety measure, as the caller should have checked the returned value
	// indicating the validity of the request when calling on QoS instance's ParseHTTPRequest

	response := unmarshalResponse(rc.logger, rc.JSONRPCReq, responseBz)

	// TODO_MVP(@adshmh): Drop the unmarshaling error: the returned response interface should provide methods to allow the caller to:
	// 1. Check if the response from the endpoint was valid or malformed. This is needed to support retrying with a different endpoint if
	// the originally selected one fails to return a valid response to the user's request.
	// 2. Return a generic but valid JSONRPC response to the user.
	rc.endpointResponses = append(rc.endpointResponses,
		endpointResponse{
			EndpointAddr: endpointAddr,
			response:     response,
		},
	)
}

// TODO_MVP(@adshmh): add `Content-Type: application/json` header.
// GetHTTPResponse builds the HTTP response that should be returned for
// a Solana blockchain service request.
func (rc requestContext) GetHTTPResponse() gateway.HTTPResponse {
	// No responses received: this is an internal error:
	// e.g. protocol-level errors like endpoint timing out.
	if len(rc.endpointResponses) == 0 {
		// Build the JSONRPC response indicating a protocol-level error.
		jsonrpcErrorResponse := jsonrpc.NewErrResponseInternalErr(rc.JSONRPCReq.ID, errors.New("protocol-level error: no endpoint responses received"))
		return qos.BuildHTTPResponseFromJSONRPCResponse(rc.logger, jsonrpcErrorResponse)
	}

	// Use the most recent endpoint response.
	// As of PR #253 there is no retry, meaning there is at most 1 endpoint response.
	selectedResponse := rc.endpointResponses[len(rc.endpointResponses)-1].GetJSONRPCResponse()
	return qos.BuildHTTPResponseFromJSONRPCResponse(rc.logger, selectedResponse)
}

// GetObservations returns all the observations contained in the request context.
// Implements the gateway.RequestQoSContext interface.
func (rc requestContext) GetObservations() qosobservations.Observations {
	// Set the observation fields common for all requests: successful or failed.
	observations := &qosobservations.SolanaRequestObservations{
		ChainId:              rc.chainID,
		ServiceId:            string(rc.serviceID),
		RequestPayloadLength: uint32(rc.requestPayloadLength),
		RequestOrigin:        rc.requestOrigin,
		JsonrpcRequest:       rc.JSONRPCReq.GetObservation(),
	}

	// No endpoint responses received.
	// Set request error.
	if len(rc.endpointResponses) == 0 {
		observations.RequestError = qos.GetRequestErrorForProtocolError()

		return qosobservations.Observations{
			ServiceObservations: &qosobservations.Observations_Solana{
				Solana: observations,
			},
		}
	}

	// Build the endpoint(s) observations.
	endpointObservations := make([]*qosobservations.SolanaEndpointObservation, len(rc.endpointResponses))
	for idx, endpointResponse := range rc.endpointResponses {
		obs := endpointResponse.GetObservation()
		obs.EndpointAddr = string(endpointResponse.EndpointAddr)
		endpointObservations[idx] = &obs
	}

	// Set the endpoint observations fields.
	observations.EndpointObservations = endpointObservations

	return qosobservations.Observations{
		ServiceObservations: &qosobservations.Observations_Solana{
			Solana: observations,
		},
	}
}

// GetEndpointSelector is required to satisfy the gateway package's ResquestQoSContext interface.
// The request context is queried for the correct endpoint selector to use because this allows different
// endpoint selectors based on the request's context.
// e.g. the request context for a particular request method can potentially rank endpoints based on their latency when responding to requests with matching method.
func (rc *requestContext) GetEndpointSelector() protocol.EndpointSelector {
	return rc
}

// Select chooses an endpoint from the list of supplied endpoints, using the perceived (using endpoints' responses) state of the Solana chain.
// It is required to satisfy the protocol package's EndpointSelector interface.
func (rc *requestContext) Select(allEndpoints protocol.EndpointAddrList) (protocol.EndpointAddr, error) {
	return rc.endpointStore.Select(allEndpoints)
}
