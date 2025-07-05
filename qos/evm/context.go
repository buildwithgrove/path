package evm

import (
	"encoding/json"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_MVP(@adshmh): Support individual configuration of timeout for every service that uses EVM QoS.
// The default timeout when sending a request to an EVM blockchain endpoint.
const defaultServiceRequestTimeoutMillisec = 10_500

// requestContext provides the support required by the gateway
// package for handling service requests.
var _ gateway.RequestQoSContext = &requestContext{}

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

	// TODO_TECHDEBT(@adshmh): support batch JSONRPC requests
	jsonrpcReq jsonrpc.Request

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
	reqBz, err := json.Marshal(rc.jsonrpcReq)
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

		// Path field is not used for EVM-based blockchains.

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

	response, err := unmarshalResponse(rc.logger, rc.jsonrpcReq, responseBz)

	rc.endpointResponses = append(rc.endpointResponses,
		endpointResponse{
			EndpointAddr: endpointAddr,
			response:     response,
			unmarshalErr: err,
		},
	)
}

// TODO_TECHDEBT: support batch JSONRPC requests by breaking them into
// single JSONRPC requests and tracking endpoints' response(s) to each.
// This would also require combining the responses into a single, valid
// response to the batch JSONRPC request.
// See the following link for more details:
// https://www.jsonrpc.org/specification#batch
//
// GetHTTPResponse builds the HTTP response that should be returned for
// an EVM blockchain service request.
// Implements the gateway.RequestQoSContext interface.
func (rc requestContext) GetHTTPResponse() gateway.HTTPResponse {
	// Use a noResponses struct if no responses were reported by the protocol from any endpoints.
	if len(rc.endpointResponses) == 0 {
		responseNoneObj := responseNone{
			logger:     rc.logger,
			jsonrpcReq: rc.jsonrpcReq,
		}

		return responseNoneObj.GetHTTPResponse()
	}

	// return the last endpoint response reported to the context.
	return rc.endpointResponses[len(rc.endpointResponses)-1].GetHTTPResponse()
}

// GetObservations returns all endpoint observations from the request context.
// Implements gateway.RequestQoSContext interface.
func (rc requestContext) GetObservations() qosobservations.Observations {
	var observations []*qosobservations.EVMEndpointObservation

	// If no (zero) responses were received, create a single observation for the no-response scenario.
	if len(rc.endpointResponses) == 0 {
		responseNoneObj := responseNone{
			logger:     rc.logger,
			jsonrpcReq: rc.jsonrpcReq,
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

	// Return the set of observations for the single JSONRPC request.
	return qosobservations.Observations{
		ServiceObservations: &qosobservations.Observations_Evm{
			Evm: &qosobservations.EVMRequestObservations{
				JsonrpcRequest:       rc.jsonrpcReq.GetObservation(),
				ChainId:              rc.chainID,
				ServiceId:            string(rc.serviceID),
				RequestPayloadLength: uint32(rc.requestPayloadLength),
				RequestOrigin:        rc.requestOrigin,
				EndpointObservations: observations,
			},
		},
	}
}

func (rc *requestContext) GetEndpointSelector() protocol.EndpointSelector {
	return rc
}

// Select returns the address of an endpoint using the request context's endpoint store.
// Implements the protocol.EndpointSelector interface.
func (rc *requestContext) Select(allEndpoints protocol.EndpointAddrList) (protocol.EndpointAddr, error) {
	return rc.serviceState.Select(allEndpoints)
}

// SelectMultiple returns multiple endpoint addresses using the request context's endpoint store.
// Implements the protocol.EndpointSelector interface.
func (rc *requestContext) SelectMultiple(allEndpoints protocol.EndpointAddrList, maxCount int) (protocol.EndpointAddrList, error) {
	return rc.serviceState.SelectMultiple(allEndpoints, maxCount)
}
