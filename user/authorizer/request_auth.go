package authorizer

import (
	"context"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	reqCtx "github.com/buildwithgrove/path/request/context"
	"github.com/buildwithgrove/path/user"
)

// user.RequestAuthorizer is used to authorize service requests by users.
// It performs user data authentication and rate limiting on requests and
// returns a gateway.HTTPResponse error message to the client if auth fails.
type RequestAuthorizer struct {
	cache       cache
	authorizers []authorizer
	logger      polylog.Logger
}

type (
	cache interface {
		GetUserApp(ctx context.Context, userAppID user.UserAppID) (user.UserApp, bool)
	}
	authorizer interface {
		authorizeRequest(ctx context.Context, reqDetails reqCtx.HTTPDetails, userApp user.UserApp) *failedAuth
	}
)

func NewRequestAuthorizer(cache cache, redisAddr string, logger polylog.Logger) *RequestAuthorizer {
	return &RequestAuthorizer{
		cache: cache,
		authorizers: []authorizer{
			newUserAppAuthenticator(logger),
			newRateLimiter(redisAddr, logger),
		},
		logger: logger.With("component", "request_authorizer"),
	}
}

// AuthorizeRequest performs authorization using all configured authenticators on the service request.
//
// It returns a failedAuth struct with and error message and 401 status code to the client if auth fails or nil if auth succeeds.
func (a *RequestAuthorizer) AuthorizeRequest(ctx context.Context, req *http.Request, userAppID user.UserAppID) gateway.HTTPResponse {

	reqDetails := reqCtx.GetHTTPDetailsFromCtx(ctx)

	userApp, ok := a.cache.GetUserApp(ctx, userAppID)
	if !ok {
		return &userAppNotFound
	}

	for _, auth := range a.authorizers {
		if resp := auth.authorizeRequest(ctx, reqDetails, userApp); resp != nil {
			return resp
		}
	}

	return nil
}
