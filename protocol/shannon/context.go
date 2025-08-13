package shannon

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/pokt-network/poktroll/pkg/relayer/proxy"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
	sdk "github.com/pokt-network/shannon-sdk"

	"github.com/buildwithgrove/path/gateway"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/websockets"
)

// TODO_TECHDEBT(@adshmh): Make this threshold configurable.
//
// Maximum time to wait before using a fallback endpoint.
const maxWaitBeforeFallbackMillisecond = 1000

// Maximum endpoint payload length for error logging (100 chars)
const maxEndpointPayloadLenForLogging = 100

// requestContext provides all the functionality required by the gateway package
// for handling a single service request.
var _ gateway.ProtocolRequestContext = &requestContext{}

// RelayRequestSigner:
//   - Used by requestContext to sign relay requests
//   - Takes an unsigned relay request and an application
//   - Returns a relay request signed by the gateway (with delegation from the app)
//   - In future Permissionless Gateway Mode, may use the app's own private key for signing
type RelayRequestSigner interface {
	SignRelayRequest(req *servicetypes.RelayRequest, app apptypes.Application) (*servicetypes.RelayRequest, error)
}

// requestContext captures all data required for handling a single service request.
type requestContext struct {
	logger polylog.Logger

	// Upstream context for timeout propagation and cancellation
	context context.Context

	fullNode FullNode
	// TODO_TECHDEBT(@adshmh): add sanctionedEndpointsStore to the request context.
	serviceID protocol.ServiceID

	relayRequestSigner RelayRequestSigner

	// selectedEndpoint:
	// 	 - Endpoint selected for sending a relay.
	// 	 - Must be set via SelectEndpoint before sending a relay (otherwise sending fails).
	selectedEndpoint endpoint

	// requestErrorObservation:
	//   - Tracks any errors encountered during request processing.
	requestErrorObservation *protocolobservations.ShannonRequestError

	// endpointObservations:
	//   - Captures observations about endpoints used during request handling.
	//   - Includes enhanced error classification for raw payload analysis.
	endpointObservations []*protocolobservations.ShannonEndpointObservation

	// currentRelayMinerError:
	//   - Tracks RelayMinerError data from the current relay response for reporting.
	//   - Set by trackRelayMinerError method and used when building observations.
	currentRelayMinerError *protocolobservations.ShannonRelayMinerError

	// HTTP client used for sending relay requests to endpoints while also capturing various debug metrics
	httpClient *httpClientWithDebugMetrics

	fallbackEndpoints map[protocol.EndpointAddr]endpoint
}

// HandleServiceRequest:
//   - Satisfies gateway.ProtocolRequestContext interface.
//   - Uses supplied payload to send a relay request to an endpoint.
//   - Verifies and returns the response.
//   - Captures RelayMinerError data when available for reporting purposes.
func (rc *requestContext) HandleServiceRequest(payload protocol.Payload) (protocol.Response, error) {
	// Internal error: No endpoint selected.
	if rc.selectedEndpoint == nil {
		return rc.handleInternalError(fmt.Errorf("HandleServiceRequest: no endpoint has been selected on service %s", rc.serviceID))
	}

	// Record endpoint query time.
	endpointQueryTime := time.Now()

	var (
		relayResponse protocol.Response
		err           error
	)
	if rc.selectedEndpoint.IsFallback() {
		// If the selected endpoint is a fallback endpoint, send the relay request to the fallback endpoint.
		// This will bypass protocol-level request processing and validation, meaning the request is not sent to a RelayMiner.
		relayResponse, err = rc.sendFallbackRelay(rc.logger, rc.selectedEndpoint, payload)
	} else {
		// TODO_TECHDEBT(@adshmh): Separate error handling for fallback and Shannon endpoints.
		// If the selected endpoint is not a fallback endpoint, send the relay request to the selected protocolendpoint.
		relayResponse, err = rc.sendRelayWithFallback(payload)
	}

	// Handle endpoint error and capture RelayMinerError data if available.
	if err != nil {
		// Pass the response (which may contain RelayMinerError data) to error handler.
		return rc.handleEndpointError(endpointQueryTime, err)
	}

	// Success:
	// 	 - Record observation
	// 	 - Return response received from endpoint.
	err = rc.handleEndpointSuccess(endpointQueryTime, &relayResponse)
	return relayResponse, err
}

