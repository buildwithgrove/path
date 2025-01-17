package auth

import (
	"context"
	"fmt"
	"testing"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
)

func Test_Check(t *testing.T) {
	tests := []struct {
		name                string
		checkReq            *envoy_auth.CheckRequest
		expectedResp        *envoy_auth.CheckResponse
		endpointID          string
		endpointIDExtractor EndpointIDExtractor
		mockEndpointReturn  *proto.GatewayEndpoint
	}{
		{
			name: "should return OK check response if check request is valid and user is authorized to access endpoint with rate limit headers set",
			checkReq: &envoy_auth.CheckRequest{
				Attributes: &envoy_auth.AttributeContext{
					Request: &envoy_auth.AttributeContext_Request{
						Http: &envoy_auth.AttributeContext_HttpRequest{
							Path: "/v1/endpoint_free",
							Headers: map[string]string{
								reqHeaderJWTUserID: "auth0|ulfric_stormcloak",
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
							{Header: &envoy_core.HeaderValue{Key: reqHeaderRateLimitThroughput, Value: "30"}},
						},
					},
				},
			},
			endpointIDExtractor: &URLPathExtractor{},
			endpointID:          "endpoint_free",
			mockEndpointReturn: &proto.GatewayEndpoint{
				EndpointId: "endpoint_free",
				Auth: &proto.Auth{
					AuthType: &proto.Auth_Jwt{
						Jwt: &proto.JWT{
							AuthorizedUsers: map[string]*proto.Empty{
								"auth0|ulfric_stormcloak": {},
							},
						},
					},
				},
				RateLimiting: &proto.RateLimiting{
					ThroughputLimit: 30,
				},
				Metadata: &proto.Metadata{
					AccountId: "account_1",
					PlanType:  "PLAN_FREE",
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
								reqHeaderJWTUserID: "auth0|frodo_baggins",
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
			endpointIDExtractor: &URLPathExtractor{},
			endpointID:          "endpoint_unlimited",
			mockEndpointReturn: &proto.GatewayEndpoint{
				EndpointId: "endpoint_unlimited",
				Auth: &proto.Auth{
					AuthType: &proto.Auth_Jwt{
						Jwt: &proto.JWT{
							AuthorizedUsers: map[string]*proto.Empty{
								"auth0|frodo_baggins": {},
							},
						},
					},
				},
				Metadata: &proto.Metadata{
					AccountId: "account_2",
					PlanType:  "PLAN_UNLIMITED",
				},
			},
		},
		{
			name: "should return ok check response if endpoint requires API key auth",
			checkReq: &envoy_auth.CheckRequest{
				Attributes: &envoy_auth.AttributeContext{
					Request: &envoy_auth.AttributeContext_Request{
						Http: &envoy_auth.AttributeContext_HttpRequest{
							Path: "/v1/api_key_endpoint",
							Headers: map[string]string{
								reqHeaderAPIKey: "api_key_good",
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
							{Header: &envoy_core.HeaderValue{Key: reqHeaderEndpointID, Value: "api_key_endpoint"}},
						},
					},
				},
			},
			endpointIDExtractor: &URLPathExtractor{},
			endpointID:          "api_key_endpoint",
			mockEndpointReturn: &proto.GatewayEndpoint{
				EndpointId: "api_key_endpoint",
				Auth: &proto.Auth{
					AuthType: &proto.Auth_StaticApiKey{
						StaticApiKey: &proto.StaticAPIKey{
							ApiKey: "api_key_good",
						},
					},
				},
			},
		},
		{
			name: "should return ok check response if endpoint requires JWT auth",
			checkReq: &envoy_auth.CheckRequest{
				Attributes: &envoy_auth.AttributeContext{
					Request: &envoy_auth.AttributeContext_Request{
						Http: &envoy_auth.AttributeContext_HttpRequest{
							Path: "/v1/jwt_endpoint",
							Headers: map[string]string{
								reqHeaderJWTUserID: "auth0|yennefer_of_vengerberg",
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
							{Header: &envoy_core.HeaderValue{Key: reqHeaderEndpointID, Value: "jwt_endpoint"}},
						},
					},
				},
			},
			endpointIDExtractor: &URLPathExtractor{},
			endpointID:          "jwt_endpoint",
			mockEndpointReturn: &proto.GatewayEndpoint{
				EndpointId: "jwt_endpoint",
				Auth: &proto.Auth{
					AuthType: &proto.Auth_Jwt{
						Jwt: &proto.JWT{
							AuthorizedUsers: map[string]*proto.Empty{
								"auth0|yennefer_of_vengerberg": {},
							},
						},
					},
				},
			},
		},
		{
			name: "should return ok check response if endpoint does not require auth",
			checkReq: &envoy_auth.CheckRequest{
				Attributes: &envoy_auth.AttributeContext{
					Request: &envoy_auth.AttributeContext_Request{
						Http: &envoy_auth.AttributeContext_HttpRequest{
							Path: "/v1/public_endpoint",
							Headers: map[string]string{
								reqHeaderJWTUserID: "auth0|ulfric_stormcloak",
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
							{Header: &envoy_core.HeaderValue{Key: reqHeaderEndpointID, Value: "public_endpoint"}},
						},
					},
				},
			},
			endpointIDExtractor: &URLPathExtractor{},
			endpointID:          "public_endpoint",
			mockEndpointReturn: &proto.GatewayEndpoint{
				EndpointId: "public_endpoint",
				Auth: &proto.Auth{
					AuthType: &proto.Auth_NoAuth{
						NoAuth: &proto.NoAuth{},
					},
				},
			},
		},
		{
			name: "should return ok check response if endpoint ID is passed via header",
			checkReq: &envoy_auth.CheckRequest{
				Attributes: &envoy_auth.AttributeContext{
					Request: &envoy_auth.AttributeContext_Request{
						Http: &envoy_auth.AttributeContext_HttpRequest{
							Path: "/v1",
							Headers: map[string]string{
								reqHeaderEndpointID: "endpoint_id_from_header",
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
							{Header: &envoy_core.HeaderValue{Key: reqHeaderEndpointID, Value: "endpoint_id_from_header"}},
						},
					},
				},
			},
			endpointIDExtractor: &HeaderExtractor{},
			endpointID:          "endpoint_id_from_header",
			mockEndpointReturn: &proto.GatewayEndpoint{
				EndpointId: "endpoint_id_from_header",
				Auth: &proto.Auth{
					AuthType: &proto.Auth_NoAuth{
						NoAuth: &proto.NoAuth{},
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
								reqHeaderJWTUserID: "auth0|ellen_ripley",
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
			endpointIDExtractor: &URLPathExtractor{},
			endpointID:          "endpoint_not_found",
			mockEndpointReturn:  &proto.GatewayEndpoint{},
		},
		{
			name: "should return denied check response if user is not authorized to access endpoint using API key auth",
			checkReq: &envoy_auth.CheckRequest{
				Attributes: &envoy_auth.AttributeContext{
					Request: &envoy_auth.AttributeContext_Request{
						Http: &envoy_auth.AttributeContext_HttpRequest{
							Path: "/v1/endpoint_api_key",
							Headers: map[string]string{
								reqHeaderAPIKey: "api_key_123",
							},
						},
					},
				},
			},
			expectedResp: &envoy_auth.CheckResponse{
				Status: &status.Status{
					Code:    int32(codes.PermissionDenied),
					Message: errUnauthorized.Error(),
				},
				HttpResponse: &envoy_auth.CheckResponse_DeniedResponse{
					DeniedResponse: &envoy_auth.DeniedHttpResponse{
						Status: &envoy_type.HttpStatus{
							Code: envoy_type.StatusCode_Unauthorized,
						},
						Body: fmt.Sprintf(`{"code": 401, "message": "%s"}`, errUnauthorized.Error()),
					},
				},
			},
			endpointIDExtractor: &URLPathExtractor{},
			endpointID:          "endpoint_api_key",
			mockEndpointReturn: &proto.GatewayEndpoint{
				EndpointId: "endpoint_api_key",
				Auth: &proto.Auth{
					AuthType: &proto.Auth_StaticApiKey{
						StaticApiKey: &proto.StaticAPIKey{
							ApiKey: "api_key_not_this_one",
						},
					},
				},
			},
		},
		{
			name: "should return denied check response if user is not authorized to access endpoint using JWT auth",
			checkReq: &envoy_auth.CheckRequest{
				Attributes: &envoy_auth.AttributeContext{
					Request: &envoy_auth.AttributeContext_Request{
						Http: &envoy_auth.AttributeContext_HttpRequest{
							Path: "/v1/endpoint_jwt_auth",
							Headers: map[string]string{
								reqHeaderJWTUserID: "auth0|ulfric_stormcloak",
							},
						},
					},
				},
			},
			expectedResp: &envoy_auth.CheckResponse{
				Status: &status.Status{
					Code:    int32(codes.PermissionDenied),
					Message: errUnauthorized.Error(),
				},
				HttpResponse: &envoy_auth.CheckResponse_DeniedResponse{
					DeniedResponse: &envoy_auth.DeniedHttpResponse{
						Status: &envoy_type.HttpStatus{
							Code: envoy_type.StatusCode_Unauthorized,
						},
						Body: fmt.Sprintf(`{"code": 401, "message": "%s"}`, errUnauthorized.Error()),
					},
				},
			},
			endpointIDExtractor: &URLPathExtractor{},
			endpointID:          "endpoint_jwt_auth",
			mockEndpointReturn: &proto.GatewayEndpoint{
				EndpointId: "endpoint_jwt_auth",
				Auth: &proto.Auth{
					AuthType: &proto.Auth_Jwt{
						Jwt: &proto.JWT{
							AuthorizedUsers: map[string]*proto.Empty{
								"auth0|chrisjen_avasarala": {},
							},
						},
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

			mockStore := NewMockEndpointStore(ctrl)
			if test.endpointID != "" {
				mockStore.EXPECT().GetGatewayEndpoint(test.endpointID).Return(test.mockEndpointReturn, test.mockEndpointReturn.EndpointId != "")
			}

			authHandler := &AuthHandler{
				Logger: polyzero.NewLogger(),

				EndpointStore:       mockStore,
				APIKeyAuthorizer:    &APIKeyAuthorizer{},
				JWTAuthorizer:       &JWTAuthorizer{},
				EndpointIDExtractor: test.endpointIDExtractor,
			}

			resp, err := authHandler.Check(context.Background(), test.checkReq)
			c.NoError(err)
			c.Equal(test.expectedResp, resp)
		})
	}
}
