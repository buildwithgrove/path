package main

import (
	"context"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/metrics"
)

const (
	// TODO_TECHDEBT(@adshmh): Support configurable pprof server address/port.
	//
	// pprofAddr is the address at which pprof server will be listening.
	// This address is selected based on the following link's examples:
	// https://pkg.go.dev/net/http/pprof
	pprofAddr = ":6060"

	// TODO_TECHDEBT(@adshmh): Support configurable metrics server address/port.
	//
	// prometheusMetricsServerAddr is the address at which the prometheus metrics server will be listening.
	prometheusMetricsServerAddr = ":9090"
)

// setupMetricsServer initializes and starts the Prometheus metrics server at the default address.
func setupMetricsServer(logger polylog.Logger) (*metrics.PrometheusMetricsReporter, error) {
	pmr := &metrics.PrometheusMetricsReporter{
		Logger: logger,
	}

	if err := pmr.ServeMetrics(prometheusMetricsServerAddr); err != nil {
		return nil, err
	}

	return pmr, nil
}

// setupPprofServer starts the metric package's pprof server, at the default address.
func setupPprofServer(ctx context.Context, logger polylog.Logger) {
	metrics.ServePprof(ctx, logger, pprofAddr)
}
