package gateway

import (
	"context"
	"time"

	"github.com/buildwithgrove/path/observation"
	"github.com/buildwithgrove/path/protocol"
)

// performWebSocketConnectionCheck performs a WebSocket connection establishment check
// for the given endpoint. It constructs a websocketRequestContext, attempts to establish
// a connection, and immediately terminates it to test connectivity.
//
// This method:
// 1. Creates a synthetic HTTP WebSocket upgrade request
// 2. Builds a websocketRequestContext with a short timeout
// 3. Attempts connection establishment
// 4. Cancels the connection immediately after establishment (or timeout)
// 5. Broadcasts connection observations
func (eph *EndpointHydrator) performWebSocketConnectionCheck(
	serviceID protocol.ServiceID,
	serviceQoS QoSService,
	endpointAddr protocol.EndpointAddr,
) error {
	logger := eph.Logger.With(
		"method", "performWebSocketConnectionCheck",
		"service_id", string(serviceID),
		"endpoint_addr", string(endpointAddr),
	)

	// Create a context with a short timeout for the WebSocket connection check
	// This ensures we don't wait too long for unresponsive endpoints
	checkTimeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()

	// Build a websocketRequestContext for the connection check
	// We manually set the required fields since we already have the serviceQoS and serviceID
	websocketRequestCtx := &websocketRequestContext{
		logger:              logger,
		context:             ctx,
		gatewayObservations: getSyntheticRequestGatewayObservations(),
		protocol:            eph.Protocol,
		// We don't need httpRequestParser since we already have serviceQoS and serviceID
		httpRequestParser: nil,
		metricsReporter:   eph.MetricsReporter,
		dataReporter:      eph.DataReporter,
		// Set the service details directly since we're bypassing the normal HTTP parsing
		serviceID:  serviceID,
		serviceQoS: serviceQoS,
		// Create a channel for message observations (required but not used for connection checks)
		messageObservationsChan: make(chan *observation.RequestResponseObservations, 10),
	}

	// Ensure connection observations are broadcast when the check completes
	defer func() {
		logger.Debug().Msg("Broadcasting WebSocket connection check observations")
		websocketRequestCtx.BroadcastWebsocketConnectionRequestObservations()

		// Note: The bridge will close messageObservationsChan during its shutdown process,
		// so we don't need to close it manually here to avoid double-close panics
	}()

	// Set the service ID in the logger since we're not going through the normal initialization
	websocketRequestCtx.logger = websocketRequestCtx.logger.With("service_id", serviceID)

	// Build the QoS context for the WebSocket connection check
	err := websocketRequestCtx.buildQoSContextFromHTTP(nil) // No HTTP request for synthetic connection check
	if err != nil {
		logger.Error().Err(err).Msg("Failed to build QoS context for WebSocket connection check")
		return err
	}

	// Build the protocol context specifically for this endpoint
	err = eph.buildWebSocketProtocolContextForEndpoint(websocketRequestCtx, endpointAddr)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to build protocol context for WebSocket connection check")
		return err
	}

	// For hydrator checks, we only need to verify that the protocol context can be built
	// and that the endpoint is available for WebSocket connections. We don't actually
	// need to establish a full WebSocket connection with message handling.
	// The protocol context setup already validates endpoint connectivity.

	// Start a goroutine that will cancel the context after a brief delay
	// This simulates a connection attempt timeout for testing endpoint responsiveness
	go func() {
		time.Sleep(5 * time.Second) // Give the setup 5 seconds to validate endpoint
		logger.Debug().Msg("Canceling WebSocket connection check after validation delay")
		cancel()
	}()

	completionChan, protocolObservations, err := websocketRequestCtx.protocolCtx.StartWebSocketBridge(
		websocketRequestCtx.context,
		// Pass a nil client connection as the client conn is not used for a synthetic connection check.
		nil,
		websocketRequestCtx, // Pass the context as message processor
		websocketRequestCtx.messageObservationsChan,
	)

	if err != nil && !isContextCancelledError(err) {
		logger.Warn().Err(err).Msg("WebSocket connection check failed during bridge setup")
		// Update protocol observations with the error
		websocketRequestCtx.updateProtocolObservations(protocolObservations)
		return err
	}

	// Update protocol observations with the setup results
	websocketRequestCtx.updateProtocolObservations(protocolObservations)

	// If bridge setup succeeded, wait briefly for potential connection establishment
	// then cancel to avoid maintaining a persistent connection
	if completionChan != nil {
		select {
		case <-completionChan:
			// Bridge completed (likely due to our context cancellation)
		case <-time.After(500 * time.Millisecond):
			// Additional timeout to ensure we don't wait indefinitely
			cancel()
		}
	}

	logger.Info().Msg("WebSocket connection check completed successfully")
	return nil
}

// buildWebSocketProtocolContextForEndpoint builds a WebSocket protocol context
// for a specific endpoint, bypassing the normal endpoint selection process.
func (eph *EndpointHydrator) buildWebSocketProtocolContextForEndpoint(
	websocketRequestCtx *websocketRequestContext,
	endpointAddr protocol.EndpointAddr,
) error {
	logger := websocketRequestCtx.logger.With("method", "buildWebSocketProtocolContextForEndpoint")

	// Build protocol context for the specific endpoint (no selection needed)
	protocolCtx, protocolCtxSetupErrObs, err := eph.Protocol.BuildWebsocketRequestContextForEndpoint(
		websocketRequestCtx.context,
		websocketRequestCtx.serviceID,
		endpointAddr,
		nil,
	)
	if err != nil {
		// Update protocol observations with the error
		websocketRequestCtx.updateProtocolObservations(&protocolCtxSetupErrObs)
		logger.Error().Err(err).Str("endpoint_addr", string(endpointAddr)).Msg("Failed to build protocol context for websocket endpoint")
		return err
	}

	websocketRequestCtx.protocolCtx = protocolCtx
	logger.Debug().Msgf("Successfully built protocol context for websocket endpoint: %s", endpointAddr)
	return nil
}

// isContextCancelledError checks if the error is due to context cancellation,
// which is expected behavior for our connection checks.
func isContextCancelledError(err error) bool {
	return err == context.Canceled || err == context.DeadlineExceeded
}
