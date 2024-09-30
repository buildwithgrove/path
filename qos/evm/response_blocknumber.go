package evm

import (
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_IN_THIS_COMMIT: implement this unmarshaller
func responseUnmarshallerBlockNumber(data []byte) (response, error) {
	return responseToBlockNumber{}, nil
}

// responseToBlockNumber captures the fields expected in a
// response to an `eth_blockNumber` request.
type responseToBlockNumber struct {
	ID      jsonrpc.ID      `json:"id"`
	JSONRPC jsonrpc.Version `json:"jsonrpc"`
	Result  string          `json:"result"`

	// TODO_FUTURE: build the response payload instead of keeping a copy.
	responsePayload []byte
}

func (r responseToBlockNumber) GetObservation() (observation, bool) {
	return observation{
		BlockHeight: r.Result,
	}, true
}

func (r responseToBlockNumber) GetResponsePayload() []byte {
	// TODO_INCOMPLETE: return a JSONRPC response indicating the error,
	// if the unmarshalling failed.
	return r.responsePayload
}
