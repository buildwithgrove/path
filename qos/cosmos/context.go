package cosmos

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/pokt-network/poktroll/pkg/relayer/proxy"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_MVP(@adshmh): Support individual configuration of timeout for every service that uses CosmosSDK QoS.
// The default timeout when sending a request to a CosmosSDK blockchain endpoint.
const defaultServiceRequestTimeoutMillisec = 10000

// requestContext provides the support required by the gateway
// package for handling service requests.
var _ gateway.RequestQoSContext = &requestContext{}

// TODO_REFACTOR: Improve naming clarity by distinguishing between interfaces and adapters
// in the metrics/qos/cosmos and qos/cosmos packages, and elsewhere names like `response` are used.
// Consider renaming:
//   - metrics/qos/cosmos: response → CosmosSDKMetricsResponse
//   - qos/cosmos: response → CosmosSDKQoSResponse
//   - observation/cosmos: observation -> CosmosSDKObservation
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
	GetObservation() qosobservations.CosmosSDKEndpointObservation

	// GetHTTPResponse returns the HTTP response to be sent back to the client.
	GetHTTPResponse() httpResponse
}

var _ response = &endpointResponse{}

type endpointResponse struct {
	protocol.EndpointAddr
	response
	unmarshalErr error
}

// requestContext implements the functionality for CosmosSDK-based blockchain services.
type requestContext struct {
	logger polylog.Logger

	// httpReq is the original HTTP request from the user
	httpReq *http.Request

	// chainID is the chain identifier for CosmosSDK QoS implementation.
	chainID string

	// service_id is the identifier for the CosmosSDK QoS implementation.
	// It is the "alias" or human readable interpretation of the chain_id.
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

	// For JSON-RPC POST requests (when applicable)
	jsonrpcReq jsonrpc.Request

	// endpointResponses is the set of responses received from one or
	// more endpoints as part of handling this service request.
	endpointResponses []endpointResponse
}

// GetServicePayload returns the payload for the service request.
// It accounts for both REST-like and JSON-RPC requests.
// Implements gateway.RequestQoSContext interface.
func (rc requestContext) GetServicePayload() protocol.Payload {
	payload := protocol.Payload{
		Method:          rc.httpReq.Method,
		TimeoutMillisec: defaultServiceRequestTimeoutMillisec,
		Headers:         getCosmosSDKHeaders(rc.httpReq.URL.Path),
	}

	// If the request is REST-like set the path including query parameters.
	if rc.httpReq.URL.Path != "" {
		payload.Path = rc.httpReq.URL.Path

		if rc.httpReq.URL.RawQuery != "" {
			payload.Path += "?" + rc.httpReq.URL.RawQuery
		}
	}

	// Determine if request is a JSON-RPC request by checking if:
	//  - The request method is POST
	//  - The JSON-RPC request is not empty.
	if rc.isJsonRpcRequest() {
		reqBz, err := json.Marshal(rc.jsonrpcReq)
		if err == nil {
			payload.Data = string(reqBz)
		}
	}

	return payload
}

// TODO_IN_THIS_PR(@commoddity): productionize this determination of how to set RPC-Type header.
// eg. save strings as consts, etc.
func getCosmosSDKHeaders(urlPath string) map[string]string {
	// If the URL path starts with /cosmos/, set the "RPC-Type" header to "rest"
	// All Cosmos SDK endpoints start with /cosmos/
	// Ref: https://docs.cosmos.network/api
	if strings.HasPrefix(urlPath, "/cosmos/") {
		return map[string]string{
			// eg. "Rpc-Type: 4" -> "REST"
			proxy.RPCTypeHeader: strconv.Itoa(int(sharedtypes.RPCType_REST)),
		}
	}

	return map[string]string{}
}

// isEmptyJSONRPCRequest checks if the JSON-RPC request is empty/uninitialized.
func (rc requestContext) isJsonRpcRequest() bool {
	return rc.httpReq.Method == http.MethodPost && !rc.jsonrpcReq.ID.IsEmpty()
}

// UpdateWithResponse is NOT safe for concurrent use
func (rc *requestContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, responseBz []byte) {
	// TODO_IMPROVE: check whether the request was valid, and return an error if it was not.
	// This would be an extra safety measure, as the caller should have checked the returned value
	// indicating the validity of the request when calling on QoS instance's ParseHTTPRequest

	response, err := unmarshalResponse(rc.logger, rc.httpReq.URL.Path, responseBz)

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
// a CosmosSDK blockchain service request.
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
	var observations []*qosobservations.CosmosSDKEndpointObservation

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
		observations = make([]*qosobservations.CosmosSDKEndpointObservation, len(rc.endpointResponses))
		for idx, endpointResponse := range rc.endpointResponses {
			obs := endpointResponse.GetObservation()
			obs.EndpointAddr = string(endpointResponse.EndpointAddr)
			observations[idx] = &obs
		}
	}

	// Return the set of observations for the single request.
	var routeRequest string
	if rc.httpReq.URL.Path != "" {
		// For REST-like requests, use the path
		routeRequest = rc.httpReq.URL.Path
		if rc.httpReq.URL.RawQuery != "" {
			routeRequest += "?" + rc.httpReq.URL.RawQuery
		}
	} else if rc.isJsonRpcRequest() {
		// For JSON-RPC requests, serialize the request
		routeRequestBytes, _ := json.Marshal(rc.jsonrpcReq)
		routeRequest = string(routeRequestBytes)
	}

	return qosobservations.Observations{
		ServiceObservations: &qosobservations.Observations_Cosmos{
			Cosmos: &qosobservations.CosmosSDKRequestObservations{
				RouteRequest:         routeRequest,
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
