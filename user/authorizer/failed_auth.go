package authorizer

import (
	"fmt"
	"net/http"

	"github.com/buildwithgrove/path/gateway"
)

// TODO_TECHDEBT - use the original "id" field from the request in JSON-RPC 2.0 response
var failedAuthTemplate = `{"jsonrpc":"2.0","error":{"code":%d,"message":"%s"},"id":0}`

var (
	userAppNotFoundCode = http.StatusNotFound
	userAppNotFound     = failedAuth{body: fmt.Sprintf(failedAuthTemplate, userAppNotFoundCode, "user app not found")}
)
var (
	userAuthFailCode           = http.StatusUnauthorized
	userAuthFailAPIKeyRequired = failedAuth{body: fmt.Sprintf(failedAuthTemplate, userAuthFailCode, "secret key is required")}
	userAuthFailInvalidAPIKey  = failedAuth{body: fmt.Sprintf(failedAuthTemplate, userAuthFailCode, "invalid secret key")}
)

// failedAuth contains a JSON-RPC 2.0 response body, including an error code and message,
// for an authentication failure to be returned to the client.
type failedAuth struct {
	body string
}

// failedAuth satisfies the gateway.HTTPResponse interface.
var _ gateway.HTTPResponse = &failedAuth{}

func (r *failedAuth) GetPayload() []byte {
	return []byte(r.body)
}

func (r *failedAuth) GetHTTPStatusCode() int {
	return http.StatusUnauthorized
}

func (r *failedAuth) GetHTTPHeaders() map[string]string {
	return map[string]string{}
}
