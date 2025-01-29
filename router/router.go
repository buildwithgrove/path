package router

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/config"
	"github.com/buildwithgrove/path/health"
)

const (
	// apiVersionPrefix is the prefix for the API version and is used by
	// the `removePrefixMiddleware` to remove the API version from the
	// request path that is forwarded to the service endpoint.
	// eg. /v1/path/segment -> /path/segment
	apiVersionPrefix = "/v1"

	// reqHeaderEndpointID is the header key for the endpoint ID, and is
	// used by the `removePrefixMiddleware` to ensure the endpoint ID is
	// not present in the request path that is forwarded to the endpoint.
	// eg. s/1a2b3c4d/path/segment -> /path/segment
	reqHeaderEndpointID = "endpoint-id"
)

type (
	router struct {
		logger polylog.Logger

		config config.RouterConfig

		mux           *http.ServeMux
		gateway       gateway
		healthChecker *health.Checker
	}
	gateway interface {
		HandleHTTPServiceRequest(ctx context.Context, httpReq *http.Request, w http.ResponseWriter)
	}
)

/* --------------------------------- Init -------------------------------- */

// NewRouter creates a new router instance
func NewRouter(logger polylog.Logger, gateway gateway, healthChecker *health.Checker, config config.RouterConfig) *router {
	r := &router{
		logger: logger.With("package", "router"),

		config: config,

		mux:           http.NewServeMux(),
		gateway:       gateway,
		healthChecker: healthChecker,
	}
	r.handleRoutes()
	return r
}

func (r *router) handleRoutes() {
	// GET /healthz - returns a JSON health check response indicating the ready status of PATH
	r.mux.HandleFunc("GET /healthz", methodCheckMiddleware(r.healthChecker.HealthzHandler))

	requestHandlerFn := r.corsMiddleware(r.removePrefixMiddleware(r.handleServiceRequest))

	// */v1/ - handles service requests with trailing slash, including REST services with additional path segments
	r.mux.HandleFunc(fmt.Sprintf("%s/", apiVersionPrefix), requestHandlerFn)

	// */v1 - handles service requests without trailing slash
	r.mux.HandleFunc(apiVersionPrefix, requestHandlerFn)
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

// removePrefixMiddleware removes the API version and endpoint ID prefixes from the URL path
// to allow REST-based services to pass the cleaned path to the selected endpoint.
//
// Example:
//
//	Input:  /v1/endpoint/path/123
//	Output: /path/123
func (r *router) removePrefixMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Remove API version prefix (e.g. /v1/path -> /path)
		req.URL.Path = strings.TrimPrefix(req.URL.Path, apiVersionPrefix)

		// Remove endpoint ID prefix if present (e.g. /1a2b3c4d/path -> /path)
		if endpointID := req.Header.Get(reqHeaderEndpointID); endpointID != "" {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/"+endpointID)
			delete(req.Header, reqHeaderEndpointID)
		}

		next(w, req)
	}
}

/* --------------------------------- Handlers -------------------------------- */

// handleServiceRequest sets the request ID and HTTP details in the request context
// from the HTTP request and passes it to the gateway handler, which processes the request.
func (r *router) handleServiceRequest(w http.ResponseWriter, req *http.Request) {
	r.gateway.HandleHTTPServiceRequest(req.Context(), req, w)
}
