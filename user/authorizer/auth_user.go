package authorizer

import (
	"context"

	"github.com/pokt-network/poktroll/pkg/polylog"

	reqCtx "github.com/buildwithgrove/path/request/context"
	"github.com/buildwithgrove/path/user"
)

type endpointAuthorizer struct {
	logger polylog.Logger
}

func newGatewayEndpointAuthorizer(logger polylog.Logger) *endpointAuthorizer {
	return &endpointAuthorizer{
		logger: logger.With("component", "user_authorizer"),
	}
}

func (a *endpointAuthorizer) authorizeRequest(ctx context.Context, reqDetails reqCtx.HTTPDetails, endpoint user.GatewayEndpoint) *failedAuth {

	if failedAPIKeyAuth := authAPIKey(reqDetails, endpoint); failedAPIKeyAuth != nil {
		return failedAPIKeyAuth
	}

	return nil
}

func authAPIKey(reqDetails reqCtx.HTTPDetails, endpoint user.GatewayEndpoint) *failedAuth {
	if apiKey, authRequired := endpoint.GetAuth(); authRequired {
		if reqDetails.APIKey == "" {
			return &userAuthFailAPIKeyRequired
		}
		if reqDetails.APIKey != apiKey {
			return &userAuthFailInvalidAPIKey
		}
	}

	return nil
}