// HandleWebsocketRequest:
//   - Opens a persistent websocket connection to the selected endpoint.
//   - Satisfies gateway.ProtocolRequestContext interface.
func (rc *requestContext) HandleWebsocketRequest(logger polylog.Logger, req *http.Request, w http.ResponseWriter) error {
	if rc.selectedEndpoint == nil {
		return fmt.Errorf("handleWebsocketRequest: no endpoint has been selected on service %s", rc.serviceID)
	}

	wsLogger := logger.With(
		"endpoint_url", rc.selectedEndpoint.PublicURL(),
		"endpoint_addr", rc.selectedEndpoint.Addr(),
		"service_id", rc.serviceID,
	)

	// Upgrade HTTP request from client to websocket connection.
	// - Connection is passed to websocket bridge for Client <-> Gateway communication.
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	clientConn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		wsLogger.Error().Err(err).Msg("Error upgrading websocket connection request")
		return err
	}

	bridge, err := websockets.NewBridge(
		wsLogger,
		clientConn,
		rc.selectedEndpoint,
		rc.relayRequestSigner,
		rc.fullNode,
	)
	if err != nil {
		wsLogger.Error().Err(err).Msg("Error creating websocket bridge")
		return err
	}

	// Run bridge in goroutine to avoid blocking main thread.
	go bridge.Run()

	wsLogger.Info().Msg("websocket connection established")

	return nil
}

// GetObservations:
// - Returns Shannon protocol-level observations for the current request context.
// - Enhanced observations include detailed error classification for metrics generation.
// - Used to:
//   - Update Shannon's endpoint store
//   - Report PATH metrics (metrics package)
//   - Report requests to the data pipeline
//
// - Implements gateway.ProtocolRequestContext interface.
func (rc *requestContext) GetObservations() protocolobservations.Observations {
	return protocolobservations.Observations{
		Shannon: &protocolobservations.ShannonObservationsList{
			Observations: []*protocolobservations.ShannonRequestObservations{
				{
					ServiceId:            string(rc.serviceID),
					RequestError:         rc.requestErrorObservation,
					EndpointObservations: rc.endpointObservations,
				},
			},
		},
	}
}

// buildHeaders creates the headers map including the RPCType header
func buildHeaders(payload protocol.Payload) map[string]string {
	headers := make(map[string]string)

	// Copy existing headers from payload
	maps.Copy(headers, payload.Headers)

	// Set the RPCType HTTP header, if set on the payload.
	// Used by endpoint/relay miner to determine correct backend service.
	if payload.RPCType != sharedtypes.RPCType_UNKNOWN_RPC {
		headers[proxy.RPCTypeHeader] = strconv.Itoa(int(payload.RPCType))
	}

	return headers
}

// sendRelayWithFallback:
// - Attempts Shannon endpoint with timeout
// - Falls back to random fallback endpoint on failure/timeout
// - Shields user from endpoint errors
func (rc *requestContext) sendRelayWithFallback(payload protocol.Payload) (protocol.Response, error) {
	// TODO_TECHDEBT(@adshmh): Replace this with intelligent fallback.

	// Setup Shannon endpoint request:
	// - Create channel for async response
	// - Initialize response variables
	shannonEndpointResponseReceivedChan := make(chan error, 1)
	var (
		shannonEndpointResponse protocol.Response
		shannonEndpointErr      error
	)

	// Send Shannon relay in parallel:
	// - Execute request asynchronously
	// - Signal completion via channel
	go func() {
		shannonEndpointResponse, shannonEndpointErr = rc.sendProtocolRelay(payload)
		shannonEndpointResponseReceivedChan <- shannonEndpointErr
	}()

	logger := rc.logger.With("timeout_ms", maxWaitBeforeFallbackMillisecond)

	// Wait for Shannon response or timeout:
	// - Return Shannon response if successful
	// - Fall back on error or timeout
	select {
	case err := <-shannonEndpointResponseReceivedChan:
		if err == nil {
			return shannonEndpointResponse, nil
		}

		logger.Info().Err(err).Msg("Error getting a valid response from the selected Shannon endpoint. Using a fallback endpoint.")

		// Shannon endpoint failed, use fallback
		return rc.sendRelayToARandomFallbackEndpoint(payload)

	// Shannon endpoint timeout, use fallback
	case <-time.After(time.Duration(maxWaitBeforeFallbackMillisecond) * time.Millisecond):
		logger.Info().Msg("Timed out waiting for Shannon endpoint to respond. Using a fallback endpoint.")

		// Use a random fallback endpoint
		return rc.sendRelayToARandomFallbackEndpoint(payload)
	}
}

