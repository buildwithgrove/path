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
		userAppID      user.EndpointID
		userApp        user.GatewayEndpoint
		req            *http.Request
		appExists      bool
		expectedResult gateway.HTTPResponse
	}{
		{
			name:      "should authenticate valid user app ID",
			userAppID: "user_app_1",
			userApp: user.GatewayEndpoint{
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
			name:      "should not authenticate user app ID when app does not exist",
			userAppID: "user_app_2",
			req: &http.Request{
				URL:    &url.URL{Path: "/v1/user_app_2"},
				Header: http.Header{"Authorization": []string{"user_app_2"}},
			},
			userApp:        user.GatewayEndpoint{},
			appExists:      false,
			expectedResult: &userAppNotFound,
		},
		{
			name:      "should not authenticate missing secret key",
			userAppID: "user_app_3",
			req: &http.Request{
				URL: &url.URL{Path: "/v1/user_app_3"},
			},
			userApp: user.GatewayEndpoint{
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
			name:      "should not authenticate invalid secret key",
			userAppID: "user_app_4",
			req: &http.Request{
				URL:    &url.URL{Path: "/v1/user_app_4"},
				Header: http.Header{"Authorization": []string{"user_app_whoops"}},
			},
			userApp: user.GatewayEndpoint{
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

			ctx := reqCtx.SetCtxFromRequest(test.req.Context(), test.req, test.userAppID)

			mockCache := NewMockcache(ctrl)
			mockCache.EXPECT().GetGatewayEndpoint(ctx, test.userAppID).Return(test.userApp, test.appExists)

			logger := polyzero.NewLogger()

			authorizer := &RequestAuthorizer{
				cache: mockCache,
				authorizers: []authorizer{
					newGatewayEndpointAuthorizer(logger),
					// TODO_IMPROVE: add test cases for rate limiting
				},
			}

			result := authorizer.AuthorizeRequest(ctx, test.req, test.userAppID)
			c.Equal(test.expectedResult, result)
		})
	}
}
