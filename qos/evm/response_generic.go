package evm

import (
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_IN_THIS_COMMIT: implement this unmarshaller
func responseUnmarshallerGeneric(data []byte) (response, error) {
	return responseGeneric{}, nil
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

	payload []byte
}

func (r responseGeneric) GetObservation() (observation, bool) {
	return observation{}, false
}

// TODO_IN_THIS_COMMIT: implement this method and add unit tests.
func (r responseGeneric) GetResponsePayload() []byte {
	return nil
}

// TODO_IN_THIS_COMMIT: implement this method and add unit tests.
func (r responseGeneric) buildPayload() []byte {
	return nil
}
