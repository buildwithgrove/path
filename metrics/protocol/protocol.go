// Package protocol handles exporting of all protocol-related observation based metrics.
package protocol

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/metrics/protocol/morse"
	"github.com/buildwithgrove/path/metrics/protocol/shannon"
	"github.com/buildwithgrove/path/observation/protocol"
)

// PublishMetrics builds and exports all protocol-related metrics using protocol-level observations.
func PublishMetrics(
	logger polylog.Logger,
	protocolObservations *protocol.Observations,
) {
	hydratedLogger := logger.With("method", "PublishProtocolMetrics")
	if protocolObservations == nil {
		hydratedLogger.Warn().Msg("SHOULD NEVER HAPPEN: received nil set of Protocol observations.")
		return
	}

	// Publish Morse metrics.
	if morseObservations := protocolObservations.GetMorse(); morseObservations != nil {
		morse.PublishMetrics(morseObservations)
		return
	}

	// Publish Shannon metrics.
	if shannonObservations := protocolObservations.GetShannon(); shannonObservations != nil {
		shannon.PublishMetrics(logger, shannonObservations)
		return
	}

	// Log warning if no matching observation types were found
	hydratedLogger.Warn().Msgf("SHOULD NEVER HAPPEN: supplied observations do not match any known Protocol: %+v", protocolObservations)
}
