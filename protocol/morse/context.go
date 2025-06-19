package morse

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pokt-foundation/pocket-go/provider"
	sdkrelayer "github.com/pokt-foundation/pocket-go/relayer"
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/protocol"
)

var _ gateway.ProtocolRequestContext = &requestContext{}

// TODO_TECHDEBT(@adshmh): Make this configurable via either an env variable or YAML config.
const defaultRelayTimeoutMillisec = 5000

// requestContext captures all the data required for handling a single service request.
type requestContext struct {
	logger    polylog.Logger
	fullNode  FullNode
	serviceID protocol.ServiceID

	// selectedEndpoint is the endpoint that has been selected for sending a relay.
	selectedEndpoint *endpoint

	// tracks any errors encountered processing the request
	requestErrorObservation *protocolobservations.MorseRequestError

	// endpointObservations captures observations about endpoints used during request handling
	endpointObservations []*protocolobservations.MorseEndpointObservation
}

// HandleServiceRequest handles incoming service request.
// It uses the supplied payload to send a relay request to an endpoint, and verifies and returns the response.
func (rc *requestContext) HandleServiceRequest(payload protocol.Payload) (protocol.Response, error) {
	// TODO_IMPROVE(@adshmh): use the same pattern for hydrated loggers in any packages with large number of logger fields.
	hydratedLogger := rc.getHydratedLogger("HandleServiceRequest")

	// Internal error: no endpoint selected.
	// record reuqest error due to internal error.
	// no endpoint to sanction.
	if rc.selectedEndpoint == nil {
		return rc.handleInternalError(fmt.Errorf("HandleServiceRequest: no endpoint has been selected on service %s", rc.serviceID))
	}

	// match the selected endpoint against the session.
	morseEndpoint, err := getEndpoint(rc.selectedEndpoint.session, rc.selectedEndpoint.Addr())

	// Internal error: endpoint not in session.
	// record request error due to internal error.
	// no endpoint to sanction.
	if err != nil {
		return rc.handleInternalError(fmt.Errorf("error matching endpoint against session's nodes: %w", err))
	}

	// record the endpoint query time.
	endpointQueryTime := time.Now()

	// send the request to the endpoint.
	output, err := rc.sendRelay(
		hydratedLogger,
		morseEndpoint,
		rc.selectedEndpoint.session,
		rc.selectedEndpoint.app.aat,
		// TODO_FUTURE(@adshmh): support service-specific timeouts, passed from the Qos instance.
		0, // SDK to use the default timeout.
		payload,
	)

	// Handle endpoint error:
	// - Record observation
	// - Return an error
	if err != nil {
		return rc.handleEndpointError(endpointQueryTime, err, []byte(output.Response), output.StatusCode)
	}

	// Success:
	// - Record observation
	// - Return the response received from the endpoint.
	return rc.handleEndpointSuccess(endpointQueryTime, []byte(output.Response), output.StatusCode)
}

// HandleWebsocketRequest handles incoming WebSocket network request.
// Morse does not support WebSocket connections so this method will always return an error.
// Satisfies the gateway.ProtocolRequestContext interface.
func (rc *requestContext) HandleWebsocketRequest(_ polylog.Logger, _ *http.Request, _ http.ResponseWriter) error {
	return fmt.Errorf("HandleWebsocketRequest: Morse does not support WebSocket connections")
}

// GetObservations returns the observations that have been collected during the protocol request processing.
// These observations can be used by the caller to make decisions on endpoint reliability and quality.
//
// Implements gateway.ProtocolRequestContext interface.
func (rc *requestContext) GetObservations() protocolobservations.Observations {
	return protocolobservations.Observations{
		Protocol: &protocolobservations.Observations_Morse{
			Morse: &protocolobservations.MorseObservationsList{
				Observations: []*protocolobservations.MorseRequestObservations{
					{
						ServiceId:            string(rc.serviceID),
						EndpointObservations: rc.endpointObservations,
					},
				},
			},
		},
	}
}

