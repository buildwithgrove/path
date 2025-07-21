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

	// Protocol-level error handlers
	//
	// Builds a response to return to the user.
	// Used only if no endpoint responses are received.
	protocolErrorResponseBuilder func() gateway.HTTPResponse

	// Builds a request error observation indicating protocol-level error.
	// Used only if no endpoint responses are received.
	protocolErrorObservationBuilder func() *qosobservations.RequestError

	// Validator to use to build user response/endpoint observations from the endpoint response.
	endpointResponseValidator func(polylog.Logger, [byte) response

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
func (rc *requestContext) GetServicePayload() protocol.Payload {
	return rc.servicePayload
}

// UpdateWithResponse processes a JSONRPC response from an endpoint
// Uses the existing response unmarshaling system
// NOT safe for concurrent use
func (rc *requestContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, responseBz []byte) {
	logger := rc.logger.With(
		"method", "UpdateWithResponse",
		"endpoint_addr", endpointAddr
	)

	// Parse and validate the endpoint response.
	parsedEndpointResponse := rc.endpointResponseValidator(logger, responseBz)

	rc.endpointResponses = append(jc.endpointResponses, endpointResponse{
		endpointAddr: endpointAddr,
		response:     parsedEndpointResponse,
	})
}

// GetHTTPResponse builds the HTTP response to return to the client
func (rc *requestContext) GetHTTPResponse() gateway.HTTPResponse {
	// No responses received - this is a protocol-level error
	if len(rc.endpointResponses) == 0 {
		rc.logger.Error().Msg("No endpoint responses received for JSONRPC request")
		return rc.protocolErrorResponseBuilder(rc.logger)
	}

	// Use the most recent response (as of current implementation, there's only one)
	latestResponse := rc.endpointResponses[len(rc.endpointResponses)-1]

	// Use the existing response system's HTTP response building
	return latestResponse.response.GetHTTPResponse()
}

// GetObservations returns QoS observations for JSONRPC requests
func (rc *requestContext) GetObservations() qosobservations.Observations {
	// Handle case where no endpoint responses were received
	if len(rc.endpointResponses) == 0 {
		rc.observations.RequestError = rc.protocolErrorObservationBuilder()

		return qosobservations.Observations{
			ServiceObservations: &qosobservations.Observations_Cosmos{
				Cosmos: rc.observations,
			},
		}
	}

	// Build endpoint observations using the existing response system
	endpointObservations := make([]*qosobservations.CosmosEndpointObservation, 0, len(rc.endpointResponses))
	for _, endpointResp := range rc.endpointResponses {
		endpointObs := endpointResp.response.GetObservation()
		endpointObs.EndpointAddr = string(endpointResp.endpointAddr)
		endpointObservations = append(endpointObservations, &endpointObs)
	}

	rc.observations.EndpointObservations = endpointObservations

	return qosobservations.Observations{
		ServiceObservations: &qosobservations.Observations_Cosmos{
			Cosmos: rc.observations,
		},
	}
}

// GetEndpointSelector returns the endpoint selector for the request context.
// Implements the gateway.RequestQoSContext interface.
func (rc *requestContext) GetEndpointSelector() protocol.EndpointSelector {
	return rc
}

// Select returns the address of an endpoint using the request context's service state.
// Implements the protocol.EndpointSelector interface.
func (rc *requestContext) Select(allEndpoints protocol.EndpointAddrList) (protocol.EndpointAddr, error) {
	return rc.serviceState.Select(allEndpoints)
}

// SelectMultiple returns multiple endpoint addresses using the request context's service state.
// Implements the protocol.EndpointSelector interface.
func (rc *requestContext) SelectMultiple(allEndpoints protocol.EndpointAddrList, numEndpoints uint) (protocol.EndpointAddrList, error) {
	// Select multiple endpoints from the available endpoints using the service state.
	return rc.serviceState.SelectMultiple(allEndpoints, numEndpoints)
}
