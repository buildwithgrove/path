package authenticator

import (
	"testing"

	"github.com/stretchr/testify/require"

	reqCtx "github.com/buildwithgrove/path/request/context"
	"github.com/buildwithgrove/path/user"
)

func Test_authSecretKey(t *testing.T) {
	tests := []struct {
		name           string
		reqDetails     reqCtx.HTTPDetails
		userApp        user.UserApp
		expectedResult *failedAuth
	}{
		{
			name: "should return nil for valid secret key",
			reqDetails: reqCtx.HTTPDetails{
				SecretKey: "validKey",
			},
			userApp: user.UserApp{
				SecretKeyRequired: true,
				SecretKey:         "validKey",
			},
			expectedResult: nil,
		},
		{
			name: "should return authFailSecretKeyRequired for empty request secret key",
			reqDetails: reqCtx.HTTPDetails{
				SecretKey: "",
			},
			userApp: user.UserApp{
				SecretKeyRequired: true,
				SecretKey:         "validKey",
			},
			expectedResult: &userAuthFailSecretKeyRequired,
		},
		{
			name: "should return authFailInvalidSecretKey for invalid request secret key",
			reqDetails: reqCtx.HTTPDetails{
				SecretKey: "invalidKey",
			},
			userApp: user.UserApp{
				SecretKeyRequired: true,
				SecretKey:         "validKey",
			},
			expectedResult: &userAuthFailInvalidSecretKey,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			result := authSecretKey(test.reqDetails, test.userApp)
			c.Equal(test.expectedResult, result)
		})
	}
}
