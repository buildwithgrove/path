package authorizer

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	"github.com/buildwithgrove/path/gateway"
	reqCtx "github.com/buildwithgrove/path/request/context"
	"github.com/buildwithgrove/path/user"
)

func Test_AuthorizeRequest(t *testing.T) {
	tests := []struct {
		name           string
		endpointID     user.EndpointID
		endpoint       user.GatewayEndpoint
		req            *http.Request
		appExists      bool
		expectedResult gateway.HTTPResponse
	}{
		{
			name:       "should authorize valid gateway endpoint ID",
			endpointID: "user_app_1",
			endpoint: user.GatewayEndpoint{
				EndpointID: "user_app_1",
				Auth: user.Auth{
					APIKeyRequired: true,
					APIKey:         "test_key_1",
				},
				UserAccount: user.UserAccount{
					AccountID: "account_1",
				},
			},
			req: &http.Request{
				URL:    &url.URL{Path: "/v1/user_app_1"},
				Header: http.Header{"Authorization": []string{"test_key_1"}},
			},
			appExists:      true,
			expectedResult: nil,
		},
		{
			name:       "should not authorize gateway endpoint ID when app does not exist",
			endpointID: "user_app_2",
			req: &http.Request{
				URL:    &url.URL{Path: "/v1/user_app_2"},
				Header: http.Header{"Authorization": []string{"user_app_2"}},
			},
			endpoint:       user.GatewayEndpoint{},
			appExists:      false,
			expectedResult: &endpointNotFound,
		},
		{
			name:       "should not authorize missing secret key",
			endpointID: "user_app_3",
			req: &http.Request{
				URL: &url.URL{Path: "/v1/user_app_3"},
			},
			endpoint: user.GatewayEndpoint{
				EndpointID: "user_app_3",
				Auth: user.Auth{
					APIKeyRequired: true,
					APIKey:         "test_key_3",
				},
				UserAccount: user.UserAccount{
					AccountID: "account_3",
				},
			},
			appExists:      true,
			expectedResult: &userAuthFailAPIKeyRequired,
		},
		{
			name:       "should not authorize invalid secret key",
			endpointID: "user_app_4",
			req: &http.Request{
				URL:    &url.URL{Path: "/v1/user_app_4"},
				Header: http.Header{"Authorization": []string{"user_app_whoops"}},
			},
			endpoint: user.GatewayEndpoint{
				EndpointID: "user_app_4",
				Auth: user.Auth{
					APIKeyRequired: true,
					APIKey:         "test_key_4",
				},
				UserAccount: user.UserAccount{
					AccountID: "account_4",
				},
			},
			appExists:      true,
			expectedResult: &userAuthFailInvalidAPIKey,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			ctrl := gomock.NewController(t)

			ctx := reqCtx.SetCtxFromRequest(test.req.Context(), test.req, test.endpointID)

			mockCache := NewMockcache(ctrl)
			mockCache.EXPECT().GetGatewayEndpoint(ctx, test.endpointID).Return(test.endpoint, test.appExists)

			logger := polyzero.NewLogger()

			authorizer := &RequestAuthorizer{
				cache: mockCache,
				authorizers: []authorizer{
					newGatewayEndpointAuthorizer(logger),
					// TODO_IMPROVE: add test cases for rate limiting
				},
			}

			result := authorizer.AuthorizeRequest(ctx, test.req, test.endpointID)
			c.Equal(test.expectedResult, result)
		})
	}
}
