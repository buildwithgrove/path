package router

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/config"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/health"
	"github.com/buildwithgrove/path/metrics/devtools"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/request"
)

// Reserve time for system overhead, i.e. time spent on non-business logic operations.
// Examples:
// - time required to read the HTTP request's body.
// - time required to write the prepared HTTP response.
const systemOverheadAllowance = 5 * time.Second

type (
	router struct {
		logger polylog.Logger

		config config.RouterConfig

		mux                           *http.ServeMux
		gateway                       gatewayHandler
		disqualifiedEndpointsReporter disqualifiedEndpointsReporter
		healthChecker                 *health.Checker
	}
	gatewayHandler interface {
		HandleServiceRequest(context.Context, *http.Request, http.ResponseWriter)
	}
	disqualifiedEndpointsReporter interface {
		ReportEndpointStatus(protocol.ServiceID, *http.Request) (devtools.DisqualifiedEndpointResponse, error)
	}
)

/* --------------------------------- Init -------------------------------- */

// NewRouter creates a new router instance
func NewRouter(
	logger polylog.Logger,
	gateway gatewayHandler,
	disqualifiedEndpointsReporter disqualifiedEndpointsReporter,
	healthChecker *health.Checker,
	config config.RouterConfig,
) *router {
	r := &router{
		logger: logger.With("package", "router"),

		config: config,

		mux:                           http.NewServeMux(),
		gateway:                       gateway,
		disqualifiedEndpointsReporter: disqualifiedEndpointsReporter,
		healthChecker:                 healthChecker,
	}
	r.handleRoutes()
	return r
}

func (r *router) handleRoutes() {
	// GET /healthz - returns a JSON health check response indicating the ready status of PATH
	r.mux.HandleFunc("GET /healthz", methodCheckMiddleware(r.healthChecker.HealthzHandler))

	// GET /v1/disqualified_endpoints/{service_id} - returns a JSON list of disqualified endpoints for a given service ID
	r.mux.HandleFunc("GET /disqualified_endpoints", methodCheckMiddleware(r.handleDisqualifiedEndpoints))

	// requestHandlerFn defines the middleware chain for all service requests
	requestHandlerFn := r.corsMiddleware(r.removeGrovePortalPrefixMiddleware(r.handleServiceRequest))

	// */v1/ - handles service requests with trailing slash, including REST services with additional path segments
	r.mux.HandleFunc(gateway.APIVersionPrefix+"/", requestHandlerFn)

	// */v1 - handles service requests without trailing slash
	r.mux.HandleFunc(gateway.APIVersionPrefix, requestHandlerFn)
}

// Start starts the API server on the specified port
func (r *router) Start() (*http.Server, error) {
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", r.config.Port),
		Handler:        r.mux,
		ReadTimeout:    r.config.ReadTimeout,
		WriteTimeout:   r.config.WriteTimeout,
		IdleTimeout:    r.config.IdleTimeout,
		MaxHeaderBytes: r.config.MaxRequestHeaderBytes,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	return server, nil
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

// TODO_TECHDEBT: Rename removeGrovePortalPrefixMiddleware to removePrefixMiddleware
// after removing all of Grove's portal specific logic from the router.
//
// removeGrovePortalPrefixMiddleware removes the API version and endpoint ID prefixes from the URL path
// to allow REST-based services to pass the cleaned path to the selected endpoint.
//
// Example:
//
//	Input:  /v1/{portal_app_id}/{rest_path_1}/{rest_path_2}/...
//	Output: /{rest_path_1}/{rest_path_2}/...
//
// Reference: The `Portal-Application-ID` header is set by the PATH External Auth Server (PEAS):
//   - https://github.com/buildwithgrove/path-external-auth-server/blob/main/auth/auth_handler.go#L173
func (r *router) removeGrovePortalPrefixMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		originalURLPath := req.URL.Path

		// Remove API version prefix (e.g. /v1/{request_path} -> /{request_path)
		// TODO_TECHDEBT(@okdas): Move this logic into envoy and rewrite the URL via helm-charts.
		// Ref: https://github.com/buildwithgrove/helm-charts/pull/96/files
		req.URL.Path = strings.TrimPrefix(req.URL.Path, gateway.APIVersionPrefix)

		// Check if the portal app ID is present in the request header.
		// In production, this should always be set by the PATH External Auth Server (PEAS).
		if portalAppID := req.Header.Get(gateway.HttpHeaderPortalAppID); portalAppID != "" {
			if strings.Contains(req.URL.Path, portalAppID) {
				// Trim the portal app ID prefix from the request path.
				req.URL.Path = strings.TrimPrefix(req.URL.Path, "/"+portalAppID)

				// Remove the portal app ID header from the request headers.
				req.Header.Del(gateway.HttpHeaderPortalAppID)

				r.logger.Debug().
					Str("original_url_path", originalURLPath).
					Str("new_url_path", req.URL.Path).
					Str("portal_app_id", portalAppID).
					Msg("✂️ Removed portal app ID from request path")
			}
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

// handleDisqualifiedEndpoints returns a JSON list of disqualified endpoints
func (r *router) handleDisqualifiedEndpoints(w http.ResponseWriter, req *http.Request) {
	serviceID := protocol.ServiceID(req.Header.Get(request.HTTPHeaderTargetServiceID))
	if serviceID == "" {
		errMsg := `{"error": "400 Bad Request", "message": "Target-Service-Id header is required"}`
		r.logger.Error().Msg(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	disqualifiedEndpointResponses, err := r.disqualifiedEndpointsReporter.ReportEndpointStatus(serviceID, req)
	if err != nil {
		errMsg := fmt.Sprintf(`{"error": "400 Bad Request", "message": "invalid service ID: %v"}`, err)
		r.logger.Error().Msg(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	// Set content type header
	w.Header().Set("Content-Type", "application/json")

	// Set status code
	w.WriteHeader(http.StatusOK)

	// Marshal and write JSON response
	if err := json.NewEncoder(w).Encode(disqualifiedEndpointResponses); err != nil {
		// If encoding fails, log the error but we can't change the status code since it's already written
		r.logger.Error().Err(err).Msg("failed to encode JSON response")
		return
	}
}
