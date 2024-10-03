package main

import (
	"errors"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/relayer"
)

// setupEndpointHydrator initializes and starts an instance of
// the EndpointHydrator matching the configuration settings.
// The EndpointHydrator will not be started if no
// service QoS generators are specified.
func setupEndpointHydrator(
	protocol gateway.Protocol,
	relayer *relayer.Relayer,
	qosPublisher gateway.QoSPublisher,
	qosGenerators map[relayer.ServiceID]gateway.QoSEndpointCheckGenerator,
	logger polylog.Logger,
) error {
	if logger == nil {
		return errors.New("no logger provided")
	}

	if len(qosGenerators) == 0 {
		logger.Warn().Msg("endpoint hydrator is disabled: no service QoS generators are specified")
		return nil
	}

	if qosPublisher == nil {
		return errors.New("endpoint hydrator enabled but no QoS publishers provided")
	}

	if protocol == nil {
		return errors.New("endpoint hydrator enabled but no protocol instance provided")
	}

	if relayer == nil {
		return errors.New("endpoint hydrator enabled but no relayer provided")
	}

	if qosPublisher == nil {
		return errors.New("endpoint hydrator enabled but no publisher provided")
	}

	endpointHydrator := gateway.EndpointHydrator{
		Protocol:             protocol,
		Relayer:              relayer,
		QoSPublisher:         qosPublisher,
		ServiceQoSGenerators: qosGenerators,
		Logger:               logger,
	}

	return endpointHydrator.Start()
}
