package shannon

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/pokt-network/poktroll/pkg/relayer/proxy"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
	sdk "github.com/pokt-network/shannon-sdk"

	"github.com/buildwithgrove/path/gateway"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/request"
)

// The requestContext implements the gateway.ProtocolRequestContextWebsocket interface.
// It handles protocol-level WebSocket message processing for both client and endpoint messages.
// For example, client messages are signed and endpoint messages are validated.
var _ gateway.ProtocolRequestContextWebsocket = &requestContext{}

// ---------- Connection Establishment ----------

// GetWebsocketConnectionHeaders returns headers for the websocket connection:
//   - Target-Service-Id: The service ID of the target service
//   - App-Address: The address of the session's application
//   - Rpc-Type: Always "websocket" for websocket connection requests
func (rc *requestContext) GetWebsocketConnectionHeaders() (http.Header, error) {
	// Requests to fallback endpoints bypass the protocol so RelayMiner headers are not needed.
	// TODO_IMPROVE(@commoddity,@adshmh): Cleanly separate fallback endpoint handling from the protocol package.
	// TODO_ARCHITECTURE: Extract fallback endpoint handling from protocol package
	// Current: Fallback logic is scattered with if rc.selectedEndpoint.IsFallback() checks
	// Suggestion: Use strategy pattern or separate fallback handler to cleanly separate concerns
	if rc.selectedEndpoint.IsFallback() {
		return http.Header{}, nil
	}

	// If the selected endpoint is a protocol endpoint, add the headers
	// that the RelayMiner requires to forward the request to the Endpoint.
	return rc.getRelayMinerConnectionHeaders()
}

// getRelayMinerConnectionHeaders returns headers for RelayMiner websocket connections.
func (rc *requestContext) getRelayMinerConnectionHeaders() (http.Header, error) {
	sessionHeader := rc.selectedEndpoint.Session().GetHeader()

	if sessionHeader == nil {
		rc.logger.Error().Msg("❌ SHOULD NEVER HAPPEN: Error getting relay miner connection headers: session header is nil")
		return http.Header{}, fmt.Errorf("session header is nil")
	}

	return http.Header{
		request.HTTPHeaderTargetServiceID: {sessionHeader.ServiceId},
		request.HTTPHeaderAppAddress:      {sessionHeader.ApplicationAddress},
		proxy.RPCTypeHeader:               {strconv.Itoa(int(sharedtypes.RPCType_WEBSOCKET))},
	}, nil
}

// GetWebsocketEndpointURL returns the websocket URL for the selected endpoint.
// This URL is used to establish the websocket connection to the endpoint.
func (rc *requestContext) GetWebsocketEndpointURL() (string, error) {
	websocketURL, err := rc.selectedEndpoint.WebsocketURL()
	if err != nil {
		rc.logger.Error().Err(err).Msg("❌ Selected endpoint does not support websocket RPC type")
		return "", err
	}

	return websocketURL, nil
}

// ---------- Client Message Processing ----------

// ProcessProtocolClientWebsocketMessage processes a message from the client.
func (rc *requestContext) ProcessProtocolClientWebsocketMessage(msgData []byte) ([]byte, error) {
	logger := rc.logger.With("method", "ProcessClientWebsocketMessage")

	logger.Debug().Msgf("received message from client: %s", string(msgData))

	// If the selected endpoint is a fallback endpoint, skip signing the message.
	// Fallback endpoints bypass the protocol so the raw message is sent to the endpoint.
	// TODO_IMPROVE(@commoddity,@adshmh): Cleanly separate fallback endpoint handling from the protocol package.
	if rc.selectedEndpoint.IsFallback() {
		return msgData, nil
	}

	// If the selected endpoint is a protocol endpoint, we need to sign the message.
	signedRelayRequest, err := rc.signClientWebsocketMessage(msgData)
	if err != nil {
		logger.Error().Err(err).Msg("❌ failed to sign request")
		return nil, err
	}

	return signedRelayRequest, nil
}

