package request

import (
	"errors"
	"fmt"
)

const parserErrorTemplate = `{"code":%d,"message":"%s"}`

var (
	errNoServiceIDProvided = errors.New("no service ID provided")
)

/* Parser Error Response */

type parserErrorResponse struct {
	err  string
	code int
}

func (r *parserErrorResponse) GetPayload() []byte {
	return []byte(fmt.Sprintf(parserErrorTemplate, r.code, r.err))
}

func (r *parserErrorResponse) GetHTTPStatusCode() int {
	return r.code
}

func (r *parserErrorResponse) GetHTTPHeaders() map[string]string {
	return map[string]string{}
}