// sendRelay is a helper function for handling the low-level details of a Morse relay.
func (rc *requestContext) sendRelay(
	logger polylog.Logger,
	node provider.Node,
	session provider.Session,
	aat provider.PocketAAT,
	timeoutMillisec int,
	payload protocol.Payload,
) (provider.RelayOutput, error) {
	fullNodeInput := &sdkrelayer.Input{
		Blockchain: string(rc.serviceID),
		Node:       &node,
		Session:    &session,
		PocketAAT:  &aat,
		Data:       payload.Data,
		Method:     payload.Method,
		Path:       payload.Path,
	}

	logger.Debug().
		Str("endpoint", node.PublicKey).
		Msg("Sending relay to endpoint")

	timeout := timeoutMillisec
	if timeout == 0 {
		timeout = defaultRelayTimeoutMillisec
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	output, err := rc.fullNode.SendRelay(ctx, fullNodeInput)
	if err != nil {
		return provider.RelayOutput{}, err
	}

	if output == nil || output.RelayOutput == nil {
		return provider.RelayOutput{}, NewNullRelayResponseError("null RelayOutput field in the relay response from the SDK")
	}

	// TODO_TECHDEBT: complete the following items regarding the node and proof structs
	// 1. Verify their correctness
	// 2. Pass a logger to the request context to log them in debug mode.
	if output.Node == nil {
		return provider.RelayOutput{}, NewNullRelayResponseError("null Node field in the relay response from the SDK")
	}

	if output.Proof == nil {
		return provider.RelayOutput{}, NewNullRelayResponseError("null Proof field in the relay response from the SDK")
	}

	return *output.RelayOutput, nil
}

// handleEndpointSuccess records a successful endpoint observation and returns the response.
// - Tracks the endpoint success in observations
// - Builds and returns the protocol response from the endpoint's returned data.
func (rc *requestContext) handleEndpointSuccess(
	endpointQueryTime time.Time,
	endpointResponsePayload []byte,
	endpointResponseHTTPStatusCode int,
) (protocol.Response, error) {
	// Track the endpoint success observation.
	rc.endpointObservations = append(rc.endpointObservations,
		buildEndpointSuccessObservation(
			*rc.selectedEndpoint,
			endpointQueryTime,
			time.Now(), // Timestamp: endpoint query completed.
		),
	)

	// Build and return the relay response received from the endpoint.
	return protocol.Response{
		EndpointAddr:   rc.selectedEndpoint.Addr(),
		Bytes:          endpointResponsePayload,
		HTTPStatusCode: endpointResponseHTTPStatusCode,
	}, nil
}

// handleEndpointError records an endpoint error observation and returns the response.
// - Tracks the endpoint error in observations
// - Builds and returns the protocol response from the endpoint's returned data.
func (rc *requestContext) handleEndpointError(
	endpointQueryTime time.Time,
	endpointErr error,
	endpointResponsePayload []byte,
	endpointResponseHTTPStatusCode int,
) (protocol.Response, error) {
	// Classify the endpoint's error for the observation.
	// Determine any applicable sanctions.
	endpointErrorType, recommendedSanctionType := classifyRelayError(rc.logger, endpointErr)

	// Log the endpoint error.
	hydratedLogger := rc.getHydratedLogger("HandleServiceRequest")
	hydratedLogger.Error().
		Str("error_type", endpointErrorType.String()).
		Str("sanction_type", recommendedSanctionType.String()).
		Err(endpointErr).
		Msg("relay error occurred.")

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

	// Return the endpoint's response as-is, along with the encountered error.
	return protocol.Response{
		EndpointAddr:   rc.selectedEndpoint.Addr(),
		Bytes:          endpointResponsePayload,
		HTTPStatusCode: endpointResponseHTTPStatusCode,
	}, endpointErr
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
