package shannon

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/pokt-network/poktroll/pkg/relayer/proxy"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
	sdk "github.com/pokt-network/shannon-sdk"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/observation"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/request"
	"github.com/buildwithgrove/path/websockets"
)

// The requestContext implements the gateway.ProtocolRequestContextWebsocket interface.
// It handles protocol-level WebSocket message processing for both client and endpoint messages.
// For example, client messages are signed and endpoint messages are validated.
var _ gateway.ProtocolRequestContextWebsocket = &websocketRequestContext{}

type websocketRequestContext struct {
	logger polylog.Logger

	// fullNode is used for retrieving onchain data.
	fullNode FullNode

	// serviceID is the service ID for the request.
	serviceID protocol.ServiceID

	// relayRequestSigner is used for signing relay requests.
	relayRequestSigner RelayRequestSigner

	// selectedEndpoint:
	//   - Endpoint selected for sending a relay.
	//   - Must be set via setSelectedEndpoint before sending a relay (otherwise sending fails).
	//   - Protected by selectedEndpointMutex for thread safety.
	selectedEndpoint endpoint
}

// ---------- Websocket Request Context Setup  ----------

// BuildWebsocketRequestContextForEndpoint creates a new WebSocket protocol request context for a specified service and endpoint.
//
// Parameters:
//   - ctx: Context for cancellation, deadlines, and logging.
//   - serviceID: The unique identifier of the target service.
//   - selectedEndpointAddr: The address of the endpoint to use for the request.
//   - httpReq: ONLY used in Delegated mode to extract the selected app from headers.
func (p *Protocol) BuildWebsocketRequestContextForEndpoint(
	ctx context.Context,
	serviceID protocol.ServiceID,
	selectedEndpointAddr protocol.EndpointAddr,
	httpReq *http.Request,
) (gateway.ProtocolRequestContextWebsocket, protocolobservations.Observations, error) {
	logger := p.logger.With(
		"method", "BuildWebsocketRequestContextForEndpoint",
		"service_id", serviceID,
		"endpoint_addr", selectedEndpointAddr,
	)

	activeSessions, err := p.getActiveGatewaySessions(ctx, serviceID, httpReq)
	if err != nil {
		logger.Error().Err(err).Msgf("Relay request will fail due to error retrieving active sessions for service %s", serviceID)
		return nil, buildProtocolContextSetupErrorObservation(serviceID, err), err
	}

	// Retrieve the list of endpoints (i.e. backend service URLs by external operators)
	// that can service RPC requests for the given service ID for the given apps.
	// This includes fallback logic if session endpoints are unavailable.
	// The final boolean parameter sets whether to filter out sanctioned endpoints.
	endpoints, err := p.getUniqueEndpoints(ctx, serviceID, activeSessions, true)
	if err != nil {
		logger.Error().Err(err).Msg(err.Error())
		return nil, buildProtocolContextSetupErrorObservation(serviceID, err), err
	}

	// Select the endpoint that matches the pre-selected address.
	// This ensures QoS checks are performed on the selected endpoint.
	selectedEndpoint, ok := endpoints[selectedEndpointAddr]
	if !ok {
		// Wrap the context setup error.
		// Used to generate the observation.
		err := fmt.Errorf("%w: service %s endpoint %s", errRequestContextSetupInvalidEndpointSelected, serviceID, selectedEndpointAddr)
		logger.Error().Err(err).Msg("Selected endpoint is not available.")
		return nil, buildProtocolContextSetupErrorObservation(serviceID, err), err
	}

	// Retrieve the relay request signer for the current gateway mode.
	permittedSigner, err := p.getGatewayModePermittedRelaySigner(p.gatewayMode)
	if err != nil {
		// Wrap the context setup error.
		// Used to generate the observation.
		err = fmt.Errorf("%w: gateway mode %s: %w", errRequestContextSetupErrSignerSetup, p.gatewayMode, err)
		return nil, buildProtocolContextSetupErrorObservation(serviceID, err), err
	}

	// Return new WebSocket request context for the pre-selected endpoint
	return &websocketRequestContext{
			logger:             logger,
			fullNode:           p.FullNode,
			selectedEndpoint:   selectedEndpoint,
			serviceID:          serviceID,
			relayRequestSigner: permittedSigner,
		},
		// If successful, return an empty observation list.
		// Websocket connection success observations are added when
		// the Bridge is started successfully in `StartWebSocketBridge`.
		protocolobservations.Observations{}, nil
}