// sendRelayToARandomFallbackEndpoint:
// - Selects random fallback endpoint
// - Routes payload via selected endpoint
// - Returns error if no endpoints available
func (rc *requestContext) sendRelayToARandomFallbackEndpoint(payload protocol.Payload) (protocol.Response, error) {
	if len(rc.fallbackEndpoints) == 0 {
		rc.logger.Warn().Msg("SHOULD HAPPEN RARELY: no fallback endpoints available for the service")
		return protocol.Response{}, fmt.Errorf("no fallback endpoints available")
	}

	logger := rc.logger.With("method", "sendRelayToARandomFallbackEndpoint")

	// Select random fallback endpoint:
	// - Convert map to slice for random selection
	// - Pick random index
	allFallbackEndpoints := make([]endpoint, 0, len(rc.fallbackEndpoints))
	for _, endpoint := range rc.fallbackEndpoints {
		allFallbackEndpoints = append(allFallbackEndpoints, endpoint)
	}
	fallbackEndpoint := allFallbackEndpoints[rand.Intn(len(allFallbackEndpoints))]

	// Send relay and handle response:
	// - Use selected fallback endpoint
	// - Log unexpected errors
	relayResponse, err := rc.sendFallbackRelay(logger, fallbackEndpoint, payload)
	if err != nil {
		logger.Warn().Err(err).Msg("SHOULD NEVER HAPPEN: fallback endpoint returned an error.")
	}

	return relayResponse, err
}

// sendProtocolRelay:
//   - Sends the supplied payload as a relay request to the endpoint selected via SelectEndpoint.
//   - Enhanced error handling for more fine-grained endpoint error type classification.
//   - Captures RelayMinerError data for reporting (but doesn't use it for classification).
//   - Required to fulfill the FullNode interface.
func (rc *requestContext) sendProtocolRelay(payload protocol.Payload) (protocol.Response, error) {
	hydratedLogger := rc.getHydratedLogger("sendRelay")
	hydratedLogger = hydrateLoggerWithPayload(hydratedLogger, &payload)

	// Validate endpoint and session
	app, err := rc.validateEndpointAndSession(hydratedLogger)
	if err != nil {
		return protocol.Response{}, err
	}

	// Build and sign the relay request
	signedRelayReq, err := rc.buildAndSignRelayRequest(hydratedLogger, payload, app)
	if err != nil {
		return protocol.Response{}, err
	}

	// Send the HTTP relay request
	// Marshal relay request to bytes
	relayRequestBz, err := signedRelayReq.Marshal()
	if err != nil {
		return protocol.Response{}, fmt.Errorf("SHOULD NEVER HAPPEN: failed to marshal relay request: %w", err)
	}

	// Send the HTTP request to the protocol endpoint.
	httpRelayResponseBz, _, err := rc.sendHTTPRequest(hydratedLogger, payload, rc.selectedEndpoint.PublicURL(), relayRequestBz)
	if err != nil {
		return protocol.Response{
			EndpointAddr: rc.selectedEndpoint.Addr(),
		}, err
	}

	// Validate and process the response
	response, err := rc.validateAndProcessResponse(hydratedLogger, httpRelayResponseBz)
	if err != nil {
		return protocol.Response{
			EndpointAddr: rc.selectedEndpoint.Addr(),
		}, err
	}

	// Deserialize the response
	deserializedResponse, err := rc.deserializeRelayResponse(response)
	if err != nil {
		return protocol.Response{
			EndpointAddr: rc.selectedEndpoint.Addr(),
		}, err
	}

	deserializedResponse.EndpointAddr = rc.selectedEndpoint.Addr()
	return deserializedResponse, nil
}

