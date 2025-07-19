package cosmos

import (
	"encoding/json"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
)

// responseRESTHealth provides the functionality required from a response by a requestContext instance
var _ response = responseRESTHealth{}

// responseUnmarshalerRESTHealth deserializes the provided payload
// into a responseRESTHealth struct, adding any encountered errors
// to the returned struct. Always returns a valid response interface, never returns an error.
func responseUnmarshalerRESTHealth(
	logger polylog.Logger,
	data []byte,
) response {
	logger = logger.With("response_processor", "health")

	// Handle empty responses
	if len(data) == 0 {
		logger.Error().
			Str("endpoint", "/health").
			Msg("Received empty response from /health endpoint")

		return getRESTHealthEmptyErrorResponse(logger)
	}

	// Validate that the response is valid JSON
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		logger.Error().
			Err(err).
			Str("raw_payload", string(data)).
			Msg("Failed to unmarshal REST /health response as JSON")

		return getRESTHealthUnmarshalErrorResponse(logger, err)
	}

	// REST `/health` endpoint typically returns an empty JSON object {} on success
	// or any valid JSON response indicating the service is healthy
	// We consider any valid JSON response as healthy
	logger.Debug().
		Interface("health_response", jsonData).
		Msg("Successfully parsed /health response as healthy")

	return responseRESTHealth{
		logger:  logger,
		healthy: true,
	}
}

// responseRESTHealth captures a Cosmos REST API /health endpoint response
// Reference: https://docs.cometbft.com/v1.0/spec/rpc/#health
type responseRESTHealth struct {
	logger polylog.Logger

	// healthy indicates if the endpoint is healthy
	healthy bool

	// validationError tracks any validation issues with the response
	validationError *qosobservations.CosmosSDKResponseValidationError
}

// GetObservation returns a CosmosSDK-based /health observation for REST endpoints
// Implements the response interface
func (r responseRESTHealth) GetObservation() qosobservations.CosmosSDKEndpointObservation {
	return qosobservations.CosmosSDKEndpointObservation{
		ResponseObservation: &qosobservations.CosmosSDKEndpointObservation_RestObservation{
			RestObservation: &qosobservations.CosmosSDKEndpointRestObservation{
				ParsedResponse: &qosobservations.CosmosSDKEndpointRestObservation_HealthResponse{
					HealthResponse: &qosobservations.CosmosSDKRESTHealthResponse{
						HttpStatusCode:       int32(r.GetResponseStatusCode()),
						HealthStatusResponse: r.healthy,
					},
				},
			},
		},
	}
}

// GetResponsePayload returns the payload for the response to a REST `/health` request
// Implements the response interface
func (r responseRESTHealth) GetResponsePayload() []byte {
	return r.getResponsePayload()
}

// getResponsePayload returns the appropriate response payload based on health status
func (r responseRESTHealth) getResponsePayload() []byte {
	if r.healthy {
		// Return empty JSON object for healthy status
		return []byte("{}")
	}

	// Return error response for unhealthy status
	errorBody := map[string]interface{}{
		"error": "Service is unhealthy",
	}
	errorBz, _ := json.Marshal(errorBody)
	return errorBz
}

// GetResponseStatusCode returns an HTTP status code corresponding to the health status
// Implements the response interface
func (r responseRESTHealth) GetResponseStatusCode() int {
	// If we have a validation error, return 500
	if r.validationError != nil {
		return http.StatusInternalServerError
	}

	// Return 200 for healthy endpoints, 503 for unhealthy
	if r.healthy {
		return http.StatusOK
	}

	return http.StatusServiceUnavailable
}

// GetHTTPResponse builds and returns the httpResponse matching the responseRESTHealth instance
// Implements the response interface
func (r responseRESTHealth) GetHTTPResponse() httpResponse {
	return httpResponse{
		responsePayload: r.GetResponsePayload(),
		httpStatusCode:  r.GetResponseStatusCode(),
	}
}

// getRESTHealthEmptyErrorResponse creates an error response for empty /health responses
func getRESTHealthEmptyErrorResponse(logger polylog.Logger) responseRESTHealth {
	validationError := qosobservations.CosmosSDKResponseValidationError_COSMOS_SDK_RESPONSE_VALIDATION_ERROR_EMPTY

	return responseRESTHealth{
		logger:          logger,
		healthy:         false,
		validationError: &validationError,
	}
}

// getRESTHealthUnmarshalErrorResponse creates an error response for REST health unmarshaling failures
func getRESTHealthUnmarshalErrorResponse(
	logger polylog.Logger,
	err error,
) responseRESTHealth {
	validationError := qosobservations.CosmosSDKResponseValidationError_COSMOS_SDK_RESPONSE_VALIDATION_ERROR_UNMARSHAL

	return responseRESTHealth{
		logger:          logger,
		healthy:         false,
		validationError: &validationError,
	}
}
