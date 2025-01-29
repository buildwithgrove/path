// package metrics provides and encapsulates all the functionality related to exporting metrics.
// As of PR #72, it uses Grafana as the metrics exporting system.
package metrics

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/metrics/qos"
	"github.com/buildwithgrove/path/observation"
)

// PrometheusMetricsReporter provides the functionality required by the gateway package for publishing metrics on requests and their corresponding response.
var _ gateway.RequestResponseReporter = &PrometheusMetricsReporter{}

// PrometheusMetricsReporter provides the functionality required for exporting PATH metrics to Grafana.
type PrometheusMetricsReporter struct {
	Logger polylog.Logger
}

// Publish exports the details of the service request and response(s) to Grafana.
// Implements the gateway.RequestResponseReporter interface.
func (pmr *PrometheusMetricsReporter) Publish(observations *observation.RequestResponseObservations) {
	// TODO_MVP(@adshmh): complete the set of published metrics to match the notion doc below:
	// https://www.notion.so/buildwithgrove/PATH-Metrics-130a36edfff680febab5d31ee871af87

	// Publish Gateway observations
	publishGatewayMetrics(observations.GetGateway())

	// Publish QoS observations
	qos.PublishQoSMetrics(observations.GetQos())
}
