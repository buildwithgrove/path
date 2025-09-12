package solana

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"

	"github.com/buildwithgrove/path/gateway"
	pathhttp "github.com/buildwithgrove/path/network/http"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// batchJSONRPCRequestContext provides the support required by the gateway
// package for handling service requests.
var _ gateway.RequestQoSContext = &batchJSONRPCRequestContext{}

type endpointJSONRPCResponse struct {
	protocol.EndpointAddr
	jsonrpc.Response
}

// batchJSONRPCRequestContext provides the functionality required
// to support QoS for a Solana blockchain service.
type batchJSONRPCRequestContext struct {
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

	JSONRPCBatchRequest jsonrpc.BatchRequest

	// The origin of the request handled by the context.
	// Either:
	// - User: user requests
	// - QoS: requests built by the QoS service to get additional data points on endpoints.
	requestOrigin qosobservations.RequestOrigin

	// endpointJSONRPCResponses is the set of responses received from one or
	// more endpoints as part of handling this service request.
	endpointJSONRPCResponses []endpointJSONRPCResponse
}

// TODO_NEXT(@commoddity): handle batch requests for Solana
// TODO_MVP(@adshmh): Ensure the JSONRPC request struct
// can handle all valid service requests.
func (brc batchJSONRPCRequestContext) GetServicePayloads() []protocol.Payload {
	protocolPayloads := make([]protocol.Payload, len(brc.JSONRPCBatchRequest.Requests))

	for i, jsonrpcRequestPayload := range brc.JSONRPCBatchRequest.GetRequestsPayloads() {
		// TODO_TECHDEBT(@adshmh): Set method-specific timeouts on protocol payload entry.
		protocolPayloads[i] = protocol.Payload{
			Data:    string(jsonrpcRequestPayload),
			Method:  http.MethodPost, // Method is alway POST for Solana.
			Path:    "",              // Path field is not used for Solana.
			RPCType: sharedtypes.RPCType_JSON_RPC,
		}
	}

	return protocolPayloads
}

// TODO_TECHDEBT(@adshmh): Refactor once the QoS context interface is updated to receive an array of responses.
// UpdateWithResponse is NOT safe for concurrent use
func (brc *batchJSONRPCRequestContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, responseBz []byte) {
	// TODO_TECHDEBT(@adshmh): Refactor this once the QoS context interface is updated to accept all endpoint responses at once.
	// This would make it possible to map each JSONRPC request of a batch to its corresponding endpoint response.
	// This is required to enable request method-specific esponse validation: e.g. format of result field in response to a `getHealth` request.
	//
	// Parse and track the endpoint payload as a JSONRPC response.
	var jsonrpcResponse jsonrpc.Response
	if err := json.Unmarshal(responseBz, &jsonrpcResponse); err != nil {
		// TODO_UPNEXT(@adshmh): Include a preview of malformed payload in the response.
		//
		// Parsing failed, store a generic error JSONRPC response
		jsonrpcResponse = jsonrpc.GetErrorResponse(jsonrpc.ID{}, errCodeUnmarshaling, errMsgUnmarshaling, nil)
	}

	// Store the response: will be processed later by the JSONRPC batch request struct.
	brc.endpointJSONRPCResponses = append(brc.endpointJSONRPCResponses, endpointJSONRPCResponse{
		EndpointAddr: endpointAddr,
		Response:     jsonrpcResponse,
	})
}

// TODO_MVP(@adshmh): add `Content-Type: application/json` header.
// GetHTTPResponse builds the HTTP response that should be returned for
// a Solana blockchain service request.
func (brc batchJSONRPCRequestContext) GetHTTPResponse() pathhttp.HTTPResponse {
	// TODO_UPNEXT(@adshmh): Return an error response matching the batch of JSONRPC requests.
	//
	// No responses received: this is an internal error:
	// e.g. protocol-level errors like endpoint timing out.
	if len(brc.endpointJSONRPCResponses) == 0 {
		// Build the JSONRPC response indicating a protocol-level error.
		jsonrpcErrorResponse := jsonrpc.NewErrResponseInternalErr(jsonrpc.ID{}, errors.New("protocol-level error: no endpoint responses received"))
		return qos.BuildHTTPResponseFromJSONRPCResponse(brc.logger, jsonrpcErrorResponse)
	}

	// assemble the array of JSONRPC responses
	jsonrpcResponses := make([]jsonrpc.Response, len(brc.endpointJSONRPCResponses))
	for i, jsonrpcResponse := range brc.endpointJSONRPCResponses {
		jsonrpcResponses[i] = jsonrpcResponse.Response
	}

	// Use the Batch JSONRPC request to assemble the JSONRPC batch response.
	batchResponseBz := brc.JSONRPCBatchRequest.BuildResponseBytes(jsonrpcResponses)

	// TODO_UPNEXT(@adshmh): Adjust HTTP status code according to responses in the batch.
	return jsonrpc.HTTPResponse{
		ResponsePayload: batchResponseBz,
		// According to the JSON-RPC 2.0 specification, even if individual responses
		// in a batch contain errors, the entire batch should still return HTTP 200 OK.
		HTTPStatusCode: http.StatusOK,
	}
}

