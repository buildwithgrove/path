package authenticator

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

func Test_AuthenticateReq(t *testing.T) {
	tests := []struct {
		name           string
		userAppID      user.UserAppID
		userApp        user.UserApp
		req            *http.Request
		appExists      bool
		expectedResult gateway.HTTPResponse
	}{
		{
			name:      "should authenticate valid user app ID",
			userAppID: "user_app_1",
			userApp: user.UserApp{
				ID:                "user_app_1",
				AccountID:         "account_1",
				SecretKey:         "test_key_1",
				SecretKeyRequired: true,
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
			userApp:        user.UserApp{},
			appExists:      false,
			expectedResult: &userAppNotFound,
		},
		{
			name:      "should not authenticate missing secret key",
			userAppID: "user_app_3",
			req: &http.Request{
				URL: &url.URL{Path: "/v1/user_app_3"},
			},
			userApp: user.UserApp{
				ID:                "user_app_3",
				AccountID:         "account_3",
				SecretKey:         "test_key_3",
				SecretKeyRequired: true,
			},
			appExists:      true,
			expectedResult: &userAuthFailSecretKeyRequired,
		},
		{
			name:      "should not authenticate invalid secret key",
			userAppID: "user_app_4",
			req: &http.Request{
				URL:    &url.URL{Path: "/v1/user_app_4"},
				Header: http.Header{"Authorization": []string{"user_app_whoops"}},
			},
			userApp: user.UserApp{
				ID:                "user_app_4",
				AccountID:         "account_4",
				SecretKey:         "test_key_4",
				SecretKeyRequired: true,
			},
			appExists:      true,
			expectedResult: &userAuthFailInvalidSecretKey,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			ctrl := gomock.NewController(t)

			ctx := reqCtx.SetCtxFromRequest(test.req.Context(), test.req, test.userAppID)

			mockCache := NewMockcache(ctrl)
			mockCache.EXPECT().GetUserApp(ctx, test.userAppID).Return(test.userApp, test.appExists)

			logger := polyzero.NewLogger()

			authenticator := &RequestAuthenticator{
				cache: mockCache,
				authenticators: []authenticator{
					newUserAppAuthenticator(logger),
					// TODO_IMPROVE: add test cases for rate limiting
				},
			}

			result := authenticator.AuthenticateReq(ctx, test.req, test.userAppID)
			c.Equal(test.expectedResult, result)
		})
	}
}
