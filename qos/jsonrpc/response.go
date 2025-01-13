package jsonrpc

import (
	"encoding/json"
	"fmt"
)

// Response captures all the fields of a JSONRPC response.
// See the following link for more details:
// https://www.jsonrpc.org/specification#response_object
type Response struct {
	// ID member is required.
	// It must be the same as the value of the id member in the Request Object.
	// If there was an error in detecting the id in the Request object (e.g. Parse error/Invalid Request), it MUST be Null.
	ID `json:"id"`
	// Version must be exactly "2.0"
	Version `json:"jsonrpc"`
	// Result captures the result field of the JSONRPC spec.
	// It is allowed to be any arbitrary value as permitted by the spec.
	// It is required on success and must not exist if there was an error invoking the method.
	Result any `json:"result,omitempty"`
	// Error captures the error field of the JSONRPC spec.
	// Is is required on error and must not exist if there was no error triggered during invocation.
	Error *ResponseError `json:"error,omitempty"`
}

func (r Response) Validate() error {
	if r.Version != Version2 {
		return fmt.Errorf("invalid JSONRPC response: jsonrpc field is %q, expected %q", r.Version, Version2)
	}
	if r.Result == nil && r.Error == nil {
		return fmt.Errorf("invalid JSONRPC response: either the result or error must be included")
	}
	if r.Result != nil && r.Error != nil {
		return fmt.Errorf("invalid JSONRPC response: both result and error must not be included")
	}
	return nil
}

func (r Response) GetResultAsBytes() ([]byte, error) {
	return json.Marshal(r.Result)
}

// GetErrorResponse is a helper function that builds a JSONRPC Response using the supplied ID and error values.
func GetErrorResponse(id ID, errCode int, errMsg string, errData map[string]string) Response {
	return Response{
		ID:      id,
		Version: Version2,
		Error: &ResponseError{
			Code:    errCode,
			Message: errMsg,
			Data:    errData,
		},
	}
}

func (r *Response) IsError() bool {
	return r.Error != nil
}
