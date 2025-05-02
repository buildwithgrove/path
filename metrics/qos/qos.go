// Package qos handles exporting of all qos-related metrics.
package qos

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/metrics/qos/evm"
	"github.com/buildwithgrove/path/observation/qos"
)

// PublishMetrics builds and exports all qos-related metrics using qos-level observations.
func PublishQoSMetrics(
	logger polylog.Logger,
	qosObservations *qos.Observations,
) {
	hydratedLogger := logger.With("method", "PublishQoSMetrics")

	if qosObservations == nil {
		hydratedLogger.Warn().Msg("received nil set of QoS observations.")
		return
	}

	var hasProcessedObservations bool

	// Publish EVM metrics.
	if evmObservations := qosObservations.GetEvm(); evmObservations != nil {
		hasProcessedObservations = true
		evm.PublishMetrics(hydratedLogger, evmObservations)
	}
	// TODO_MVP(@adshmh): add calls to metric exporter functions for Solana QoS

	// Log warning if no matching observation types were found
	if !hasProcessedObservations {
		hydratedLogger.Warn().Msgf("supplied observations do not match any known QoS service: %+v", qosObservations)
	}
}
