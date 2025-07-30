package shannon

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/pokt-network/poktroll/pkg/relayer/proxy"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
	sdk "github.com/pokt-network/shannon-sdk"

	"github.com/buildwithgrove/path/gateway"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/websockets"
)

// Maximum endpoint payload length for error logging (100 chars)
const maxEndpointPayloadLenForLogging = 100

// Minimum number of payloads to trigger batch processing
// For fewer payloads, single request processing is used for better performance
const minPayloadsForBatchProcessing = 2

// requestContext provides all the functionality required by the gateway package
// for handling a single service request.
var _ gateway.ProtocolRequestContext = &requestContext{}

// RelayRequestSigner:
// - Used by requestContext to sign relay requests
// - Takes an unsigned relay request and an application
// - Returns a relay request signed by the gateway (with delegation from the app)
// - In future Permissionless Gateway Mode, may use the app's own private key for signing
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
	// - Endpoint selected for sending a relay.
	// - Must be set via SelectEndpoint before sending a relay (otherwise sending fails).
	selectedEndpoint *endpoint

	// requestErrorObservation:
	// - Tracks any errors encountered during request processing.
	requestErrorObservation *protocolobservations.ShannonRequestError

	// endpointObservations:
	// - Captures observations about endpoints used during request handling.
	// - Includes enhanced error classification for raw payload analysis.
	endpointObservations []*protocolobservations.ShannonEndpointObservation

	// currentRelayMinerError:
	// - Tracks RelayMinerError data from the current relay response for reporting.
	// - Set by trackRelayMinerError method and used when building observations.
	currentRelayMinerError *protocolobservations.ShannonRelayMinerError

	// HTTP client used for sending relay requests to endpoints while also capturing various debug metrics
	httpClient *httpClientWithDebugMetrics
}

// HandleServiceRequest:
// - Satisfies gateway.ProtocolRequestContext interface.
// - Uses supplied payloads to send relay requests to an endpoint.
// - Handles both single requests and JSON-RPC batch requests concurrently when beneficial.
// - Returns responses as an array to match interface, but gateway currently expects single response.
// - Captures RelayMinerError data when available for reporting purposes.
func (rc *requestContext) HandleServiceRequest(payloads []protocol.Payload) ([]protocol.Response, error) {
	// Internal error: No endpoint selected.
	if rc.selectedEndpoint == nil {
		response, err := rc.handleInternalError(fmt.Errorf("HandleServiceRequest: no endpoint has been selected on service %s", rc.serviceID))
		return []protocol.Response{response}, err
	}

	// Handle empty payloads.
	if len(payloads) == 0 {
		response, err := rc.handleInternalError(fmt.Errorf("HandleServiceRequest: no payloads provided for service %s", rc.serviceID))
		return []protocol.Response{response}, err
	}

	// For single payload, handle directly without additional overhead.
	if len(payloads) == 1 {
		response, err := rc.sendSingleRelay(payloads[0])
		return []protocol.Response{response}, err
	}

	// For multiple payloads, decide whether to use batch processing.
	// Use batch processing only when we have enough payloads to benefit from concurrency.
	if len(payloads) >= minPayloadsForBatchProcessing {
		return rc.sendBatchRelays(payloads)
	}

	// For fewer payloads, process sequentially to avoid goroutine overhead.
	return rc.sendSequentialRelays(payloads)
}

// sendBatchRelays:
// - Sends multiple relay requests concurrently to the selected endpoint.
// - Handles JSON-RPC batch requests and returns individual responses.
// - Uses goroutines for concurrent processing while maintaining response order.
// - Returns all responses as separate entries in the response array.
func (rc *requestContext) sendBatchRelays(payloads []protocol.Payload) ([]protocol.Response, error) {

	// Create channels for collecting results.
	resultChan := make(chan relayResult, len(payloads))
	var wg sync.WaitGroup

	// Record batch start time.
	batchStartTime := time.Now()

	// Launch goroutines for each payload.
	for i, payload := range payloads {
		wg.Add(1)
		go func(index int, p protocol.Payload) {
			defer wg.Done()

			response, err := rc.sendSingleRelay(p)

			resultChan <- relayResult{
				index:    index,
				response: response,
				err:      err,
			}
		}(i, payload)
	}

	// Wait for all goroutines to complete.
	wg.Wait()
	close(resultChan)

	// Collect results in order.
	results := make([]relayResult, len(payloads))
	for result := range resultChan {
		results[result.index] = result
	}

	// Log batch completion for debugging.
	rc.logger.Debug().
		Int("batch_size", len(payloads)).
		Dur("batch_duration", time.Since(batchStartTime)).
		Msg("Concurrent batch relay processing completed")

	// Convert results to response array and return first error if any.
	return rc.convertResultsToResponses(results)
}