// signClientWebsocketMessage signs a message from the client using the Relay Request Signer.
func (rc *requestContext) signClientWebsocketMessage(msgData []byte) ([]byte, error) {
	unsignedRelayRequest := &servicetypes.RelayRequest{
		Meta: servicetypes.RelayRequestMetadata{
			SessionHeader:           rc.selectedEndpoint.Session().GetHeader(),
			SupplierOperatorAddress: rc.selectedEndpoint.Supplier(),
		},
		Payload: msgData,
	}

	app := rc.selectedEndpoint.Session().GetApplication()
	if app == nil {
		rc.logger.Error().Msg("❌ SHOULD NEVER HAPPEN: session application is nil")
		return nil, fmt.Errorf("session application is nil")
	}

	signedRelayRequest, err := rc.relayRequestSigner.SignRelayRequest(unsignedRelayRequest, *app)
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
func (rc *requestContext) ProcessProtocolEndpointWebsocketMessage(
	msgData []byte,
) ([]byte, protocolobservations.Observations, error) {
	logger := rc.logger.With("method", "ProcessEndpointWebsocketMessage")

	logger.Debug().Msgf("received message from endpoint: %s", string(msgData))

	// If the selected endpoint is a fallback endpoint, skip validation.
	// Fallback endpoints bypass the protocol so the raw message is sent to the endpoint.
	// TODO_IMPROVE(@commoddity,@adshmh): Cleanly separate fallback endpoint handling from the protocol package.
	if rc.selectedEndpoint.IsFallback() {
		return msgData, rc.getWebsocketMessageSuccessObservation(msgData), nil
	}

	// If the selected endpoint is a protocol endpoint, we need to validate the message.
	validatedRelayResponse, err := rc.validateEndpointWebsocketMessage(msgData)
	if err != nil {
		logger.Error().Err(err).Msg("❌ failed to validate relay response")
		return nil, rc.getWebsocketMessageErrorObservation(msgData, err), err
	}

	return validatedRelayResponse, rc.getWebsocketMessageSuccessObservation(msgData), nil
}

// validateEndpointWebsocketMessage validates a message from the endpoint using the Shannon FullNode.
// TODO_IMPROVE(@adshmh): Compare this to 'validateAndProcessResponse' and align the two implementations
// w.r.t design, error handling, etc...
func (rc *requestContext) validateEndpointWebsocketMessage(msgData []byte) ([]byte, error) {
	// Validate the relay response using the Shannon FullNode
	relayResponse, err := rc.fullNode.ValidateRelayResponse(sdk.SupplierAddress(rc.selectedEndpoint.Supplier()), msgData)
	if err != nil {
		rc.logger.Error().Err(err).Msg("❌ failed to validate relay response in websocket message")
		return nil, fmt.Errorf("%w: %s", errRelayResponseInWebsocketMessageValidationFailed, err.Error())
	}
	rc.logger.Debug().Msgf("received message from protocol endpoint: %s", string(relayResponse.Payload))

	return relayResponse.Payload, nil
}

// getWebsocketMessageSuccessObservation updates the observations for the current message
// if the message handler does not return an error.
func (rc *requestContext) getWebsocketMessageSuccessObservation(
	msgData []byte,
) protocolobservations.Observations {
	// Create a new WebSocket message observation for success
	wsMessageObs := buildWebsocketMessageSuccessObservation(
		rc.logger,
		rc.selectedEndpoint,
		int64(len(msgData)),
	)

	// Update the observations to use the WebSocket message observation
	return protocolobservations.Observations{
		Shannon: &protocolobservations.ShannonObservationsList{
			Observations: []*protocolobservations.ShannonRequestObservations{
				{
					ServiceId:    string(rc.serviceID),
					RequestError: rc.requestErrorObservation,
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
func (rc *requestContext) getWebsocketMessageErrorObservation(
	msgData []byte,
	messageError error,
) protocolobservations.Observations {
	// Error classification based on trusted error sources only
	endpointErrorType, recommendedSanctionType := classifyRelayError(rc.logger, messageError)

	// Create a new WebSocket message observation for error
	wsMessageObs := buildWebsocketMessageErrorObservation(
		rc.selectedEndpoint,
		int64(len(msgData)),
		endpointErrorType,
		fmt.Sprintf("websocket message error: %v", messageError),
		recommendedSanctionType,
		rc.currentRelayMinerError,
	)

	return protocolobservations.Observations{
		Shannon: &protocolobservations.ShannonObservationsList{
			Observations: []*protocolobservations.ShannonRequestObservations{
				{
					ServiceId:    string(rc.serviceID),
					RequestError: rc.requestErrorObservation,
					ObservationData: &protocolobservations.ShannonRequestObservations_WebsocketMessageObservation{
						WebsocketMessageObservation: wsMessageObs,
					},
				},
			},
		},
	}
}
