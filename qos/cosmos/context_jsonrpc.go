package cosmos

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/pokt-network/poktroll/pkg/polylog"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
)

const (
	// Default timeout for JSONRPC requests to Cosmos endpoints
	defaultJSONRPCRequestTimeoutMillisec = 15_000
)

// jsonrpcContext provides specialized context for JSONRPC requests
// Implements gateway.RequestQoSContext interface
type jsonrpcContext struct {
	logger    polylog.Logger
	chainID   string
	serviceID protocol.ServiceID

	// JSONRPC-specific fields
	jsonrpcReq           jsonrpc.Request
	rpcType              sharedtypes.RPCType // JSON_RPC or COMET_BFT
	requestPayloadLength uint
	requestOrigin        qosobservations.RequestOrigin

	// Service state for endpoint selection
	serviceState protocol.EndpointSelector

	// Endpoint response tracking
	endpointResponses []endpointResponse
}

// endpointResponse tracks a response from a specific endpoint
type endpointResponse struct {
	endpointAddr protocol.EndpointAddr
	response     response
}

// GetServicePayload builds the JSONRPC payload to send to blockchain endpoints
func (jc *jsonrpcContext) GetServicePayload() protocol.Payload {
	reqBz, err := json.Marshal(jc.jsonrpcReq)
	if err != nil {
		jc.logger.Error().Err(err).Msg("Failed to marshal JSONRPC request")
		return protocol.Payload{}
	}

	return protocol.Payload{
		Data:            string(reqBz),
		Method:          http.MethodPost, // JSONRPC always uses POST
		TimeoutMillisec: defaultJSONRPCRequestTimeoutMillisec,
		// Path is not used for JSONRPC (typically sent to root endpoint)
	}
}

// UpdateWithResponse processes a JSONRPC response from an endpoint
// Uses the existing response unmarshaling system
// NOT safe for concurrent use
func (jc *jsonrpcContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, responseBz []byte) {
	// Use the existing response unmarshaling system
	apiPath := string(jc.jsonrpcReq.Method)
	resp, err := unmarshalResponse(jc.logger, apiPath, responseBz, true, endpointAddr)
	if err != nil {
		jc.logger.Error().
			Err(err).
			Str("endpoint", string(endpointAddr)).
			Str("method", apiPath).
			Msg("Failed to unmarshal JSONRPC response from endpoint")

		// Create a generic error response for tracking
		errorResp := jsonrpc.NewErrResponseInternalErr(
			jc.jsonrpcReq.ID,
			errors.New("failed to unmarshal endpoint response"),
		)
		resp = responseGeneric{
			logger:          jc.logger,
			jsonRPCResponse: errorResp,
			isRestResponse:  false,
		}
	}

	jc.endpointResponses = append(jc.endpointResponses, endpointResponse{
		endpointAddr: endpointAddr,
		response:     resp,
	})
}

// GetHTTPResponse builds the HTTP response to return to the client
func (jc *jsonrpcContext) GetHTTPResponse() gateway.HTTPResponse {
	// No responses received - this is a protocol-level error
	if len(jc.endpointResponses) == 0 {
		jc.logger.Error().Msg("No endpoint responses received for JSONRPC request")
		errorResp := jsonrpc.NewErrResponseInternalErr(
			jc.jsonrpcReq.ID,
			errors.New("protocol-level error: no endpoint responses received"),
		)
		return qos.BuildHTTPResponseFromJSONRPCResponse(jc.logger, errorResp)
	}

	// Use the most recent response (as of current implementation, there's only one)
	latestResponse := jc.endpointResponses[len(jc.endpointResponses)-1]

	// Use the existing response system's HTTP response building
	httpResp := latestResponse.response.GetHTTPResponse()

	return gateway.HTTPResponse{
		Body:       httpResp.responsePayload,
		StatusCode: httpResp.httpStatusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// GetObservations returns QoS observations for JSONRPC requests
func (jc *jsonrpcContext) GetObservations() qosobservations.Observations {
	observations := &qosobservations.CosmosSDKRequestObservations{
		ChainId:              jc.chainID,
		ServiceId:            string(jc.serviceID),
		RequestPayloadLength: uint32(jc.requestPayloadLength),
		RequestOrigin:        jc.requestOrigin,
		RpcType:              convertToProtoRPCType(jc.rpcType),
		JsonrpcRequest:       jc.jsonrpcReq.GetObservation(),
	}

	// Handle case where no endpoint responses were received
	if len(jc.endpointResponses) == 0 {
		observations.RequestError = &qosobservations.RequestError{
			ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_INTERNAL_PROTOCOL_ERROR,
			ErrorDetails:   "No endpoint responses received",
			HttpStatusCode: int32(http.StatusInternalServerError),
		}

		return qosobservations.Observations{
			ServiceObservations: &qosobservations.Observations_Cosmos{
				Cosmos: observations,
			},
		}
	}

	// Build endpoint observations using the existing response system
	endpointObservations := make([]*qosobservations.CosmosSDKEndpointObservation, 0, len(jc.endpointResponses))
	for _, endpointResp := range jc.endpointResponses {
		endpointObs := endpointResp.response.GetObservation()
		endpointObs.EndpointAddr = string(endpointResp.endpointAddr)
		endpointObservations = append(endpointObservations, &endpointObs)
	}

	observations.EndpointObservations = endpointObservations

	return qosobservations.Observations{
		ServiceObservations: &qosobservations.Observations_Cosmos{
			Cosmos: observations,
		},
	}
}

// GetEndpointSelector returns the endpoint selector for the request context.
// Implements the gateway.RequestQoSContext interface.
func (jc *jsonrpcContext) GetEndpointSelector() protocol.EndpointSelector {
	return jc
}

// Select returns the address of an endpoint using the request context's service state.
// Implements the protocol.EndpointSelector interface.
func (jc *jsonrpcContext) Select(allEndpoints protocol.EndpointAddrList) (protocol.EndpointAddr, error) {
	return jc.serviceState.Select(allEndpoints)
}

// SelectMultiple returns multiple endpoint addresses using the request context's service state.
// Implements the protocol.EndpointSelector interface.
func (jc *jsonrpcContext) SelectMultiple(allEndpoints protocol.EndpointAddrList, numEndpoints uint) (protocol.EndpointAddrList, error) {
	// Select multiple endpoints from the available endpoints using the service state.
	return jc.serviceState.SelectMultiple(allEndpoints, numEndpoints)
}
