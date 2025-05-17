package solana

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// Maximum length for error messages stored in validation failure logs/observations.
// - Prevents overly verbose error messages in logs/metrics
// - Reduces memory usage and cost
const maxErrMessageLen = 1000

// TODO_MVP(@adshmh): Drop this struct once JUDGE framework is merged.
//
// requestValidator:
// - Handles request validation for Solana JSONRPC requests
// - Generates error contexts if validation fails (e.g. error parsing JSONRPC request)
// - Generates request context if validation succeeds
// - Used as the entry point for HTTP request validation
type requestValidator struct {
	logger        polylog.Logger
	chainID       string
	serviceID     protocol.ServiceID
	endpointStore *EndpointStore
}

// validateHTTPRequest:
// - Validates an HTTP request for a Solana JSONRPC payload
// - Extracts and validates the JSONRPC request from the HTTP body
// - Returns (errorContext, false) if validation fails
// - Returns (requestContext, true) if validation succeeds
func (rv *requestValidator) validateHTTPRequest(req *http.Request) (gateway.RequestQoSContext, bool) {
	logger := rv.logger.With(
		"qos", "Solana",
		"method", "validateHTTPRequest",
	)

	// Read the HTTP request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Warn().Err(err).Msg("HTTP request body read failed - returning generic error response")
		return rv.createHTTPBodyReadFailureContext(err), false
	}

	// Parse and validate the JSONRPC request
	jsonrpcReq, err := parseJSONRPCFromRequestBody(logger, body)
	if err != nil {
		return rv.createRequestUnmarshalingFailureContext(jsonrpcReq.ID, err), false
	}

	// Request is valid, return a fully initialized requestContext
	return &requestContext{
		logger:               rv.logger,
		chainID:              rv.chainID,
		serviceID:            rv.serviceID,
		requestPayloadLength: uint(len(body)),
		JSONRPCReq:           jsonrpcReq,
		// Set the origin of the request as USER (i.e. organic relay)
		// The request is from a user.
		requestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
		endpointStore: rv.endpointStore,
	}, true
}

// createHTTPBodyReadFailureContext:
// - Creates an error context for HTTP body read failures
// - Used when the HTTP request body cannot be read
func (rv *requestValidator) createHTTPBodyReadFailureContext(err error) gateway.RequestQoSContext {
	// TODO_IMPROVE(@adshmh): Propagate a request ID parameter on internal errors
	// that occur after successful request parsing.
	// There are no such cases as of PR #186.
	//
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

// createRequestUnmarshalingFailureContext:
// - Creates an error context for JSONRPC request unmarshaling failures
// - Used when the request body cannot be parsed into a valid JSONRPC request
func (rv *requestValidator) createRequestUnmarshalingFailureContext(id jsonrpc.ID, err error) gateway.RequestQoSContext {
	// Create the JSON-RPC error response
	response := jsonrpc.NewErrResponseInvalidRequest(id, err)

	// Create the observations object with the request unmarshaling failure observation
	observations := createRequestUnmarshalingFailureObservation(id, rv.serviceID, rv.chainID, err, response)

	// Build and return the error context
	return &qos.RequestErrorContext{
		Logger:   rv.logger,
		Response: response,
		Observations: &qosobservations.Observations{
			ServiceObservations: observations,
		},
	}
}

// createRequestUnmarshalingFailureObservation:
// - Creates an observation for a Solana request that failed to unmarshal from JSON
// - Captures details for metrics and debugging:
//   - Request ID
//   - Error message
//   - Chain ID
//   - HTTP status code
//
// Parameters:
//   - id: JSON-RPC request ID
//   - serviceID: Service identifier
//   - chainID: Solana chain identifier
//   - err: Error from unmarshaling
//   - jsonrpcResponse: Associated JSONRPC error response
//
// Returns:
//   - Structured observation with failure details
func createRequestUnmarshalingFailureObservation(
	_ jsonrpc.ID,
	serviceID protocol.ServiceID,
	chainID string,
	err error,
	jsonrpcResponse jsonrpc.Response,
) *qosobservations.Observations_Solana {
	return &qosobservations.Observations_Solana{
		Solana: &qosobservations.SolanaRequestObservations{
			ServiceId: string(serviceID),
			ChainId:   chainID,
			RequestError: &qosobservations.RequestError{
				ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_USER_ERROR_JSONRPC_PARSE_ERROR,
				ErrorDetails:   err.Error(),
				HttpStatusCode: int32(jsonrpcResponse.GetRecommendedHTTPStatusCode()),
			},
		},
	}
}

// createHTTPBodyReadFailureObservation:
// - Creates an observation for Solana HTTP request body read failures
// - Captures details for troubleshooting:
//   - Chain ID
//   - Error message
//   - HTTP status code
//
// - Useful for diagnosing connectivity or parsing issues
//
// Parameters:
//   - serviceID: Service identifier
//   - chainID: Solana chain identifier
//   - err: Error from HTTP body read
//   - jsonrpcResponse: Associated JSONRPC error response
//
// Returns:
//   - Structured observation with HTTP read failure details
func createHTTPBodyReadFailureObservation(
	serviceID protocol.ServiceID,
	chainID string,
	err error,
	jsonrpcResponse jsonrpc.Response,
) *qosobservations.Observations_Solana {
	return &qosobservations.Observations_Solana{
		Solana: &qosobservations.SolanaRequestObservations{
			ServiceId: string(serviceID),
			ChainId:   chainID,
			RequestError: &qosobservations.RequestError{
				ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_USER_ERROR_JSONRPC_PARSE_ERROR,
				ErrorDetails:   err.Error(),
				HttpStatusCode: int32(jsonrpcResponse.GetRecommendedHTTPStatusCode()),
			},
		},
	}
}

// TODO_TECHDEBT(@adshmh): Support Batch JSONRPC requests per spec:
// https://www.jsonrpc.org/specification#batch
//
// TODO_MVP(@adshmh): Add a JSON-RPC request validator to reject invalid/unsupported method calls early.
//
// parseJSONRPCFromRequestBody:
// - Attempts to unmarshal HTTP request body into a JSONRPC request structure
// - On failure:
//   - Logs a preview of the request body (truncated for security/performance)
//   - Logs the specific error
//
// Parameters:
//   - logger: Logger for structured logging
//   - requestBody: Raw HTTP request body bytes
//
// Returns:
//   - jsonrpc.Request: Parsed request (empty on error)
//   - error: Any error encountered during parsing
func parseJSONRPCFromRequestBody(
	logger polylog.Logger,
	requestBody []byte,
) (jsonrpc.Request, error) {
	var jsonrpcRequest jsonrpc.Request
	err := json.Unmarshal(requestBody, &jsonrpcRequest)
	if err != nil {
		// Only log a preview of the request body (first 1000 bytes or less) to avoid excessive logging
		requestPreview := string(requestBody[:min(maxErrMessageLen, len(requestBody))])
		logger.Info().Err(err).Msgf("Request failed JSON-RPC validation - returning generic error response. Request preview: %s", requestPreview)
	}

	return jsonrpcRequest, err
}
