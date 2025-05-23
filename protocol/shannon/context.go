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

// requestContext provides all the functionality required by the gateway package
// for handling a single service request.
var _ gateway.ProtocolRequestContext = &requestContext{}

// RelayRequestSigner is used by the request context to sign the relay request.
// It takes an unsigned relay request and an application, and returns a relay request signed either by the gateway that has delegation from the app.
// If/when the Permissionless Gateway Mode is supported by the Shannon integration, the app's own private key may also be used for signing the relay request.
type RelayRequestSigner interface {
	SignRelayRequest(req *servicetypes.RelayRequest, app apptypes.Application) (*servicetypes.RelayRequest, error)
}

// requestContext captures all the data required for handling a single service request.
type requestContext struct {
	logger polylog.Logger

	fullNode FullNode
	// TODO_TECHDEBT(@adshmh): add sanctionedEndpointsStore to the request context.
	serviceID protocol.ServiceID

	relayRequestSigner RelayRequestSigner

	// selectedEndpoint is the endpoint that has been selected for sending a relay.
	// Sending a relay will fail if this field is not set through a call to the SelectEndpoint method.
	selectedEndpoint *endpoint

	// tracks any errors encountered processing the request
	requestErrorObservation *protocolobservations.ShannonRequestError

	// endpointObservations captures observations about endpoints used during request handling
	endpointObservations []*protocolobservations.ShannonEndpointObservation
}

// HandleServiceRequest satisfies the gateway package's ProtocolRequestContext interface.
// It uses the supplied payload to send a relay request to an endpoint, and verifies and returns the response.
func (rc *requestContext) HandleServiceRequest(payload protocol.Payload) (protocol.Response, error) {
	// Internal error: no endpoint selected.
	// record reuqest error due to internal error.
	// no endpoint to sanction.
	if rc.selectedEndpoint == nil {
		return rc.handleInternalError(fmt.Errorf("HandleServiceRequest: no endpoint has been selected on service %s", rc.serviceID))
	}

	// record the endpoint query time.
	endpointQueryTime := time.Now()

	// send the relay request.
	response, err := rc.sendRelay(payload)

	// Handle endpoint error:
	// - Record observation
	// - Return an error
	if err != nil {
		return rc.handleEndpointError(endpointQueryTime, err)
	}

	// The Payload field of the response received from the endpoint, i.e. the relay miner,
	// is a serialized http.Response struct. It needs to be deserialized into an HTTP Response struct
	// to access the Service's response body, status code, etc.
	relayResponse, err := deserializeRelayResponse(response.Payload)
	relayResponse.EndpointAddr = rc.selectedEndpoint.Addr()
	if err != nil {
		// Wrap the error with a detailed message.
		deserializeErr := fmt.Errorf("error deserializing endpoint into a POKTHTTP response: %w", err)
		return rc.handleEndpointError(endpointQueryTime, deserializeErr)
	}

	// Success:
	// - Record observation
	// - Return the response received from the endpoint.
	return rc.handleEndpointSuccess(endpointQueryTime, relayResponse)

}

// HandleWebsocketRequest opens a persistent websocket connection to the selected endpoint.
// Satisfies the gateway.ProtocolRequestContext interface.
func (rc *requestContext) HandleWebsocketRequest(logger polylog.Logger, req *http.Request, w http.ResponseWriter) error {
	if rc.selectedEndpoint == nil {
		return fmt.Errorf("handleWebsocketRequest: no endpoint has been selected on service %s", rc.serviceID)
	}

	wsLogger := logger.With(
		"endpoint_url", rc.selectedEndpoint.PublicURL(),
		"endpoint_addr", rc.selectedEndpoint.Addr(),
		"service_id", rc.serviceID,
	)

	// Upgrade the HTTP request from the client to a websocket connection.
	// This connection is then passed to the websocket bridge to handle the Client<->Gateway communication.
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

	// run bridge in a goroutine to avoid blocking the main thread
	go bridge.Run()

	wsLogger.Info().Msg("websocket connection established")

	return nil
}

// GetObservations returns the set of Shannon protocol-level observations for the current request context.
// The returned observations are used to:
// 1. Update the Shannon's endpoint store.
// 2. Report metrics on the operation of PATH (in the metrics package)
// 3. Report the requests to the data pipeline.
//
// Implements the gateway.ProtocolRequestContext interface.
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

