package authenticator

import (
	"context"

	"github.com/pokt-network/poktroll/pkg/polylog"

	reqCtx "github.com/buildwithgrove/path/request/context"
	"github.com/buildwithgrove/path/user"
)

type userAuthenticator struct {
	logger polylog.Logger
}

func newUserAuthenticator(logger polylog.Logger) *userAuthenticator {
	return &userAuthenticator{
		logger: logger.With("component", "user_authenticator"),
	}
}

func (a *userAuthenticator) authenticate(ctx context.Context, reqDetails reqCtx.HTTPDetails, userApp user.UserApp) *invalidResp {

	if failedSecretKeyAuth := authSecretKey(reqDetails, userApp); failedSecretKeyAuth != nil {
		return failedSecretKeyAuth
	}

	return nil
}

func authSecretKey(reqDetails reqCtx.HTTPDetails, userApp user.UserApp) *invalidResp {
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
