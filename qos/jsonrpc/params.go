package jsonrpc

import (
	"encoding/json"
	"fmt"
)

// Params represents the 'params' field in a JSON-RPC request. It accepts any valid JSON data, including:
//   - Objects (single or array)
//   - Strings (single or array)
//   - Basic Go types (single value or array, can be mixed types)
//
// Params only validates JSON formatting - it does not perform method-specific validation.
// Individual request handlers must implement their own parameter validation logic.
//
// See the below link on JSONRPC spec for more details:
// https://www.jsonrpc.org/specification#parameter_structures
type Params struct {
	// Stores the JSON-formatted payload.
	// Declared as private to ensure only valid JSON is accepted as the value, through the custom unmarshaler.
	rawMessage json.RawMessage
}

// Custom marshaler enforces JSON-RPC 2.0 param validation by keeping raw data private and only allowing params to be set through validated unmarshaling.
func (p Params) MarshalJSON() ([]byte, error) {
	return p.rawMessage, nil
}

// Custom unmarshaler ensures incoming data complies with JSON-RPC 2.0 specification
func (p *Params) UnmarshalJSON(data []byte) error {
	// First validate the input is valid JSON.
	var rawMessage json.RawMessage
	if err := json.Unmarshal(data, &rawMessage); err != nil {
		return fmt.Errorf("failed to unmarshal params field: %v", err)
	}

	// Then validate it's either array or object
	var checkType interface{}
	if err := json.Unmarshal(data, &checkType); err != nil {
		return err
	}

	switch checkType.(type) {
	// The only valid types for params are an array or an object.
	case []interface{}, map[string]interface{}:
		p.rawMessage = rawMessage
		return nil
	default:
		return fmt.Errorf("params must be either array or object")
	}
}

// IsEmpty returns true when params contains no data.
// The JSON marshaler uses this to completely omit the params field from the JSON output when empty.
func (p Params) IsEmpty() bool {
	return len(p.rawMessage) == 0
}
