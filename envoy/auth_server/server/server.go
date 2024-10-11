//go:build auth_server

package server

import (
	"context"
	"fmt"
	"strings"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/pokt-network/poktroll/pkg/polylog"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"

	"github.com/buildwithgrove/auth-server/user"
)

const (
	pathPrefix = "/v1/"

	reqHeaderAccountUserID = "x-jwt-user-id" // Defined in envoy.yaml

	reqHeaderEndpointID          = "x-endpoint-id"    // Set on all service requests
	reqHeaderRateLimitEndpointID = "x-rl-endpoint-id" // Set only on service requests that should be rate limited
	reqHeaderRateLimitPlan       = "x-rl-plan"        // Set only on service requests that should be rate limited
)

// The endpointDataCache contains an in-memory cache of GatewayEndpoints
// and their associated data from the connected Postgres database.
type endpointDataCache interface {
	GetGatewayEndpoint(user.EndpointID) (user.GatewayEndpoint, bool)
}

// The Authorizer interface performs requests authorization, for example using
// API key authentication to ensures a downstream (client) request is authorized.
type Authorizer interface {
	authorizeRequest(user.AccountUserID, user.GatewayEndpoint) error
}

// struct with check method
type AuthServer struct {
	// The endpointDataCache contains an in-memory cache of GatewayEndpoints
	// and their associated data from the connected Postgres database.
	Cache endpointDataCache
	// The authorizers represents a list of authorization types that must
	// pass before a request may be forwarded to the PATH service.
	// Configured in `main.go` and passed to the filter.
	Authorizers []Authorizer
	Logger      polylog.Logger
}

// Check satisfies the implementation of the Envoy External Authorization gRPC service.
// It performs the following steps:
// - Extracts the endpoint ID from the path
// - Extracts the account user ID from the headers
// - Fetches the GatewayEndpoint from the database
// - Performs all configured authorization checks
// - Returns a response with the HTTP headers set
func (a *AuthServer) Check(ctx context.Context, checkReq *envoy_auth.CheckRequest,
) (*envoy_auth.CheckResponse, error) {

	req := checkReq.GetAttributes().GetRequest().GetHttp()

	// Get the request path
	path := req.GetPath()
	if path == "" {
		return &envoy_auth.CheckResponse{
			Status: &status.Status{
				Code:    int32(codes.InvalidArgument),
				Message: "path not provided",
			},
		}, nil
	}

	// Get the request headers
	headers := req.GetHeaders()
	if len(headers) == 0 {
		return &envoy_auth.CheckResponse{
			Status: &status.Status{
				Code:    int32(codes.InvalidArgument),
				Message: "headers not found",
			},
		}, nil
	}

	// Get the account user ID from the headers set from the JWT sub claim
	accountUserIDHeader, ok := headers[reqHeaderAccountUserID]
	if !ok || accountUserIDHeader == "" {
		return &envoy_auth.CheckResponse{
			Status: &status.Status{
				Code:    int32(codes.Unauthenticated),
				Message: "account user ID not found in JWT",
			},
		}, nil
	}
	accountUserID := user.AccountUserID(accountUserIDHeader)

	// Extract the endpoint ID from the path
	endpointID, err := extractEndpointID(path)
	if err != nil {
		return &envoy_auth.CheckResponse{
			Status: &status.Status{
				Code:    int32(codes.InvalidArgument),
				Message: err.Error(),
			},
		}, nil
	}

	// If GatewayEndpoint is not found send an error response downstream (client)
	gatewayEndpoint, ok := a.getGatewayEndpoint(endpointID)
	if !ok {
		return &envoy_auth.CheckResponse{
			Status: &status.Status{
				Code:    int32(codes.NotFound),
				Message: "endpoint not found",
			},
		}, nil
	}

	// Perform all configured authorization checks
	if err := a.authGatewayEndpoint(accountUserID, gatewayEndpoint); err != nil {
		return &envoy_auth.CheckResponse{
			Status: &status.Status{
				Code:    int32(codes.PermissionDenied),
				Message: err.Error(),
			},
		}, nil
	}

	// Return a valid response with the HTTP headers set
	return &envoy_auth.CheckResponse{
		Status: &status.Status{
			Code:    int32(codes.OK),
			Message: "ok",
		},
		HttpResponse: &envoy_auth.CheckResponse_OkResponse{
			OkResponse: &envoy_auth.OkHttpResponse{
				// Add endpoint ID and rate limiting values to the headers
				// to be passed along the filter chain to the rate limiter.
				Headers: a.getHTTPHeaders(gatewayEndpoint),
			},
		},
	}, nil
}

/* --------------------------------- Helpers -------------------------------- */

// extractEndpointID extracts the endpoint ID from the URL path.
// The endpoint ID is the part of the path after "/v1/" and is used to identify the GatewayEndpoint.
func extractEndpointID(urlPath string) (user.EndpointID, error) {
	if strings.HasPrefix(urlPath, pathPrefix) {
		if endpointID := strings.TrimPrefix(urlPath, pathPrefix); endpointID != "" {
			return user.EndpointID(endpointID), nil
		}
		return "", fmt.Errorf("endpoint ID not provided")
	}
	return "", fmt.Errorf("invalid path: %s", urlPath)
}

// getGatewayEndpoint fetches the GatewayEndpoint from the database and a bool indicating if it was found
func (a *AuthServer) getGatewayEndpoint(endpointID user.EndpointID) (user.GatewayEndpoint, bool) {
	return a.Cache.GetGatewayEndpoint(endpointID)
}

// authGatewayEndpoint performs all configured authorization checks on the request
func (a *AuthServer) authGatewayEndpoint(accountUserID user.AccountUserID, gatewayEndpoint user.GatewayEndpoint) error {
	for _, auth := range a.Authorizers {
		if err := auth.authorizeRequest(accountUserID, gatewayEndpoint); err != nil {
			return err
		}
	}
	return nil
}

// getHTTPHeaders sets all HTTP headers required by the PATH services on the request being forwarded
func (a *AuthServer) getHTTPHeaders(gatewayEndpoint user.GatewayEndpoint) []*envoy_core.HeaderValueOption {

	// Set endpoint ID header on all requests
	headers := []*envoy_core.HeaderValueOption{
		{
			Header: &envoy_core.HeaderValue{
				Key:   reqHeaderEndpointID,
				Value: string(gatewayEndpoint.EndpointID),
			},
		},
	}

	// Set rate limit headers if the gateway endpoint should be rate limited
	if gatewayEndpoint.RateLimiting.ThroughputLimit > 0 {

		// Set the rate limit endpoint ID header
		headers = append(headers, &envoy_core.HeaderValueOption{
			Header: &envoy_core.HeaderValue{
				Key:   reqHeaderRateLimitEndpointID,
				Value: string(gatewayEndpoint.EndpointID),
			},
		})

		// Set the account plan type header
		headers = append(headers, &envoy_core.HeaderValueOption{
			Header: &envoy_core.HeaderValue{
				Key:   reqHeaderRateLimitPlan,
				Value: string(gatewayEndpoint.UserAccount.PlanType),
			},
		})

	}

	return headers
}
