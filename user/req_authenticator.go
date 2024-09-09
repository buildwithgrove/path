package user

import (
	"context"
	"fmt"
	"net/http"

	"github.com/buildwithgrove/path/gateway"
	reqCtx "github.com/buildwithgrove/path/request/context"
)

type RequestAuthenticator struct {
	Cache cache
}

type (
	cache interface {
		GetUserApp(ctx context.Context, userAppID UserAppID) (UserApp, bool)
	}
	authFailedResponse struct {
		reason string
	}
)

func (r *authFailedResponse) GetPayload() []byte {
	return []byte(r.reason)
}

func (r *authFailedResponse) GetHTTPStatusCode() int {
	return http.StatusUnauthorized
}

func (r *authFailedResponse) GetHTTPHeaders() map[string]string {
	return map[string]string{}
}

func (a *RequestAuthenticator) AuthenticateReq(ctx context.Context, req *http.Request, appID string) gateway.HTTPResponse {
	userAppID := UserAppID(appID)

	reqDetails := reqCtx.GetHTTPDetailsFromCtx(ctx)

	userApp, ok := a.Cache.GetUserApp(ctx, userAppID)
	if !ok {
		return &authFailedResponse{reason: fmt.Sprintf("user app %s not found", userAppID)}
	}

	if userApp.SecretKeyRequired {
		if failedAuthResp := authenticateSecretKey(reqDetails.SecretKey, userApp.SecretKey); failedAuthResp != nil {
			return failedAuthResp
		}
	}

	// if len(userApp.Allowlists) > 0 {
	// 	resp = authenticateAllowlists(ctx, req, userApp)
	// }

	return nil
}

func authenticateSecretKey(reqSecretKey, userSecretKey string) *authFailedResponse {
	if reqSecretKey == "" {
		return &authFailedResponse{reason: "secret key is required"}
	}
	if reqSecretKey != userSecretKey {
		return &authFailedResponse{reason: "invalid secret key"}
	}
	return nil
}

// func authenticateAllowlists(ctx context.Context, req *http.Request, userApp UserApp) *authFailedResponse {
// 	if len(userApp.Allowlists[AllowlistTypeContracts]) > 0 {

// 	}

// }
