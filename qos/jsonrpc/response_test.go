package jsonrpc

import (
	"encoding/json"
	"testing"
)

func TestResponse_Validate(t *testing.T) {
	tests := []struct {
		name     string
		reqID    ID
		response Response
		wantErr  bool
	}{
		{
			name:  "should validate successfully with correct version and result",
			reqID: IDFromStr("1"),
			response: Response{
				ID:      IDFromStr("1"),
				Version: Version2,
				Result:  &json.RawMessage(`"success"`),
			},
			wantErr: false,
		},
		{
			name:  "should validate successfully with null result",
			reqID: IDFromStr("1"),
			response: Response{
				ID:      IDFromStr("1"),
				Version: Version2,
				Result:  &json.RawMessage(`null`),
			},
			wantErr: false,
		},
		{
			name:  "should validate successfully with complex result",
			reqID: IDFromStr("1"),
			response: Response{
				ID:      IDFromStr("1"),
				Version: Version2,
				Result:  &json.RawMessage(`{"blockHash":"0x1234","status":"0x1"}`),
			},
			wantErr: false,
		},
		{
			name:  "should fail validation with incorrect version",
			reqID: IDFromStr("1"),
			response: Response{
				ID:      IDFromStr("1"),
				Version: "1.0",
				Result:  &json.RawMessage(`"success"`),
			},
			wantErr: true,
		},
		{
			name:  "should fail validation with both result and error",
			reqID: IDFromStr("1"),
			response: Response{
				ID:      IDFromStr("1"),
				Version: Version2,
				Result:  &json.RawMessage(`"success"`),
				Error:   &ResponseError{Code: 1, Message: "error"},
			},
			wantErr: true,
		},
		{
			name:  "should fail validation with both null result and error",
			reqID: IDFromStr("1"),
			response: Response{
				ID:      IDFromStr("1"),
				Version: Version2,
				Result:  &json.RawMessage(`null`),
				Error:   &ResponseError{Code: 1, Message: "error"},
			},
			wantErr: true,
		},
		{
			name:  "should fail validation with neither result nor error",
			reqID: IDFromStr("1"),
			response: Response{
				ID:      IDFromStr("1"),
				Version: Version2,
				Result:  nil, // no result field
				Error:   nil, // no error field
			},
			wantErr: true,
		},
		{
			name:  "should fail validation with incorrect id",
			reqID: IDFromStr("2"),
			response: Response{
				ID:      IDFromStr("1"),
				Version: Version2,
				Result:  &json.RawMessage(`"success"`),
			},
			wantErr: true,
		},
		{
			name:  "should validate successfully with null request ID and null response ID",
			reqID: ID{}, // empty ID (will serialize as null)
			response: Response{
				ID:      ID{}, // empty ID (will serialize as null)
				Version: Version2,
				Result:  &json.RawMessage(`"success"`),
			},
			wantErr: false,
		},
		{
			name:  "should validate successfully with integer ID",
			reqID: IDFromInt(42),
			response: Response{
				ID:      IDFromInt(42),
				Version: Version2,
				Result:  &json.RawMessage(`"success"`),
			},
			wantErr: false,
		},
		{
			name:  "should validate successfully with zero integer ID",
			reqID: IDFromInt(0),
			response: Response{
				ID:      IDFromInt(0),
				Version: Version2,
				Result:  &json.RawMessage(`"success"`),
			},
			wantErr: false,
		},
		{
			name:  "should fail validation when request ID is null but response ID is not",
			reqID: ID{}, // empty ID (null)
			response: Response{
				ID:      IDFromStr("1"),
				Version: Version2,
				Result:  &json.RawMessage(`"success"`),
			},
			wantErr: true,
		},
		{
			name:  "should fail validation when request ID is set but response ID is null",
			reqID: IDFromStr("1"),
			response: Response{
				ID:      ID{}, // empty ID (null)
				Version: Version2,
				Result:  &json.RawMessage(`"success"`),
			},
			wantErr: true,
		},
		{
			name:  "should validate successfully with error response (no result field)",
			reqID: IDFromStr("1"),
			response: Response{
				ID:      IDFromStr("1"),
				Version: Version2,
				Result:  nil, // no result field for error responses
				Error:   &ResponseError{Code: -32000, Message: "Server error"},
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.response.Validate(test.reqID)
			if (err != nil) != test.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestResponse_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		expected string
	}{
		{
			name: "should marshal response with string result",
			response: Response{
				ID:      IDFromStr("1"),
				Version: Version2,
				Result:  &json.RawMessage(`"success"`),
			},
			expected: `{"id":"1","jsonrpc":"2.0","result":"success"}`,
		},
		{
			name: "should marshal response with null result",
			response: Response{
				ID:      IDFromStr("1"),
				Version: Version2,
				Result:  &json.RawMessage(`null`),
			},
			expected: `{"id":"1","jsonrpc":"2.0","result":null}`,
		},
		{
			name: "should marshal response with numeric result",
			response: Response{
				ID:      IDFromInt(42),
				Version: Version2,
				Result:  &json.RawMessage(`123`),
			},
			expected: `{"id":42,"jsonrpc":"2.0","result":123}`,
		},
		{
			name: "should marshal error response without result field",
			response: Response{
				ID:      IDFromStr("1"),
				Version: Version2,
				Result:  nil, // omitted due to omitempty
				Error:   &ResponseError{Code: -32000, Message: "Server error"},
			},
			expected: `{"id":"1","jsonrpc":"2.0","error":{"code":-32000,"message":"Server error"}}`,
		},
		{
			name: "should marshal response with complex result",
			response: Response{
				ID:      IDFromStr("1"),
				Version: Version2,
				Result:  &json.RawMessage(`{"blockHash":"0x1234","status":"0x1"}`),
			},
			expected: `{"id":"1","jsonrpc":"2.0","result":{"blockHash":"0x1234","status":"0x1"}}`,
		},
		{
			name: "should marshal response with array result",
			response: Response{
				ID:      IDFromStr("1"),
				Version: Version2,
				Result:  &json.RawMessage(`["item1","item2"]`),
			},
			expected: `{"id":"1","jsonrpc":"2.0","result":["item1","item2"]}`,
		},
		{
			name: "should marshal response with boolean result",
			response: Response{
				ID:      IDFromStr("1"),
				Version: Version2,
				Result:  &json.RawMessage(`true`),
			},
			expected: `{"id":"1","jsonrpc":"2.0","result":true}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := json.Marshal(test.response)
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
				return
			}
			if string(data) != test.expected {
				t.Errorf("Marshal() = %s, expected %s", string(data), test.expected)
			}
		})
	}
}

func TestResponse_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		validate func(t *testing.T, response Response)
		wantErr  bool
	}{
		{
			name:     "should unmarshal response with string result",
			jsonData: `{"id":"1","jsonrpc":"2.0","result":"success"}`,
			validate: func(t *testing.T, response Response) {
				if response.ID.String() != "1" {
					t.Errorf("ID = %v, expected '1'", response.ID)
				}
				if response.Version != Version2 {
					t.Errorf("Version = %v, expected %v", response.Version, Version2)
				}
				if response.Result == nil {
					t.Error("Result should not be nil")
					return
				}
				if string(*response.Result) != `"success"` {
					t.Errorf("Result = %s, expected '\"success\"'", string(*response.Result))
				}
			},
			wantErr: false,
		},
		{
			name:     "should unmarshal response with null result",
			jsonData: `{"id":"1","jsonrpc":"2.0","result":null}`,
			validate: func(t *testing.T, response Response) {
				if response.Result == nil {
					t.Error("Result pointer should not be nil for explicit null")
					return
				}
				if string(*response.Result) != "null" {
					t.Errorf("Result = %s, expected 'null'", string(*response.Result))
				}
			},
			wantErr: false,
		},
		{
			name:     "should unmarshal response with missing result field",
			jsonData: `{"id":"1","jsonrpc":"2.0"}`,
			validate: func(t *testing.T, response Response) {
				if response.Result != nil {
					t.Error("Result should be nil when field is absent")
				}
			},
			wantErr: false,
		},
		{
			name:     "should unmarshal response with numeric result",
			jsonData: `{"id":42,"jsonrpc":"2.0","result":123}`,
			validate: func(t *testing.T, response Response) {
				if response.ID.String() != "42" {
					t.Errorf("ID = %v, expected '42'", response.ID)
				}
				if string(*response.Result) != "123" {
					t.Errorf("Result = %s, expected '123'", string(*response.Result))
				}
			},
			wantErr: false,
		},
		{
			name:     "should unmarshal error response",
			jsonData: `{"id":"1","jsonrpc":"2.0","error":{"code":-32000,"message":"Server error"}}`,
			validate: func(t *testing.T, response Response) {
				if response.Result != nil {
					t.Error("Result should be nil for error responses")
				}
				if response.Error == nil {
					t.Error("Error should not be nil")
					return
				}
				if response.Error.Code != -32000 {
					t.Errorf("Error.Code = %d, expected -32000", response.Error.Code)
				}
				if response.Error.Message != "Server error" {
					t.Errorf("Error.Message = %s, expected 'Server error'", response.Error.Message)
				}
			},
			wantErr: false,
		},
		{
			name:     "should unmarshal response with complex result",
			jsonData: `{"id":"1","jsonrpc":"2.0","result":{"blockHash":"0x1234","status":"0x1"}}`,
			validate: func(t *testing.T, response Response) {
				expected := `{"blockHash":"0x1234","status":"0x1"}`
				if string(*response.Result) != expected {
					t.Errorf("Result = %s, expected %s", string(*response.Result), expected)
				}
			},
			wantErr: false,
		},
		{
			name:     "should unmarshal response with array result",
			jsonData: `{"id":"1","jsonrpc":"2.0","result":["item1","item2"]}`,
			validate: func(t *testing.T, response Response) {
				expected := `["item1","item2"]`
				if string(*response.Result) != expected {
					t.Errorf("Result = %s, expected %s", string(*response.Result), expected)
				}
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var response Response
			err := json.Unmarshal([]byte(test.jsonData), &response)
			if (err != nil) != test.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if err == nil && test.validate != nil {
				test.validate(t, response)
			}
		})
	}
}

func TestResponse_GetResultAsBytes(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		expected []byte
		wantErr  bool
	}{
		{
			name: "should return result bytes for valid result",
			response: Response{
				Result: &json.RawMessage(`{"status":"0x1"}`),
			},
			expected: []byte(`{"status":"0x1"}`),
			wantErr:  false,
		},
		{
			name: "should return null bytes for null result",
			response: Response{
				Result: &json.RawMessage(`null`),
			},
			expected: []byte(`null`),
			wantErr:  false,
		},
		{
			name: "should return error for missing result",
			response: Response{
				Result: nil,
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.response.GetResultAsBytes()
			if (err != nil) != test.wantErr {
				t.Errorf("GetResultAsBytes() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if string(result) != string(test.expected) {
				t.Errorf("GetResultAsBytes() = %s, expected %s", string(result), string(test.expected))
			}
		})
	}
}

func TestResponse_UnmarshalResult(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		target   interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name: "should unmarshal string result",
			response: Response{
				Result: &json.RawMessage(`"hello"`),
			},
			target:   new(string),
			expected: "hello",
			wantErr:  false,
		},
		{
			name: "should unmarshal object result",
			response: Response{
				Result: &json.RawMessage(`{"status":"0x1","blockHash":"0x1234"}`),
			},
			target: new(map[string]string),
			expected: map[string]string{
				"status":    "0x1",
				"blockHash": "0x1234",
			},
			wantErr: false,
		},
		{
			name: "should unmarshal null result",
			response: Response{
				Result: &json.RawMessage(`null`),
			},
			target:   new(*string),
			expected: (*string)(nil),
			wantErr:  false,
		},
		{
			name: "should return error for missing result",
			response: Response{
				Result: nil,
			},
			target:  new(string),
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.response.UnmarshalResult(test.target)
			if (err != nil) != test.wantErr {
				t.Errorf("UnmarshalResult() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !test.wantErr {
				// Dereference pointer to get actual value
				switch v := test.target.(type) {
				case *string:
					if *v != test.expected.(string) {
						t.Errorf("UnmarshalResult() = %v, expected %v", *v, test.expected)
					}
				case *map[string]string:
					expected := test.expected.(map[string]string)
					for k, expectedV := range expected {
						if (*v)[k] != expectedV {
							t.Errorf("UnmarshalResult() key %s = %v, expected %v", k, (*v)[k], expectedV)
						}
					}
				case **string:
					if *v != test.expected.(*string) {
						t.Errorf("UnmarshalResult() = %v, expected %v", *v, test.expected)
					}
				}
			}
		})
	}
}

func TestResponse_HelperFunctions(t *testing.T) {
	t.Run("GetErrorResponse", func(t *testing.T) {
		response := GetErrorResponse(IDFromStr("1"), -32000, "Server error", nil)

		if response.Result != nil {
			t.Error("Result should be nil for error response")
		}
		if response.Error == nil {
			t.Error("Error should not be nil")
		}
		if response.Error.Code != -32000 {
			t.Errorf("Error.Code = %d, expected -32000", response.Error.Code)
		}
	})
}

func TestResponse_RealWorldEthereumExample(t *testing.T) {
	// Test the exact case from the original issue
	jsonData := `{"jsonrpc":"2.0","id":1,"result":null}`

	var response Response
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal real-world Ethereum response: %v", err)
		return
	}

	// Validate against the request ID
	reqID := IDFromInt(1)
	err = response.Validate(reqID)
	if err != nil {
		t.Errorf("Real-world Ethereum response should be valid: %v", err)
	}

	// Ensure result field is present but contains null
	if response.Result == nil {
		t.Error("Result field should be present for explicit null")
		return
	}
	if string(*response.Result) != "null" {
		t.Errorf("Result = %s, expected 'null'", string(*response.Result))
	}

	// Ensure it marshals back correctly
	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Failed to marshal response: %v", err)
		return
	}

	expected := `{"id":1,"jsonrpc":"2.0","result":null}`
	if string(data) != expected {
		t.Errorf("Marshal() = %s, expected %s", string(data), expected)
	}
}

func TestResponse_FieldPresenceDistinction(t *testing.T) {
	t.Run("missing result field vs explicit null", func(t *testing.T) {
		// Case 1: Missing result field - should be invalid
		missingJSON := `{"jsonrpc":"2.0","id":1}`
		var missingResponse Response
		err := json.Unmarshal([]byte(missingJSON), &missingResponse)
		if err != nil {
			t.Errorf("Unmarshal failed: %v", err)
			return
		}

		if missingResponse.Result != nil {
			t.Error("Result should be nil when field is absent")
		}

		err = missingResponse.Validate(IDFromInt(1))
		if err == nil {
			t.Error("Validation should fail for missing result field")
		}

		// Case 2: Explicit null result - should be valid
		nullJSON := `{"jsonrpc":"2.0","id":1,"result":null}`
		var nullResponse Response
		err = json.Unmarshal([]byte(nullJSON), &nullResponse)
		if err != nil {
			t.Errorf("Unmarshal failed: %v", err)
			return
		}

		if nullResponse.Result == nil {
			t.Error("Result should not be nil for explicit null field")
		}

		err = nullResponse.Validate(IDFromInt(1))
		if err != nil {
			t.Errorf("Validation should pass for explicit null result: %v", err)
		}
	})
}
