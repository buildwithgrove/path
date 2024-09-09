package context

import (
	"context"
	"net/http"
)

type ctxKey string

const (
	ctxKeyHttpDetails ctxKey = "http_details"
	ctxKeyUserAppID   ctxKey = "user_app_id"
)

// HttpDetails contains HTTP details from an http.Request to be used
// throughout the service request lifecycle, including the auth token, if set.
type HttpDetails struct {
	Method    string
	Path      string
	Origin    string
	UserAgent string
	SecretKey string
}

// SetCtxFromRequest sets HTTP details and user app ID in the context from an
// http.Request and returns the updated context to be used in the service
// request lifecycle. This data is used for user app-specific request authentication.
func SetCtxFromRequest(ctx context.Context, req *http.Request, userAppID string) context.Context {
	ctx = context.WithValue(ctx, ctxKeyHttpDetails, HttpDetails{
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

func GetHTTPDetailsFromCtx(ctx context.Context) HttpDetails {
	if httpDetails, ok := ctx.Value(ctxKeyHttpDetails).(HttpDetails); ok {
		return httpDetails
	}
	return HttpDetails{}
}

func GetUserAppIDFromCtx(ctx context.Context) string {
	if userAppID, ok := ctx.Value(ctxKeyUserAppID).(string); ok {
		return userAppID
	}
	return ""
}