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
	// rawMessage stores the actual value of the params field (e.g., ["0x1b4", true]), not the entire JSON-RPC request.
	// It is kept private to ensure all values pass through JSON validation during unmarshaling.
	//
	// According to JSON-RPC 2.0 spec, params must be a structured value.
	// Common blockchain examples:
	//  - Block by number:  {"params": ["0x1b4", true]}  // [blockNum, includeTx]
	//  - Get balance:      {"params": ["0x407d73d8a49eeb85d32cf465507dd71d507100c1", "latest"]} // [address, block]
	rawMessage json.RawMessage
}

func NewParams(rawMessage json.RawMessage) Params {
	return Params{rawMessage: rawMessage}
}

// Custom marshaler allows Params to be serialized while keeping rawMessage private.
// This is needed because Go's default JSON marshaler only processes public fields, but we want to keep rawMessage private
// to enforce JSON-RPC 2.0 validation during unmarshaling.
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

	// Validate that params follows JSON-RPC 2.0 spec: must be array or object.
	// json.Unmarshal into interface{} fails for primitive types as they are not valid top-level JSON structures.
	// Examples:
	//   Valid:   [1, "test"] or {"foo": "bar"}
	//   Invalid: "test" or 42 or true
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
		return fmt.Errorf("params must be either array or object, got %T", checkType)
	}
}

// IsEmpty returns true when params contains no data.
// The JSON marshaler uses this to completely omit the params field from the JSON output when empty.
func (p Params) IsEmpty() bool {
	return len(p.rawMessage) == 0
}
