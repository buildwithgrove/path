package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const endpointMetrics = "/metrics"

// Starts a metrics server on the given address.
func (pmr *PrometheusMetricsReporter) ServeMetrics(addr string) error {
	// Start the server in a new goroutine
	go func() {
		pmr.Logger.Info().Str("endpoint_addr", addr).Msg("starting Prometheus reporter to serve metrics asynchronously.")
		http.Handle(endpointMetrics, promhttp.Handler())
		if err := http.ListenAndServe(addr, nil); err != nil {
			pmr.Logger.Error().Err(err).Msg("prometheus metrics reporter failed starting server")
			return
		}
	}()

	return nil
}
