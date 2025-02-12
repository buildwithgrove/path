package metrics

import (
	"context"
	"net/http"
	"net/http/pprof"

	"github.com/pokt-network/poktroll/pkg/polylog"
)

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
		logger.Info().Str("endpoint_addr", addr).Msg("starting pprof endpoint to serve go runtime debugging info asynchronously.")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Str("endpoint_addr", addr).Msg("unable to asynchronously start a pprof server for serving go runtime debugging info.")
		}
	}()

	go func() {
		<-ctx.Done()
		logger = logger.With("endpoint_addr", addr)
		logger.Info().Msg("stopping the asynchronous pprof server for serving go runtime debugging info.")
		err := server.Shutdown(ctx)
		if err != nil {
			logger.Error().Err(err).Msg("error stopping the asynchronous go runtime debugging info pprof server.")
		}
	}()
}
