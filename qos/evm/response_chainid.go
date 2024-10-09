package evm

import (
	"encoding/json"
	"fmt"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseUnmarshallerChainID deserializes the provided byte slice
// into a responseToChainID struct, adding any encountered errors
// to the returned struct for constructing a response payload.
func responseUnmarshallerChainID(data []byte) (response, error) {
	var response responseToChainID
	err := json.Unmarshal(data, &response)
	if err != nil {
		response.unmarshallingErr = err
	}

	return response, nil
}

// responseToChainID captures the fields expected in a
// response to an `eth_chainId` request.
type responseToChainID struct {
	ID      jsonrpc.ID      `json:"id"`
	JSONRPC jsonrpc.Version `json:"jsonrpc"`
	Result  string          `json:"result"`

	// unmarshallingErr captures any unmarshalling errors
	// that may have occurred when constructing this instance.
	unmarshallingErr error
}

func (r responseToChainID) GetObservation() (observation, bool) {
	return observation{
		ChainID: r.Result,
	}, true
}

func (r responseToChainID) GetResponsePayload() []byte {
	if r.unmarshallingErr != nil {
		// TODO_UPNEXT(@adshmh): return a JSONRPC response indicating the error,
		// if unmarshalling failed.
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

func (r responseToChainID) Validate(id jsonrpc.ID) error {
	if r.ID != id {
		return fmt.Errorf("validate chainID response: invalid ID; expected %v, got %v", id, r.ID)
	}

	return nil
}
