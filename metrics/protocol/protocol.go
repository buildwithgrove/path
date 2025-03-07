// Package protocol handles exporting of all protocol-related observation based metrics.
package protocol

import (
	"github.com/buildwithgrove/path/metrics/protocol/morse"
	"github.com/buildwithgrove/path/observation/protocol"
)

// PublishMetrics builds and exports all protocol-related metrics using protocol-level observations.
func PublishMetrics(protocolObservations *protocol.Observations) {
	if protocolObservations == nil {
		return
	}

	// Publish Morse metrics.
	if morseObservations := protocolObservations.GetMorse(); morseObservations != nil {
		morse.PublishMetrics(morseObservations)
	}
	// TODO_MVP(@adshmh): add calls to metric exporter functions for the Shannon protocol.
}
