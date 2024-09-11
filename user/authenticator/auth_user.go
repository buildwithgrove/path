package authenticator

import (
	"context"

	"github.com/pokt-network/poktroll/pkg/polylog"

	reqCtx "github.com/buildwithgrove/path/request/context"
	"github.com/buildwithgrove/path/user"
)

type userAppAuthenticator struct {
	logger polylog.Logger
}

func newUserAppAuthenticator(logger polylog.Logger) *userAppAuthenticator {
	return &userAppAuthenticator{
		logger: logger.With("component", "user_authenticator"),
	}
}

func (a *userAppAuthenticator) authenticate(ctx context.Context, reqDetails reqCtx.HTTPDetails, userApp user.UserApp) *failedAuth {

	if failedSecretKeyAuth := authSecretKey(reqDetails, userApp); failedSecretKeyAuth != nil {
		return failedSecretKeyAuth
	}

	return nil
}

func authSecretKey(reqDetails reqCtx.HTTPDetails, userApp user.UserApp) *failedAuth {
	if userApp.SecretKeyRequired {
		if reqDetails.SecretKey == "" {
			return &userAuthFailSecretKeyRequired
		}
		if reqDetails.SecretKey != userApp.SecretKey {
			return &userAuthFailInvalidSecretKey
		}
	}

	return nil
}
