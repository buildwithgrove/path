package evm

import (
	"encoding/json"
	"fmt"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

func responseUnmarshallerChainID(data []byte) (response, error) {
	var response responseToChainID
	if err := json.Unmarshal(data, &response); err != nil {
		return responseToChainID{}, err
	}

	return response, nil
}

// responseToChainID captures the fields expected in a
// response to an `eth_chainId` request.
type responseToChainID struct {
	ID      jsonrpc.ID      `json:"id"`
	JSONRPC jsonrpc.Version `json:"jsonrpc"`
	Result  string          `json:"result"`

	// TODO_FUTURE: build the response payload instead of keeping a copy.
	responsePayload []byte
}

func (r responseToChainID) GetObservation() (observation, bool) {
	return observation{
		ChainID: r.Result,
	}, true
}

func (r responseToChainID) GetResponsePayload() []byte {
	// TODO_INCOMPLETE: return a JSONRPC response indicating the error,
	// if the unmarshalling failed.
	return r.responsePayload
}

func (r responseToChainID) Validate(id jsonrpc.ID) error {
	if r.ID != id {
		return fmt.Errorf("validate chainID response: invalid ID; expected %v, got %v", id, r.ID)
	}

	return nil
}
