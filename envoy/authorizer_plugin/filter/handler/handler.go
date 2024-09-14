package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/ardikabs/gonvoy"
	"github.com/buildwithgrove/path-authorizer/user"
)

type Handler struct {
	gonvoy.PassthroughHttpFilterHandler
	Cache cache
}

type cache interface {
	GetGatewayEndpoint(ctx context.Context, userAppID user.EndpointID) (user.GatewayEndpoint, bool)
}

func (h *Handler) OnRequestHeader(c gonvoy.Context) error {
	if h.Cache == nil {
		return fmt.Errorf("cache is not initialized")
	}

	req := c.Request()

	endpointID := user.EndpointID(extractV1Path(req.URL.Path))

	gatewayEndpoint, ok := h.Cache.GetGatewayEndpoint(req.Context(), endpointID)
	if !ok {
		return fmt.Errorf("gateway endpoint not found")
	}

	c.RequestHeader().Add("x-endpoint-id", string(gatewayEndpoint.EndpointID))
	c.RequestHeader().Add("x-account-id", string(gatewayEndpoint.UserAccount.AccountID))
	c.RequestHeader().Add("x-plan", string(gatewayEndpoint.UserAccount.PlanType))
	c.RequestHeader().Add("x-rate-limit-throughput", fmt.Sprintf("%d", gatewayEndpoint.GetThroughputLimit()))

	return nil
}

func (h *Handler) OnResponseHeader(c gonvoy.Context) error {
	return nil
}

func extractV1Path(urlPath string) string {
	const prefix = "/v1/"
	if idx := strings.Index(urlPath, prefix); idx != -1 {
		return urlPath[idx+len(prefix):]
	}
	return ""
}
