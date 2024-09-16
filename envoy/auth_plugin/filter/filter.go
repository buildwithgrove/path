//go:build auth_plugin

package filter

import (
	"context"
	"fmt"
	"strings"

	"github.com/buildwithgrove/authorizer-plugin/types"
	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
)

// The callbacks in the filter, like `DecodeHeaders`, can be implemented on demand.
// Because api.PassThroughStreamFilter provides a default implementation.
type HTTPFilter struct {
	api.PassThroughStreamFilter

	Callbacks api.FilterCallbackHandler
	Config    *EnvoyConfig
	Cache     userDataCache
}

type userDataCache interface {
	GetGatewayEndpoint(ctx context.Context, userAppID types.EndpointID) (types.GatewayEndpoint, bool)
}

// Callbacks which are called in request path
// The endStream is true if the request doesn't have body
func (f *HTTPFilter) DecodeHeaders(header api.RequestHeaderMap, endStream bool) api.StatusType {
	path, ok := header.Get(":path")
	if !ok || path == "" {
		return sendErrPathNotProvided(f.Callbacks.DecoderFilterCallbacks())
	}

	// If the path is "/healthz", we don't need to authenticate
	if path == "/healthz" {
		return api.Continue
	}

	// If the path is "/v1/{gateway_endpoint_id}", we need to authenticate
	endpointID, ok := extractEndpointID(path)
	if !ok {
		return sendErrEndpointIDNotProvided(f.Callbacks.DecoderFilterCallbacks())
	}

	// If the code is time-consuming, to avoid blocking the Envoy,
	// we need to run the code in a background goroutine
	// and suspend & resume the filter
	go func() {
		defer f.Callbacks.DecoderFilterCallbacks().RecoverPanic()

		gatewayEndpoint, ok := f.Cache.GetGatewayEndpoint(context.Background(), endpointID)
		if !ok {
			sendAsyncErrResponse(f.Callbacks.DecoderFilterCallbacks(), errEndpointNotFound)
		}

		header.Set("x-endpoint-id", string(gatewayEndpoint.EndpointID))
		header.Set("x-account-id", string(gatewayEndpoint.UserAccount.AccountID))
		header.Set("x-plan", string(gatewayEndpoint.UserAccount.PlanType))
		header.Set("x-rate-limit-throughput", fmt.Sprintf("%d", gatewayEndpoint.RateLimiting.ThroughputLimit))

		// Continue the filter
		f.Callbacks.DecoderFilterCallbacks().Continue(api.Continue)
	}()

	// suspend the filter
	return api.Running
}

func extractEndpointID(urlPath string) (types.EndpointID, bool) {
	const prefix = "/v1/"
	if strings.HasPrefix(urlPath, prefix) {
		return types.EndpointID(urlPath[len(prefix):]), true
	}
	return "", false
}

/* --------------------------------- Unused -------------------------------- */

// DecodeData might be called multiple times during handling the request body.
// The endStream is true when handling the last piece of the body.
func (f *HTTPFilter) DecodeData(buffer api.BufferInstance, endStream bool) api.StatusType {
	// support suspending & resuming the filter in a background goroutine
	return api.Continue
}

func (f *HTTPFilter) DecodeTrailers(trailers api.RequestTrailerMap) api.StatusType {
	// support suspending & resuming the filter in a background goroutine
	return api.Continue
}

// Callbacks which are called in response path
// The endStream is true if the response doesn't have body
func (f *HTTPFilter) EncodeHeaders(header api.ResponseHeaderMap, endStream bool) api.StatusType {
	// support suspending & resuming the filter in a background goroutine
	return api.Continue
}

// EncodeData might be called multiple times during handling the response body.
// The endStream is true when handling the last piece of the body.
func (f *HTTPFilter) EncodeData(buffer api.BufferInstance, endStream bool) api.StatusType {
	// support suspending & resuming the filter in a background goroutine
	return api.Continue
}

func (f *HTTPFilter) EncodeTrailers(trailers api.ResponseTrailerMap) api.StatusType {
	return api.Continue
}

// OnLog is called when the HTTP stream is ended on HTTP Connection Manager filter.
func (f *HTTPFilter) OnLog() {}

// OnLogDownstreamStart is called when HTTP Connection Manager filter receives a new HTTP request
// (required the corresponding access log type is enabled)
func (f *HTTPFilter) OnLogDownstreamStart() {}

// OnLogDownstreamPeriodic is called on any HTTP Connection Manager periodic log record
// (required the corresponding access log type is enabled)
func (f *HTTPFilter) OnLogDownstreamPeriodic() {}

func (f *HTTPFilter) OnDestroy(reason api.DestroyReason) {
	// One should not access f.callbacks here because the FilterCallbackHandler
	// is released. But we can still access other Go fields in the filter f.

	// goroutine can be used everywhere.
}
