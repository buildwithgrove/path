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
	endpointLister gateway.EndpointLister,
	relayer *relayer.Relayer,
	qosPublisher gateway.QoSPublisher,
	qosGenerators map[relayer.ServiceID]gateway.QoSEndpointCheckGenerator,
	logger polylog.Logger,
) (*gateway.EndpointHydrator, error) {
	if logger == nil {
		return nil, errors.New("no logger provided")
	}

	if len(qosGenerators) == 0 {
		logger.Warn().Msg("endpoint hydrator is disabled: no service QoS generators are specified")
		return nil, nil
	}

	if qosPublisher == nil {
		return nil, errors.New("endpoint hydrator enabled but no QoS publishers provided")
	}

	if endpointLister == nil {
		return nil, errors.New("endpoint hydrator enabled but no endpointLister instance provided")
	}

	if relayer == nil {
		return nil, errors.New("endpoint hydrator enabled but no relayer provided")
	}

	if qosPublisher == nil {
		return nil, errors.New("endpoint hydrator enabled but no publisher provided")
	}

	endpointHydrator := gateway.EndpointHydrator{
		EndpointLister:       endpointLister,
		Relayer:              relayer,
		QoSPublisher:         qosPublisher,
		ServiceQoSGenerators: qosGenerators,
		Logger:               logger,
	}

	err := endpointHydrator.Start()
	if err != nil {
		return nil, err
	}

	return &endpointHydrator, nil
}
