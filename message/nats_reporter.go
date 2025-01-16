package message

import (
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/observation"
)

// NATSMetricsReporter provides the functionality required by the gateway package for publishing metrics on requests and their corresponding response.
// It uses NATS as its messaging platform.
var _ gateway.RequestResponseReporter = &NATSMetricsReporter{}

// NATSMetricsReporter provides the functionality required for exporting PATH metrics to NATS messaging platform.
type NATSMetricsReporter struct{}

// Publish exports the details of the service request and response(s) to NATS messaging system.
// Any entity interested in this data, e.g. the data pipeline for PATH once it is built, should subscribe to NATS to receive the exported data.
// This method implements the gateway.RequestResponseReporter interface.
func (nmr *NATSMetricsReporter) Publish(_ *observation.RequestResponseObservations) {
	// TODO_MVP(@adshmh): implement the Publish method below by building and exporting the metrics as specified in the notion doc below:
	// https://www.notion.so/buildwithgrove/PATH-Metrics-130a36edfff680febab5d31ee871af87
}
