package cosmos

import (
	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// jsonrpcUnrecognizedResponse handles unrecognized JSONRPC responses
// Implements the response interface for cases where the response cannot be properly parsed
// but we still have a valid JSONRPC response structure
type jsonrpcUnrecognizedResponse struct {
	logger          polylog.Logger
	jsonrpcResponse jsonrpc.Response
	validationErr   qosobservations.CosmosResponseValidationError
}

// GetHTTPResponse builds the HTTP response to return to the client
// Uses the existing QoS helper to build response from JSONRPC response
func (r *jsonrpcUnrecognizedResponse) GetHTTPResponse() gateway.HTTPResponse {
	return qos.BuildHTTPResponseFromJSONRPCResponse(r.logger, r.jsonrpcResponse)
}

// GetObservation returns the QoS observation for this unrecognized response
func (r *jsonrpcUnrecognizedResponse) GetObservation() qosobservations.CosmosEndpointObservation {
	return qosobservations.CosmosEndpointObservation{
		EndpointResponseValidationResult: &qosobservations.CosmosEndpointResponseValidationResult{
			ResponseValidationType: qosobservations.CosmosResponseValidationType_COSMOS_RESPONSE_VALIDATION_TYPE_JSONRPC,
			HttpStatusCode:         int32(r.jsonrpcResponse.GetRecommendedHTTPStatusCode()),
			ValidationError:        &r.validationErr,
			ParsedResponse: &qosobservations.CosmosEndpointResponseValidationResult_ResponseJsonrpc{
				ResponseJsonrpc: r.jsonrpcResponse.GetObservation(),
			},
		},
	}
}
