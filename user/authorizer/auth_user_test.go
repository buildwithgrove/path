package authorizer

import (
	"testing"

	"github.com/stretchr/testify/require"

	reqCtx "github.com/buildwithgrove/path/request/context"
	"github.com/buildwithgrove/path/user"
)

func Test_authAPIKey(t *testing.T) {
	tests := []struct {
		name           string
		reqDetails     reqCtx.HTTPDetails
		endpoint       user.GatewayEndpoint
		expectedResult *failedAuth
	}{
		{
			name: "should return nil for valid secret key",
			reqDetails: reqCtx.HTTPDetails{
				APIKey: "validKey",
			},
			endpoint: user.GatewayEndpoint{
				Auth: user.Auth{
					APIKeyRequired: true,
					APIKey:         "validKey",
				},
			},
			expectedResult: nil,
		},
		{
			name: "should return authFailAPIKeyRequired for empty request secret key",
			reqDetails: reqCtx.HTTPDetails{
				APIKey: "",
			},
			endpoint: user.GatewayEndpoint{
				Auth: user.Auth{
					APIKeyRequired: true,
					APIKey:         "validKey",
				},
			},
			expectedResult: &userAuthFailAPIKeyRequired,
		},
		{
			name: "should return authFailInvalidAPIKey for invalid request secret key",
			reqDetails: reqCtx.HTTPDetails{
				APIKey: "invalidKey",
			},
			endpoint: user.GatewayEndpoint{
				Auth: user.Auth{
					APIKeyRequired: true,
					APIKey:         "validKey",
				},
			},
			expectedResult: &userAuthFailInvalidAPIKey,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			result := authAPIKey(test.reqDetails, test.endpoint)
			c.Equal(test.expectedResult, result)
		})
	}
}
