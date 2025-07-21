package cosmos

import (
	"net/http"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
	"github.com/pokt-network/poktroll/pkg/polylog"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
)

// requestValidator handles validation for all Cosmos service requests
// Coordinates between different protocol validators (JSONRPC, REST)
type requestValidator struct {
	logger        polylog.Logger
	chainID       string
	serviceID     protocol.ServiceID
	supportedAPIs map[sharedtypes.RPCType]struct{}
	serviceState  protocol.EndpointSelector
}

// validateHTTPRequest validates an HTTP request and routes to appropriate sub-validator
// Returns (context, true) on success or (errorContext, false) on failure
func (rv *requestValidator) validateHTTPRequest(req *http.Request) (gateway.RequestQoSContext, bool) {
	logger := rv.logger.With(
		"qos", "Cosmos",
		"method", "validateHTTPRequest",
		"path", req.URL.Path,
		"http_method", req.Method,
	)

	// Read the request body.
	// This is necessary to distinguish REST vs. JSONRPC on request with POST HTTP method.
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to parse JSONRPC request")
		// Return a context with a JSONRPC-formatted response, as we cannot detect the request type.
		return createHTTPBodyReadFailureContext(err), false
	}

	// Determine request type and route to appropriate validator
	if isJSONRPCRequest(req.Method, body) {
		logger.Debug().Msg("Routing to JSONRPC validator")

		// Validate the JSONRPC request.
		// Builds and returns a context to handle the request.
		// Uses a specialized context for handling invalid requests.
		return rv.validateJSONRPCRequest(body)
	} else {
		logger.Debug().Msg("Routing to REST validator")

		// Build and returns a request context to handle the REST request.
		// Uses a specialized context for handling invalid requests.
		return rv.validateRESTRequest(req.URL, req.Method, body)
	}
}

// isJSONRPCRequest determines if the incoming HTTP request is a JSONRPC request
// Uses simple heuristics: POST method and specific content.
func isJSONRPCRequest(httpMethod string, httpRequestBody []byte) bool {
	// Stage 1: Non-POST requests are always REST
	if httpMethod != http.MethodPost {
		return false
	}

	// Stage 2: POST requests - check for JSONRPC payload
	if strings.Contains(string(httpRequestBody), "jsonrpc") {
		return true
	}

	// Stage 3: POST without jsonrpc field is REST
	return false
}

// createHTTPBodyReadFailureContext:
// - Creates an error context for HTTP body read failures
// - Used when the HTTP request body cannot be read
func (rv *requestValidator) createHTTPBodyReadFailureContext(err error) gateway.RequestQoSContext {
	// Create the JSON-RPC error response
	response := jsonrpc.NewErrResponseInternalErr(jsonrpc.ID{}, err)

	// Create the observations object with the HTTP body read failure observation
	observations := createHTTPBodyReadFailureObservation(rv.serviceID, rv.chainID, err, response)

	// Build and return the error context
	return &qos.RequestErrorContext{
		Logger:   rv.logger,
		Response: response,
		Observations: &qosobservations.Observations{
			ServiceObservations: observations,
		},
	}
}

func createHTTPBodyReadFailureObservation(
	serviceID protocol.ServiceID,
	chainID string,
	err error,
	jsonrpcResponse jsonrpc.Response,
) *qosobservations.Observations_Cosmos {
	return &qosobservations.Observations_Cosmos{
		Cosmos: &qosobservations.CosmosRequestObservations{
			ServiceId: string(serviceID),
			ChainId:   chainID,
			RequestError: &qosobservations.RequestError{
				ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_INTERNAL_READ_HTTP_ERROR,
				ErrorDetails:   err.Error(),
				HttpStatusCode: int32(jsonrpcResponse.GetRecommendedHTTPStatusCode()),
			},
		},
	}
}
