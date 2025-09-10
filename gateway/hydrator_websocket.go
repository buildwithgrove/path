package gateway

import (
	"context"
	"sync"
	"time"

	"github.com/buildwithgrove/path/protocol"
)

// websocketCheckInterval is the interval at which WebSocket connection checks are performed.
const websocketCheckInterval = 10 * time.Minute

// runWebSocketChecks performs WebSocket connection checks for all services and endpoints.
func (eph *EndpointHydrator) runWebSocketChecks() {
	logger := eph.Logger.With(
		"services_count", len(eph.ActiveQoSServices),
		"check_type", "websocket",
	)
	logger.Info().Msg("Running WebSocket Endpoint Hydrator checks")

	// TODO_TECHDEBT: ensure every outgoing request (or the goroutine checking a service ID)
	// has a timeout set.
	var wg sync.WaitGroup

	for svcID, svcQoS := range eph.ActiveQoSServices {
		logger := logger.With("service_id", string(svcID))

		// Skip if WebSocket checks are not enabled for this service.
		if !svcQoS.CheckWebsocketConnection() {
			logger.Debug().Msg("Service is not configured to run WebSocket checks. Skipping.")
			continue
		}

		wg.Add(1)
		go func(serviceID protocol.ServiceID, serviceQoS QoSService) {
			defer wg.Done()

			logger := eph.Logger.With("serviceID", serviceID)

			err := eph.performWebSocketChecks(serviceID, serviceQoS)
			if err != nil {
				logger.Warn().Err(err).Msg("failed to run WebSocket checks for service")
				return
			}

			logger.Info().Msg("successfully completed WebSocket checks for service")
		}(svcID, svcQoS)
	}
	wg.Wait()
}

// performWebSocketChecks performs WebSocket connection checks for a specific service.
func (eph *EndpointHydrator) performWebSocketChecks(serviceID protocol.ServiceID, serviceQoS QoSService) error {
	logger := eph.Logger.With(
		"method", "performWebSocketChecks",
		"service_id", string(serviceID),
	)

	// Passing a nil as the HTTP request, because we assume the hydrator uses "Centralized Operation Mode".
	// TODO_TECHDEBT(@adshmh): support specifying the app(s) used for sending/signing synthetic relay requests by the hydrator.
	// TODO_FUTURE(@adshmh): consider publishing observations if endpoint lookup fails.
	availableEndpoints, _, err := eph.AvailableWebsocketEndpoints(context.TODO(), serviceID, nil)
	if err != nil || len(availableEndpoints) == 0 {
		// No session found or no endpoints available for service: skip.
		logger.Warn().Msg("no session found or no endpoints available for service when running WebSocket hydrator checks.")
		// do NOT return an error: hydrator and PATH should not report unhealthy status if a single service is unavailable.
		return nil
	}

	logger = logger.With("number_of_endpoints", len(availableEndpoints))

	// Prepare a channel that will keep track of all the parallel async job to perform WebSocket checks on every endpoint.
	endpointCheckChan := make(chan protocol.EndpointAddr, len(availableEndpoints))

	var wgEndpoints sync.WaitGroup
	for range eph.MaxEndpointCheckWorkers {
		wgEndpoints.Add(1)

		go func() {
			defer wgEndpoints.Done()

			for endpointAddr := range endpointCheckChan {
				endpointLogger := logger.With("endpoint_addr", string(endpointAddr))
				endpointLogger.Info().Msg("Running WebSocket connection check for endpoint")

				err := eph.performWebSocketConnectionCheck(serviceID, endpointAddr)
				if err != nil {
					endpointLogger.Warn().Err(err).Msg("WebSocket connection check failed")
					// Continue with other endpoints even if one WebSocket check fails
				}
			}
		}()
	}

	// Kick off the workers above for every unique endpoint.
	for _, endpointAddr := range availableEndpoints {
		endpointCheckChan <- endpointAddr
	}

	close(endpointCheckChan)

	// Wait for all workers to finish processing the endpoints.
	wgEndpoints.Wait()

	// TODO_FUTURE: publish aggregated WebSocket check reports
	return nil
}

// performWebSocketConnectionCheck performs a WebSocket connection establishment check
// for the given endpoint. It performs a simplified version of the websocket bridge connection process
// to determine if an endpoint can support websocket connections.
func (eph *EndpointHydrator) performWebSocketConnectionCheck(
	serviceID protocol.ServiceID,
	endpointAddr protocol.EndpointAddr,
) error {
	// Create a context with a short timeout for the WebSocket connection check
	// This ensures we don't wait too long for unresponsive endpoints
	checkTimeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()

	obs := eph.Protocol.CheckWebsocketConnection(ctx, serviceID, endpointAddr)
	if obs != nil {
		eph.Protocol.ApplyWebSocketObservations(obs)
	}

	return nil
}
