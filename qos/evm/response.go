package evm

import (
	"encoding/json"
	"fmt"
)

// responseUnmarshaller is the entrypoint function for any
// new supported response types.
// E.g. to handle "eth_getBalance" requests, the following need to be fined:
// 	1. A new custom responseUnmarshaller
//      2. A new custom struct  to handle the details of the particular response.
type responseUnmarshaller func([]byte) (response, error)

var (
	_ response = &responseToChainID{}
	_ response = &responseToBlockHeight{}

	methodResponseMappings = map[method]responseUnmarshaller{
		methodChainID:     responseUnmarshallerChainID,
		methodBlockNumber: responseUnmarshallerBlockHeight,
	}
)

func unmarshalResponse(method method, data []byte) (response, error) {
	unmarshaller, found := methodResponseMappings[method]
	if found {
		return unmarshaller(data)
	}

	return genericUnmarshaller(data)
}

func responseUnmarshallerChainID(data []byte) (responseToChainID, error) {
	var response responseToChainID
	if err := json.Unmarshal(data, &response); err != nil {
		return responseToChainID{}, err
	}

	return response, nil
}

// TODO_IN_THIS_COMMIT: implement this unmarshaller
func responseUnmarshallerBlockHeight(data []byte) (responseToBlockHeight, error) {

}

// TODO_IN_THIS_COMMIT: implement this unmarshaller
func responseUnmarshallerGeneric(data []byte) (responseGeneric, error) {

}

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
		return fmt.Errorf("validate chainID response: invalid ID; expected %s, got %s", id, r.ID)
	}

	return nil
}

type responseToBlockHeight struct {
	ID      jsonrpc.ID      `json:"id"`
	JSONRPC jsonrpc.Version `json:"jsonrpc"`
	Result  string          `json:"result"`

	// TODO_FUTURE: build the response payload instead of keeping a copy.
	responsePayload []byte
}

func (r responseToBlockHeight) GetObservation() (observation, bool) {
	return observation{
		BlockHeight: r.Result,
	}, true
}

func (r responseToBlockHeight) GetResponsePayload() []byte {
	// TODO_INCOMPLETE: return a JSONRPC response indicating the error,
	// if the unmarshalling failed.
	return r.responsePayload
}

type responseGeneric struct {
	ID      jsonrpc.ID      `json:"id"`
	JSONRPC jsonrpc.Version `json:"jsonrpc"`
	json.RawMessage

	payload []byte
}

func (r responseGeneric) GetObservation() (observation, bool) {
	return observation{}, false
}

func (r responseGeneric) GetResponsePayload() []byte {
	return r.buildPayload()
}

func (r responseGeneric) buildPayload() []byte {

}