// ---------- Connection Establishment ----------

// StartWebSocketBridge creates and starts a WebSocket bridge between client and endpoint.
// It handles all protocol-specific setup including headers, URL generation, and connection establishment.
//
// The messageProcessor handles the actual message processing (typically the gateway's websocketRequestContext).
//
// This method sends establishment observation immediately, blocks until bridge completes, then sends closure observation.
func (wrc *websocketRequestContext) StartWebSocketBridge(
	ctx context.Context,
	httpRequest *http.Request,
	httpResponseWriter http.ResponseWriter,
	messageProcessor websockets.WebsocketMessageProcessor,
	messageObservationsChan chan *observation.RequestResponseObservations,
	establishmentObservationsChan, closureObservationsChan chan *protocolobservations.Observations,
) error {
	wrc.hydratedLogger("StartWebSocketBridge")

	// Get the websocket-specific URL from the selected endpoint.
	websocketEndpointURL, err := wrc.getWebsocketEndpointURL()
	if err != nil {
		// Build error observation for connection failure
		errorObs := wrc.getWebsocketConnectionErrorObservation(err, "selected endpoint does not support websocket RPC type")
		wrc.logger.Error().Err(err).Msg("❌ Selected endpoint does not support websocket RPC type")
		// Send error observation to establishment channel (since connection failed to establish)
		establishmentObservationsChan <- errorObs
		return fmt.Errorf("selected endpoint does not support websocket RPC type: %w", err)
	}
	wrc.logger = wrc.logger.With("websocket_url", websocketEndpointURL)

	// Get the headers for the websocket connection that will be sent to the endpoint.
	endpointConnectionHeaders, err := wrc.getWebsocketConnectionHeaders()
	if err != nil {
		// Build error observation for connection failure
		errorObs := wrc.getWebsocketConnectionErrorObservation(err, "failed to get websocket connection headers")
		wrc.logger.Error().Err(err).Msg("❌ Failed to get websocket connection headers")
		// Send error observation to establishment channel (since connection failed to establish)
		establishmentObservationsChan <- errorObs
		return fmt.Errorf("failed to get websocket connection headers: %w", err)
	}

	// Start the websocket bridge and get a completion channel.
	// The messageProcessor (typically from the gateway layer) handles message processing.
	bridgeCompletionChan, err := websockets.StartBridge(
		ctx,
		wrc.logger,
		httpRequest,
		httpResponseWriter,
		websocketEndpointURL,
		endpointConnectionHeaders,
		messageProcessor,
		messageObservationsChan,
	)
	if err != nil {
		// Build error observation for connection failure
		errorObs := wrc.getWebsocketConnectionErrorObservation(err, "failed to start websocket bridge")
		wrc.logger.Error().Err(err).Msg("Failed to start WebSocket bridge")
		// Send error observation to establishment channel (since connection failed to establish)
		establishmentObservationsChan <- errorObs
		return fmt.Errorf("failed to start websocket bridge: %w", err)
	}

	// Send establishment observation immediately so gateway can broadcast it
	wrc.logger.Info().Msg("✅ WebSocket bridge started successfully, sending establishment observation")
	establishmentObservationsChan <- wrc.getWebsocketConnectionEstablishedObservation()

	// Wait for the bridge to complete (blocks until WebSocket connection terminates)
	<-bridgeCompletionChan

	// Send closure observation so gateway can broadcast it
	wrc.logger.Info().Msg("🔌 WebSocket connection closed, sending closure observation")
	closureObservationsChan <- wrc.getWebsocketConnectionClosedObservation()

	return nil
}

// getWebsocketConnectionHeaders returns headers for the websocket connection:
//   - Target-Service-Id: The service ID of the target service
//   - App-Address: The address of the session's application
//   - Rpc-Type: Always "websocket" for websocket connection requests
func (wrc *websocketRequestContext) getWebsocketConnectionHeaders() (http.Header, error) {
	// Requests to fallback endpoints bypass the protocol so RelayMiner headers are not needed.
	// TODO_IMPROVE(@commoddity,@adshmh): Cleanly separate fallback endpoint handling from the protocol package.
	// TODO_ARCHITECTURE: Extract fallback endpoint handling from protocol package
	// Current: Fallback logic is scattered with if wrc.selectedEndpoint.IsFallback() checks
	// Suggestion: Use strategy pattern or separate fallback handler to cleanly separate concerns
	if wrc.selectedEndpoint.IsFallback() {
		return http.Header{}, nil
	}

	// If the selected endpoint is a protocol endpoint, add the headers
	// that the RelayMiner requires to forward the request to the Endpoint.
	return wrc.getRelayMinerConnectionHeaders()
}