// TODO_IMPROVE(@adshmh): Track the method field of each request in a JSONRPC batch request:
// - Update proto/path/qos/solana.proto to include request details in each endpoint observation.
// - Map each request in a batch to its corresponding response: needs gateway.QoSRequestContext interface update to handle slice of response.
// - Update the endpoint observation building code below to include details of the corresponding request.
//
// GetObservations returns all the observations contained in the request context.
// Implements the gateway.RequestQoSContext interface.
func (rc batchJSONRPCRequestContext) GetObservations() qosobservations.Observations {
	// Set the observation fields common for all requests: successful or failed.
	observations := &qosobservations.SolanaRequestObservations{
		ChainId:              rc.chainID,
		ServiceId:            string(rc.serviceID),
		RequestPayloadLength: uint32(rc.requestPayloadLength),
		RequestOrigin:        rc.requestOrigin,
		// TODO_UPNEXT(@adshmh): Add a Batch JSONRPC request observation.
	}

	// No endpoint responses received.
	// Set request error.
	if len(rc.endpointJSONRPCResponses) == 0 {
		observations.RequestError = qos.GetRequestErrorForProtocolError()

		return qosobservations.Observations{
			ServiceObservations: &qosobservations.Observations_Solana{
				Solana: observations,
			},
		}
	}

	// Add one endpoint observation per request in the JSONRPC batch request.
	endpointObservations := make([]*qosobservations.SolanaEndpointObservation, len(rc.endpointJSONRPCResponses))
	for index, endpointResp := range rc.endpointJSONRPCResponses {
		// TODO_TECHDEBT(@adshmh): Support method-specific JSONRPC responses on batch requests.
		// This requires mapping each endpoint response to its corresponding request in the batch.
		//
		endpointObs := &qosobservations.SolanaEndpointObservation{
			// TODO_DOCUMENT(@adshmh): Add a reference for the choice of HTTP status code on batch requests.
			//
			// HTTP status code 200 for batch requests.
			HttpStatusCode: int32(http.StatusOK),
			// Track response as an unrecognized response, since QoS does not currently use batch requests to evaluate endpoints.
			ResponseObservation: &qosobservations.SolanaEndpointObservation_UnrecognizedResponse{
				UnrecognizedResponse: &qosobservations.SolanaUnrecognizedResponse{
					// Track details of the JSONRPC response: e.g. ID and a preview of result.
					JsonrpcResponse: endpointResp.GetObservation(),
				},
			},
		}

		// Store in the list of endpoint observations.
		endpointObservations[index] = endpointObs
	}

	observations.EndpointObservations = endpointObservations

	return qosobservations.Observations{
		ServiceObservations: &qosobservations.Observations_Solana{
			Solana: observations,
		},
	}
}

// GetEndpointSelector is required to satisfy the gateway package's RequestQoSContext interface.
// The request context is queried for the correct endpoint selector.
// This allows different endpoint selectors based on the request's context.
// e.g. the request context for a particular request method can potentially rank endpoints based on their latency when responding to requests with matching method.
func (rc *batchJSONRPCRequestContext) GetEndpointSelector() protocol.EndpointSelector {
	return rc
}

// TODO_TECHDEBT(@adshmh): Enhance endpoint selection to consider endpoint quality specific to batch requests.
//
// Select chooses an endpoint from the list of supplied endpoints.
// It uses the perceived state of the Solana chain using other endpoints' responses.
// It is required to satisfy the protocol package's EndpointSelector interface.
func (rc *batchJSONRPCRequestContext) Select(allEndpoints protocol.EndpointAddrList) (protocol.EndpointAddr, error) {
	return rc.endpointStore.Select(allEndpoints)
}

// SelectMultiple chooses multiple endpoints from the list of supplied endpoints.
// It uses the perceived state of the Solana chain using other endpoints' responses.
// It is required to satisfy the protocol package's EndpointSelector interface.
func (rc *batchJSONRPCRequestContext) SelectMultiple(allEndpoints protocol.EndpointAddrList, numEndpoints uint) (protocol.EndpointAddrList, error) {
	return rc.endpointStore.SelectMultiple(allEndpoints, numEndpoints)
}
