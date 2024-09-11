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
		authenticate(ctx context.Context, reqDetails reqCtx.HTTPDetails, userApp user.UserApp) *invalidResp
	}
)

func NewRequestAuthenticator(cache cache, redisAddr string, logger polylog.Logger) *RequestAuthenticator {
	return &RequestAuthenticator{
		cache: cache,
		authenticators: []authenticator{
			newUserAuthenticator(logger),
			newRateLimitAuthenticator(redisAddr, logger),
		},
		logger: logger.With("component", "request_authenticator"),
	}
}

// AuthenticateReq erforms authentication using all configured authenticators on the service request.
//
// It returns an invalidResp struct containing a failure message to be returned
// to the client if authentication fails.
func (a *RequestAuthenticator) AuthenticateReq(ctx context.Context, req *http.Request, appID string) gateway.HTTPResponse {
	userAppID := user.UserAppID(appID)

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
