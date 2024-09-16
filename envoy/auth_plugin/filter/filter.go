//go:build auth_plugin

package filter

import (
	"fmt"
	"strings"

	"github.com/buildwithgrove/auth-plugin/types"
	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
)

var _ api.StreamFilter = &HTTPFilter{}

// The HTTPFilter struct implements the Envoy `api.StreamFilter` interface and is responsible
// for authorizing and settings headers for requests to the PATH service.
//
// See https://github.com/envoyproxy/envoy/blob/main/contrib/golang/common/go/api/filter.go
type HTTPFilter struct {
	// api.PassThrhoughStreamFilter is provided by Envoy.
	api.PassThroughStreamFilter
	// Callbacks are provided by Envoy and are used to control the filter's behavior
	Callbacks api.FilterCallbackHandler
	// Config is provided by Envoy and represents the hardcoded configuration from `envoy.yaml`.
	// NOTE: Currently we do not use any fields from `envoy.yaml` for configuring the filter.
	Config *EnvoyConfig

	// The userDataCache contains an in-memory cache of GatewayEndpoints
	// and their associated data from the connected Postgres database.
	Cache userDataCache
	// The authorizers represents a list of authorization types that must
	// pass before a request may be forwarded to the PATH service.
	// Configured in `main.go` and passed to the filter.
	Authorizers []Authorizer
}

// The userDataCache contains an in-memory cache of GatewayEndpoints
// and their associated data from the connected Postgres database.
type userDataCache interface {
	GetGatewayEndpoint(types.EndpointID) (types.GatewayEndpoint, bool)
}

// The Authorizer interface performs requests authorization, for example using
// API key authentication to ensures a downstream (client) request is authorized.
type Authorizer interface {
	authorizeRequest(api.RequestHeaderMap, types.GatewayEndpoint) *errorResponse
}

/* --------------------------------- Service Request Processing -------------------------------- */

const (
	reqHeaderEndpointID = "x-endpoint-id"
	reqHeaderThroughput = "x-rate-limit-throughput"
)

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
		f.sendErrResponse(errPathNotProvided)
		return api.LocalReply
	}

	// If the path is "/healthz", we don't need to authenticate
	if path == "/healthz" {
		return api.Continue
	}

	// If the request is for `/v1/{gateway_endpoint_id}`, we need to authenticate

	// Extract the endpoint ID from the path
	endpointID, ok := extractEndpointID(path)
	if !ok {
		f.sendErrResponse(errEndpointIDNotProvided)
		return api.LocalReply
	}

	// To avoid blocking the Envoy thread, run the code in a background goroutine
	// and suspend & resume the filter while handling GatewayEndpoint data
	go func() {
		defer f.Callbacks.DecoderFilterCallbacks().RecoverPanic()
		f.handleGatewayEndpoint(req, endpointID)
	}()

	// Suspend the filter while the background goroutine handles the GatewayEndpoint
	return api.Running
}

// handleGatewayEndpoints performs all required handling on GatewayEndpoint data associated with a service request
func (f *HTTPFilter) handleGatewayEndpoint(req api.RequestHeaderMap, endpointID types.EndpointID) {

	// If GatewayEndpoint is not found send an error response downstream (client)
	gatewayEndpoint, ok := f.getGatewayEndpoint(endpointID)
	if !ok {
		f.sendErrResponse(errEndpointNotFound)
		return
	}

	// If the request is not authorized, send an error response downstream (client)
	if errResp := f.authGatewayEndpoint(req, gatewayEndpoint); errResp != nil {
		f.sendErrResponse(*errResp)
		return
	}

	// Set the headers on the request to be forwarded to the PATH service
	f.setHeaders(req, gatewayEndpoint)

	// Resume the filter when done
	f.Callbacks.DecoderFilterCallbacks().Continue(api.Continue)
}

// getGatewayEndpoint fetches the GatewayEndpoint from the database and a bool indicating if it was found
func (f *HTTPFilter) getGatewayEndpoint(endpointID types.EndpointID) (types.GatewayEndpoint, bool) {
	return f.Cache.GetGatewayEndpoint(endpointID)
}

// authGatewayEndpoint performs all configured authorization checks on the request
func (f *HTTPFilter) authGatewayEndpoint(req api.RequestHeaderMap, gatewayEndpoint types.GatewayEndpoint) *errorResponse {
	for _, auth := range f.Authorizers {
		if errResp := auth.authorizeRequest(req, gatewayEndpoint); errResp != nil {
			return errResp
		}
	}
	return nil
}

// setHeaders sets all headers required by the PATH services on the request being forwarded
func (f *HTTPFilter) setHeaders(req api.RequestHeaderMap, gatewayEndpoint types.GatewayEndpoint) {
	// Set endpoint ID in the headers
	req.Set(reqHeaderEndpointID, string(gatewayEndpoint.EndpointID))

	// Set rate limit headers if the gateway endpoint should be rate limited
	if gatewayEndpoint.RateLimiting.ThroughputLimit > 0 {
		req.Set(reqHeaderThroughput, fmt.Sprintf("%d", gatewayEndpoint.RateLimiting.ThroughputLimit))
	}
}

// sendErrResponse sends a local reply error response for requests that failed authorization checks
func (f *HTTPFilter) sendErrResponse(err errorResponse) {
	f.Callbacks.DecoderFilterCallbacks().SendLocalReply(
		err.code,                            // HTTP status code
		getErrString(err),                   // error body string
		nil,                                 // headers
		0,                                   // gRPC status
		"return error response to upstream", // error details
	)
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