// getRelayMinerConnectionHeaders returns headers for RelayMiner websocket connections.
func (wrc *websocketRequestContext) getRelayMinerConnectionHeaders() (http.Header, error) {
	wrc.hydratedLogger("getRelayMinerConnectionHeaders")

	sessionHeader := wrc.selectedEndpoint.Session().GetHeader()

	if sessionHeader == nil {
		wrc.logger.Error().Msg("❌ SHOULD NEVER HAPPEN: Error getting relay miner connection headers: session header is nil")
		return http.Header{}, fmt.Errorf("session header is nil")
	}

	return http.Header{
		request.HTTPHeaderTargetServiceID: {sessionHeader.ServiceId},
		request.HTTPHeaderAppAddress:      {sessionHeader.ApplicationAddress},
		proxy.RPCTypeHeader:               {strconv.Itoa(int(sharedtypes.RPCType_WEBSOCKET))},
	}, nil
}

// getWebsocketEndpointURL returns the websocket URL for the selected endpoint.
// This URL is used to establish the websocket connection to the endpoint.
func (wrc *websocketRequestContext) getWebsocketEndpointURL() (string, error) {
	wrc.hydratedLogger("getWebsocketEndpointURL")

	websocketURL, err := wrc.selectedEndpoint.WebsocketURL()
	if err != nil {
		wrc.logger.Error().Err(err).Msg("❌ Selected endpoint does not support websocket RPC type")
		return "", err
	}

	return websocketURL, nil
}

// ---------- Client Message Processing ----------

// ProcessProtocolClientWebsocketMessage processes a message from the client.
// Implements gateway.ProtocolRequestContextWebsocket interface.
func (wrc *websocketRequestContext) ProcessProtocolClientWebsocketMessage(msgData []byte) ([]byte, error) {
	wrc.hydratedLogger("ProcessClientWebsocketMessage")

	wrc.logger.Debug().Msgf("received message from client: %s", string(msgData))

	// If the selected endpoint is a fallback endpoint, skip signing the message.
	// Fallback endpoints bypass the protocol so the raw message is sent to the endpoint.
	// TODO_IMPROVE(@commoddity,@adshmh): Cleanly separate fallback endpoint handling from the protocol package.
	if wrc.selectedEndpoint.IsFallback() {
		return msgData, nil
	}

	// If the selected endpoint is a protocol endpoint, we need to sign the message.
	signedRelayRequest, err := wrc.signClientWebsocketMessage(msgData)
	if err != nil {
		wrc.logger.Error().Err(err).Msg("❌ failed to sign request")
		return nil, err
	}

	return signedRelayRequest, nil
}

// signClientWebsocketMessage signs a message from the client using the Relay Request Signer.
func (wrc *websocketRequestContext) signClientWebsocketMessage(msgData []byte) ([]byte, error) {
	wrc.hydratedLogger("signClientWebsocketMessage")

	unsignedRelayRequest := &servicetypes.RelayRequest{
		Meta: servicetypes.RelayRequestMetadata{
			SessionHeader:           wrc.selectedEndpoint.Session().GetHeader(),
			SupplierOperatorAddress: wrc.selectedEndpoint.Supplier(),
		},
		Payload: msgData,
	}

	app := wrc.selectedEndpoint.Session().GetApplication()
	if app == nil {
		wrc.logger.Error().Msg("❌ SHOULD NEVER HAPPEN: session application is nil")
		return nil, fmt.Errorf("session application is nil")
	}

	signedRelayRequest, err := wrc.relayRequestSigner.SignRelayRequest(unsignedRelayRequest, *app)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errRelayRequestWebsocketMessageSigningFailed, err.Error())
	}

	relayRequestBz, err := signedRelayRequest.Marshal()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errRelayRequestWebsocketMessageSigningFailed, err.Error())
	}

	return relayRequestBz, nil
}

// ---------- Endpoint Message Processing ----------

