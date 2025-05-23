package router

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	jsonresponse "github.com/pokt-foundation/utils-go/json-response"
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/config"
	"github.com/buildwithgrove/path/health"
	"github.com/buildwithgrove/path/metrics/devtools"
	"github.com/buildwithgrove/path/protocol"
)

const (
	// apiVersionPrefix is the prefix for the API version and is used by
	// the `removePrefixMiddleware` to remove the API version from the
	// request path that is forwarded to the service endpoint.
	//
	// Example:
	//
	//  /v1/path/segment -> /path/segment
	//  /v1/path -> /path
	apiVersionPrefix = "/v1"

	// reqHeaderEndpointID is the header key for the endpoint ID, and is
	// used by the `removePrefixMiddleware` to ensure the endpoint ID is
	// not present in the request path that is forwarded to the endpoint.
	//
	// Example:
	//
	//  /1a2b3c4d/path/segment -> /path/segment
	//  /1a2b3c4d/path -> /path
	reqHeaderEndpointID = "endpoint-id"

	// Reserve time for system overhead, i.e. time spent on non-business logic operations.
	// Examples:
	// - time required to read the HTTP request's body.
	// - time required to write the prepared HTTP response.
	systemOverheadAllowance = 5 * time.Second
)

type (
	router struct {
		logger polylog.Logger

		config config.RouterConfig

		mux           *http.ServeMux
		gateway       gateway
		dataReporter  dataReporter
		healthChecker *health.Checker
	}
	gateway interface {
		HandleServiceRequest(ctx context.Context, httpReq *http.Request, w http.ResponseWriter)
	}
	dataReporter interface {
		GetSanctionedEndpoints(protocol.ServiceID) devtools.SanctionDetailsResponse
	}
)

/* --------------------------------- Init -------------------------------- */

// NewRouter creates a new router instance
func NewRouter(
	logger polylog.Logger,
	gateway gateway,
	dataReporter dataReporter,
	healthChecker *health.Checker,
	config config.RouterConfig,
) *router {
	r := &router{
		logger: logger.With("package", "router"),

		config: config,

		mux:           http.NewServeMux(),
		gateway:       gateway,
		dataReporter:  dataReporter,
		healthChecker: healthChecker,
	}
	r.handleRoutes()
	return r
}

func (r *router) handleRoutes() {
	// GET /healthz - returns a JSON health check response indicating the ready status of PATH
	r.mux.HandleFunc("GET /healthz", methodCheckMiddleware(r.healthChecker.HealthzHandler))

	// GET /v1/sanctioned_endpoints - returns a JSON list of sanctioned endpoints
	// This will eventually be removed in favour of a metrics-based approach.
	r.mux.HandleFunc("GET /sanctioned_endpoints", r.handleSanctionedEndpoints)

	// GET /v1/sanctioned_endpoints/{service_id} - returns a JSON list of sanctioned endpoints for a given service ID
	// This will eventually be removed in favour of a metrics-based approach.
	r.mux.HandleFunc("GET /sanctioned_endpoints/{service_id}", r.handleSanctionedEndpoints)

	// requestHandlerFn defines the middleware chain for all service requests
	requestHandlerFn := r.corsMiddleware(r.removePrefixMiddleware(r.handleServiceRequest))

	// */v1/ - handles service requests with trailing slash, including REST services with additional path segments
	r.mux.HandleFunc(apiVersionPrefix+"/", requestHandlerFn)

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

// handleServiceRequest:
// 1. Creates timeout context before WriteTimeout expires
// 2. Prevents empty responses on long operations
// 3. Forwards request to gateway handler
func (r *router) handleServiceRequest(w http.ResponseWriter, req *http.Request) {
	// Reserve time for system overhead
	processingTimeout := r.config.WriteTimeout - systemOverheadAllowance

	if processingTimeout <= 0 {
		// Use original context if timeout calculation invalid
		r.gateway.HandleServiceRequest(req.Context(), req, w)
		return
	}

	// Apply timeout to business logic operations
	// DEV_NOTE: Assumes request body read time is negligible.
	// If body read is slow, little time remains for business logic since WriteTimeout resets after body read:
	// https://pkg.go.dev/net/http#Server (ReadTimeout/WriteTimeout)
	reqCtx, cancel := context.WithTimeout(req.Context(), processingTimeout)
	defer cancel()
	r.gateway.HandleServiceRequest(reqCtx, req, w)
}

// handleSanctionedEndpoints returns a JSON list of sanctioned endpoints
// This will eventually be removed in favour of a metrics-based approach.
func (r *router) handleSanctionedEndpoints(w http.ResponseWriter, req *http.Request) {
	serviceID := protocol.ServiceID(req.PathValue("service_id"))

	sanctionedEndpointDetails := r.dataReporter.GetSanctionedEndpoints(serviceID)

	jsonresponse.RespondWithJSON(w, http.StatusOK, sanctionedEndpointDetails)
}
