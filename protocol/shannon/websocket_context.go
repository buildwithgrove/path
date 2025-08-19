package shannon

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pokt-network/poktroll/pkg/relayer/proxy"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
	sdk "github.com/pokt-network/shannon-sdk"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/observation"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/request"
)

var _ gateway.ProtocolRequestContext = &requestContext{}

// GetWebsocketConnectionHeaders returns headers for the websocket connection:
//   - Target-Service-Id: The service ID of the target service
//   - App-Address: The address of the session's application
//   - Rpc-Type: Always "websocket" for websocket connection requests
func (rc *requestContext) GetWebsocketConnectionHeaders() (http.Header, error) {
	// Requests to fallback endpoints bypass the protocol so RelayMiner headers are not needed.
	if rc.selectedEndpoint.IsFallback() {
		return http.Header{}, nil
	}

	// If the selected endpoint is a protocol endpoint, add the headers
	// that the RelayMiner requires to forward the request to the Endpoint.
	return rc.getRelayMinerConnectionHeaders(rc.getSelectedEndpoint().Session().GetHeader())
}

// getRelayMinerConnectionHeaders returns headers for RelayMiner websocket connections:
//   - Target-Service-Id: The service ID of the target service
//   - App-Address: The address of the session's application
//   - Rpc-Type: Always "websocket" for websocket connection requests
func (rc *requestContext) getRelayMinerConnectionHeaders(sessionHeader *sessiontypes.SessionHeader) (http.Header, error) {
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

// ProcessClientWebsocketMessage processes a message from the client.
func (rc *requestContext) ProcessClientWebsocketMessage(msgData []byte) ([]byte, error) {
	logger := rc.logger.With("method", "ProcessClientWebsocketMessage")

	logger.Debug().Msgf("received message from client: %s", string(msgData))

	// If the selected endpoint is a fallback endpoint, skip signing the message.
	// Fallback endpoints bypass the protocol so the raw message is sent to the endpoint.
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
		return nil, fmt.Errorf("%w: %s", errRelayRequestSigningFailed, err.Error())
	}

	relayRequestBz, err := signedRelayRequest.Marshal()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errRelayRequestSigningFailed, err.Error())
	}

	return relayRequestBz, nil
}

// ---------- Endpoint Message Processing ----------

// ProcessEndpointWebsocketMessage processes a message from the endpoint.
func (rc *requestContext) ProcessEndpointWebsocketMessage(msgData []byte) ([]byte, error) {
	logger := rc.logger.With("method", "ProcessEndpointWebsocketMessage")

	logger.Debug().Msgf("received message from endpoint: %s", string(msgData))

	// If the selected endpoint is a fallback endpoint, skip validation.
	// Fallback endpoints bypass the protocol so the raw message is sent to the endpoint.
	if rc.selectedEndpoint.IsFallback() {
		return msgData, nil
	}

	// If the selected endpoint is a protocol endpoint, we need to validate the message.
	validatedRelayResponse, err := rc.validateEndpointWebsocketMessage(msgData)
	if err != nil {
		logger.Error().Err(err).Msg("❌ failed to validate relay response")
		return nil, err
	}

	return validatedRelayResponse, nil
}

// validateEndpointWebsocketMessage validates a message from the endpoint using the Shannon FullNode.
func (rc *requestContext) validateEndpointWebsocketMessage(msgData []byte) ([]byte, error) {
	// Validate the relay response using the Shannon FullNode
	relayResponse, err := rc.fullNode.ValidateRelayResponse(sdk.SupplierAddress(rc.selectedEndpoint.Supplier()), msgData)
	if err != nil {
		rc.logger.Error().Err(err).Msg("❌ failed to validate relay response")
		return nil, fmt.Errorf("%w: %s", errRelayResponseValidationFailed, err.Error())
	}

	rc.logger.Debug().Msgf("received message from protocol endpoint: %s", string(relayResponse.Payload))

	return relayResponse.Payload, nil
}

// TODO_IN_THIS_PR(@commoddity): clean up the observation initialization logic below
// TODO_IN_THIS_PR(@commoddity): create a new Shannon observation specific to websockets,
// which must differentiate between a failed connection attempt and a failed message.

// UpdateMessageObservationsFromSuccess updates the observations for the current message
// if the message handler does not return an error.
func (rc *requestContext) UpdateMessageObservationsFromSuccess(
	observations *observation.RequestResponseObservations,
) *protocolobservations.Observations {
	// Get the websocket endpoint observation to update
	endpointObs, err := rc.getWebsocketEndpointObservation(observations)
	if err != nil {
		rc.logger.Error().Err(err).Msg("❌ SHOULD NEVER HAPPEN: failed to get websocket endpoint observation")
		return nil
	}

	return &protocolobservations.Observations{
		Shannon: &protocolobservations.ShannonObservationsList{
			Observations: []*protocolobservations.ShannonRequestObservations{
				{
					EndpointObservations: []*protocolobservations.ShannonEndpointObservation{
						buildWebsocketMessageSuccessObservation(endpointObs),
					},
				},
			},
		},
	}
}

// UpdateMessageObservationsFromError updates the observations for the current message
// if the message handler returns an error.
func (rc *requestContext) UpdateMessageObservationsFromError(
	observations *observation.RequestResponseObservations,
	messageError error,
) *protocolobservations.Observations {
	// Set the endpoint observations for the current message
	rc.handleEndpointWebsocketError(time.Now(), messageError)

	return &protocolobservations.Observations{
		Shannon: &protocolobservations.ShannonObservationsList{
			Observations: []*protocolobservations.ShannonRequestObservations{
				{EndpointObservations: rc.endpointObservations},
			},
		},
	}
}

// getWebsocketEndpointObservation safely retrieves the websocket
// endpoint observation from the request-response observations.
//
// This method is primarily a sanity check as Bridge obervations should
// always have only one request observation with one endpoint observation.
func (rc *requestContext) getWebsocketEndpointObservation(
	observations *observation.RequestResponseObservations,
) (*protocolobservations.ShannonEndpointObservation, error) {
	// Validate observation structure
	if observations == nil ||
		observations.Protocol == nil ||
		observations.Protocol.Shannon == nil {
		return nil, fmt.Errorf("observations are nil")
	}

	shannonObs := observations.Protocol.Shannon

	// For websocket connections, we expect exactly one request observation
	if len(shannonObs.Observations) != 1 {
		return nil, fmt.Errorf("observations have more than one request observation")
	}

	requestObs := shannonObs.Observations[0]

	// Each websocket connection should have exactly one endpoint observation
	if len(requestObs.EndpointObservations) != 1 {
		return nil, fmt.Errorf("request observation has more than one endpoint observation")
	}

	return requestObs.EndpointObservations[0], nil
}
