//go:build auth_plugin

package filter

import (
	"fmt"
	"net/http"
)

const (
	errorResponseTemplate = `{"code": %d, "message": "%s"}`
)

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

type errorResponse struct {
	code    int
	message string
}

func getErrString(err errorResponse) string {
	return fmt.Sprintf(errorResponseTemplate, err.code, err.message)
}
