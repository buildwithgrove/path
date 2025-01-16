// package metrics provides and encapsulates all the functionality related to exporting metrics.
// As of PR #72, it uses Grafana as the metrics exporting system.
package metrics

import (
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/observation"
)

// GrafanaMetricsReporter provides the functionality required by the gateway package for publishing metrics on requests and their corresponding response.
var _ gateway.RequestResponseReporter = &GrafanaMetricsReporter{}

// GrafanaMetricsReporter provides the functionality required for exporting PATH metrics to Grafana.
type GrafanaMetricsReporter struct{}

// Publish exports the details of the service request and response(s) to Grafana.
// This method implements the gateway.RequestResponseReporter interface.
func (gmr *GrafanaMetricsReporter) Publish(_ *observation.RequestResponseObservations) {
	// TODO_MVP(@adshmh): implement the Publish method below by building and exporting the metrics as specified in the notion doc below:
	// https://www.notion.so/buildwithgrove/PATH-Metrics-130a36edfff680febab5d31ee871af87
}
