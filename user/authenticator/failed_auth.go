package authenticator

import (
	"fmt"
	"net/http"

	"github.com/buildwithgrove/path/gateway"
)

// TODO_IMPROVE - use correct "id" field in response
// TODO_IMPROVE - formalize error codes
var failedAuthTemplate = `{"jsonrpc":"2.0","error":{"code":%d,"message":"%s"},"id":0}`

var (
	userAppNotFoundCode = -32005
	userAppNotFound     = failedAuth{body: fmt.Sprintf(failedAuthTemplate, userAppNotFoundCode, "user app not found")}
)
var (
	userAuthFailCode              = -32006
	userAuthFailSecretKeyRequired = failedAuth{body: fmt.Sprintf(failedAuthTemplate, userAuthFailCode, "secret key is required")}
	userAuthFailInvalidSecretKey  = failedAuth{body: fmt.Sprintf(failedAuthTemplate, userAuthFailCode, "invalid secret key")}
)

// failedAuth contains a response body for an authentication failure to be
// returned to the client. It satisfies the gateway.HTTPResponse interface.
type failedAuth struct {
	body string
}

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
