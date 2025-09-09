package jsonrpc

import (
	"encoding/json"
	"fmt"
)

// Method is the method specified by a JSONRPC request.
// See the following link for more details:
// https://www.jsonrpc.org/specification
type Method string
type Version string

const Version2 = Version("2.0")

// Request represents a JSON-RPC 2.0 request.
//
// Specification requirements:
//   - jsonrpc: must be "2.0"
//   - method: string containing the method name
//   - params: structured values (array or object), optional
//   - id: identifier for correlation, always included (null if unset)
//
// Reference: https://www.jsonrpc.org/specification#request_object
type Request struct {
	ID      ID      `json:"id"` // Always include in JSON
	JSONRPC Version `json:"jsonrpc"`
	Method  Method  `json:"method"`
	Params  Params  `json:"params,omitempty"`
}

// MarshalJSON implements json.Marshaler interface.
// Always includes the ID field for JSON-RPC 2.0 compliance.
// Unset IDs are automatically serialized as null.
func (r Request) MarshalJSON() ([]byte, error) {
	// Use a simple struct for marshaling since ID is now always present
	type requestAlias struct {
		JSONRPC Version `json:"jsonrpc"`
		Method  Method  `json:"method"`
		Params  *Params `json:"params,omitempty"`
		ID      ID      `json:"id"`
	}

	out := requestAlias{
		JSONRPC: r.JSONRPC,
		Method:  r.Method,
		ID:      r.ID, // ID.MarshalJSON() handles null case automatically
	}

	if !r.Params.IsEmpty() {
		out.Params = &r.Params
	}

	return json.Marshal(out)
}

// SetParams sets the params field directly from a byte array
func (r *Request) SetParams(params []byte) {
	r.Params = Params{rawMessage: params}
}

// -----------------
// The following functions build Params objects from various input types.
// These are individually defined in order to allow type-safe param construction.
//
// JSON-RPC spec reference: https://www.jsonrpc.org/specification#parameter_structures
// -----------------

// TODO_TECHDEBT(@commoddity): A single method on Request, e.g. SetParams([]byte), should be sufficient.
// These special case methods can then live in the client code.

// BuildParamsFromString builds a Params object from a single string.
//
// For example, for an EVM `eth_getBalance` request, the params would look like:
// params - ["0x28C6c06298d514Db089934071355E5743bf21d60"]
//
// Used for eth_getTransactionReceipt and eth_getTransactionByHash
func BuildParamsFromString(stringParam string) (Params, error) {
	if stringParam == "" {
		return Params{}, fmt.Errorf("param is empty")
	}
	jsonParams, err := json.Marshal([1]string{stringParam})
	if err != nil {
		return Params{}, err
	}
	return Params{rawMessage: jsonParams}, nil
}

// BuildParamsFromStringArray builds a Params object from an array of strings.
//
// For example, for an EVM `eth_getBalance` request, the params would look like:
// params - ["0x28C6c06298d514Db089934071355E5743bf21d60", "0xe71e1d"]]
//
// Used for eth_getBalance, eth_getTransactionCount, and eth_getTransactionReceipt
func BuildParamsFromStringArray(params [2]string) (Params, error) {
	for i, param := range params {
		if param == "" {
			return Params{}, fmt.Errorf("param at index %d is empty", i)
		}
	}
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return Params{}, err
	}
	return Params{rawMessage: jsonParams}, nil
}

// BuildParamsFromStringAndBool builds a Params object from a single string and a boolean.
//
// For example, for an EVM `eth_getBlockByNumber` request, the params would look like:
// params - ["0xe71e1d", false]
//
// Used for eth_getBlockByNumber
func BuildParamsFromStringAndBool(stringParam string, boolParam bool) (Params, error) {
	if stringParam == "" {
		return Params{}, fmt.Errorf("string param is empty")
	}
	jsonParams, err := json.Marshal([2]any{stringParam, boolParam})
	if err != nil {
		return Params{}, err
	}
	return Params{rawMessage: jsonParams}, nil
}

// BuildParamsFromObjectAndString builds a Params object from a map and a string.
//
// For example, for an EVM `eth_call` request, the params would look like:
// params - [{"to":"0xdAC17F958D2ee523a2206206994597C13D831ec7","data":"0x18160ddd"}, "latest"]
//
// Used for eth_call
func BuildParamsFromObjectAndString(objectParam map[string]string, stringParam string) (Params, error) {
	if stringParam == "" {
		return Params{}, fmt.Errorf("string param is empty")
	}
	jsonParams, err := json.Marshal([2]any{objectParam, stringParam})
	if err != nil {
		return Params{}, err
	}
	return Params{rawMessage: jsonParams}, nil
}

// BuildParamsFromStringAndObject builds a Params object from a single string and a map.
//
// For example, for a Solana `getSignaturesForAddress` request, the params would look like:
// params - ["Vote111111111111111111111111111111111111111",{"limit":1}]
//
// Used for getSignaturesForAddress
func BuildParamsFromStringAndObject(stringParam string, objectParam map[string]any) (Params, error) {
	if stringParam == "" {
		return Params{}, fmt.Errorf("string param is empty")
	}
	jsonParams, err := json.Marshal([2]any{stringParam, objectParam})
	if err != nil {
		return Params{}, err
	}
	return Params{rawMessage: jsonParams}, nil
}

// BuildParamsFromUint64AndObject builds a Params object from a single uint64 and a map.
//
// For example, for a Solana `getBlock` request, the params would look like:
// params - [430, {"encoding": "json", "transactionDetails": "full", "maxSupportedTransactionVersion": 0}]
//
// Used for getBlock
func BuildParamsFromUint64AndObject(uint64Param uint64, objectParam map[string]any) (Params, error) {
	if uint64Param == 0 {
		return Params{}, fmt.Errorf("uint64 param is empty")
	}
	jsonParams, err := json.Marshal([2]any{uint64Param, objectParam})
	if err != nil {
		return Params{}, err
	}
	return Params{rawMessage: jsonParams}, nil
}

// BuildParamsFromEmptyObject builds a Params object from an empty object.
// Results in params: {}
func BuildParamsFromEmptyObject() Params {
	return Params{rawMessage: []byte("{}")}
}
