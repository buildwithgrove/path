package authenticator

import (
	"context"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	reqCtx "github.com/buildwithgrove/path/request/context"
	"github.com/buildwithgrove/path/user"
)

// user.RequestAuthenticator is used to authenticate service requests by users.
// It performs authentication and allowlist validation on requests and returns a
// failure response message to the client when authentication fails.
type RequestAuthenticator struct {
	cache          cache
	authenticators []authenticator
	logger         polylog.Logger
}

type (
	cache interface {
		GetUserApp(ctx context.Context, userAppID user.UserAppID) (user.UserApp, bool)
	}
	authenticator interface {
		authenticate(ctx context.Context, reqDetails reqCtx.HTTPDetails, userApp user.UserApp) *failedAuth
	}
)

func NewRequestAuthenticator(cache cache, logger polylog.Logger) *RequestAuthenticator {
	return &RequestAuthenticator{
		cache: cache,
		authenticators: []authenticator{
			newUserAppAuthenticator(logger),
			// TODO_NEXT: add rate limit authenticator
		},
		logger: logger.With("component", "request_authenticator"),
	}
}

// AuthenticateReq erforms authentication using all configured authenticators on the service request.
//
// It returns a failedAuth struct with and error message and 401 status code to the client if auth fails or nil if auth succeeds.
func (a *RequestAuthenticator) AuthenticateReq(ctx context.Context, req *http.Request, userAppID user.UserAppID) gateway.HTTPResponse {

	reqDetails := reqCtx.GetHTTPDetailsFromCtx(ctx)

	userApp, ok := a.cache.GetUserApp(ctx, userAppID)
	if !ok {
		return &userAppNotFound
	}

	for _, auth := range a.authenticators {
		if resp := auth.authenticate(ctx, reqDetails, userApp); resp != nil {
			return resp
		}
	}

	return nil
}
