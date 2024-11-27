package auth

import (
	"fmt"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
)

// errUnauthorized is returned when a request is not authorized.
// It is left intentionally vague to avoid leaking information to the client.
var errUnauthorized = fmt.Errorf("unauthorized")

// The Authorizer interface performs requests authorization, for example using
// API key authentication to ensures a downstream (client) request is authorized.
type Authorizer interface {
	authorizeRequest(headers map[string]string, endpoint *proto.GatewayEndpoint) error
}
