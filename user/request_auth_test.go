package user

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	"github.com/buildwithgrove/path/gateway"
	reqCtx "github.com/buildwithgrove/path/request/context"
)

func Test_AuthenticateReq(t *testing.T) {
	tests := []struct {
		name           string
		userAppID      UserAppID
		userApp        UserApp
		req            *http.Request
		appExists      bool
		expectedResult gateway.HTTPResponse
	}{
		{
			name:      "should authenticate valid user app ID",
			userAppID: "user_app_1",
			userApp: UserApp{
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
			userApp:        UserApp{},
			appExists:      false,
			expectedResult: &authFailUserAppNotFound,
		},
		{
			name:      "should not authenticate missing secret key",
			userAppID: "user_app_3",
			req: &http.Request{
				URL: &url.URL{Path: "/v1/user_app_3"},
			},
			userApp: UserApp{
				ID:                "user_app_3",
				AccountID:         "account_3",
				SecretKey:         "test_key_3",
				SecretKeyRequired: true,
			},
			appExists:      true,
			expectedResult: &authFailSecretKeyRequired,
		},
		{
			name:      "should not authenticate invalid secret key",
			userAppID: "user_app_4",
			req: &http.Request{
				URL:    &url.URL{Path: "/v1/user_app_4"},
				Header: http.Header{"Authorization": []string{"user_app_whoops"}},
			},
			userApp: UserApp{
				ID:                "user_app_4",
				AccountID:         "account_4",
				SecretKey:         "test_key_4",
				SecretKeyRequired: true,
			},
			appExists:      true,
			expectedResult: &authFailInvalidSecretKey,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			ctrl := gomock.NewController(t)

			ctx := reqCtx.SetCtxFromRequest(test.req.Context(), test.req, string(test.userAppID))

			mockCache := NewMockcache(ctrl)
			mockCache.EXPECT().GetUserApp(ctx, test.userAppID).Return(test.userApp, test.appExists)

			authenticator := &RequestAuthenticator{Cache: mockCache}

			result := authenticator.AuthenticateReq(ctx, test.req, string(test.userAppID))
			c.Equal(test.expectedResult, result)
		})
	}
}

func Test_isSecretKeyValid(t *testing.T) {
	tests := []struct {
		name           string
		reqSecretKey   string
		userSecretKey  string
		expectedResult *invalidResp
	}{
		{
			name:           "should return nil for valid secret key",
			reqSecretKey:   "validKey",
			userSecretKey:  "validKey",
			expectedResult: nil,
		},
		{
			name:           "should return authFailSecretKeyRequired for empty request secret key",
			reqSecretKey:   "",
			userSecretKey:  "validKey",
			expectedResult: &authFailSecretKeyRequired,
		},
		{
			name:           "should return authFailInvalidSecretKey for invalid request secret key",
			reqSecretKey:   "invalidKey",
			userSecretKey:  "validKey",
			expectedResult: &authFailInvalidSecretKey,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			result := isSecretKeyValid(test.reqSecretKey, test.userSecretKey)
			c.Equal(test.expectedResult, result)
		})
	}
}

func Test_invalidResp(t *testing.T) {
	tests := []struct {
		name            string
		resp            *invalidResp
		expectedBody    []byte
		expectedStatus  int
		expectedHeaders map[string]string
	}{
		{
			name:            "should return correct values",
			resp:            &invalidResp{body: "there was a button. I pushed it."},
			expectedBody:    []byte("there was a button. I pushed it."),
			expectedStatus:  http.StatusUnauthorized,
			expectedHeaders: map[string]string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			c.Equal(test.expectedBody, test.resp.GetPayload())
			c.Equal(test.expectedStatus, test.resp.GetHTTPStatusCode())
			c.Equal(test.expectedHeaders, test.resp.GetHTTPHeaders())
		})
	}
}
