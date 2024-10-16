//go:build auth_server

package server

import (
	"context"
	"testing"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"

	"github.com/buildwithgrove/auth-server/user"
)

func Test_Check(t *testing.T) {
	tests := []struct {
		name               string
		checkReq           *envoy_auth.CheckRequest
		expectedResp       *envoy_auth.CheckResponse
		endpointID         user.EndpointID
		mockEndpointReturn user.GatewayEndpoint
	}{
		{
			name: "should return OK check response if check request is valid and user is authorized to access endpoint with rate limit headers set",
			checkReq: &envoy_auth.CheckRequest{
				Attributes: &envoy_auth.AttributeContext{
					Request: &envoy_auth.AttributeContext_Request{
						Http: &envoy_auth.AttributeContext_HttpRequest{
							Path: "/v1/endpoint_free",
							Headers: map[string]string{
								reqHeaderAccountUserID: "auth0|ulfric_stormcloak",
							},
						},
					},
				},
			},
			expectedResp: &envoy_auth.CheckResponse{
				Status: &status.Status{
					Code:    int32(codes.OK),
					Message: "ok",
				},
				HttpResponse: &envoy_auth.CheckResponse_OkResponse{
					OkResponse: &envoy_auth.OkHttpResponse{
						Headers: []*envoy_core.HeaderValueOption{
							{Header: &envoy_core.HeaderValue{Key: reqHeaderEndpointID, Value: "endpoint_free"}},
							{Header: &envoy_core.HeaderValue{Key: reqHeaderRateLimitEndpointID, Value: "endpoint_free"}},
							{Header: &envoy_core.HeaderValue{Key: reqHeaderRateLimitPlan, Value: "PLAN_FREE"}},
						},
					},
				},
			},
			endpointID: "endpoint_free",
			mockEndpointReturn: user.GatewayEndpoint{
				EndpointID: "endpoint_free",
				Auth: user.Auth{
					AuthorizedUsers: map[user.ProviderUserID]struct{}{
						"auth0|ulfric_stormcloak": {},
					},
				},
				RateLimiting: user.RateLimiting{
					ThroughputLimit: 30,
				},
				UserAccount: user.UserAccount{
					PlanType: user.PlanType("PLAN_FREE"),
				},
			},
		},
		{
			name: "should return OK check response if check request is valid and user is authorized to access endpoint with no rate limit headers set",
			checkReq: &envoy_auth.CheckRequest{
				Attributes: &envoy_auth.AttributeContext{
					Request: &envoy_auth.AttributeContext_Request{
						Http: &envoy_auth.AttributeContext_HttpRequest{
							Path: "/v1/endpoint_unlimited",
							Headers: map[string]string{
								reqHeaderAccountUserID: "auth0|frodo_baggins",
							},
						},
					},
				},
			},
			expectedResp: &envoy_auth.CheckResponse{
				Status: &status.Status{
					Code:    int32(codes.OK),
					Message: "ok",
				},
				HttpResponse: &envoy_auth.CheckResponse_OkResponse{
					OkResponse: &envoy_auth.OkHttpResponse{
						Headers: []*envoy_core.HeaderValueOption{
							{Header: &envoy_core.HeaderValue{Key: reqHeaderEndpointID, Value: "endpoint_unlimited"}},
						},
					},
				},
			},
			endpointID: "endpoint_unlimited",
			mockEndpointReturn: user.GatewayEndpoint{
				EndpointID: "endpoint_unlimited",
				Auth: user.Auth{
					AuthorizedUsers: map[user.ProviderUserID]struct{}{
						"auth0|frodo_baggins": {},
					},
				},
				UserAccount: user.UserAccount{
					PlanType: user.PlanType("PLAN_UNLIMITED"),
				},
			},
		},
		{
			name: "should return denied check response if HTTP request not found",
			checkReq: &envoy_auth.CheckRequest{
				Attributes: &envoy_auth.AttributeContext{},
			},
			expectedResp: &envoy_auth.CheckResponse{
				Status: &status.Status{
					Code:    int32(codes.PermissionDenied),
					Message: "HTTP request not found",
				},
				HttpResponse: &envoy_auth.CheckResponse_DeniedResponse{
					DeniedResponse: &envoy_auth.DeniedHttpResponse{
						Status: &envoy_type.HttpStatus{
							Code: envoy_type.StatusCode_BadRequest,
						},
						Body: `{"code": 400, "message": "HTTP request not found"}`,
					},
				},
			},
		},
		{
			name: "should return denied check response if path not found",
			checkReq: &envoy_auth.CheckRequest{
				Attributes: &envoy_auth.AttributeContext{
					Request: &envoy_auth.AttributeContext_Request{
						Http: &envoy_auth.AttributeContext_HttpRequest{},
					},
				},
			},
			expectedResp: &envoy_auth.CheckResponse{
				Status: &status.Status{
					Code:    int32(codes.PermissionDenied),
					Message: "path not provided",
				},
				HttpResponse: &envoy_auth.CheckResponse_DeniedResponse{
					DeniedResponse: &envoy_auth.DeniedHttpResponse{
						Status: &envoy_type.HttpStatus{
							Code: envoy_type.StatusCode_BadRequest,
						},
						Body: `{"code": 400, "message": "path not provided"}`,
					},
				},
			},
		},
		{
			name: "should return denied check response if headers not found",
			checkReq: &envoy_auth.CheckRequest{
				Attributes: &envoy_auth.AttributeContext{
					Request: &envoy_auth.AttributeContext_Request{
						Http: &envoy_auth.AttributeContext_HttpRequest{
							Path: "/v1/test",
						},
					},
				},
			},
			expectedResp: &envoy_auth.CheckResponse{
				Status: &status.Status{
					Code:    int32(codes.PermissionDenied),
					Message: "headers not found",
				},
				HttpResponse: &envoy_auth.CheckResponse_DeniedResponse{
					DeniedResponse: &envoy_auth.DeniedHttpResponse{
						Status: &envoy_type.HttpStatus{
							Code: envoy_type.StatusCode_BadRequest,
						},
						Body: `{"code": 400, "message": "headers not found"}`,
					},
				},
			},
		},
		{
			name: "should return denied check response if provider user ID not found in JWT sub claim",
			checkReq: &envoy_auth.CheckRequest{
				Attributes: &envoy_auth.AttributeContext{
					Request: &envoy_auth.AttributeContext_Request{
						Http: &envoy_auth.AttributeContext_HttpRequest{
							Path: "/v1/test",
							Headers: map[string]string{
								"x-not-the-right-jwt-claim": "auth0|who_did_this",
							},
						},
					},
				},
			},
			expectedResp: &envoy_auth.CheckResponse{
				Status: &status.Status{
					Code:    int32(codes.PermissionDenied),
					Message: "provider user ID not found in JWT",
				},
				HttpResponse: &envoy_auth.CheckResponse_DeniedResponse{
					DeniedResponse: &envoy_auth.DeniedHttpResponse{
						Status: &envoy_type.HttpStatus{
							Code: envoy_type.StatusCode_Unauthorized,
						},
						Body: `{"code": 401, "message": "provider user ID not found in JWT"}`,
					},
				},
			},
		},
		{
			name: "should return denied check response if path does not have /v1/ prefix",
			checkReq: &envoy_auth.CheckRequest{
				Attributes: &envoy_auth.AttributeContext{
					Request: &envoy_auth.AttributeContext_Request{
						Http: &envoy_auth.AttributeContext_HttpRequest{
							Path: "/invalidprefix/test",
							Headers: map[string]string{
								reqHeaderAccountUserID: "auth0|james_holden",
							},
						},
					},
				},
			},
			expectedResp: &envoy_auth.CheckResponse{
				Status: &status.Status{
					Code:    int32(codes.PermissionDenied),
					Message: "invalid path: /invalidprefix/test",
				},
				HttpResponse: &envoy_auth.CheckResponse_DeniedResponse{
					DeniedResponse: &envoy_auth.DeniedHttpResponse{
						Status: &envoy_type.HttpStatus{
							Code: envoy_type.StatusCode_Forbidden,
						},
						Body: `{"code": 403, "message": "invalid path: /invalidprefix/test"}`,
					},
				},
			},
		},
		{
			name: "should return denied check response if path is invalid",
			checkReq: &envoy_auth.CheckRequest{
				Attributes: &envoy_auth.AttributeContext{
					Request: &envoy_auth.AttributeContext_Request{
						Http: &envoy_auth.AttributeContext_HttpRequest{
							Path: "/v1/",
							Headers: map[string]string{
								reqHeaderAccountUserID: "auth0|paul_atreides",
							},
						},
					},
				},
			},
			expectedResp: &envoy_auth.CheckResponse{
				Status: &status.Status{
					Code:    int32(codes.PermissionDenied),
					Message: "endpoint ID not provided",
				},
				HttpResponse: &envoy_auth.CheckResponse_DeniedResponse{
					DeniedResponse: &envoy_auth.DeniedHttpResponse{
						Status: &envoy_type.HttpStatus{
							Code: envoy_type.StatusCode_Forbidden,
						},
						Body: `{"code": 403, "message": "endpoint ID not provided"}`,
					},
				},
			},
		},
		{
			name: "should return denied check response if gateway endpoint not found",
			checkReq: &envoy_auth.CheckRequest{
				Attributes: &envoy_auth.AttributeContext{
					Request: &envoy_auth.AttributeContext_Request{
						Http: &envoy_auth.AttributeContext_HttpRequest{
							Path: "/v1/endpoint_not_found",
							Headers: map[string]string{
								reqHeaderAccountUserID: "auth0|ellen_ripley",
							},
						},
					},
				},
			},
			expectedResp: &envoy_auth.CheckResponse{
				Status: &status.Status{
					Code:    int32(codes.PermissionDenied),
					Message: "endpoint not found",
				},
				HttpResponse: &envoy_auth.CheckResponse_DeniedResponse{
					DeniedResponse: &envoy_auth.DeniedHttpResponse{
						Status: &envoy_type.HttpStatus{
							Code: envoy_type.StatusCode_NotFound,
						},
						Body: `{"code": 404, "message": "endpoint not found"}`,
					},
				},
			},
			endpointID:         "endpoint_not_found",
			mockEndpointReturn: user.GatewayEndpoint{},
		},
		{
			name: "should return denied check response if user is not authorized to access endpoint",
			checkReq: &envoy_auth.CheckRequest{
				Attributes: &envoy_auth.AttributeContext{
					Request: &envoy_auth.AttributeContext_Request{
						Http: &envoy_auth.AttributeContext_HttpRequest{
							Path: "/v1/endpoint_found",
							Headers: map[string]string{
								reqHeaderAccountUserID: "auth0|ulfric_stormcloak",
							},
						},
					},
				},
			},
			expectedResp: &envoy_auth.CheckResponse{
				Status: &status.Status{
					Code:    int32(codes.PermissionDenied),
					Message: "user is not authorized to access this endpoint",
				},
				HttpResponse: &envoy_auth.CheckResponse_DeniedResponse{
					DeniedResponse: &envoy_auth.DeniedHttpResponse{
						Status: &envoy_type.HttpStatus{
							Code: envoy_type.StatusCode_Unauthorized,
						},
						Body: `{"code": 401, "message": "user is not authorized to access this endpoint"}`,
					},
				},
			},
			endpointID: "endpoint_found",
			mockEndpointReturn: user.GatewayEndpoint{
				EndpointID: "endpoint_found",
				Auth: user.Auth{
					AuthorizedUsers: map[user.ProviderUserID]struct{}{
						"auth0|chrisjen_avasarala": {},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCache := NewMockendpointDataCache(ctrl)
			if test.endpointID != "" {
				mockCache.EXPECT().GetGatewayEndpoint(test.endpointID).Return(test.mockEndpointReturn, test.mockEndpointReturn.EndpointID != "")
			}

			server := &AuthServer{
				Cache: mockCache,
				Authorizers: []Authorizer{
					&ProviderUserIDAuthorizer{},
				},
				Logger: polyzero.NewLogger(),
			}

			resp, err := server.Check(context.Background(), test.checkReq)
			c.NoError(err)
			c.Equal(test.expectedResp, resp)
		})
	}
}
