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
		pmr.Logger.Info().Str("endpoint", addr).Msg("serving metrics")
		http.Handle(endpointMetrics, promhttp.Handler())
		if err := http.ListenAndServe(addr, nil); err != nil {
			pmr.Logger.Error().Err(err).Msg("metrics server failed")
			return
		}
	}()

	return nil
}
