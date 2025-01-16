package main

import (
	"errors"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/config"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
)

// setupEndpointHydrator initializes and starts an instance of
// the EndpointHydrator matching the configuration settings.
// The EndpointHydrator will not be started if no
// service QoS generators are specified.
func setupEndpointHydrator(
	protocol gateway.Protocol,
	qosPublisher gateway.QoSPublisher,
	qosGenerators map[protocol.ServiceID]gateway.QoSEndpointCheckGenerator,
	hydratorConfig config.EndpointHydratorConfig,
	logger polylog.Logger,
) (*gateway.EndpointHydrator, error) {
	if logger == nil {
		return nil, errors.New("no logger provided")
	}

	if len(qosGenerators) == 0 {
		logger.Warn().Msg("endpoint hydrator is disabled: no service QoS generators are specified")
		return nil, nil
	}

	if protocol == nil {
		return nil, errors.New("endpoint hydrator enabled but no protocol provided")
	}

	if qosPublisher == nil {
		return nil, errors.New("endpoint hydrator enabled but no publisher provided")
	}

	endpointHydrator := gateway.EndpointHydrator{
		Protocol:                protocol,
		QoSPublisher:            qosPublisher,
		ServiceQoSGenerators:    qosGenerators,
		RunInterval:             hydratorConfig.RunInterval,
		MaxEndpointCheckWorkers: hydratorConfig.MaxEndpointCheckWorkers,
		Logger:                  logger,
	}

	err := endpointHydrator.Start()
	if err != nil {
		return nil, err
	}

	return &endpointHydrator, nil
}
