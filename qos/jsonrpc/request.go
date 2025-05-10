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

// Request represents a request as specificed
// by the JSONRPC spec.
// See the following link for more details:
// https://www.jsonrpc.org/specification#request_object
type Request struct {
	ID      ID      `json:"id,omitempty"`
	JSONRPC Version `json:"jsonrpc"`
	Method  Method  `json:"method"`
	Params  Params  `json:"params,omitempty"`
}

// MarshalJSON customizes the JSON serialization of a Request.
// It returns a serialized version of the receiver with empty fields (e.g. ID, Params, etc) omitted
func (r Request) MarshalJSON() ([]byte, error) {
	// Define a structure that makes ID and Params optional in the JSON output
	type requestAlias struct {
		JSONRPC Version `json:"jsonrpc"`
		Method  Method  `json:"method"`
		Params  *Params `json:"params,omitempty"` // Optional in JSON output
		ID      *ID     `json:"id,omitempty"`     // Optional in JSON output
	}

	// Build the serializable version of the request
	out := requestAlias{
		JSONRPC: r.JSONRPC,
		Method:  r.Method,
	}

	// Only include non-empty fields
	if !r.ID.IsEmpty() {
		out.ID = &r.ID
	}
	if !r.Params.IsEmpty() {
		out.Params = &r.Params
	}

	// Marshal and return the serializable version of the request
	return json.Marshal(out)
}

// -----------------
// The following functions build Params objects from various input types.
// These are individually defined in order to allow type-safe param construction.
//
// JSON-RPC spec reference: https://www.jsonrpc.org/specification#parameter_structures
// -----------------

// BuildParamsFromStringArray builds a Params object from an array of strings.
//
// For example, for an `eth_getBalance` request, the params would look like:
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
// For example, for an `eth_getBalance` request, the params would look like:å
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
// For example, for an `eth_getBlockByNumber` request, the params would look like:
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
// For example, for an `eth_call` request, the params would look like:
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
