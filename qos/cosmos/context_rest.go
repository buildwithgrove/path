package cosmos

import (
	"encoding/json"
	"net/http"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/pokt-network/poktroll/pkg/polylog"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
)

const (
	// Default timeout for REST requests to Cosmos endpoints
	defaultRESTRequestTimeoutMillisec = 15_000
)

// restContext provides specialized context for REST API requests
// Implements gateway.RequestQoSContext interface
type restContext struct {
	logger    polylog.Logger
	chainID   string
	serviceID protocol.ServiceID

	// REST-specific fields
	httpMethod           string
	urlPath              string
	requestBody          []byte
	rpcType              sharedtypes.RPCType // REST or COMET_BFT
	requestPayloadLength uint
	requestOrigin        qosobservations.RequestOrigin

	// HTTP headers from original request (for forwarding)
	headers map[string]string

	// Service state for endpoint selection
	serviceState protocol.EndpointSelector

	// Endpoint response tracking
	endpointResponses []endpointResponse
}

// GetServicePayload builds the REST payload to send to blockchain endpoints
func (rc *restContext) GetServicePayload() protocol.Payload {
	var data string
	if rc.requestBody != nil {
		data = string(rc.requestBody)
	}

	return protocol.Payload{
		Data:            data,
		Method:          rc.httpMethod,
		Path:            rc.urlPath,
		TimeoutMillisec: defaultRESTRequestTimeoutMillisec,
		// TODO: Forward relevant headers if needed
	}
}

// UpdateWithResponse processes a REST response from an endpoint
// Uses the existing response unmarshaling system for REST responses
// NOT safe for concurrent use
func (rc *restContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, responseBz []byte) {
	// Use the existing response unmarshaling system
	// For REST requests, we pass the URL path as the apiPath and set isJSONRPC to false
	resp, err := unmarshalResponse(rc.logger, rc.urlPath, responseBz, false, endpointAddr)
	if err != nil {
		rc.logger.Error().
			Err(err).
			Str("endpoint", string(endpointAddr)).
			Str("path", rc.urlPath).
			Msg("Failed to unmarshal REST response from endpoint")

		// Create a generic error response for tracking
		resp = responseGeneric{
			logger:         rc.logger,
			rawData:        responseBz,
			isRestResponse: true,
		}
	}

	rc.endpointResponses = append(rc.endpointResponses, endpointResponse{
		endpointAddr: endpointAddr,
		response:     resp,
	})
}

// GetHTTPResponse builds the HTTP response to return to the client
func (rc *restContext) GetHTTPResponse() gateway.HTTPResponse {
	// No responses received - this is a protocol-level error
	if len(rc.endpointResponses) == 0 {
		rc.logger.Error().Msg("No endpoint responses received for REST request")

		errorBody := map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Protocol-level error: no endpoint responses received",
				"code":    "INTERNAL_ERROR",
			},
		}

		errorBz, _ := json.Marshal(errorBody)

		return gateway.HTTPResponse{
			Body: errorBz,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			StatusCode: http.StatusInternalServerError,
		}
	}

	// Use the most recent response (as of current implementation, there's only one)
	latestResponse := rc.endpointResponses[len(rc.endpointResponses)-1]

	// Use the existing response system's HTTP response building
	httpResp := latestResponse.response.GetHTTPResponse()

	return gateway.HTTPResponse{
		Body:       httpResp.responsePayload,
		StatusCode: httpResp.httpStatusCode,
		Headers: map[string]string{
			"Content-Type": "application/json", // Default for Cosmos REST APIs
		},
	}
}

// GetObservations returns QoS observations for REST requests
func (rc *restContext) GetObservations() qosobservations.Observations {
	observations := &qosobservations.CosmosSDKRequestObservations{
		ChainId:              rc.chainID,
		ServiceId:            string(rc.serviceID),
		RequestPayloadLength: uint32(rc.requestPayloadLength),
		RequestOrigin:        rc.requestOrigin,
		RpcType:              convertToProtoRPCType(rc.rpcType),
		// Note: JsonrpcRequest is nil for REST requests
	}

	// Handle case where no endpoint responses were received
	if len(rc.endpointResponses) == 0 {
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
	endpointObservations := make([]*qosobservations.CosmosSDKEndpointObservation, 0, len(rc.endpointResponses))
	for _, endpointResp := range rc.endpointResponses {
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
func (rc *restContext) GetEndpointSelector() protocol.EndpointSelector {
	return rc
}

// Select returns the address of an endpoint using the request context's service state.
// Implements the protocol.EndpointSelector interface.
func (rc *restContext) Select(allEndpoints protocol.EndpointAddrList) (protocol.EndpointAddr, error) {
	return rc.serviceState.Select(allEndpoints)
}

// SelectMultiple returns multiple endpoint addresses using the request context's service state.
// Implements the protocol.EndpointSelector interface.
func (rc *restContext) SelectMultiple(allEndpoints protocol.EndpointAddrList, numEndpoints uint) (protocol.EndpointAddrList, error) {
	// Select multiple endpoints from the available endpoints using the service state.
	return rc.serviceState.SelectMultiple(allEndpoints, numEndpoints)
}
