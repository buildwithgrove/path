package context

import (
	"context"
	"net/http"
	"strings"

	"github.com/buildwithgrove/path/user"
)

type ctxKey string

const (
	ctxKeyHTTPDetails ctxKey = "http_details"
	ctxKeyEndpointID  ctxKey = "endpoint_id"
)

// HTTPDetails contains HTTP details from an http.Request to be used
// throughout the service request lifecycle, including the auth token, if set.
type HTTPDetails struct {
	Method    string
	Path      string
	Origin    string
	UserAgent string
	APIKey    string
}

// SetCtxFromRequest sets HTTP details and gateway endpoint ID in the context from an
// http.Request and returns the updated context to be used in the service
// request lifecycle. This data is used for user app-specific request authentication.
func SetCtxFromRequest(ctx context.Context, req *http.Request, endpointID user.EndpointID) context.Context {
	ctx = context.WithValue(ctx, ctxKeyHTTPDetails, HTTPDetails{
		Method:    req.Method,
		Path:      req.URL.Path,
		Origin:    req.Header.Get("Origin"),
		UserAgent: req.Header.Get("User-Agent"),
		APIKey:    getAPIKeyFromAuthHeader(req),
	})
	ctx = context.WithValue(ctx, ctxKeyEndpointID, endpointID)
	return ctx
}

// getAPIKeyFromAuthHeader allows setting the API key in the header both on its own or with the "Bearer " prefix.
func getAPIKeyFromAuthHeader(req *http.Request) string {
	if authHeader := req.Header.Get("Authorization"); authHeader != "" {
		return strings.TrimPrefix(strings.ToLower(authHeader), "bearer ")
	}
	return ""
}

/* --------------------------------- Getters -------------------------------- */

func GetHTTPDetailsFromCtx(ctx context.Context) HTTPDetails {
	if httpDetails, ok := ctx.Value(ctxKeyHTTPDetails).(HTTPDetails); ok {
		return httpDetails
	}
	return HTTPDetails{}
}

func GetEndpointIDFromCtx(ctx context.Context) user.EndpointID {
	if endpointID, ok := ctx.Value(ctxKeyEndpointID).(user.EndpointID); ok {
		return endpointID
	}
	return ""
}
