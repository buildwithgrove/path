//go:build auth_plugin

package filter

import (
	"context"
	"fmt"
	"strings"

	"github.com/buildwithgrove/authorizer-plugin/types"
	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
)

// The HTTPFilter struct implements the Envoy `api.StreamFilter` interface and is responsible
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

// All processing of the service request is done in DecodeHeaders. This includes:
//
// - extracting the endpoint ID from the path
//
// - performing authorization checks on the request
//
// - setting the appropriate headers (x-endpoint-id, x-account-id, x-rate-limit-throughput)
//
// - sending an error response if the request is not valid
//
// - forwarding the request to the PATH service if the request is valid
//
// endStream is true if the request doesn't have a body.
func (f *HTTPFilter) DecodeHeaders(req api.RequestHeaderMap, endStream bool) api.StatusType {
	path := req.Path()
	if path == "" {
		return sendErrPathNotProvided(f.Callbacks)
	}

	// If the path is "/healthz", we don't need to authenticate
	if path == "/healthz" {
		return api.Continue
	}

	// If the path is "/v1/{gateway_endpoint_id}", we need to authenticate
	endpointID, ok := extractEndpointID(path)
	if !ok {
		return sendErrEndpointIDNotProvided(f.Callbacks)
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
			return
		}

		// Then, check if the API key is required and valid
		if apiKey, authRequired := gatewayEndpoint.GetAuth(); authRequired {

			// If the API key is required, check if it is provided in the req auth header
			reqAPIKey, ok := req.Get("Authorization")
			if !ok || reqAPIKey == "" {
				sendAsyncErrAPIKeyRequired(f.Callbacks)
				return
			}

			// If the API key in the req header does not match the endpoint's API key, return an error
			if reqAPIKey != apiKey {
				sendAsyncErrAPIKeyInvalid(f.Callbacks)
				return
			}
		}

		// Set endpoint ID in the headers
		req.Set("x-endpoint-id", string(gatewayEndpoint.EndpointID))

		// Set rate limit headers if the gateway endpoint should be rate limited
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

/* --------------------------------- Unused -------------------------------- */
// Present only to satisfy Envoy's `api.StreamFilter` interface

func (f *HTTPFilter) DecodeData(buffer api.BufferInstance, endStream bool) api.StatusType {
	return api.Continue
}
func (f *HTTPFilter) DecodeTrailers(trailers api.RequestTrailerMap) api.StatusType {
	return api.Continue
}
func (f *HTTPFilter) EncodeData(buffer api.BufferInstance, endStream bool) api.StatusType {
	return api.Continue
}
func (f *HTTPFilter) EncodeHeaders(header api.ResponseHeaderMap, endStream bool) api.StatusType {
	return api.Continue
}
func (f *HTTPFilter) EncodeTrailers(trailers api.ResponseTrailerMap) api.StatusType {
	return api.Continue
}
func (f *HTTPFilter) OnDestroy(reason api.DestroyReason) {}
func (f *HTTPFilter) OnLog()                             {}
func (f *HTTPFilter) OnLogDownstreamPeriodic()           {}
func (f *HTTPFilter) OnLogDownstreamStart()              {}