// sendSequentialRelays:
// - Sends multiple relay requests sequentially to the selected endpoint.
// - Used for smaller batch sizes where goroutine overhead exceeds benefits.
// - Returns individual responses as separate entries in the response array.
func (rc *requestContext) sendSequentialRelays(payloads []protocol.Payload) ([]protocol.Response, error) {
	// Record batch start time.
	batchStartTime := time.Now()

	// Process payloads sequentially.
	results := make([]relayResult, len(payloads))
	for i, payload := range payloads {
		response, err := rc.sendSingleRelay(payload)
		results[i] = relayResult{
			index:    i,
			response: response,
			err:      err,
		}
	}

	// Log batch completion for debugging.
	rc.logger.Debug().
		Int("batch_size", len(payloads)).
		Dur("batch_duration", time.Since(batchStartTime)).
		Msg("Sequential batch relay processing completed")

	// Convert results to response array and return first error if any.
	return rc.convertResultsToResponses(results)
}

// relayResult holds the result of a single relay request for batch processing.
type relayResult struct {
	index    int
	response protocol.Response
	err      error
}

// sendSingleRelay:
// - Handles a single relay request with full error handling and observation tracking.
// - Extracted from original HandleServiceRequest logic for reuse in batch processing.
func (rc *requestContext) sendSingleRelay(payload protocol.Payload) (protocol.Response, error) {
	// Record endpoint query time.
	endpointQueryTime := time.Now()

	// Send the relay request.
	response, err := rc.sendRelay(payload)

	// Handle endpoint error and capture RelayMinerError data if available
	if err != nil {
		// Pass the response (which may contain RelayMinerError data) to error handler
		return rc.handleEndpointError(endpointQueryTime, err)
	}

	// The Payload field of the response from the endpoint (relay miner):
	// - Is a serialized http.Response struct.
	// - Needs to be deserialized to access the service's response body, status code, etc.
	relayResponse, err := deserializeRelayResponse(response.Payload)
	relayResponse.EndpointAddr = rc.selectedEndpoint.Addr()
	if err != nil {
		// Wrap error with detailed message.
		deserializeErr := fmt.Errorf("error deserializing endpoint into a POKTHTTP response: %w", err)
		return rc.handleEndpointError(endpointQueryTime, deserializeErr)
	}

	// Success:
	// - Record observation
	// - Return response received from endpoint.
	err = rc.handleEndpointSuccess(endpointQueryTime, &relayResponse)
	return relayResponse, err
}

// convertResultsToResponses:
// - Converts relay results into an array of protocol responses.
// - Maintains the order of responses to match the order of input payloads.
// - Returns the first error encountered, or nil if all succeeded.
func (rc *requestContext) convertResultsToResponses(results []relayResult) ([]protocol.Response, error) {
	if len(results) == 0 {
		response, err := rc.handleInternalError(fmt.Errorf("convertResultsToResponses: no results to convert"))
		return []protocol.Response{response}, err
	}

	// Create response array in the same order as input payloads.
	responses := make([]protocol.Response, len(results))
	var firstErr error

	// Process results in order.
	for i, result := range results {
		responses[i] = result.response

		// Track the first error encountered.
		if result.err != nil && firstErr == nil {
			firstErr = result.err
		}
	}

	rc.logger.Debug().
		Int("num_responses", len(responses)).
		Bool("has_errors", firstErr != nil).
		Msg("Response conversion completed")

	return responses, firstErr
}

