package authenticator

import (
	"fmt"
	"net/http"
)

// TODO_IMPROVE - use correct "id" field in response
// TODO_IMPROVE - formalize error codes
var invalidRespTemplate = `{"jsonrpc":"2.0","error":{"code":%d,"message":"%s"},"id":0}`

var (
	userAppNotFoundCode = -32005
	userAppNotFound     = invalidResp{body: fmt.Sprintf(invalidRespTemplate, userAppNotFoundCode, "user app not found")}
)
var (
	userAuthFailCode              = -32006
	userAuthFailSecretKeyRequired = invalidResp{body: fmt.Sprintf(invalidRespTemplate, userAuthFailCode, "secret key is required")}
	userAuthFailInvalidSecretKey  = invalidResp{body: fmt.Sprintf(invalidRespTemplate, userAuthFailCode, "invalid secret key")}
)
var (
	rateLimitExceededCode   = -32007
	throughputLimitExceeded = invalidResp{body: fmt.Sprintf(invalidRespTemplate, rateLimitExceededCode, "throughput limit exceeded")}
)

// invalidResp contains a response body for an authentication failure to be
// returned to the client. It satisfies the gateway.HTTPResponse interface.
type invalidResp struct {
	body string
}

func (r *invalidResp) GetPayload() []byte {
	return []byte(r.body)
}

func (r *invalidResp) GetHTTPStatusCode() int {
	return http.StatusUnauthorized
}

func (r *invalidResp) GetHTTPHeaders() map[string]string {
	return map[string]string{}
}
