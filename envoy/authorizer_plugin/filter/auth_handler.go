//go:build authorizer_plugin

package filter

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ardikabs/gonvoy"

	"github.com/buildwithgrove/authorizer-plugin/user"
)

type AuthorizationHandler struct {
	gonvoy.PassthroughHttpFilterHandler
	cache cache
}

const jsonError = `{"code":%d,"message":"%s"}`

func (h *AuthorizationHandler) OnRequestHeader(c gonvoy.Context) error {
	req := c.Request()

	endpointID := user.EndpointID(extractV1Path(req.URL.Path))

	gatewayEndpoint, ok := h.cache.GetGatewayEndpoint(req.Context(), endpointID)
	if !ok {
		return c.JSON(http.StatusNotFound, []byte(fmt.Sprintf(jsonError, http.StatusNotFound, fmt.Sprintf("endpoint %s not found", endpointID))), nil)
	}

	c.RequestHeader().Add("x-endpoint-id", string(gatewayEndpoint.EndpointID))
	c.RequestHeader().Add("x-account-id", string(gatewayEndpoint.UserAccount.AccountID))
	c.RequestHeader().Add("x-plan", string(gatewayEndpoint.UserAccount.PlanType))
	c.RequestHeader().Add("x-rate-limit-throughput", fmt.Sprintf("%d", gatewayEndpoint.GetThroughputLimit()))

	return nil
}

func (h *AuthorizationHandler) OnResponseHeader(c gonvoy.Context) error {
	return nil
}

func extractV1Path(urlPath string) string {
	const prefix = "/v1/"
	if idx := strings.Index(urlPath, prefix); idx != -1 {
		return urlPath[idx+len(prefix):]
	}
	return ""
}
