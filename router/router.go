package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/config"
	reqCtx "github.com/buildwithgrove/path/request/context"
)

const (
	imageTagEnvVar  = "IMAGE_TAG"
	defaultImageTag = "development"
)

const userAppIDPathParam = "userAppID"

type (
	router struct {
		mux             *http.ServeMux
		gateway         gateway
		config          config.RouterConfig
		userDataEnabled bool
		logger          polylog.Logger
	}
	gateway interface {
		HandleHTTPServiceRequest(ctx context.Context, httpReq *http.Request, w http.ResponseWriter)
	}
)

/* --------------------------------- Init -------------------------------- */

// NewRouter creates a new router instance
func NewRouter(gateway gateway, config config.RouterConfig, userDataEnabled bool, logger polylog.Logger) *router {
	r := &router{
		mux:             http.NewServeMux(),
		gateway:         gateway,
		config:          config,
		userDataEnabled: userDataEnabled,
		logger:          logger.With("package", "router"),
	}
	r.handleRoutes()
	return r
}

func (r *router) handleRoutes() {
	// GET /healthz - handleHealthz returns a simple health check response
	r.mux.HandleFunc("GET /healthz", methodCheckMiddleware(r.handleHealthz))

	// * /v1... - is the entrypoint for all service requests
	if r.userDataEnabled {
		// * /v1/{userAppID} - handles service requests for a specific user app ID only
		r.mux.HandleFunc(fmt.Sprintf("/v1/{%s}", userAppIDPathParam), r.corsMiddleware(r.handleServiceRequestWithUserAppID))
	} else {
		// * /v1 - handles service requests without any user data handling
		r.mux.HandleFunc("/v1", r.corsMiddleware(r.handleServiceRequest))
	}
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

// GET - /healthz - handleHealthz returns a simple health check response
func (r *router) handleHealthz(w http.ResponseWriter, req *http.Request) {

	imageTag := os.Getenv(imageTagEnvVar)
	if imageTag == "" {
		imageTag = defaultImageTag
	}

	// TODO_IMPROVE: return component ready states for components that must initialize before the service is ready
	responseBytes, err := json.Marshal(struct {
		Status   string `json:"status"`
		ImageTag string `json:"imageTag"`
	}{
		Status:   "ok",
		ImageTag: imageTag,
	})
	if err != nil {
		r.logger.Error().Str("error", err.Error()).Msg("error marshalling health check response")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(responseBytes)
	if err != nil {
		r.logger.Error().Str("error", err.Error()).Msg("error writing health check response")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

// * - /v1/{userAppID} - handleServiceRequestWithUserAppID sets the HTTP details and user app ID
// in the request context and passes it to the gateway handler, which processes the request.
func (r *router) handleServiceRequestWithUserAppID(w http.ResponseWriter, req *http.Request) {
	userAppID := req.PathValue(userAppIDPathParam)
	if userAppID == "" {
		// TODO_TECHDEBT: return proper HTTP error response
		r.logger.Error().Msg("user app ID is required")
		http.Error(w, "user app ID is required", http.StatusBadRequest)
		return
	}

	ctx := reqCtx.SetCtxFromRequest(req.Context(), req, userAppID)

	r.gateway.HandleHTTPServiceRequest(ctx, req, w)
}

// * - /v1 - handleServiceRequest passes the HTTP request to the gateway handler without any user app ID
// or request ID set in the request context. This means to no user authentication or rate limiting is applied.
func (r *router) handleServiceRequest(w http.ResponseWriter, req *http.Request) {
	r.gateway.HandleHTTPServiceRequest(req.Context(), req, w)
}
