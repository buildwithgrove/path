// package metrics provides functionality for metrics collection and export via Grafana
// As of PR #72, it uses Grafana as the metrics exporting system.
package metrics

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/metrics/qos"
	"github.com/buildwithgrove/path/metrics/protocol"
	"github.com/buildwithgrove/path/observation"
)

// PrometheusMetricsReporter provides the functionality required by the gateway package for publishing metrics on requests and responses.
var _ gateway.RequestResponseReporter = &PrometheusMetricsReporter{}

// PrometheusMetricsReporter provides the functionality required for exporting PATH metrics to Grafana.
type PrometheusMetricsReporter struct {
	Logger polylog.Logger
}

// Publish exports service request and response metrics to Prometheus/Grafana
// Implements the gateway.RequestResponseReporter interface.
func (pmr *PrometheusMetricsReporter) Publish(observations *observation.RequestResponseObservations) {
	// TODO_MVP(@adshmh): complete the set of published metrics to match the notion doc below:
	// https://www.notion.so/buildwithgrove/PATH-Metrics-130a36edfff680febab5d31ee871af87

	// Publish Gateway observations
	publishGatewayMetrics(observations.GetGateway())

	// Publish QoS observations
	qos.PublishQoSMetrics(pmr.Logger, observations.GetQos())

	// Publish Protocol observations
	protocol.PublishMetrics(observations.GetProtocol())
}
