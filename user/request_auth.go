package user

import (
	"context"
	"fmt"
	"net/http"

	"github.com/buildwithgrove/path/gateway"
	reqCtx "github.com/buildwithgrove/path/request/context"
)

var (
	invalidRespTemplate = `{"code":%d,"reason":"%s"}`

	authFailCode              = -32006
	authFailUserAppNotFound   = invalidResp{body: fmt.Sprintf(invalidRespTemplate, authFailCode, "user app not found")}
	authFailSecretKeyRequired = invalidResp{body: fmt.Sprintf(invalidRespTemplate, authFailCode, "secret key is required")}
	authFailInvalidSecretKey  = invalidResp{body: fmt.Sprintf(invalidRespTemplate, authFailCode, "invalid secret key")}

	rateLimitExceededCode   = -32007
	throughputLimitExceeded = invalidResp{body: fmt.Sprintf(invalidRespTemplate, rateLimitExceededCode, "throughput limit exceeded")}
)

// user.RequestAuthenticator is used to authenticate service requests by users.
// It performs authentication and allowlist validation on requests and returns a
// failure response message to the client when authentication fails.
type RequestAuthenticator struct {
	Cache             cache
	throughputLimiter *limiterManager
}
type cache interface {
	GetUserApp(ctx context.Context, userAppID UserAppID) (UserApp, bool)
}

func NewRequestAuthenticator(cache cache) *RequestAuthenticator {
	return &RequestAuthenticator{
		Cache:             cache,
		throughputLimiter: newLimiterManager(),
	}
}

// invalidResp contains a response body for an authentication failure to be
// returned to the client. It satisfies the gateway.HTTPResponse interface.
type invalidResp struct {
	body string
}

func (r *invalidResp) GetPayload() []byte {
	return []byte(r.body)
}

func (r *invalidResp) GetHTTPStatusCode() int {
	return http.StatusUnauthorized
}

func (r *invalidResp) GetHTTPHeaders() map[string]string {
	return map[string]string{}
}

// AuthenticateReq authenticates a service request made by a user. It performs all required validation on the service request, including:
// secret key authentication if the user app requires a secret key,
// allowlist validation if the user app has an allowlist configured,
// and throughput rate limiting if the user app is for a plan with a throughput limit configured.
//
// It returns an invalidResp struct containing a failure message to be returned
// to the client if authentication fails.
func (a *RequestAuthenticator) AuthenticateReq(ctx context.Context, req *http.Request, appID string) gateway.HTTPResponse {
	userAppID := UserAppID(appID)

	reqDetails := reqCtx.GetHTTPDetailsFromCtx(ctx)

	userApp, ok := a.Cache.GetUserApp(ctx, userAppID)
	if !ok {
		return &authFailUserAppNotFound
	}

	// TODO - move user auth to own func
	if userApp.SecretKeyRequired {
		if invalidResp := isSecretKeyValid(reqDetails.SecretKey, userApp.SecretKey); invalidResp != nil {
			return invalidResp
		}
	}

	// TODO - move rate limit auth to own func
	if userApp.RateLimitThroughput > 0 {
		if limiter := a.throughputLimiter.getLimiter(userAppID, userApp.RateLimitThroughput); limiter != nil {
			if !limiter.Allow() {
				fmt.Println("rate limit exceeded")
				return &throughputLimitExceeded
			}
		}
		fmt.Println("rate limit not exceeded")
	}

	return nil
}

func isSecretKeyValid(reqSecretKey, userSecretKey string) *invalidResp {
	if reqSecretKey == "" {
		return &authFailSecretKeyRequired
	}
	if reqSecretKey != userSecretKey {
		return &authFailInvalidSecretKey
	}
	return nil
}
