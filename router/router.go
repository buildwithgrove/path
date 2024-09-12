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
	"github.com/buildwithgrove/path/user"
)

const (
	imageTagEnvVar  = "IMAGE_TAG"
	defaultImageTag = "development"
)

const endpointIDPathParam = "endpoint_id"

type (
	router struct {
		mux     *http.ServeMux
		gateway gateway
		config  config.RouterConfig
		logger  polylog.Logger
	}
	gateway interface {
		HandleHTTPServiceRequest(ctx context.Context, httpReq *http.Request, w http.ResponseWriter)
	}
)

type RouterParams struct {
	Gateway         gateway
	Config          config.RouterConfig
	UserDataEnabled bool
	Logger          polylog.Logger
}

/* --------------------------------- Init -------------------------------- */

// NewRouter creates a new router instance
func NewRouter(params RouterParams) *router {
	r := &router{
		mux:     http.NewServeMux(),
		gateway: params.Gateway,
		config:  params.Config,
		logger:  params.Logger.With("package", "router"),
	}
	r.handleRoutes(params.UserDataEnabled)
	return r
}

func (r *router) handleRoutes(userDataEnabled bool) {
	// GET /healthz - handleHealthz returns a simple health check response
	r.mux.HandleFunc("GET /healthz", methodCheckMiddleware(r.handleHealthz))

	// * /v1... - is the entrypoint for all service requests
	if userDataEnabled {
		// * /v1/{endpoint_id} - handles service requests for a specific gateway endpoint ID only
		r.mux.HandleFunc(fmt.Sprintf("/v1/{%s}", endpointIDPathParam), r.corsMiddleware(r.handleServiceRequest))
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

// handleServiceRequest sets the request ID and HTTP details in the request context
// from the HTTP request and passes it to the gateway handler, which processes the request.
// * - /v1  - user data not enabled: handles requests for all gateway endpoint IDs
// * - /v1/{endpoint_id} - user data enabled: handles requests for a specific gateway endpoint ID only
func (r *router) handleServiceRequest(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	// if user data is enabled set the gateway endpoint ID and HTTP details in request ctx
	if endpointID := req.PathValue(endpointIDPathParam); endpointID != "" {
		ctx = reqCtx.SetCtxFromRequest(ctx, req, user.EndpointID(endpointID))
	}

	r.gateway.HandleHTTPServiceRequest(ctx, req, w)
}
