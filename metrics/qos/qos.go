// Package qos handles exporting of all qos-related metrics.
package qos

import (
	"github.com/buildwithgrove/path/observation/qos"
)

// PublishMetrics builds and exports all qos-related metrics using qos-level observations.
func PublishQoSMetrics(qosObservations *qos.Observations) {
	if qosObservations == nil {
		return
	}

	if evmObservations := qosObservations.GetEvm(); evmObservations != nil {
		PublishEVMMetrics(evmObservations)
	}
	// TODO_MVP(@adshmh): add calls to metric exporter functions for Solana QoS
}
