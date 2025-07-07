package shannon

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"

	"github.com/buildwithgrove/path/gateway"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/websockets"
)

// Maximum length of the endpoint payload logged on error.
const maxEndpointPayloadLenForLogging = 100

// requestContext provides all the functionality required by the gateway package
// for handling a single service request.
var _ gateway.ProtocolRequestContext = &requestContext{}

// RelayRequestSigner:
// - Used by requestContext to sign relay requests.
// - Takes an unsigned relay request and an application.
// - Returns a relay request signed by the gateway (with delegation from the app).
// - In future Permissionless Gateway Mode, may use the app's own private key for signing.
type RelayRequestSigner interface {
	SignRelayRequest(req *servicetypes.RelayRequest, app apptypes.Application) (*servicetypes.RelayRequest, error)
}

// requestContext:
// - Captures all data required for handling a single service request.
type requestContext struct {
	logger polylog.Logger

	// context:
	// - Upstream context for proper timeout propagation and cancellation.
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
	endpointObservations []*protocolobservations.ShannonEndpointObservation
}

// HandleServiceRequest:
// - Satisfies gateway.ProtocolRequestContext interface.
// - Uses supplied payload to send a relay request to an endpoint.
// - Verifies and returns the response.
func (rc *requestContext) HandleServiceRequest(payload protocol.Payload) (protocol.Response, error) {
	// Internal error: No endpoint selected.
	// - Record request error due to internal error.
	// - No endpoint to sanction.
	if rc.selectedEndpoint == nil {
		return rc.handleInternalError(fmt.Errorf("HandleServiceRequest: no endpoint has been selected on service %s", rc.serviceID))
	}

	// Record endpoint query time.
	endpointQueryTime := time.Now()

	// Send the relay request.
	response, err := rc.sendRelay(payload)

	// Handle endpoint error:
	// - Record observation
	// - Return error
	if err != nil {
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
	rc.handleEndpointSuccess(endpointQueryTime, &relayResponse)
	return relayResponse, nil
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

// sendRelay:
// - Sends the supplied payload as a relay request to the endpoint selected via SelectEndpoint.
// - Required to fulfill the FullNode interface.
func (rc *requestContext) sendRelay(payload protocol.Payload) (*servicetypes.RelayResponse, error) {
	hydratedLogger := rc.getHydratedLogger("sendRelay").With("method", "sendRelay")
	hydratedLogger = hydrateLoggerWithPayload(hydratedLogger, &payload)

	// TODO_MVP(@adshmh): enhance Shannon metrics, e.g. request error kind, to capture all potential errors via metrics.
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

	payloadBz := []byte(payload.Data)
	relayRequest, err := buildUnsignedRelayRequest(*rc.selectedEndpoint, session, payloadBz, payload.Path)
	if err != nil {
		hydratedLogger.Warn().Err(err).Msg("SHOULD NEVER HAPPEN: Failed to build the unsigned relay request. Relay request will fail.")
		return nil, err
	}

	signedRelayReq, err := rc.signRelayRequest(relayRequest, app)
	if err != nil {
		hydratedLogger.Warn().Err(err).Msg("SHOULD NEVER HAPPEN: Failed to sign the relay request. Relay request will fail.")
		return nil, fmt.Errorf("sendRelay: error signing the relay request for app %s: %w", app.Address, err)
	}

	timeout := time.Duration(payload.TimeoutMillisec) * time.Millisecond
	ctxWithTimeout, cancelFn := context.WithTimeout(rc.context, timeout)
	defer cancelFn()

	// TODO_MVP(@adshmh): Check the HTTP status code returned by the endpoint.
	responseBz, err := sendHttpRelay(ctxWithTimeout, rc.selectedEndpoint.url, signedRelayReq, timeout)
	if err != nil {
		// endpoint failed to respond before the timeout expires.
		hydratedLogger.Error().Err(err).Msgf("❌ Failed to receive a response from the selected endpoint: '%s'. Relay request will FAIL 😢", rc.selectedEndpoint.Addr())
		return nil, fmt.Errorf("error sending request to endpoint %s: %w", rc.selectedEndpoint.Addr(), err)
	}

	// Validate the response.
	response, err := rc.fullNode.ValidateRelayResponse(sdk.SupplierAddress(rc.selectedEndpoint.supplier), responseBz)
	if err != nil {
		// TODO_TECHDEBT(@adshmh): Complete the following steps to track endpoint errors and sanction as needed:
		// 1. Enhance the `RelayResponse` struct with an error field:
		// 	https://github.com/pokt-network/poktroll/blob/2ba8b60d6bd8d21949211844161f932dd383bb76/proto/pocket/service/relay.proto#L46
		// 2. Update the classifyRelayError function to sanction endpoints depending on the error.
		// 3. Enhance the Shannon metrics: proto/path/protocol/shannon.proto, specifically the RequestErrorType enum, to track the errors.
		// 4. Update the files in `metrics.protocol.shannon` package to add/update metrics according to the above.
		//
		// Log raw payload for error tracking:
		// - RelayResponse lacks error field (see TODO above)
		// - RelayMiner returns generic HTTP on errors (expired sessions, etc.)
		// - Enables error analysis via PATH logs
		responseStr := string(responseBz)
		responseStrForLogging := responseStr[:min(len(responseStr), maxEndpointPayloadLenForLogging)]
		hydratedLogger.With("endpoint_payload", responseStrForLogging).Warn().Err(err).Msg("Failed to validate the payload from the selected endpoint. Relay request will fail.")
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
	payload []byte,
	path string,
) (*servicetypes.RelayRequest, error) {
	// If path is not empty (e.g. for REST service request), append to endpoint URL.
	url := endpoint.url
	if path != "" {
		url = fmt.Sprintf("%s%s", url, path)
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
// - Records endpoint error observation and returns the response.
// - Tracks endpoint error in observations.
// - Builds and returns protocol response from endpoint's returned data.
func (rc *requestContext) handleEndpointError(
	endpointQueryTime time.Time,
	endpointErr error,
) (protocol.Response, error) {
	hydratedLogger := rc.getHydratedLogger("handleEndpointError")
	selectedEndpointAddr := rc.selectedEndpoint.Addr()

	// Classify endpoint error for observation.
	// Determine any applicable sanctions.
	endpointErrorType, recommendedSanctionType := classifyRelayError(hydratedLogger, endpointErr)

	// Log endpoint error.
	hydratedLogger.Error().
		Err(endpointErr).
		Str("error_type", endpointErrorType.String()).
		Str("sanction_type", recommendedSanctionType.String()).
		Msg("relay error occurred. Service request will fail.")

	// Track endpoint error observation.
	rc.endpointObservations = append(rc.endpointObservations,
		buildEndpointErrorObservation(
			rc.logger,
			*rc.selectedEndpoint,
			endpointQueryTime,
			time.Now(), // Timestamp: endpoint query completed.
			endpointErrorType,
			fmt.Sprintf("relay error: %v", endpointErr),
			recommendedSanctionType,
		),
	)

	// Return error.
	return protocol.Response{EndpointAddr: selectedEndpointAddr},
		fmt.Errorf("relay: error sending relay for service %s endpoint %s: %w",
			rc.serviceID, selectedEndpointAddr, endpointErr,
		)
}

// handleEndpointSuccess:
// - Records successful endpoint observation and returns the response.
// - Tracks endpoint success in observations.
// - Builds and returns protocol response from endpoint's returned data.
func (rc *requestContext) handleEndpointSuccess(
	endpointQueryTime time.Time,
	endpointResponse *protocol.Response) {
	hydratedLogger := rc.getHydratedLogger("handleEndpointSuccess")
	hydratedLogger = hydratedLogger.With("endpoint_response_payload_len", len(endpointResponse.Bytes))
	hydratedLogger.Debug().Msg("Successfully deserialized the response received from the selected endpoint.")

	// Track endpoint success observation.
	rc.endpointObservations = append(rc.endpointObservations,
		buildEndpointSuccessObservation(
			rc.logger,
			*rc.selectedEndpoint,
			endpointQueryTime,
			time.Now(), // Timestamp: endpoint query completed.
			endpointResponse,
		),
	)
}