// validateEndpointAndSession validates that the endpoint and session are properly configured
func (rc *requestContext) validateEndpointAndSession(hydratedLogger polylog.Logger) (apptypes.Application, error) {
	if rc.selectedEndpoint == nil {
		hydratedLogger.Warn().Msg("SHOULD NEVER HAPPEN: No endpoint has been selected. Relay request will fail.")
		return apptypes.Application{}, fmt.Errorf("sendRelay: no endpoint has been selected on service %s", rc.serviceID)
	}

	session := rc.selectedEndpoint.Session()
	if session.Application == nil {
		hydratedLogger.Warn().Msg("SHOULD NEVER HAPPEN: selected endpoint session has nil Application. Relay request will fail.")
		return apptypes.Application{}, fmt.Errorf("sendRelay: nil app on session %s for service %s", session.SessionId, rc.serviceID)
	}

	return *session.Application, nil
}

// buildAndSignRelayRequest builds and signs the relay request
func (rc *requestContext) buildAndSignRelayRequest(
	hydratedLogger polylog.Logger,
	payload protocol.Payload,
	app apptypes.Application,
) (*servicetypes.RelayRequest, error) {
	// Prepare the relay request
	relayRequest, err := buildUnsignedRelayRequest(rc.selectedEndpoint, payload)
	if err != nil {
		hydratedLogger.Warn().Err(err).Msg("SHOULD NEVER HAPPEN: Failed to build the unsigned relay request. Relay request will fail.")
		return nil, err
	}

	// Sign the relay request
	signedRelayReq, err := rc.signRelayRequest(relayRequest, app)
	if err != nil {
		hydratedLogger.Warn().Err(err).Msg("SHOULD NEVER HAPPEN: Failed to sign the relay request. Relay request will fail.")
		return nil, fmt.Errorf("sendRelay: error signing the relay request for app %s: %w", app.Address, err)
	}

	return signedRelayReq, nil
}

// validateAndProcessResponse validates the relay response and tracks relay miner errors
func (rc *requestContext) validateAndProcessResponse(
	hydratedLogger polylog.Logger,
	httpRelayResponseBz []byte,
) (*servicetypes.RelayResponse, error) {
	// Validate the response - check for specific validation errors that indicate raw payload issues
	supplierAddr := sdk.SupplierAddress(rc.selectedEndpoint.Supplier())
	response, err := rc.fullNode.ValidateRelayResponse(supplierAddr, httpRelayResponseBz)

	// Track RelayMinerError data for tracking, regardless of validation result
	// Cross referenced against endpoint payload parse results via metrics
	rc.trackRelayMinerError(response)

	if err != nil {
		// Log raw payload for error tracking
		responseStr := string(httpRelayResponseBz)
		hydratedLogger.With(
			"endpoint_payload", responseStr[:min(len(responseStr), maxEndpointPayloadLenForLogging)],
			"endpoint_payload_length", len(httpRelayResponseBz),
			"validation_error", err.Error(),
		).Warn().Err(err).Msg("Failed to validate the payload from the selected endpoint. Relay request will fail.")

		// Check if this is a validation error that requires raw payload analysis
		if errors.Is(err, sdk.ErrRelayResponseValidationUnmarshal) || errors.Is(err, sdk.ErrRelayResponseValidationBasicValidation) {
			return nil, fmt.Errorf("raw_payload: %s: %w", responseStr, errMalformedEndpointPayload)
		}

		return nil, fmt.Errorf("relay: error verifying the relay response for app %s, endpoint %s: %w",
			rc.selectedEndpoint.Session().Application.Address, rc.selectedEndpoint.PublicURL(), err)
	}

	return response, nil
}

