//go:build auth_plugin

package filter

import (
	"context"
	"fmt"
	"strings"

	"github.com/buildwithgrove/authorizer-plugin/types"
	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
)

// The HTTPFilter struct implements the Envoy filter interface and is responsible
// for authorizing and settings headers for requests to the PATH service.
type HTTPFilter struct {
	api.PassThroughStreamFilter

	Callbacks api.FilterCallbackHandler
	Config    *EnvoyConfig

	Cache userDataCache
}

// The userDataCache contains an in-memory cache of GatewayEndpoints
// and their associated data from the connected Postgres database.
type userDataCache interface {
	GetGatewayEndpoint(context.Context, types.EndpointID) (types.GatewayEndpoint, bool)
}

/* --------------------------------- DecodeHeaders -------------------------------- */

// All processing of the request headers is done in DecodeHeaders. This includes:
//
// - extracting the endpoint ID from the path
//
// - performing authorization checks on the request
//
// - setting the appropriate headers (x-endpoint-id, x-account-id, x-rate-limit-throughput)
//
// - continuing the filter chain or sending an error response
//
// endStream is true if the request doesn't have a body.
func (f *HTTPFilter) DecodeHeaders(req api.RequestHeaderMap, endStream bool) api.StatusType {
	path := req.Path()
	if path == "" {
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

		// TODO_IMPROVE - move all auth handling to own files/structs

		// First, get the gateway endpoint from the cache and return an error if not found
		gatewayEndpoint, ok := f.Cache.GetGatewayEndpoint(context.Background(), endpointID)
		if !ok {
			sendAsyncErrEndpointNotFound(f.Callbacks)
		}

		// Then, check if the API key is required and valid
		if apiKey, authRequired := gatewayEndpoint.GetAuth(); authRequired {

			// If the API key is required, check if it is provided in the req auth header
			reqAPIKey, ok := req.Get("Authorization")
			if !ok || reqAPIKey == "" {
				sendAsyncErrAPIKeyRequired(f.Callbacks)
			}

			// If the API key in the req header does not match the endpoint's API key, return an error
			if reqAPIKey != apiKey {
				sendAsyncErrAPIKeyInvalid(f.Callbacks)
			}
		}

		// Set endpoint ID in the headers
		req.Set("x-endpoint-id", string(gatewayEndpoint.EndpointID))
		req.Set("x-account-id", string(gatewayEndpoint.UserAccount.AccountID))

		// Set rate limiting headers if the gateway endpoint has a throughput limit
		if gatewayEndpoint.RateLimiting.ThroughputLimit > 0 {
			req.Set("x-rate-limit-throughput", fmt.Sprintf("%d", gatewayEndpoint.RateLimiting.ThroughputLimit))
		}

		// Continue the filter and forward the request to the PATH service
		f.Callbacks.DecoderFilterCallbacks().Continue(api.Continue)
	}()

	// suspend the filter while the background goroutine that handles GatewayEndpoint auth is running
	return api.Running
}

// extractEndpointID extracts the endpoint ID from the URL path.
// The endpoint ID is the part of the path after "/v1/" and is used to identify the GatewayEndpoint.
//
// TODO_IMPROVE - see if there is a better way to extract the endpoint ID from the path.
func extractEndpointID(urlPath string) (types.EndpointID, bool) {
	const prefix = "/v1/"
	if strings.HasPrefix(urlPath, prefix) {
		return types.EndpointID(urlPath[len(prefix):]), true
	}
	return "", false
}

/* --------------------------------- Unused - Only to satisfy interface -------------------------------- */

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
