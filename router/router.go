package router

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/config"
)

const (
	imageTagEnvVar  = "IMAGE_TAG"
	defaultImageTag = "development"
)

type (
	router struct {
		mux         *http.ServeMux
		gateway     gateway
		healthCheck *healthCheck
		config      config.RouterConfig
		logger      polylog.Logger
	}
	gateway interface {
		HandleHTTPServiceRequest(ctx context.Context, httpReq *http.Request, w http.ResponseWriter)
	}
)

/* --------------------------------- Init -------------------------------- */

// NewRouter creates a new router instance
func NewRouter(gateway gateway, healthCheckComponents []HealthCheckComponent, config config.RouterConfig, logger polylog.Logger) *router {
	r := &router{
		mux:     http.NewServeMux(),
		gateway: gateway,
		healthCheck: &healthCheck{
			components: healthCheckComponents,
			logger:     logger,
		},
		config: config,
		logger: logger.With("package", "router"),
	}
	r.handleRoutes()
	return r
}

func (r *router) handleRoutes() {
	// GET /healthz - handleHealthz returns a simple health check response
	r.mux.HandleFunc("GET /healthz", methodCheckMiddleware(r.healthCheck.healthCheckHandler))

	// * /v1 - handles service requests
	r.mux.HandleFunc("/v1", r.corsMiddleware(r.handleServiceRequest))
}

// Start starts the API server on the specified port
func (r *router) Start() error {
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", r.config.Port),
		Handler:        r.mux,
		ReadTimeout:    r.config.ReadTimeout,
		WriteTimeout:   r.config.WriteTimeout,
		IdleTimeout:    r.config.IdleTimeout,
		MaxHeaderBytes: r.config.MaxRequestBodySize,
	}

	r.logger.Info().Msgf("PATH gateway running on port %d", r.config.Port)

	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

/* --------------------------------- Middleware -------------------------------- */

// methodCheckMiddleware ensures that only GET requests are allowed for the wrapped handler
func methodCheckMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed: only GET requests are allowed", http.StatusMethodNotAllowed)
			return
		}
		next(w, r)
	}
}

// TODO_IMPROVE: gather the CORS config from the config YAML
func (r *router) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, solana-client")
		if r.Method == "OPTIONS" {
			// Handle preflight request, which is necessary for CORS to work.
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

/* --------------------------------- Handlers -------------------------------- */

// * - /v1 - handleServiceRequest sets the request ID and HTTP details in the request context
// from the HTTP request and passes it to the gateway handler, which processes the request.
func (r *router) handleServiceRequest(w http.ResponseWriter, req *http.Request) {
	r.gateway.HandleHTTPServiceRequest(req.Context(), req, w)
}
