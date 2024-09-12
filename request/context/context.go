package context

import (
	"context"
	"net/http"

	"github.com/buildwithgrove/path/user"
)

type ctxKey string

const (
	ctxKeyHttpDetails ctxKey = "http_details"
	ctxKeyEndpointID  ctxKey = "endpoint_id"
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

// SetCtxFromRequest sets HTTP details and gateway endpoint ID in the context from an
// http.Request and returns the updated context to be used in the service
// request lifecycle. This data is used for user app-specific request authentication.
func SetCtxFromRequest(ctx context.Context, req *http.Request, endpointID user.EndpointID) context.Context {
	ctx = context.WithValue(ctx, ctxKeyHttpDetails, HttpDetails{
		Method:    req.Method,
		Path:      req.URL.Path,
		Origin:    req.Header.Get("Origin"),
		UserAgent: req.Header.Get("User-Agent"),
		SecretKey: req.Header.Get("Authorization"),
	})
	ctx = context.WithValue(ctx, ctxKeyEndpointID, endpointID)
	return ctx
}

/* --------------------------------- Getters -------------------------------- */

func GetHTTPDetailsFromCtx(ctx context.Context) HttpDetails {
	if httpDetails, ok := ctx.Value(ctxKeyHttpDetails).(HttpDetails); ok {
		return httpDetails
	}
	return HttpDetails{}
}

func GetEndpointIDFromCtx(ctx context.Context) user.EndpointID {
	if endpointID, ok := ctx.Value(ctxKeyEndpointID).(user.EndpointID); ok {
		return endpointID
	}
	return ""
}
