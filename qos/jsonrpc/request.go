package jsonrpc

import (
	"encoding/json"
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

// NewRequest constructs a new JSONRPC request from the given ID, method, and optional parameters.
//
// For example:
// id - 1
// method - eth_getBalance
// params - ["0x28C6c06298d514Db089934071355E5743bf21d60", "0xe71e1d"]
//
// The above request would be serialized as:
// {"jsonrpc":"2.0","id":1,"method":"eth_getBalance","params":["0x28C6c06298d514Db089934071355E5743bf21d60", "0xe71e1d"]}
func NewRequest(id int, method Method, params ...any) Request {
	request := Request{
		JSONRPC: Version2,
		ID:      IDFromInt(id),
		Method:  method,
	}

	if len(params) > 0 {
		jsonParams, err := json.Marshal(params)
		if err == nil {
			request.Params = NewParams(jsonParams)
		}
	}

	return request
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
