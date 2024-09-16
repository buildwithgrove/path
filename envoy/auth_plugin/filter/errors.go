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
)

func sendErrPathNotProvided(callbacks api.FilterProcessCallbacks) api.StatusType {
	return sendErrResponse(callbacks, errPathNotProvided)
}

func sendErrEndpointIDNotProvided(callbacks api.FilterProcessCallbacks) api.StatusType {
	return sendErrResponse(callbacks, errEndpointIDNotProvided)
}

func sendErrEndpointNotFound(callbacks api.FilterProcessCallbacks) api.StatusType {
	return sendErrResponse(callbacks, errEndpointNotFound)
}

func sendErrResponse(callbacks api.FilterProcessCallbacks, err errorResponse) api.StatusType {
	callbacks.SendLocalReply(err.code, getErrString(err), nil, 0, "if endpoint not found, return 404")
	return api.LocalReply
}

func sendAsyncErrResponse(callbacks api.FilterProcessCallbacks, err errorResponse) {
	callbacks.SendLocalReply(err.code, getErrString(err), nil, 0, "if endpoint not found, return 404")
	callbacks.Continue(api.LocalReply)
}

func getErrString(err errorResponse) string {
	return fmt.Sprintf(errorResponseTemplate, err.code, err.message)
}
