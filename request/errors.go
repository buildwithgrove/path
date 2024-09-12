package request

import (
	"fmt"
	"net/http"
)

const (
	// TODO_TECHDEBT: return the correct id field from the request
	parserErrorTemplate = `{"jsonrpc":"2.0","id":"0","error":{"code":%d,"message":"%s"}}`
)

/* Parser Error Response */

type ParserErrorResponse struct {
	err string
}

func (r *ParserErrorResponse) GetPayload() []byte {
	return []byte(fmt.Sprintf(parserErrorTemplate, http.StatusBadRequest, r.err))
}

func (r *ParserErrorResponse) GetHTTPStatusCode() int {
	return http.StatusOK // JSON-RPC 2.0 spec requires a 200 status code even for errors.
}

func (r *ParserErrorResponse) GetHTTPHeaders() map[string]string {
	return map[string]string{}
}