// HandleWebsocketRequest:
// - Opens a persistent websocket connection to the selected endpoint.
// - Satisfies gateway.ProtocolRequestContext interface.
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
		Protocol: &protocolobservations.Observations_Shannon{
			Shannon: &protocolobservations.ShannonObservationsList{
				Observations: []*protocolobservations.ShannonRequestObservations{
					{
						ServiceId:            string(rc.serviceID),
						RequestError:         rc.requestErrorObservation,
						EndpointObservations: rc.endpointObservations,
					},
				},
			},
		},
	}
}

// buildHeaders creates the headers map including the RPCType header
func buildHeaders(payload protocol.Payload) map[string]string {
	headers := make(map[string]string)

	// Copy existing headers from payload
	for key, value := range payload.Headers {
		headers[key] = value
	}

	// Set the RPCType HTTP header, if set on the payload.
	// Used by endpoint/relay miner to determine correct backend service.
	if payload.RPCType != sharedtypes.RPCType_UNKNOWN_RPC {
		headers[proxy.RPCTypeHeader] = strconv.Itoa(int(payload.RPCType))
	}

	return headers
}

// sendRelay:
// - Sends the supplied payload as a relay request to the endpoint selected via SelectEndpoint.
// - Enhanced error handling for more fine-grained endpoint error type classification.
// - Captures RelayMinerError data for reporting (but doesn't use it for classification).
// - Required to fulfill the FullNode interface.
func (rc *requestContext) sendRelay(payload protocol.Payload) (*servicetypes.RelayResponse, error) {
	hydratedLogger := rc.getHydratedLogger("sendRelay")
	hydratedLogger = hydrateLoggerWithPayload(hydratedLogger, &payload)

	if rc.selectedEndpoint == nil {
		hydratedLogger.Warn().Msg("SHOULD NEVER HAPPEN: No endpoint has been selected. Relay request will fail.")
		return nil, fmt.Errorf("sendRelay: no endpoint has been selected on service %s", rc.serviceID)
	}

	// Hydrate the logger with endpoint/session details.
	hydratedLogger = hydrateLoggerWithEndpoint(hydratedLogger, rc.selectedEndpoint)

	session := rc.selectedEndpoint.session
	if session.Application == nil {
		hydratedLogger.Warn().Msg("SHOULD NEVER HAPPEN: selected endpoint session has nil Application. Relay request will fail.")
		return nil, fmt.Errorf("sendRelay: nil app on session %s for service %s", session.SessionId, rc.serviceID)
	}
	app := *session.Application

	// Prepare and sign the relay request.
	relayRequest, err := buildUnsignedRelayRequest(*rc.selectedEndpoint, session, payload)
	if err != nil {
		hydratedLogger.Warn().Err(err).Msg("SHOULD NEVER HAPPEN: Failed to build the unsigned relay request. Relay request will fail.")
		return nil, err
	}
	signedRelayReq, err := rc.signRelayRequest(relayRequest, app)
	if err != nil {
		hydratedLogger.Warn().Err(err).Msg("SHOULD NEVER HAPPEN: Failed to sign the relay request. Relay request will fail.")
		return nil, fmt.Errorf("sendRelay: error signing the relay request for app %s: %w", app.Address, err)
	}

	// Prepare a timeout context for the relay request.
	timeout := time.Duration(payload.TimeoutMillisec) * time.Millisecond
	ctxWithTimeout, cancelFn := context.WithTimeout(rc.context, timeout)
	defer cancelFn()

	// Build headers including RPCType header
	headers := buildHeaders(payload)

	// Send the HTTP relay request
	httpRelayResponseBz, err := rc.httpClient.SendHTTPRelay(
		ctxWithTimeout,
		hydratedLogger,
		rc.selectedEndpoint.url,
		signedRelayReq,
		headers,
	)

	if err != nil {
		// Endpoint failed to respond before the timeout expires.
		// Wrap the net/http error with our classification error
		wrappedErr := fmt.Errorf("%w: %v", errSendHTTPRelay, err)

		hydratedLogger.Error().Err(wrappedErr).Msgf("❌ Failed to receive a response from the selected endpoint: '%s'. Relay request will FAIL 😢", rc.selectedEndpoint.Addr())
		return nil, fmt.Errorf("error sending request to endpoint %s: %w", rc.selectedEndpoint.Addr(), wrappedErr)
	}

	// Validate the response - check for specific validation errors that indicate raw payload issues
	supplierAddr := sdk.SupplierAddress(rc.selectedEndpoint.supplier)
	response, err := rc.fullNode.ValidateRelayResponse(supplierAddr, httpRelayResponseBz)

	// Track RelayMinerError data for tracking, regardless of validation result.
	// Cross referenced against endpoint payload parse results via metrics.
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

		return nil, fmt.Errorf("relay: error verifying the relay response for app %s, endpoint %s: %w", app.Address, rc.selectedEndpoint.url, err)
	}

	return response, nil
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
// - Builds a ready-to-sign RelayRequest using the supplied endpoint, session, and payload.
// - Returned RelayRequest is meant to be signed and sent to the endpoint to receive its response.
func buildUnsignedRelayRequest(
	endpoint endpoint,
	session sessiontypes.Session,
	payload protocol.Payload,
) (*servicetypes.RelayRequest, error) {
	// If path is not empty (e.g. for REST service request), append to endpoint URL.
	url := endpoint.url
	if payload.Path != "" {
		url = fmt.Sprintf("%s%s", url, payload.Path)
	}

	// TODO_TECHDEBT: Select the correct underlying request (HTTP, etc.) based on selected service.
	jsonRpcHttpReq, err := shannonJsonRpcHttpRequest(payload, url)
	if err != nil {
		return nil, fmt.Errorf("error building a JSONRPC HTTP request for url %s: %w", url, err)
	}

	relayRequest, err := embedHttpRequest(jsonRpcHttpReq)
	if err != nil {
		return nil, fmt.Errorf("error embedding a JSONRPC HTTP request for url %s: %w", endpoint.url, err)
	}

	// TODO_MVP(@adshmh): Use new `FilteredSession` struct from Shannon SDK to get session and endpoint.
	relayRequest.Meta = servicetypes.RelayRequestMetadata{
		SessionHeader:           session.Header,
		SupplierOperatorAddress: string(endpoint.supplier),
	}

	return relayRequest, nil
}

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
		"selected_endpoint_supplier", rc.selectedEndpoint.supplier,
		"selected_endpoint_url", rc.selectedEndpoint.url,
	)

	sessionHeader := rc.selectedEndpoint.session.GetHeader()
	if sessionHeader == nil {
		return logger
	}

	logger = logger.With(
		"selected_endpoint_app", sessionHeader.ApplicationAddress,
	)

	return logger
}

