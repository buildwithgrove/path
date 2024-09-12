package authorizer

import (
	"context"

	"github.com/pokt-network/poktroll/pkg/polylog"

	reqCtx "github.com/buildwithgrove/path/request/context"
	"github.com/buildwithgrove/path/user"
)

type userAppAuthorizer struct {
	logger polylog.Logger
}

func newGatewayEndpointAuthorizer(logger polylog.Logger) *userAppAuthorizer {
	return &userAppAuthorizer{
		logger: logger.With("component", "user_authorizer"),
	}
}

func (a *userAppAuthorizer) authenticate(ctx context.Context, reqDetails reqCtx.HTTPDetails, userApp user.GatewayEndpoint) *failedAuth {

	if failedAPIKeyAuth := authAPIKey(reqDetails, userApp); failedAPIKeyAuth != nil {
		return failedAPIKeyAuth
	}

	return nil
}

func authAPIKey(reqDetails reqCtx.HTTPDetails, userApp user.GatewayEndpoint) *failedAuth {
	if apiKey, authRequired := userApp.GetAuth(); authRequired {
		if reqDetails.APIKey == "" {
			return &userAuthFailAPIKeyRequired
		}
		if reqDetails.APIKey != apiKey {
			return &userAuthFailInvalidAPIKey
		}
	}

	return nil
}