// ProcessProtocolEndpointWebsocketMessage processes a message from the endpoint.
func (wrc *websocketRequestContext) ProcessProtocolEndpointWebsocketMessage(
	msgData []byte,
) ([]byte, protocolobservations.Observations, error) {
	wrc.hydratedLogger("ProcessEndpointWebsocketMessage")

	wrc.logger.Debug().Msgf("received message from endpoint: %s", string(msgData))

	// If the selected endpoint is a fallback endpoint, skip validation.
	// Fallback endpoints bypass the protocol so the raw message is sent to the endpoint.
	// TODO_IMPROVE(@commoddity,@adshmh): Cleanly separate fallback endpoint handling from the protocol package.
	if wrc.selectedEndpoint.IsFallback() {
		return msgData, wrc.getWebsocketMessageSuccessObservation(msgData), nil
	}

	// If the selected endpoint is a protocol endpoint, we need to validate the message.
	validatedRelayResponse, err := wrc.validateEndpointWebsocketMessage(msgData)
	if err != nil {
		wrc.logger.Error().Err(err).Msg("❌ failed to validate relay response")
		return nil, wrc.getWebsocketMessageErrorObservation(msgData, err), err
	}

	return validatedRelayResponse, wrc.getWebsocketMessageSuccessObservation(msgData), nil
}

// validateEndpointWebsocketMessage validates a message from the endpoint using the Shannon FullNode.
// TODO_IMPROVE(@adshmh): Compare this to 'validateAndProcessResponse' and align the two implementations
// w.r.t design, error handling, etc...
func (wrc *websocketRequestContext) validateEndpointWebsocketMessage(msgData []byte) ([]byte, error) {
	wrc.hydratedLogger("validateEndpointWebsocketMessage")

	// Validate the relay response using the Shannon FullNode
	relayResponse, err := wrc.fullNode.ValidateRelayResponse(sdk.SupplierAddress(wrc.selectedEndpoint.Supplier()), msgData)
	if err != nil {
		wrc.logger.Error().Err(err).Msg("❌ failed to validate relay response in websocket message")
		return nil, fmt.Errorf("%w: %s", errRelayResponseInWebsocketMessageValidationFailed, err.Error())
	}
	wrc.logger.Debug().Msgf("received message from protocol endpoint: %s", string(relayResponse.Payload))

	return relayResponse.Payload, nil
}

// ---------- Message-Level Observations ----------

// getWebsocketMessageSuccessObservation updates the observations for the current message
// if the message handler does not return an error.
func (wrc *websocketRequestContext) getWebsocketMessageSuccessObservation(
	msgData []byte,
) protocolobservations.Observations {
	wrc.hydratedLogger("getWebsocketMessageSuccessObservation")

	// Create a new WebSocket message observation for success
	wsMessageObs := buildWebsocketMessageSuccessObservation(
		wrc.logger,
		wrc.selectedEndpoint,
		int64(len(msgData)),
	)

	// Update the observations to use the WebSocket message observation
	return protocolobservations.Observations{
		Shannon: &protocolobservations.ShannonObservationsList{
			Observations: []*protocolobservations.ShannonRequestObservations{
				{
					ServiceId:    string(wrc.serviceID),
					RequestError: nil, // WS messages do not have request errors
					ObservationData: &protocolobservations.ShannonRequestObservations_WebsocketMessageObservation{
						WebsocketMessageObservation: wsMessageObs,
					},
				},
			},
		},
	}
}

// getWebsocketMessageErrorObservation updates the observations for the current message
// if the message handler returns an error.
func (wrc *websocketRequestContext) getWebsocketMessageErrorObservation(
	msgData []byte,
	messageError error,
) protocolobservations.Observations {
	// Error classification based on trusted error sources only
	endpointErrorType, recommendedSanctionType := classifyRelayError(wrc.logger, messageError)

	// Create a new WebSocket message observation for error
	wsMessageObs := buildWebsocketMessageErrorObservation(
		wrc.selectedEndpoint,
		int64(len(msgData)),
		endpointErrorType,
		fmt.Sprintf("websocket message error: %v", messageError),
		recommendedSanctionType,
	)

	return protocolobservations.Observations{
		Shannon: &protocolobservations.ShannonObservationsList{
			Observations: []*protocolobservations.ShannonRequestObservations{
				{
					ServiceId:    string(wrc.serviceID),
					RequestError: nil, // WS messages do not have request errors
					ObservationData: &protocolobservations.ShannonRequestObservations_WebsocketMessageObservation{
						WebsocketMessageObservation: wsMessageObs,
					},
				},
			},
		},
	}
}

