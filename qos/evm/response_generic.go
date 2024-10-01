package evm

import (
	"encoding/json"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseUnmarshallerGeneric unmarshal the provided byte slice
// into a responseGeneric struct and saves any data that may be
// needed for producing a response payload into the struct.
func responseUnmarshallerGeneric(data []byte) (response, error) {
	var response responseGeneric
	err := json.Unmarshal(data, &response)
	if err != nil {
		// TODO_INCOMPLETE: implement a method-specific validator of the response.
		response.unmarshallingErr = err
	}

	response.rawBytes = data
	return response, nil
}

// TODO_UPNEXT(@adshmh): implement the generic jsonrpc response
// (with the scope limited to an EVM-based blockchain)
// responseGeneric captures the fields expected
// in response to any request on an EVM-based blockchain.
// It is intended to be used when no validation/observation
// is applicable to the corresponding request's JSONRPC method.
type responseGeneric struct {
	ID      jsonrpc.ID      `json:"id"`
	JSONRPC jsonrpc.Version `json:"jsonrpc"`

	rawBytes         []byte
	unmarshallingErr error
}

func (r responseGeneric) GetObservation() (observation, bool) {
	return observation{}, false
}

// TODO_UPNEXT(@adshmh): handle any unmarshalling errors
// TODO_INCOMPLETE: build a method-specific payload generator.
func (r responseGeneric) GetResponsePayload() []byte {
	return r.rawBytes
}