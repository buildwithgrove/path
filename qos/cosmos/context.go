package cosmos

import (
	"encoding/json"
	"net/http"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/pokt-network/poktroll/pkg/polylog"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
)

// Default timeout when sending a request to a CosmosSDK blockchain endpoint.
const defaultServiceRequestTimeoutMillisec = 10_000

// requestContext implements the functionality for CosmosSDK-based blockchain services
// with unified handling of different RPC types (REST, JSON-RPC, COMET_BFT)
type requestContext struct {
	logger polylog.Logger

	// httpReq is the original HTTP request from the user
	httpReq http.Request

	// chainID is the chain identifier for CosmosSDK QoS implementation.
	chainID string

	// serviceID is the identifier for the CosmosSDK QoS implementation.
	serviceID protocol.ServiceID

	// rpcType indicates the detected RPC type for this request
	rpcType sharedtypes.RPCType

	// The origin of the request handled by the context.
	requestOrigin qosobservations.RequestOrigin

	// The length of the request payload in bytes.
	requestPayloadLength uint

	// RPC type-specific fields (only relevant fields will be populated based on rpcType)

	// For JSON-RPC requests (both EVM JSON-RPC and CometBFT JSON-RPC)
	jsonrpcReq *jsonrpc.Request

	// For REST requests (both CosmosSDK REST and CometBFT REST-style)
	restBody []byte

	// endpointResponses is the set of responses received from endpoints
	endpointResponses []endpointResponse
}

// endpointResponse stores the response received from an endpoint.
type endpointResponse struct {
	protocol.EndpointAddr
	response
	unmarshalErr error
}

// GetServicePayload returns the payload for the service request.
// Uses the RPC type to determine appropriate payload construction.
func (rc *requestContext) GetServicePayload() protocol.Payload {
	switch rc.rpcType {
	case sharedtypes.RPCType_REST, sharedtypes.RPCType_COMET_BFT:
		return rc.buildRESTPayload()
	case sharedtypes.RPCType_JSON_RPC:
		return rc.buildJSONRPCPayload()
	default:
		rc.logger.Warn().Str("rpc_type", rc.rpcType.String()).Msg("Unknown RPC type, using default payload")
		return rc.buildDefaultPayload()
	}
}

// buildRESTPayload constructs a payload for REST-style requests
func (rc *requestContext) buildRESTPayload() protocol.Payload {
	payload := protocol.Payload{
		Method:          rc.httpReq.Method,
		Path:            rc.httpReq.URL.Path,
		TimeoutMillisec: defaultServiceRequestTimeoutMillisec,
		Headers:         make(map[string]string),
		RPCType:         rc.rpcType,
	}

	// Add query parameters to path if present
	if rc.httpReq.URL.RawQuery != "" {
		payload.Path += "?" + rc.httpReq.URL.RawQuery
	}

	// Add body data if present (for POST/PUT REST requests)
	if len(rc.restBody) > 0 {
		payload.Data = string(rc.restBody)
	}

	return payload
}

// buildJSONRPCPayload constructs a payload for JSON-RPC requests
func (rc *requestContext) buildJSONRPCPayload() protocol.Payload {
	if rc.jsonrpcReq == nil {
		rc.logger.Error().Msg("JSONRPC request context missing jsonrpcReq")
		return protocol.Payload{
			Method:          "POST",
			TimeoutMillisec: defaultServiceRequestTimeoutMillisec,
			Headers:         make(map[string]string),
			Data:            "{}",
			RPCType:         rc.rpcType,
		}
	}

	reqBz, err := json.Marshal(rc.jsonrpcReq)
	if err != nil {
		rc.logger.Error().Err(err).Msg("failed to marshal JSON-RPC request")
		reqBz = []byte("{}")
	}

	return protocol.Payload{
		Method:          "POST",
		TimeoutMillisec: defaultServiceRequestTimeoutMillisec,
		Headers:         make(map[string]string),
		Data:            string(reqBz),
		RPCType:         rc.rpcType,
	}
}

// buildDefaultPayload constructs a default payload when RPC type is unknown
func (rc *requestContext) buildDefaultPayload() protocol.Payload {
	payload := protocol.Payload{
		Method:          rc.httpReq.Method,
		Path:            rc.httpReq.URL.Path,
		TimeoutMillisec: defaultServiceRequestTimeoutMillisec,
		Headers:         make(map[string]string),
		RPCType:         sharedtypes.RPCType_UNSPECIFIED,
	}

	if rc.httpReq.URL.RawQuery != "" {
		payload.Path += "?" + rc.httpReq.URL.RawQuery
	}

	// Try to include data from either source
	if rc.jsonrpcReq != nil {
		if reqBz, err := json.Marshal(rc.jsonrpcReq); err == nil {
			payload.Data = string(reqBz)
		}
	} else if len(rc.restBody) > 0 {
		payload.Data = string(rc.restBody)
	}

	return payload
}

