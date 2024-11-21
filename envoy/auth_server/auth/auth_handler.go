// The auth package contains the implementation of the Envoy External Authorization gRPC service.
// It is responsible for receiving requests from Envoy and authorizing them based on the GatewayEndpoint
// data stored in the endpointstore package. It receives a check request from Envoy, containing a user ID parsed
// from a JWT in the previous HTTP filter defined in `envoy.yaml`.
package auth

import (
	"context"
	"fmt"
	"strings"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/pokt-network/poktroll/pkg/polylog"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"

	store "github.com/buildwithgrove/path/envoy/auth_server/endpoint_store"
	"github.com/buildwithgrove/path/envoy/auth_server/proto"
)

const (
	pathPrefix = "/v1/"

	reqHeaderEndpointID          = "x-endpoint-id"    // Set on all service requests
	reqHeaderRateLimitEndpointID = "x-rl-endpoint-id" // Set only on service requests that should be rate limited
	reqHeaderRateLimitThroughput = "x-rl-throughput"  // Set only on service requests that should be rate limited

	errBody = `{"code": %d, "message": "%s"}`
)

// The endpointStore interface contains an in-memory store of GatewayEndpoints
// and their associated data from the connected Postgres database.
type endpointStore interface {
	GetGatewayEndpoint(endpointID string) (*proto.GatewayEndpoint, bool)
}

// Enforce that the EndpointStore implements the endpointStore interface.
var _ endpointStore = &store.EndpointStore{}

// struct with check method
type AuthHandler struct {
	// The endpointStore contains an in-memory store of GatewayEndpoints
	// and their associated data from the connected Postgres database.
	EndpointStore endpointStore
	// The authorizers represents a list of authorization types that must
	// pass before a request may be forwarded to the PATH service.
	// Configured in `main.go` and passed to the filter.
	Authorizers map[proto.Auth_AuthType]Authorizer
	Logger      polylog.Logger
}

// Check satisfies the implementation of the Envoy External Authorization gRPC service.
// It performs the following steps:
// - Extracts the endpoint ID from the path
// - Extracts the account user ID from the headers
// - Fetches the GatewayEndpoint from the database
// - Performs all configured authorization checks
// - Returns a response with the HTTP headers set
func (a *AuthHandler) Check(ctx context.Context, checkReq *envoy_auth.CheckRequest,
) (*envoy_auth.CheckResponse, error) {

	// Get the HTTP request
	req := checkReq.GetAttributes().GetRequest().GetHttp()
	if req == nil {
		return getDeniedCheckResponse("HTTP request not found", envoy_type.StatusCode_BadRequest), nil
	}

	// Get the request path
	path := req.GetPath()
	if path == "" {
		return getDeniedCheckResponse("path not provided", envoy_type.StatusCode_BadRequest), nil
	}

	// Get the request headers
	headers := req.GetHeaders()
	if len(headers) == 0 {
		return getDeniedCheckResponse("headers not found", envoy_type.StatusCode_BadRequest), nil
	}

	// Extract the endpoint ID from the path
	endpointID, err := extractEndpointID(path)
	if err != nil {
		return getDeniedCheckResponse(err.Error(), envoy_type.StatusCode_Forbidden), nil
	}

	// Fetch GatewayEndpoint from endpoint store
	gatewayEndpoint, ok := a.getGatewayEndpoint(endpointID)
	if !ok {
		return getDeniedCheckResponse("endpoint not found", envoy_type.StatusCode_NotFound), nil
	}

	// Perform all configured authorization checks
	if err := a.authGatewayEndpoint(headers, gatewayEndpoint); err != nil {
		return getDeniedCheckResponse(err.Error(), envoy_type.StatusCode_Unauthorized), nil
	}

	// Add endpoint ID and rate limiting values to the headers
	// to be passed upstream along the filter chain to the rate limiter.
	httpHeaders := a.getHTTPHeaders(gatewayEndpoint)

	// Return a valid response with the HTTP headers set
	return getOKCheckResponse(httpHeaders), nil
}

/* --------------------------------- Helpers -------------------------------- */

