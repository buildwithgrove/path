package metrics

import (
	"context"
	"net"
	"net/http"
	"net/http/pprof"

	"cosmossdk.io/depinject"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/pokt-network/poktroll/pkg/polylog"
)

// Starts a metrics server on the given address.
func (pmr *PrometheusMetricsReporter) ServeMetrics(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		pmr.Logger.Error().Err(err).Msg("failed to listen on address for metrics")
		return err
	}

	// If no error, start the server in a new goroutine
	go func() {
		pmr.Logger.Info().Str("endpoint", addr).Msg("serving metrics")
		if err := http.Serve(ln, promhttp.Handler()); err != nil {
			pmr.Logger.Error().Err(err).Msg("metrics server failed")
			return
		}
	}()

	return nil
}

// ServePprof Starts a pprof server on the given address.
func ServePprof(ctx context.Context, logger polylog.Logger, addr string) {
	pprofMux := http.NewServeMux()
	pprofMux.HandleFunc("/debug/pprof/", pprof.Index)
	pprofMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	pprofMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	pprofMux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	server := &http.Server{
		Addr:    addr,
		Handler: pprofMux,
	}
	// If no error, start the server in a new goroutine
	go func() {
		logger.Info().Str("endpoint", addr).Msg("starting a pprof endpoint")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Str("endpoint", addr).Msg("unable to start a pprof endpoint")
		}
	}()

	go func() {
		<-ctx.Done()
		logger.Info().Str("endpoint", addr).Msg("stopping a pprof endpoint")
		_ = server.Shutdown(ctx)
	}()
}
