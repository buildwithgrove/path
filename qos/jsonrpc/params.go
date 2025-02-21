package jsonrpc

import (
	"encoding/json"
	"errors"
)

// Params stores the data contained in the `params` field of a JSONRPC request.
// As of PR #170, it supports:
// - An array of objects
// - An array of strings
// See the below link on JSONRPC spec for more details:
// https://www.jsonrpc.org/specification#parameter_structures
type Params struct {
	// Params stores the raw JSON-RPC parameters without validation.
	// The method-specific request handlers are responsible for validating the parameters based on the JSON-RPC method being called.
	rawPayload []byte
}

func (p Params) MarshalJSON() ([]byte, error) {
	return p.rawPayload, nil
}

func (p *Params) UnmarshalJSON(data []byte) error {
	// Try first as a structure with array of interface{}
	var genericValue []interface{}
	if err := json.Unmarshal(data, &genericValue); err == nil {
		p.rawPayload = data
		return nil
	}

	// Try second as a structure with array of strings
	var stringsValue []string
	if err := json.Unmarshal(data, &stringsValue); err == nil {
		p.rawPayload = data
		return nil
	}

	// If both failed, return the error from the first attempt
	return errors.New("failed to unmarshal as either []interface{} or []string")
}