// extractEndpointID extracts the endpoint ID from the URL path.
// The endpoint ID is the part of the path after "/v1/" and is used to identify the GatewayEndpoint.
func extractEndpointID(urlPath string) (string, error) {
	if strings.HasPrefix(urlPath, pathPrefix) {
		if endpointID := strings.TrimPrefix(urlPath, pathPrefix); endpointID != "" {
			return endpointID, nil
		}
		return "", fmt.Errorf("endpoint ID not provided")
	}
	return "", fmt.Errorf("invalid path: %s", urlPath)
}

// getGatewayEndpoint fetches the GatewayEndpoint from the endpoint store and a bool indicating if it was found
func (a *AuthHandler) getGatewayEndpoint(endpointID string) (*proto.GatewayEndpoint, bool) {
	return a.EndpointStore.GetGatewayEndpoint(endpointID)
}

// authGatewayEndpoint performs all configured authorization checks on the request
func (a *AuthHandler) authGatewayEndpoint(headers map[string]string, gatewayEndpoint *proto.GatewayEndpoint) error {
	authType := gatewayEndpoint.GetAuth().GetAuthType()

	// If the endpoint has no authorization requirements, return no error
	if authType == proto.Auth_NO_AUTH {
		return nil
	}

	// If the endpoint has authorization requirements, get the authorizer for the request
	requestAuthorizer, ok := a.Authorizers[authType]
	if !ok {
		return fmt.Errorf("invalid authorization type: %s", authType)
	}

	// Authorize the request using the authorizer configured for the gateway endpoint
	return requestAuthorizer.authorizeRequest(headers, gatewayEndpoint)
}

// getHTTPHeaders sets all HTTP headers required by the PATH services on the request being forwarded
func (a *AuthHandler) getHTTPHeaders(gatewayEndpoint *proto.GatewayEndpoint) []*envoy_core.HeaderValueOption {

	// Set endpoint ID header on all requests
	headers := []*envoy_core.HeaderValueOption{
		{
			Header: &envoy_core.HeaderValue{
				Key:   reqHeaderEndpointID,
				Value: gatewayEndpoint.GetEndpointId(),
			},
		},
	}

	// Set rate limit headers if the gateway endpoint should be rate limited
	if gatewayEndpoint.GetRateLimiting().GetThroughputLimit() > 0 {

		// Set the rate limit endpoint ID header
		headers = append(headers, &envoy_core.HeaderValueOption{
			Header: &envoy_core.HeaderValue{
				Key:   reqHeaderRateLimitEndpointID,
				Value: gatewayEndpoint.GetEndpointId(),
			},
		})

		// Set the account plan type header
		headers = append(headers, &envoy_core.HeaderValueOption{
			Header: &envoy_core.HeaderValue{
				Key:   reqHeaderRateLimitThroughput,
				Value: fmt.Sprintf("%d", gatewayEndpoint.GetRateLimiting().GetThroughputLimit()),
			},
		})

	}

	return headers
}

// getDeniedCheckResponse returns a CheckResponse with a denied status and error message
func getDeniedCheckResponse(err string, httpCode envoy_type.StatusCode) *envoy_auth.CheckResponse {
	return &envoy_auth.CheckResponse{
		Status: &status.Status{
			Code:    int32(codes.PermissionDenied),
			Message: err,
		},
		HttpResponse: &envoy_auth.CheckResponse_DeniedResponse{
			DeniedResponse: &envoy_auth.DeniedHttpResponse{
				Status: &envoy_type.HttpStatus{
					Code: httpCode,
				},
				Body: fmt.Sprintf(errBody, httpCode, err),
			},
		},
	}
}

// getOKCheckResponse returns a CheckResponse with an OK status and the provided headers
func getOKCheckResponse(headers []*envoy_core.HeaderValueOption) *envoy_auth.CheckResponse {
	return &envoy_auth.CheckResponse{
		Status: &status.Status{
			Code:    int32(codes.OK),
			Message: "ok",
		},
		HttpResponse: &envoy_auth.CheckResponse_OkResponse{
			OkResponse: &envoy_auth.OkHttpResponse{
				Headers: headers,
			},
		},
	}
}