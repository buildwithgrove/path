package cosmos

import (
	"encoding/json"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/log"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
)

// responseHealth provides the functionality required from a response by a requestContext instance
var _ response = responseHealth{}

// responseUnmarshalerHealth deserializes the provided payload
// into a responseHealth struct, adding any encountered errors
// to the returned struct.
// Always returns a valid response interface, never returns an error
// Implements the response interface used by the request context.
func responseUnmarshalerHealth(
	logger polylog.Logger,
	data []byte,
) response {
	logger = logger.With("response_validator", "health")

	jsonrpcResponse, validationErr := unmarshalAsJSONRPCResponse(logger, data)
	// endpoint payload failed to parse as a valid JSONRPC response.
	if validationErr != nil {
		return responseHealth{
			logger:              logger,
			userJSONRPCResponse: jsonrpcResponse,
			validationError:     validationErr,
		}
	}

	// Any valid, non-error JSONRPC response to `/health` is accepted as valid.
	// Reference:
	// https://docs.cometbft.com/main/spec/rpc/#health
	return responseHealth{
		logger:              logger,
		userJSONRPCResponse: jsonrpcResponse,
		healthy:             true,
	}
}

// responseHealth captures a Cosmos REST API /health endpoint response
// Reference: https://docs.cometbft.com/v1.0/spec/rpc/#health
type responseHealth struct {
	logger polylog.Logger

	// healthy indicates if the endpoint is healthy
	healthy bool

	// validationError tracks any validation issues with the response
	validationError *qosobservations.CosmosResponseValidationError

	// tracks the JSONRPC response to return to the user.
	userJSONRPCResponse jsonrpc.Response
}

// GetObservation returns a Cosmos-based /health observation for REST endpoints
// Implements the response interface
func (r responseHealth) GetObservation() qosobservations.CosmosEndpointObservation {
	result := qosobservations.CosmosEndpointResponseValidationResult{
		ResponseValidationType: qosobservations.CosmosResponseValidationType_COSMOS_RESPONSE_VALIDATION_TYPE_JSONRPC,
		// Set HTTP status code based on the JSONRPC response.
		HttpStatusCode: r.userJSONRPCResponse.GetRecommendedHTTPStatusCode(),
		// Set validation error if present
		ValidationError: r.validationError,
		ParsedResponse: &qosobservations.CosmosEndpointResponseValidationResult_ResponseHealth{
			ResponseHealth: &qosobservations.CosmosResponseHealth{
				HealthStatus: r.healthy,
			},
		},
	}

	return qosobservations.CosmosEndpointObservation{
		// EndpointAddr is set by the caller.
		EndpointResponseValidationResult: result,
	}
}

// GetHTTPResponse builds and returns the httpResponse matching the responseRESTHealth instance
// Implements the response interface
func (r responseHealth) GetHTTPResponse() httpResponse {
	return httpResponse{
		responsePayload: r.getResponsePayload(),
		httpStatusCode:  r.userJSONRPCResponse.GetRecommendedHTTPStatusCode(),
	}
}

// GetResponsePayload returns the payload for the response to a REST `/health` request
func (r responseHealth) getResponsePayload() []byte {
	bz, err := json.Marshal(r.userJSONRPCResponse)
	if err != nil {
		r.logger.With(
			"jsonrpc_response", r.userJSONRPCResponse,
			"marshal_err", err,
		).Warn().Msg("SHOULD NEVER HAPPEN: error marshaling JSONRPC response.")
	}

	// Return serialized payload regardless of any errors
	return bz
}