// sendRelay sends a the supplied payload as a relay request to the endpoint selected for the request context through the SelectEndpoint method.
// It is required to fulfill the FullNode interface.
func (rc *requestContext) sendRelay(payload protocol.Payload) (*servicetypes.RelayResponse, error) {
	if rc.selectedEndpoint == nil {
		return nil, fmt.Errorf("sendRelay: no endpoint has been selected on service %s", rc.serviceID)
	}

	session := rc.selectedEndpoint.session
	if session.Application == nil {
		return nil, fmt.Errorf("sendRelay: nil app on session %s for service %s", session.SessionId, rc.serviceID)
	}
	app := *session.Application

	payloadBz := []byte(payload.Data)
	relayRequest, err := buildUnsignedRelayRequest(*rc.selectedEndpoint, session, payloadBz, payload.Path)
	if err != nil {
		return nil, err
	}

	signedRelayReq, err := rc.signRelayRequest(relayRequest, app)
	if err != nil {
		return nil, fmt.Errorf("sendRelay: error signing the relay request for app %s: %w", app.Address, err)
	}

	ctxWithTimeout, cancelFn := context.WithTimeout(context.Background(), time.Duration(payload.TimeoutMillisec)*time.Millisecond)
	defer cancelFn()

	// TODO_MVP(@adshmh): check the HTTP status code returned by the endpoint.
	responseBz, err := sendHttpRelay(ctxWithTimeout, rc.selectedEndpoint.url, signedRelayReq)
	if err != nil {
		return nil, fmt.Errorf("relay: error sending request to endpoint %s: %w", rc.selectedEndpoint.url, err)
	}

	// Validate the response
	response, err := rc.fullNode.ValidateRelayResponse(sdk.SupplierAddress(rc.selectedEndpoint.supplier), responseBz)
	if err != nil {
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

// buildUnsignedRelayRequest builds a ready-to-sign RelayRequest struct using the
// supplied endpoint, session, and payload.
// The returned RelayRequest is intended to be signed and sent to the endpoint to
// receive the endpoint's response.
func buildUnsignedRelayRequest(
	endpoint endpoint,
	session sessiontypes.Session,
	payload []byte,
	path string,
) (*servicetypes.RelayRequest, error) {
	// If the path is not empty (ie. for a REST service request), append it to the endpoint's URL
	url := endpoint.url
	if path != "" {
		url = fmt.Sprintf("%s%s", url, path)
	}

	// TODO_TECHDEBT: need to select the correct underlying request (HTTP, etc.) based on the selected service.
	jsonRpcHttpReq, err := shannonJsonRpcHttpRequest(payload, url)
	if err != nil {
		return nil, fmt.Errorf("error building a JSONRPC HTTP request for url %s: %w", url, err)
	}

	relayRequest, err := embedHttpRequest(jsonRpcHttpReq)
	if err != nil {
		return nil, fmt.Errorf("error embedding a JSONRPC HTTP request for url %s: %w", endpoint.url, err)
	}

	// TODO_MVP(@adshmh): use the new `FilteredSession` struct provided by the Shannon SDK to get the session and the endpoint.
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

	// No endpoint specified on the request context.
	// This should never happen.
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
		"seleced_endpoint_app", sessionHeader.ApplicationAddress,
	)

	return logger
}

// handleInternalError is called if the request processing fails.
// This only happens before the request is sent to any endpoints.
// DEV_NOTE: This should NEVER happen and any logged entries by this method should be investigated.
// - Records an internal error on the request for observations.
// - Logs an error entry.
func (rc *requestContext) handleInternalError(internalErr error) (protocol.Response, error) {
	hydratedLogger := rc.getHydratedLogger("handleInternalError")

	// log the internal error.
	hydratedLogger.Error().Err(internalErr).Msg("Internal error occurred. This should be investigated as a bug.")

	// Sets the request processing error for generating observations.
	rc.requestErrorObservation = buildInternalRequestProcessingErrorObservation(internalErr)

	return protocol.Response{}, internalErr
}

// handleEndpointError records an endpoint error observation and returns the response.
// - Tracks the endpoint error in observations
// - Builds and returns the protocol response from the endpoint's returned data.
func (rc *requestContext) handleEndpointError(
	endpointQueryTime time.Time,
	endpointErr error,
) (protocol.Response, error) {
	hydratedLogger := rc.getHydratedLogger("handleEndpointError")
	selectedEndpointAddr := rc.selectedEndpoint.Addr()

	// Classify the endpoint's error for the observation.
	// Determine any applicable sanctions.
	endpointErrorType, recommendedSanctionType := classifyRelayError(hydratedLogger, endpointErr)

	// Log the endpoint error.
	hydratedLogger.Warn().
		Str("error_type", endpointErrorType.String()).
		Str("sanction_type", recommendedSanctionType.String()).
		Err(endpointErr).
		Msg("relay error occurred. Service request will fail.")

	// Track the endpoint error observation.
	rc.endpointObservations = append(rc.endpointObservations,
		buildEndpointErrorObservation(
			*rc.selectedEndpoint,
			endpointQueryTime,
			time.Now(), // Timestamp: endpoint query completed.
			endpointErrorType,
			fmt.Sprintf("relay error: %v", endpointErr),
			recommendedSanctionType,
		),
	)

	// Return an error
	return protocol.Response{EndpointAddr: selectedEndpointAddr},
		fmt.Errorf("relay: error sending relay for service %s endpoint %s: %w",
			rc.serviceID, selectedEndpointAddr, endpointErr,
		)
}

// handleEndpointSuccess records a successful endpoint observation and returns the response.
// - Tracks the endpoint success in observations
// - Builds and returns the protocol response from the endpoint's returned data.
func (rc *requestContext) handleEndpointSuccess(
	endpointQueryTime time.Time,
	endpointResponse protocol.Response,
) (protocol.Response, error) {
	hydratedLogger := rc.getHydratedLogger("handleEndpointSuccess")
	hydratedLogger = hydratedLogger.With("endpoint_response_payload_len", len(endpointResponse.Bytes))
	hydratedLogger.Debug().Msg("Successfully deserialized the response received from the selected endpoint.")

	// Track the endpoint success observation.
	rc.endpointObservations = append(rc.endpointObservations,
		buildEndpointSuccessObservation(
			*rc.selectedEndpoint,
			endpointQueryTime,
			time.Now(), // Timestamp: endpoint query completed.
		),
	)

	// Return the relay response received from the endpoint.
	return endpointResponse, nil
}
