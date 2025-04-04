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

// BuildParamsFromStrings builds a Params object from an array of strings.
//
// For example, for an `eth_getBalance` request, the params would look like:
// params - ["0x28C6c06298d514Db089934071355E5743bf21d60", "0xe71e1d"]]
//
// JSON-RPC array params must be passed in the order specified by the method.
// Reference: https://www.jsonrpc.org/specification#parameter_structures
//
// TODO_FUTURE(@commoddity): other helper methods may be required to build
// params for different JSON-RPC methods, eg. ["<string>", <bool>]
func BuildArrayParamsFromStrings(params [2]string) (Params, error) {
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