// trackRelayMinerError:
// - Tracks RelayMinerError data from the RelayResponse for reporting purposes.
// - Updates the requestContext with RelayMinerError data.
// - Will be included in observations.
// - Logs RelayMinerError details for visibility.
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
// - Called if request processing fails (before sending to any endpoints).
// - DEV_NOTE: Should NEVER happen; investigate any logged entries from this method.
// - Records internal error on request for observations.
// - Logs error entry.
func (rc *requestContext) handleInternalError(internalErr error) (protocol.Response, error) {
	hydratedLogger := rc.getHydratedLogger("handleInternalError")

	// Log the internal error.
	hydratedLogger.Error().Err(internalErr).Msg("Internal error occurred. This should be investigated as a bug.")

	// Set request processing error for generating observations.
	rc.requestErrorObservation = buildInternalRequestProcessingErrorObservation(internalErr)

	return protocol.Response{}, internalErr
}

// handleEndpointError:
// - Records endpoint error observation with enhanced classification and returns the response.
// - Tracks endpoint error in observations with detailed categorization for metrics.
// - Includes any RelayMinerError data that was captured via trackRelayMinerError.
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
		*rc.selectedEndpoint,
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
// - Records successful endpoint observation and returns the response.
// - Tracks endpoint success in observations with timing data for performance metrics.
// - Includes any RelayMinerError data that was captured via trackRelayMinerError.
// - Builds and returns protocol response from endpoint's returned data.
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
		*rc.selectedEndpoint,
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
