package context

import (
	"context"
	"net/http"

	"github.com/buildwithgrove/path/user"
)

type ctxKey string

const (
	ctxKeyHttpDetails ctxKey = "http_details"
	ctxKeyUserAppID   ctxKey = "user_app_id"
)

// HTTPDetails contains HTTP details from an http.Request to be used
// throughout the service request lifecycle, including the auth token, if set.
type HTTPDetails struct {
	Method    string
	Path      string
	Origin    string
	UserAgent string
	SecretKey string
}

// SetCtxFromRequest sets HTTP details and user app ID in the context from an
// http.Request and returns the updated context to be used in the service
// request lifecycle. This data is used for user app-specific request authentication.
func SetCtxFromRequest(ctx context.Context, req *http.Request, userAppID user.UserAppID) context.Context {
	ctx = context.WithValue(ctx, ctxKeyHttpDetails, HTTPDetails{
		Method:    req.Method,
		Path:      req.URL.Path,
		Origin:    req.Header.Get("Origin"),
		UserAgent: req.Header.Get("User-Agent"),
		SecretKey: req.Header.Get("Authorization"),
	})
	ctx = context.WithValue(ctx, ctxKeyUserAppID, userAppID)
	return ctx
}

/* --------------------------------- Getters -------------------------------- */

func GetHTTPDetailsFromCtx(ctx context.Context) HTTPDetails {
	if httpDetails, ok := ctx.Value(ctxKeyHttpDetails).(HTTPDetails); ok {
		return httpDetails
	}
	return HTTPDetails{}
}

func GetUserAppIDFromCtx(ctx context.Context) user.UserAppID {
	if userAppID, ok := ctx.Value(ctxKeyUserAppID).(user.UserAppID); ok {
		return userAppID
	}
	return ""
}
