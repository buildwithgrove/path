//go:build auth_plugin

package filter

import (
	"github.com/buildwithgrove/auth-plugin/types"
	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
)

type APIKeyAuthorizer struct{}

func (a *APIKeyAuthorizer) authorizeRequest(req api.RequestHeaderMap, endpoint types.GatewayEndpoint) *errorResponse {
	if apiKey, authRequired := endpoint.GetAuth(); authRequired {
		reqAPIKey, ok := req.Get("Authorization")
		if !ok || reqAPIKey == "" {
			return &errAPIKeyRequired
		}
		if reqAPIKey != apiKey {
			return &errAPIKeyInvalid
		}
	}
	return nil
}