// deserializeRelayResponse deserializes the relay response payload into a protocol.Response
func (rc *requestContext) deserializeRelayResponse(response *servicetypes.RelayResponse) (protocol.Response, error) {
	// The Payload field of the response from the endpoint (relay miner):
	//   - Is a serialized http.Response struct.
	//   - Needs to be deserialized to access the service's response body, status code, etc.
	deserializedResponse, err := deserializeRelayResponse(response.Payload)
	if err != nil {
		// Wrap error with detailed message
		return protocol.Response{}, fmt.Errorf("error deserializing endpoint into a POKTHTTP response: %w", err)
	}

	return deserializedResponse, nil
}

func (rc *requestContext) signRelayRequest(unsignedRelayReq *servicetypes.RelayRequest, app apptypes.Application) (*servicetypes.RelayRequest, error) {
	// Verify the relay request's metadata, specifically the session header.
	// Note: cannot use the RelayRequest's ValidateBasic() method here, as it looks for a signature in the struct, which has not been added yet at this point.
	meta := unsignedRelayReq.GetMeta()

	if meta.GetSessionHeader() == nil {
		return nil, errors.New("signRelayRequest: relay request is missing session header")
	}

	sessionHeader := meta.GetSessionHeader()
	if err := sessionHeader.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("signRelayRequest: relay request session header is invalid: %w", err)
	}

	// Sign the relay request using the selected app's private key
	return rc.relayRequestSigner.SignRelayRequest(unsignedRelayReq, app)
}

// buildUnsignedRelayRequest:
//   - Builds a ready-to-sign RelayRequest using the supplied endpoint, session, and payload.
//   - Returned RelayRequest is meant to be signed and sent to the endpoint to receive its response.
func buildUnsignedRelayRequest(
	endpoint endpoint,
	payload protocol.Payload,
) (*servicetypes.RelayRequest, error) {
	// If path is not empty (e.g. for REST service request), append to endpoint URL.
	url := prepareURLFromPayload(endpoint.PublicURL(), payload)

	// TODO_TECHDEBT: Select the correct underlying request (HTTP, etc.) based on selected service.
	jsonRpcHttpReq, err := shannonJsonRpcHttpRequest(payload, url)
	if err != nil {
		return nil, fmt.Errorf("error building a JSONRPC HTTP request for url %s: %w", url, err)
	}

	relayRequest, err := embedHttpRequest(jsonRpcHttpReq)
	if err != nil {
		return nil, fmt.Errorf("error embedding a JSONRPC HTTP request for url %s: %w", endpoint.PublicURL(), err)
	}

	// TODO_MVP(@adshmh): Use new `FilteredSession` struct from Shannon SDK to get session and endpoint.
	relayRequest.Meta = servicetypes.RelayRequestMetadata{
		SessionHeader:           endpoint.Session().Header,
		SupplierOperatorAddress: endpoint.Supplier(),
	}

	return relayRequest, nil
}

// sendFallbackRelay:
//   - Sends the supplied payload as a relay request to the fallback endpoint.
//   - This will bypass protocol-level request processing and validation,
//     meaning the request is not sent to a RelayMiner.
//   - Used when all endpoints are sanctioned for a service ID.
//   - Returns the response received from the fallback endpoint.
func (rc *requestContext) sendFallbackRelay(
	hydratedLogger polylog.Logger,
	selectedEndpoint endpoint,
	payload protocol.Payload,
) (protocol.Response, error) {
	// Get the fallback URL for the selected endpoint.
	// If the RPC type is unknown or not configured for the
	// service, `endpointFallbackURL` will be the default URL.
	endpointFallbackURL := selectedEndpoint.FallbackURL(payload.RPCType)

	// Prepare the fallback URL with optional path
	fallbackURL := prepareURLFromPayload(endpointFallbackURL, payload)

	// Send the HTTP request to the fallback endpoint.
	httpResponseBz, httpStatusCode, err := rc.sendHTTPRequest(
		hydratedLogger,
		payload,
		fallbackURL,
		[]byte(payload.Data),
	)
	if err != nil {
		return protocol.Response{
			EndpointAddr: selectedEndpoint.Addr(),
		}, err
	}

	// Build and return the fallback response
	return protocol.Response{
		Bytes:          httpResponseBz,
		HTTPStatusCode: httpStatusCode,
		EndpointAddr:   selectedEndpoint.Addr(),
	}, nil
}