// ---------- Connection-Level Observations ----------

// getWebsocketConnectionSuccessObservation builds observations for successful WebSocket connection establishment.
func (wrc *websocketRequestContext) getWebsocketConnectionEstablishedObservation() *protocolobservations.Observations {
	return &protocolobservations.Observations{
		Shannon: &protocolobservations.ShannonObservationsList{
			Observations: []*protocolobservations.ShannonRequestObservations{
				{
					ServiceId: string(wrc.serviceID),
					ObservationData: &protocolobservations.ShannonRequestObservations_WebsocketConnectionObservation{
						WebsocketConnectionObservation: buildWebsocketConnectionObservation(
							wrc.logger,
							wrc.selectedEndpoint,
							protocolobservations.ShannonWebsocketConnectionObservation_CONNECTION_ESTABLISHED,
						),
					},
				},
			},
		},
	}
}

func (wrc *websocketRequestContext) getWebsocketConnectionClosedObservation() *protocolobservations.Observations {
	return &protocolobservations.Observations{
		Shannon: &protocolobservations.ShannonObservationsList{
			Observations: []*protocolobservations.ShannonRequestObservations{
				{
					ServiceId: string(wrc.serviceID),
					ObservationData: &protocolobservations.ShannonRequestObservations_WebsocketConnectionObservation{
						WebsocketConnectionObservation: buildWebsocketConnectionObservation(
							wrc.logger,
							wrc.selectedEndpoint,
							protocolobservations.ShannonWebsocketConnectionObservation_CONNECTION_CLOSED,
						),
					},
				},
			},
		},
	}
}

// getWebsocketConnectionErrorObservation builds observations for failed WebSocket connection establishment.
func (wrc *websocketRequestContext) getWebsocketConnectionErrorObservation(
	err error,
	details string,
) *protocolobservations.Observations {
	endpointErrorType, recommendedSanctionType := classifyRelayError(wrc.logger, err)

	errorDetails := fmt.Sprintf("websocket connection error: %s: %v", details, err)

	return &protocolobservations.Observations{
		Shannon: &protocolobservations.ShannonObservationsList{
			Observations: []*protocolobservations.ShannonRequestObservations{
				{
					ServiceId: string(wrc.serviceID),
					RequestError: &protocolobservations.ShannonRequestError{
						ErrorType:    protocolobservations.ShannonRequestErrorType_SHANNON_REQUEST_ERROR_INTERNAL,
						ErrorDetails: fmt.Sprintf("websocket connection error: %s: %v", details, err),
					},
					ObservationData: &protocolobservations.ShannonRequestObservations_WebsocketConnectionObservation{
						WebsocketConnectionObservation: buildWebsocketConnectionErrorObservation(
							wrc.logger,
							wrc.selectedEndpoint,
							endpointErrorType,
							errorDetails,
							recommendedSanctionType,
							protocolobservations.ShannonWebsocketConnectionObservation_CONNECTION_ESTABLISHMENT_FAILED,
						),
					},
				},
			},
		},
	}
}

// ---------- Logger Helpers ----------

// hydratedLogger:
// - Enhances the base logger with information from the request context.
// - Includes:
//   - Method name
//   - Service ID
//   - Selected endpoint supplier
//   - Selected endpoint URL
func (wrc *websocketRequestContext) hydratedLogger(methodName string) {
	logger := wrc.logger.With(
		"request_type", "websocket",
		"method", methodName,
		"service_id", wrc.serviceID,
	)

	defer func() {
		wrc.logger = logger
	}()

	// No endpoint specified on request context.
	// - This should never happen.
	selectedEndpoint := wrc.selectedEndpoint
	if selectedEndpoint == nil {
		return
	}

	logger = logger.With(
		"selected_endpoint_supplier", selectedEndpoint.Supplier(),
		"selected_endpoint_url", selectedEndpoint.PublicURL(),
	)

	sessionHeader := selectedEndpoint.Session().Header
	if sessionHeader == nil {
		return
	}

	logger = logger.With(
		"selected_endpoint_app", sessionHeader.ApplicationAddress,
	)
}
