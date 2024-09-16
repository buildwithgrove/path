//go:build auth_plugin

package filter

import (
	"fmt"
	"net/http"

	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
)

const (
	errorResponseTemplate = `{"code": %d, "message": "%s"}`
)

type errorResponse struct {
	code    int
	message string
}

var (
	errPathNotProvided = errorResponse{
		code:    http.StatusBadRequest,
		message: "path not provided",
	}
	errEndpointIDNotProvided = errorResponse{
		code:    http.StatusBadRequest,
		message: "endpoint ID not provided",
	}
	errEndpointNotFound = errorResponse{
		code:    http.StatusNotFound,
		message: "endpoint not found",
	}
	errAPIKeyRequired = errorResponse{
		code:    http.StatusUnauthorized,
		message: "API key required",
	}
	errAPIKeyInvalid = errorResponse{
		code:    http.StatusUnauthorized,
		message: "invalid API key",
	}
)

func getErrString(err errorResponse) string {
	return fmt.Sprintf(errorResponseTemplate, err.code, err.message)
}

// Sync Errors - these errors are used to stop the filter chain by returning a api.StatusType, which is returned by DecodeHeaders

func sendErrPathNotProvided(callbacks api.FilterCallbackHandler) api.StatusType {
	return sendErrResponse(callbacks, errPathNotProvided)
}

func sendErrEndpointIDNotProvided(callbacks api.FilterCallbackHandler) api.StatusType {
	return sendErrResponse(callbacks, errEndpointIDNotProvided)
}

func sendErrResponse(callbacks api.FilterCallbackHandler, err errorResponse) api.StatusType {
	decoderCallbacks := callbacks.DecoderFilterCallbacks()
	decoderCallbacks.SendLocalReply(err.code, getErrString(err), nil, 0, "")
	return api.LocalReply
}

// Async Errors - these errors are sent asynchronously and are used to stop the filter chain
// by calling callbacks.Continue(api.LocalReply) directly while the api.Running status is returned by DecodeHeaders

func sendAsyncErrEndpointNotFound(callbacks api.FilterCallbackHandler) {
	sendAsyncErrResponse(callbacks, errEndpointNotFound)
}

func sendAsyncErrAPIKeyRequired(callbacks api.FilterCallbackHandler) {
	sendAsyncErrResponse(callbacks, errAPIKeyRequired)
}

func sendAsyncErrAPIKeyInvalid(callbacks api.FilterCallbackHandler) {
	sendAsyncErrResponse(callbacks, errAPIKeyInvalid)
}

func sendAsyncErrResponse(callbacks api.FilterCallbackHandler, err errorResponse) {
	decoderCallbacks := callbacks.DecoderFilterCallbacks()
	decoderCallbacks.SendLocalReply(err.code, getErrString(err), nil, 0, "")
}
