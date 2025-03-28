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
	fullNode                 FullNode
	sanctionedEndpointsStore *sanctionedEndpointsStore
	serviceID                protocol.ServiceID
	logger                   polylog.Logger

	// endpoints contains all the candidate endpoints available for processing a service request.
	endpoints map[protocol.EndpointAddr]endpoint
	// selectedEndpoint is the endpoint that has been selected for sending a relay.
	// NOTE: Sending a relay will fail if this field is not set through a call to the SelectEndpoint method.
	selectedEndpoint endpoint

	// endpointObservations captures observations about endpoints used during request handling
	endpointObservations []*protocolobservations.MorseEndpointObservation
}

// AvailableEndpoints returns the list of available endpoints for the current request context.
// Satisfies the gateway.ProtocolRequestContext interface.
func (rc *requestContext) AvailableEndpoints() ([]protocol.Endpoint, error) {
	if len(rc.endpoints) == 0 {
		return nil, fmt.Errorf("AvailableEndpoints: no endpoints found for service %s", rc.serviceID)
	}

	endpoints := make([]protocol.Endpoint, 0, len(rc.endpoints))
	for _, endpoint := range rc.endpoints {
		endpoints = append(endpoints, endpoint)
	}
	return endpoints, nil
}

// HandleServiceRequest handles incoming service request.
// It uses the supplied payload to send a relay request to an endpoint, and verifies and returns the response.
func (rc *requestContext) HandleServiceRequest(payload protocol.Payload) (protocol.Response, error) {
	// TODO_IMPROVE(@adshmh): use the same pattern for hydrated loggers in any packages with large number of logger fields.
	hydratedLogger := rc.getHydratedLogger("HandleServiceRequest")

	if rc.selectedEndpoint.IsEmpty() {
		// Internal error: no endpoint selected, record an observation but no sanctions
		// as no endpoint was contacted
		hydratedLogger.Error().Msg("no endpoint has been selected.")

		rc.recordEndpointObservation(
			endpoint{}, // No endpoint selected
			protocolobservations.MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_INTERNAL,
			"no endpoint has been selected, endpoint selection has not been performed or has failed",
			protocolobservations.MorseSanctionType_MORSE_SANCTION_UNSPECIFIED, // No sanction as no endpoint was contacted
		)

		return protocol.Response{}, fmt.Errorf("HandleServiceRequest: no endpoint has been selected on service %s", rc.serviceID)
	}

	morseEndpoint, err := getEndpoint(rc.selectedEndpoint.session, rc.selectedEndpoint.Addr())
	if err != nil {
		// Internal error: endpoint not in session, record an observation but no sanctions
		// as this is an internal error, not an endpoint issue
		hydratedLogger.Error().Err(err).Msg("endpoint not found in session.")

		rc.recordEndpointObservation(
			rc.selectedEndpoint,
			protocolobservations.MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_INTERNAL,
			fmt.Sprintf("error matching endpoint against session's nodes: %v", err),
			protocolobservations.MorseSanctionType_MORSE_SANCTION_UNSPECIFIED, // No sanction as this is an internal error
		)

		return protocol.Response{}, NewEndpointNotInSessionError(string(rc.selectedEndpoint.Addr()))
	}

	output, err := rc.sendRelay(
		hydratedLogger,
		morseEndpoint,
		rc.selectedEndpoint.session,
		rc.selectedEndpoint.app.aat,
		// TODO_FUTURE(@adshmh): support service-specific timeouts, passed from the Qos instance.
		0, // SDK to use the default timeout.
		payload,
	)

	// TODO_FUTURE(@adshmh): evaluate tracking successful cases (err == nil).
	// Example: measuring endpoint response times.
	//
	// Record any errors that occurred during relay
	if err != nil {
		endpointErrorType, recommendedSanctionType := classifyRelayError(rc.logger, err)

		hydratedLogger.Error().
			Str("error_type", endpointErrorType.String()).
			Str("sanction_type", recommendedSanctionType.String()).
			Err(err).
			Msg("relay error occurred.")

		rc.recordEndpointObservation(
			rc.selectedEndpoint,
			endpointErrorType,
			fmt.Sprintf("relay error: %v", err),
			recommendedSanctionType,
		)
	}

	return protocol.Response{
		EndpointAddr:   rc.selectedEndpoint.Addr(),
		Bytes:          []byte(output.Response),
		HTTPStatusCode: output.StatusCode,
	}, err
}

// HandleWebsocketRequest handles incoming WebSocket network request.
// Morse does not support WebSocket connections so this method will always return an error.
// Satisfies the gateway.ProtocolRequestContext interface.
func (rc *requestContext) HandleWebsocketRequest(_ polylog.Logger, _ *http.Request, _ http.ResponseWriter) error {
	return fmt.Errorf("HandleWebsocketRequest: Morse does not support WebSocket connections")
}

// SelectEndpoint selects an endpoint from the available endpoints using the provided EndpointSelector.
// The selected endpoint will be used for subsequent service requests.
// Satisfies the gateway.ProtocolRequestContext interface.
func (rc *requestContext) SelectEndpoint(selector protocol.EndpointSelector) error {
	endpoints := make([]protocol.Endpoint, 0, len(rc.endpoints))
	for _, endpoint := range rc.endpoints {
		endpoints = append(endpoints, endpoint)
	}

	if len(endpoints) == 0 {
		return NewNoEndpointsError(string(rc.serviceID))
	}

	selectedEndpointAddr, err := selector.Select(endpoints)
	if err != nil {
		return NewEndpointSelectionError(string(rc.serviceID), err)
	}

	selectedEndpoint, found := rc.endpoints[selectedEndpointAddr]
	if !found {
		return NewEndpointNotFoundError(string(selectedEndpointAddr), string(rc.serviceID))
	}

	rc.selectedEndpoint = selectedEndpoint
	return nil
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

// recordEndpointObservation records an observation for an endpoint.
// It only appends to the observations slice and does not return an error.
func (rc *requestContext) recordEndpointObservation(
	endpoint endpoint,
	errorType protocolobservations.MorseEndpointErrorType,
	errorDetails string,
	sanctionType protocolobservations.MorseSanctionType,
) {
	rc.endpointObservations = append(rc.endpointObservations, &protocolobservations.MorseEndpointObservation{
		AppAddress:          string(endpoint.app.Addr()),
		SessionKey:          endpoint.session.Key,
		SessionServiceId:    endpoint.session.Header.Chain,
		SessionHeight:       int32(endpoint.session.Header.SessionHeight),
		EndpointAddr:        endpoint.address,
		ErrorType:           &errorType,
		ErrorDetails:        &errorDetails,
		RecommendedSanction: &sanctionType,
	})
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
