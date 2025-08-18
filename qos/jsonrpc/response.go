package jsonrpc

import (
	"encoding/json"
	"fmt"
)

// Response captures all the fields of a JSONRPC response.
// See the following link for more details:
// https://www.jsonrpc.org/specification#response_object
//
// Design decisions:
// • Result uses *json.RawMessage to distinguish field absence vs explicit null
// • Pointer nil = field omitted in JSON (invalid per spec)
// • Pointer to null bytes = {"result":null} (valid for methods like eth_getTransactionReceipt)
// • json.RawMessage avoids double marshaling and preserves original JSON structure
// • Custom UnmarshalJSON handles the edge case where JSON null becomes Go nil
type Response struct {
	// ID member is required.
	// It must be the same as the value of the id member in the Request Object.
	// If there was an error in detecting the id in the Request object (e.g. Parse error/Invalid Request), it MUST be Null.
	ID ID `json:"id"`
	// Version must be exactly "2.0"
	Version Version `json:"jsonrpc"`
	// Result captures the result field of the JSONRPC spec.
	// It is allowed to be any arbitrary value as permitted by the spec.
	// It is required on success and must not exist if there was an error invoking the method.
	// Using a pointer to json.RawMessage to distinguish between absent field vs explicit null.
	Result *json.RawMessage `json:"result,omitempty"`
	// Error captures the error field of the JSONRPC spec.
	// Is is required on error and must not exist if there was no error triggered during invocation.
	Error *ResponseError `json:"error,omitempty"`
}

func (r Response) Validate(reqID ID) error {
	if r.Version != Version2 {
		return fmt.Errorf("invalid JSONRPC response: jsonrpc field is %q, expected %q", r.Version, Version2)
	}

	// Check if result field is present (pointer is non-nil) vs absent (pointer is nil)
	hasResult := r.Result != nil
	hasError := r.Error != nil

	if !hasResult && !hasError {
		return fmt.Errorf("invalid JSONRPC response: either the result or error must be included")
	}
	if hasResult && hasError {
		return fmt.Errorf("invalid JSONRPC response: both result and error must not be included")
	}
	if r.ID.String() != reqID.String() {
		return fmt.Errorf("invalid JSONRPC response: id field is %q, expected %q", r.ID, reqID)
	}
	return nil
}

func (r Response) GetResultAsBytes() ([]byte, error) {
	if r.Result == nil {
		return nil, fmt.Errorf("no result field present")
	}
	return *r.Result, nil
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

// TODO_TECHDEBT(@adshmh): Validate the results JSONRPC result struct:
// - Return an error if invalid: e.g. if missing both result and error fields.
// - Define and use exported errors for each validation failure scenario.
// - Add a method to construct an observation of type observation.qos.JsonRpcResponseValidationError from the above error.
//
// UnmarshalJSON implements custom unmarshaling to handle the result field presence detection
func (r *Response) UnmarshalJSON(data []byte) error {
	// Use a temporary struct to unmarshal into, avoiding infinite recursion
	type TempResponse struct {
		ID      ID              `json:"id"`
		Version Version         `json:"jsonrpc"`
		Result  json.RawMessage `json:"result"` // Note: no pointer, no omitempty
		Error   *ResponseError  `json:"error,omitempty"`
	}

	var temp TempResponse
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Copy fields
	r.ID = temp.ID
	r.Version = temp.Version
	r.Error = temp.Error

	// Check if result field was present in the original JSON
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return err
	}

	if resultValue, hasResult := rawMap["result"]; hasResult {
		// Field was present, even if null
		r.Result = &resultValue
	} else {
		// Field was absent
		r.Result = nil
	}

	return nil
}

// UnmarshalResult unmarshals the result into the provided value
func (r Response) UnmarshalResult(v any) error {
	if r.Result == nil {
		return fmt.Errorf("no result field present")
	}
	return json.Unmarshal(*r.Result, v)
}
