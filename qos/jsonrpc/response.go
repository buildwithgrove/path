package jsonrpc

import (
	"encoding/json"
	"fmt"
)

// Response captures all the fields of a JSONRPC response.
// See the following link for more details:
// https://www.jsonrpc.org/specification#response_object
type Response struct {
	ID      `json:"id"`
	Version `json:"jsonrpc"`
	// Result captures the result field of the JSONRPC spec.
	// It is allowed to be any arbitrary value as permitted by the spec.
	Result any           `json:"result"`
	Error  ResponseError `json:"error"`
}

func (r Response) Validate() error {
	if r.Version != Version2 {
		return fmt.Errorf("invalid JSONRPC response: jsonrpc field is %q, expected %q", r.Version, Version2)
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
		Error: ResponseError{
			Code:    errCode,
			Message: errMsg,
			Data:    errData,
		},
	}
}