// trackRelayMinerError:
//   - Tracks RelayMinerError data from the RelayResponse for reporting purposes.
//   - Updates the requestContext with RelayMinerError data.
//   - Will be included in observations.
//   - Logs RelayMinerError details for visibility.
func (rc *requestContext) trackRelayMinerError(relayResponse *servicetypes.RelayResponse) {
	// Check if RelayResponse contains RelayMinerError data
	if relayResponse == nil || relayResponse.RelayMinerError == nil {
		// No RelayMinerError data to track
		return
	}

	relayMinerErr := relayResponse.RelayMinerError
	hydratedLogger := rc.getHydratedLogger("trackRelayMinerError")

	// Log RelayMinerError details for visibility
	hydratedLogger.With(
		"relay_miner_error_codespace", relayMinerErr.Codespace,
		"relay_miner_error_code", relayMinerErr.Code,
		"relay_miner_error_message", relayMinerErr.Message,
	).Info().Msg("RelayMiner returned an error in RelayResponse (captured for reporting)")

	// Store RelayMinerError data in request context for use in observations
	rc.currentRelayMinerError = &protocolobservations.ShannonRelayMinerError{
		Codespace: relayMinerErr.Codespace,
		Code:      relayMinerErr.Code,
		Message:   relayMinerErr.Message,
	}
}

// handleInternalError:
//   - Called if request processing fails (before sending to any endpoints).
//   - DEV_NOTE: Should NEVER happen; investigate any logged entries from this method.
//   - Records internal error on request for observations.
//   - Logs error entry.
func (rc *requestContext) handleInternalError(internalErr error) (protocol.Response, error) {
	hydratedLogger := rc.getHydratedLogger("handleInternalError")

	// Log the internal error.
	hydratedLogger.Error().Err(internalErr).Msg("Internal error occurred. This should be investigated as a bug.")

	// Set request processing error for generating observations.
	rc.requestErrorObservation = buildInternalRequestProcessingErrorObservation(internalErr)

	return protocol.Response{}, internalErr
}

// handleEndpointError:
//   - Records endpoint error observation with enhanced classification and returns the response.
//   - Tracks endpoint error in observations with detailed categorization for metrics.
//   - Includes any RelayMinerError data that was captured via trackRelayMinerError.
func (rc *requestContext) handleEndpointError(
	endpointQueryTime time.Time,
	endpointErr error,
) (protocol.Response, error) {
	hydratedLogger := rc.getHydratedLogger("handleEndpointError")
	selectedEndpointAddr := rc.selectedEndpoint.Addr()

	// Error classification based on trusted error sources only
	endpointErrorType, recommendedSanctionType := classifyRelayError(hydratedLogger, endpointErr)

	// Enhanced logging with error type and error source classification
	isMalformedPayloadErr := isMalformedEndpointPayloadError(endpointErrorType)
	hydratedLogger.Error().
		Err(endpointErr).
		Str("error_type", endpointErrorType.String()).
		Str("sanction_type", recommendedSanctionType.String()).
		Bool("is_malformed_payload_error", isMalformedPayloadErr).
		Msg("relay error occurred. Service request will fail.")

	// Build enhanced observation with RelayMinerError data from request context
	endpointObs := buildEndpointErrorObservation(
		rc.logger,
		rc.selectedEndpoint,
		endpointQueryTime,
		time.Now(), // Timestamp: endpoint query completed.
		endpointErrorType,
		fmt.Sprintf("relay error: %v", endpointErr),
		recommendedSanctionType,
		rc.currentRelayMinerError, // Use RelayMinerError data from request context
	)

	// Track endpoint error observation for metrics and sanctioning
	rc.endpointObservations = append(rc.endpointObservations, endpointObs)

	// Return error.
	return protocol.Response{EndpointAddr: selectedEndpointAddr},
		fmt.Errorf("relay: error sending relay for service %s endpoint %s: %w",
			rc.serviceID, selectedEndpointAddr, endpointErr,
		)
}

