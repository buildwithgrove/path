package auth

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

	"github.com/buildwithgrove/auth-server/proto"
)

func Test_Check(t *testing.T) {
	tests := []struct {
		name               string
		checkReq           *envoy_auth.CheckRequest
		expectedResp       *envoy_auth.CheckResponse
		endpointID         string
		mockEndpointReturn *proto.GatewayEndpoint
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
			mockEndpointReturn: &proto.GatewayEndpoint{
				EndpointId: "endpoint_free",
				Auth: &proto.Auth{
					AuthorizedUsers: map[string]*proto.Empty{
						"auth0|ulfric_stormcloak": {},
					},
				},
				RateLimiting: &proto.RateLimiting{
					ThroughputLimit: 30,
				},
				UserAccount: &proto.UserAccount{
					PlanType: "PLAN_FREE",
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
			mockEndpointReturn: &proto.GatewayEndpoint{
				EndpointId: "endpoint_unlimited",
				Auth: &proto.Auth{
					AuthorizedUsers: map[string]*proto.Empty{
						"auth0|frodo_baggins": {},
					},
				},
				UserAccount: &proto.UserAccount{
					PlanType: "PLAN_UNLIMITED",
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
			mockEndpointReturn: &proto.GatewayEndpoint{},
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
			mockEndpointReturn: &proto.GatewayEndpoint{
				EndpointId: "endpoint_found",
				Auth: &proto.Auth{
					AuthorizedUsers: map[string]*proto.Empty{
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
				mockCache.EXPECT().GetGatewayEndpoint(test.endpointID).Return(test.mockEndpointReturn, test.mockEndpointReturn.EndpointId != "")
			}

			authHandler := &AuthHandler{
				Cache: mockCache,
				Authorizers: []Authorizer{
					&ProviderUserIDAuthorizer{},
				},
				Logger: polyzero.NewLogger(),
			}

			resp, err := authHandler.Check(context.Background(), test.checkReq)
			c.NoError(err)
			c.Equal(test.expectedResp, resp)
		})
	}
}
