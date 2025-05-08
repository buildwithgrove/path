package main

// TODO_TECHDEBT(@olshansk): Revisit the name `hydrator` to something more appropriate.

import (
	"errors"
	"fmt"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/config"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
)

// TODO_TECHDEBT: Make this configurable.
const defaultProtocolHealthTimeout = 2 * time.Minute

// setupEndpointHydrator
//
// - Initializes and starts an instance of EndpointHydrator matching the configuration settings.
// - Will NOT start the EndpointHydrator if no service QoS generators are specified.
func setupEndpointHydrator(
	cmdLogger polylog.Logger,
	protocolInstance gateway.Protocol,
	qosServices map[protocol.ServiceID]gateway.QoSService,
	metricsReporter gateway.RequestResponseReporter,
	dataReporter gateway.RequestResponseReporter,
	hydratorConfig config.EndpointHydratorConfig,
) (*gateway.EndpointHydrator, error) {
	if cmdLogger == nil {
		return nil, errors.New("no logger provided")
	}
	logger := cmdLogger.With(
		"component", "hydrator",
		"method", "setupEndpointHydrator",
	)

	// Wait for the protocol to become healthy BEFORE configuring and starting the hydrator.
	// - Ensures the protocol instance's configured service IDs are available before hydrator startup.
	err := waitForProtocolHealth(logger, protocolInstance, defaultProtocolHealthTimeout)
	if err != nil {
		return nil, err
	}

	// Get configured service IDs from the protocol instance.
	// - Used to run hydrator checks on all configured service IDs (except those manually disabled by the user).
	configuredServiceIDs := protocolInstance.ConfiguredServiceIDs()

	// Remove any service IDs that are manually disabled by the user.
	for _, disabledServiceID := range hydratorConfig.QoSDisabledServiceIDs {
		// Throw error if any manually disabled service IDs are not found in the protocol's configured service IDs.
		if _, found := configuredServiceIDs[disabledServiceID]; !found {
			return nil, fmt.Errorf("invalid configuration: QoS manually disabled for service ID: %s, but not found in protocol's configured service IDs", disabledServiceID)
		}
		logger.Info().Msgf("QoS manually disabled for service ID: %s", disabledServiceID)
		delete(configuredServiceIDs, disabledServiceID)
	}

	// Ensures the same QoS instance is used by:
	// - Hydrator: generates observations on endpoints
	// - Gateway: selects endpoints (validated using Hydrator's observations)
	hydratorQoSServices := make(map[protocol.ServiceID]gateway.QoSService)
	for serviceID := range configuredServiceIDs {
		serviceQoS, found := qosServices[serviceID]
		if !found {
			logger.Info().Msgf("QoS service not found for service ID: %s. NoOp QoS will be used for this service.", serviceID)
			continue
		}
		hydratorQoSServices[serviceID] = serviceQoS
	}

	if len(hydratorQoSServices) == 0 {
		logger.Warn().Msg("endpoint hydrator is fully disabled: no (zero) active service QoS instances are specified")
		return nil, nil
	}

	if protocolInstance == nil {
		return nil, errors.New("endpoint hydrator enabled but no protocol provided. this should never happen")
	}

	endpointHydrator := gateway.EndpointHydrator{
<<<<<<< HEAD
		Logger:                  cmdLogger.With("component", "hydrator"),
=======
		Logger:                  cmdLogger,
>>>>>>> origin/main
		Protocol:                protocolInstance,
		ActiveQoSServices:       hydratorQoSServices,
		RunInterval:             hydratorConfig.RunInterval,
		MaxEndpointCheckWorkers: hydratorConfig.MaxEndpointCheckWorkers,
		MetricsReporter:         metricsReporter,
		DataReporter:            dataReporter,
	}

	if err := endpointHydrator.Start(); err != nil {
		return nil, err
	}

	return &endpointHydrator, nil
}

// waitForProtocolHealth:
//
// - Blocks until the Protocol reports as healthy
// - Ensures hydrator only starts running once the underlying protocol layer is ready
func waitForProtocolHealth(logger polylog.Logger, protocolInstance gateway.Protocol, timeout time.Duration) error {
	logger.Info().Msg("waitForProtocolHealth: waiting for protocol to become healthy before configuring and starting hydrator")

	start := time.Now()
	for !protocolInstance.IsAlive() {
		if time.Since(start) > timeout {
			return errors.New("waitForProtocolHealth: protocol did not become healthy within timeout")
		}
		logger.Info().Msg("waitForProtocolHealth: protocol not yet healthy, waiting...")
		time.Sleep(1 * time.Second)
	}

	logger.Info().Msg("waitForProtocolHealth: protocol is now healthy, hydrator configuration and startup can proceed")
	return nil
}