// UpdateWithResponse stores the response from an endpoint in the request context.
func (rc *requestContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, responseBz []byte) {
	response, err := unmarshalResponse(
		rc.logger,
		rc.httpReq.URL.Path,
		responseBz,
		rc.isJSONRPCRequest(),
		endpointAddr,
	)

	rc.endpointResponses = append(rc.endpointResponses,
		endpointResponse{
			EndpointAddr: endpointAddr,
			response:     response,
			unmarshalErr: err,
		},
	)
}

// isJSONRPCRequest returns true if this is a JSON-RPC style request
func (rc *requestContext) isJSONRPCRequest() bool {
	return rc.jsonrpcReq != nil
}

// GetHTTPResponse builds the HTTP response for the service request.
func (rc *requestContext) GetHTTPResponse() gateway.HTTPResponse {
	// Use a noResponses struct if no responses were reported by the protocol from any endpoints.
	if len(rc.endpointResponses) == 0 {
		return rc.createNoResponseError()
	}

	// Return the last endpoint response reported to the context.
	lastResponse := rc.endpointResponses[len(rc.endpointResponses)-1]
	return lastResponse.GetHTTPResponse()
}

// createNoResponseError creates an appropriate error response when no endpoint responses are available
func (rc *requestContext) createNoResponseError() gateway.HTTPResponse {
	responseNoneObj := responseNone{
		logger:     rc.logger,
		httpReq:    rc.httpReq,
		jsonrpcReq: rc.jsonrpcReq,
		rpcType:    rc.rpcType,
	}

	return responseNoneObj.GetHTTPResponse()
}

// GetObservations returns all endpoint observations from the request context.
func (rc *requestContext) GetObservations() qosobservations.Observations {
	// Set the observation fields common for all requests: successful or failed.
	observations := &qosobservations.CosmosSDKRequestObservations{
		ChainId:       rc.chainID,
		ServiceId:     string(rc.serviceID),
		RequestOrigin: rc.requestOrigin,
		RpcType:       convertToProtoRPCType(rc.rpcType),
	}

	// Add JSON-RPC request details if available
	if rc.jsonrpcReq != nil {
		observations.JsonrpcRequest = &qosobservations.JsonRpcRequest{
			Id:     rc.jsonrpcReq.ID.String(),
			Method: string(rc.jsonrpcReq.Method),
		}
	}

	// No endpoint responses received - set request error
	if len(rc.endpointResponses) == 0 {
		observations.RequestError = &qosobservations.RequestError{
			ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_INTERNAL_PROTOCOL_ERROR,
			ErrorDetails:   "No response received from any endpoint",
			HttpStatusCode: http.StatusInternalServerError,
		}

		return qosobservations.Observations{
			ServiceObservations: &qosobservations.Observations_Cosmos{
				Cosmos: observations,
			},
		}
	}

	// Build the endpoint(s) observations.
	endpointObservations := make([]*qosobservations.CosmosSDKEndpointObservation, len(rc.endpointResponses))
	for idx, endpointResponse := range rc.endpointResponses {
		obs := endpointResponse.GetObservation()
		obs.EndpointAddr = string(endpointResponse.EndpointAddr)
		endpointObservations[idx] = &obs
	}

	// Set the endpoint observations fields.
	observations.EndpointObservations = endpointObservations

	return qosobservations.Observations{
		ServiceObservations: &qosobservations.Observations_Cosmos{
			Cosmos: observations,
		},
	}
}

// GetEndpointSelector returns the endpoint selector for the request context.
func (rc *requestContext) GetEndpointSelector() protocol.EndpointSelector {
	return rc
}

// Select returns the address of an endpoint using the request context's endpoint store.
func (rc *requestContext) Select(allEndpoints protocol.EndpointAddrList) (protocol.EndpointAddr, error) {
	// This would typically use a service state or endpoint selector
	// For now, return the first endpoint as a placeholder
	if len(allEndpoints) > 0 {
		return allEndpoints[0], nil
	}
	return "", protocol.ErrNoEndpointsAvailable
}

// SelectMultiple returns multiple endpoint addresses using the request context's endpoint store.
func (rc *requestContext) SelectMultiple(allEndpoints protocol.EndpointAddrList, numEndpoints uint) (protocol.EndpointAddrList, error) {
	// Simple implementation - return up to numEndpoints from available endpoints
	if uint(len(allEndpoints)) <= numEndpoints {
		return allEndpoints, nil
	}
	return allEndpoints[:numEndpoints], nil
}
