package authorizer

import (
	"context"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	reqCtx "github.com/buildwithgrove/path/request/context"
	"github.com/buildwithgrove/path/user"
)

// user.RequestAuthorizer is used to authenticate service requests by users.
// It performs authorization and allowlist validation on requests and returns a
// failure response message to the client when authorization fails.
type RequestAuthorizer struct {
	cache       cache
	authorizers []authorizer
	logger      polylog.Logger
}

type (
	cache interface {
		GetGatewayEndpoint(ctx context.Context, userAppID user.EndpointID) (user.GatewayEndpoint, bool)
	}
	authorizer interface {
		authenticate(ctx context.Context, reqDetails reqCtx.HTTPDetails, userApp user.GatewayEndpoint) *failedAuth
	}
)

func NewRequestAuthorizer(cache cache, logger polylog.Logger) *RequestAuthorizer {
	return &RequestAuthorizer{
		cache: cache,
		authorizers: []authorizer{
			newGatewayEndpointAuthorizer(logger),
			// TODO_NEXT: add rate limit authorizer
		},
		logger: logger.With("component", "request_authorizer"),
	}
}

// AuthorizeRequest performs authorization using all configured authorizers on the service request.
//
// It returns a failedAuth struct with and error message and 401 status code to the client if auth fails or nil if auth succeeds.
func (a *RequestAuthorizer) AuthorizeRequest(ctx context.Context, req *http.Request, userAppID user.EndpointID) gateway.HTTPResponse {

	reqDetails := reqCtx.GetHTTPDetailsFromCtx(ctx)

	userApp, ok := a.cache.GetGatewayEndpoint(ctx, userAppID)
	if !ok {
		return &userAppNotFound
	}

	for _, auth := range a.authorizers {
		if resp := auth.authenticate(ctx, reqDetails, userApp); resp != nil {
			return resp
		}
	}

	return nil
}
