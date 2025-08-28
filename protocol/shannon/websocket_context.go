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

	// Upstream context for timeout propagation and cancellation
	context context.Context

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

	// requestErrorObservation:
	//   - Tracks any errors encountered during request processing.
	requestErrorObservation *protocolobservations.ShannonRequestError

	// fallbackEndpoints is used to retrieve a fallback endpoint by an endpoint address.
	fallbackEndpoints map[protocol.EndpointAddr]endpoint
}

// ---------- Connection Establishment ----------

// StartWebSocketBridge creates and starts a WebSocket bridge between client and endpoint.
// It handles all protocol-specific setup including headers, URL generation, and connection establishment.
// The messageProcessor handles the actual message processing (typically the gateway's websocketRequestContext).
// Returns a completion channel that signals when the bridge shuts down and observations for the connection.
func (wrc *websocketRequestContext) StartWebSocketBridge(
	ctx context.Context,
	httpRequest *http.Request,
	httpResponseWriter http.ResponseWriter,
	messageProcessor websockets.WebsocketMessageProcessor,
	messageObservationsChan chan *observation.RequestResponseObservations,
) (<-chan struct{}, *protocolobservations.Observations, error) {
	logger := wrc.logger.With("method", "StartWebSocketBridge")

	// Get the websocket-specific URL from the selected endpoint.
	websocketEndpointURL, err := wrc.getWebsocketEndpointURL()
	if err != nil {
		// Build error observation for connection failure
		errorObs := wrc.getWebsocketConnectionErrorObservation(err, "selected endpoint does not support websocket RPC type")
		logger.Error().Err(err).Msg("❌ Selected endpoint does not support websocket RPC type")
		return nil, errorObs, fmt.Errorf("selected endpoint does not support websocket RPC type: %w", err)
	}
	logger = logger.With("websocket_url", websocketEndpointURL)
	wrc.logger = logger

	// Get the headers for the websocket connection that will be sent to the endpoint.
	endpointConnectionHeaders, err := wrc.getWebsocketConnectionHeaders()
	if err != nil {
		// Build error observation for connection failure
		errorObs := wrc.getWebsocketConnectionErrorObservation(err, "failed to get websocket connection headers")
		logger.Error().Err(err).Msg("❌ Failed to get websocket connection headers")
		return nil, errorObs, fmt.Errorf("failed to get websocket connection headers: %w", err)
	}

	// Start the websocket bridge and get a completion channel.
	// The messageProcessor (typically from the gateway layer) handles message processing.
	completionChan, err := websockets.StartBridge(
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
		logger.Error().Err(err).Msg("Failed to start WebSocket bridge")
		return nil, errorObs, fmt.Errorf("failed to start websocket bridge: %w", err)
	}

	// Build success observation for the established connection
	successObs := wrc.getWebsocketConnectionSuccessObservation()
	logger.Info().Msg("✅ WebSocket bridge started successfully")

	return completionChan, successObs, nil
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
	logger := wrc.logger.With("method", "ProcessClientWebsocketMessage")

	logger.Debug().Msgf("received message from client: %s", string(msgData))

	// If the selected endpoint is a fallback endpoint, skip signing the message.
	// Fallback endpoints bypass the protocol so the raw message is sent to the endpoint.
	// TODO_IMPROVE(@commoddity,@adshmh): Cleanly separate fallback endpoint handling from the protocol package.
	if wrc.selectedEndpoint.IsFallback() {
		return msgData, nil
	}

	// If the selected endpoint is a protocol endpoint, we need to sign the message.
	signedRelayRequest, err := wrc.signClientWebsocketMessage(msgData)
	if err != nil {
		logger.Error().Err(err).Msg("❌ failed to sign request")
		return nil, err
	}

	return signedRelayRequest, nil
}

// signClientWebsocketMessage signs a message from the client using the Relay Request Signer.
func (wrc *websocketRequestContext) signClientWebsocketMessage(msgData []byte) ([]byte, error) {
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
	logger := wrc.logger.With("method", "ProcessEndpointWebsocketMessage")

	logger.Debug().Msgf("received message from endpoint: %s", string(msgData))

	// If the selected endpoint is a fallback endpoint, skip validation.
	// Fallback endpoints bypass the protocol so the raw message is sent to the endpoint.
	// TODO_IMPROVE(@commoddity,@adshmh): Cleanly separate fallback endpoint handling from the protocol package.
	if wrc.selectedEndpoint.IsFallback() {
		return msgData, wrc.getWebsocketMessageSuccessObservation(msgData), nil
	}

	// If the selected endpoint is a protocol endpoint, we need to validate the message.
	validatedRelayResponse, err := wrc.validateEndpointWebsocketMessage(msgData)
	if err != nil {
		logger.Error().Err(err).Msg("❌ failed to validate relay response")
		return nil, wrc.getWebsocketMessageErrorObservation(msgData, err), err
	}

	return validatedRelayResponse, wrc.getWebsocketMessageSuccessObservation(msgData), nil
}

// validateEndpointWebsocketMessage validates a message from the endpoint using the Shannon FullNode.
// TODO_IMPROVE(@adshmh): Compare this to 'validateAndProcessResponse' and align the two implementations
// w.r.t design, error handling, etc...
func (wrc *websocketRequestContext) validateEndpointWebsocketMessage(msgData []byte) ([]byte, error) {
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
					RequestError: wrc.requestErrorObservation,
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
					RequestError: wrc.requestErrorObservation,
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
func (wrc *websocketRequestContext) getWebsocketConnectionSuccessObservation() *protocolobservations.Observations {
	return &protocolobservations.Observations{
		Shannon: &protocolobservations.ShannonObservationsList{
			Observations: []*protocolobservations.ShannonRequestObservations{
				{
					ServiceId: string(wrc.serviceID),
					ObservationData: &protocolobservations.ShannonRequestObservations_WebsocketConnectionObservation{
						WebsocketConnectionObservation: buildWebsocketConnectionSuccessObservation(
							wrc.logger,
							wrc.selectedEndpoint,
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
							err,
							details,
						),
					},
				},
			},
		},
	}
}
