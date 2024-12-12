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

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
)

const (
	// TODO_MVP(@commoddity): Eliminate the hard-coding of this prefix for all endpoints.
	// See this thread for more information: https://github.com/buildwithgrove/path/pull/52/files#r1859632536
	pathPrefix = "/v1/"

	reqHeaderEndpointID          = "x-endpoint-id"    // Set on all service requests
	reqHeaderRateLimitEndpointID = "x-rl-endpoint-id" // Set only on service requests that should be rate limited
	reqHeaderRateLimitThroughput = "x-rl-throughput"  // Set only on service requests that should be rate limited

	errBody = `{"code": %d, "message": "%s"}`
)

// The EndpointStore interface contains an in-memory store of GatewayEndpoints
// and their associated data from the PADS (PATH Auth Data Server).
// See: https://github.com/buildwithgrove/path-auth-data-server
//
// It used to allow fast lookups of authorization data for PATH when processing requests.
type EndpointStore interface {
	GetGatewayEndpoint(endpointID string) (*proto.GatewayEndpoint, bool)
}

// The AuthHandler struct contains the methods for processing requests from Envoy,
// primarily the Check method that is called by Envoy for each request.
type AuthHandler struct {
	// The EndpointStore contains an in-memory store of GatewayEndpoints
	// and their associated data from the PADS (PATH Auth Data Server).
	EndpointStore EndpointStore

	// The authorizers to be used for the request
	APIKeyAuthorizer Authorizer
	JWTAuthorizer    Authorizer

	Logger polylog.Logger
}

// Check satisfies the implementation of the Envoy External Authorization gRPC service.
// It performs the following steps:
// - Extracts the endpoint ID from the path
// - Extracts the account user ID from the headers
// - Fetches the GatewayEndpoint from the database
// - Performs all configured authorization checks
// - Returns a response with the HTTP headers set
func (a *AuthHandler) Check(
	ctx context.Context,
	checkReq *envoy_auth.CheckRequest,
) (*envoy_auth.CheckResponse, error) {
	a.Logger.Info().Str("path", checkReq.GetAttributes().GetRequest().GetHttp().GetPath()).Msg("path")

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
	// Get the authorization type for the gateway endpoint
	authType := gatewayEndpoint.GetAuth().GetAuthType()

	switch authType.(type) {
	case *proto.Auth_NoAuth:
		return nil // If the endpoint has no authorization requirements, return no error

	case *proto.Auth_StaticApiKey:
		return a.APIKeyAuthorizer.authorizeRequest(headers, gatewayEndpoint)

	case *proto.Auth_Jwt:
		return a.JWTAuthorizer.authorizeRequest(headers, gatewayEndpoint)

	default:
		return fmt.Errorf("invalid authorization type")
	}
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
