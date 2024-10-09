package evm

import (
	"encoding/json"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseUnmarshallerBlockNumber deserializes the provided payload
// into a responseToBlockNumber struct, adding any encountered errors
// to the returned struct.
func responseUnmarshallerBlockNumber(data []byte) (response, error) {
	var response responseToBlockNumber
	err := json.Unmarshal(data, &response)
	if err != nil {
		response.unmarshallingErr = err
	}

	return response, err
}

// responseToBlockNumber captures the fields expected in a
// response to an `eth_blockNumber` request.
type responseToBlockNumber struct {
	ID      jsonrpc.ID      `json:"id"`
	JSONRPC jsonrpc.Version `json:"jsonrpc"`
	Result  string          `json:"result"`

	unmarshallingErr error
}

func (r responseToBlockNumber) GetObservation() (observation, bool) {
	return observation{
		BlockHeight: r.Result,
	}, true
}

func (r responseToBlockNumber) GetResponsePayload() []byte {
	if r.unmarshallingErr != nil {
		// TODO_UPNEXT(@adshmh): return a JSONRPC response indicating the error,
		// if the unmarshalling failed.
		return []byte("{}")
	}

	bz, err := json.Marshal(r)
	if err != nil {
		// TODO_UPNEXT(@adshmh): return a JSONRPC response indicating the error,
		// if marshalling failed.
		return []byte("{}")
	}

	return bz
}