// handleEndpointSuccess:
//   - Records successful endpoint observation and returns the response.
//   - Tracks endpoint success in observations with timing data for performance metrics.
//   - Includes any RelayMinerError data that was captured via trackRelayMinerError.
//   - Builds and returns protocol response from endpoint's returned data.
func (rc *requestContext) handleEndpointSuccess(
	endpointQueryTime time.Time,
	endpointResponse *protocol.Response,
) error {
	hydratedLogger := rc.getHydratedLogger("handleEndpointSuccess")
	hydratedLogger = hydratedLogger.With("endpoint_response_payload_len", len(endpointResponse.Bytes))
	hydratedLogger.Debug().Msg("Successfully deserialized the response received from the selected endpoint.")

	// Build success observation with timing data and any RelayMinerError data from request context
	endpointObs := buildEndpointSuccessObservation(
		rc.logger,
		rc.selectedEndpoint,
		endpointQueryTime,
		time.Now(), // Timestamp: endpoint query completed.
		endpointResponse,
		rc.currentRelayMinerError, // Use RelayMinerError data from request context
	)

	// Track endpoint success observation for metrics
	rc.endpointObservations = append(rc.endpointObservations, endpointObs)

	// Return relay response received from endpoint.
	return nil
}

// sendHTTPRequest is a shared method for sending HTTP requests with common logic
func (rc *requestContext) sendHTTPRequest(
	hydratedLogger polylog.Logger,
	payload protocol.Payload,
	url string,
	requestData []byte,
) ([]byte, int, error) {
	// Prepare a timeout context for the request
	timeout := time.Duration(gateway.RelayRequestTimeout) * time.Millisecond

	// TODO_INVESTIGATE: Evaluate the impact of `rc.context` vs `context.TODO`
	// with respect to handling timeouts.
	ctxWithTimeout, cancelFn := context.WithTimeout(context.TODO(), timeout)
	defer cancelFn()

	// Build headers including RPCType header
	headers := buildHeaders(payload)

	// Send the HTTP request
	httpResponseBz, httpStatusCode, err := rc.httpClient.SendHTTPRelay(
		ctxWithTimeout,
		hydratedLogger,
		url,
		payload.Method,
		requestData,
		headers,
	)

	if err != nil {
		// Endpoint failed to respond before the timeout expires
		// Wrap the net/http error with our classification error
		wrappedErr := fmt.Errorf("%w: %v", errSendHTTPRelay, err)

		hydratedLogger.Error().Err(wrappedErr).Msgf("❌ Failed to receive a response from the selected endpoint: '%s'. Relay request will FAIL 😢", rc.selectedEndpoint.Addr())
		return nil, 0, fmt.Errorf("error sending request to endpoint %s: %w", rc.selectedEndpoint.Addr(), wrappedErr)
	}

	return httpResponseBz, httpStatusCode, nil
}

// prepareURLFromPayload constructs the URL for requests, including optional path.
// Adding the path ensures that REST requests' path is forwarded to the endpoint.
func prepareURLFromPayload(endpointURL string, payload protocol.Payload) string {
	url := endpointURL
	if payload.Path != "" {
		url = fmt.Sprintf("%s%s", url, payload.Path)
	}
	return url
}

// getHydratedLogger:
// - Enhances the base logger with information from the request context.
// - Includes:
//   - Method name
//   - Service ID
//   - Selected endpoint supplier
//   - Selected endpoint URL
func (rc *requestContext) getHydratedLogger(methodName string) polylog.Logger {
	logger := rc.logger.With(
		"method_name", methodName,
		"service_id", rc.serviceID,
	)

	// No endpoint specified on request context.
	// - This should never happen.
	if rc.selectedEndpoint == nil {
		return logger
	}

	logger = logger.With(
		"selected_endpoint_supplier", rc.selectedEndpoint.Supplier(),
		"selected_endpoint_url", rc.selectedEndpoint.PublicURL(),
	)

	sessionHeader := rc.selectedEndpoint.Session().Header
	if sessionHeader == nil {
		return logger
	}

	logger = logger.With(
		"selected_endpoint_app", sessionHeader.ApplicationAddress,
	)

	return logger
}
