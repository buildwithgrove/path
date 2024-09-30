package evm

import (
	"encoding/json"
	"fmt"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseUnmarshaller is the entrypoint function for any
// new supported response types, e.g. to add a response to an
// eth_getBalance request, a new responseUnmarshaller needs to
// be defined, along with a custom struct to handle the details
// of the particular response.
type responseUnmarshaller func([]byte) (response, error)

var (
	_ response = &responseToChainID{}
	_ response = &responseToBlockNumber{}

	methodResponseMappings = map[jsonrpc.Method]responseUnmarshaller{
		methodChainID:     responseUnmarshallerChainID,
		methodBlockNumber: responseUnmarshallerBlockNumber,
	}
)

func unmarshalResponse(method jsonrpc.Method, data []byte) (response, error) {
	unmarshaller, found := methodResponseMappings[method]
	if found {
		return unmarshaller(data)
	}

	return responseUnmarshallerGeneric(data)
}

func responseUnmarshallerChainID(data []byte) (response, error) {
	var response responseToChainID
	if err := json.Unmarshal(data, &response); err != nil {
		return responseToChainID{}, err
	}

	return response, nil
}

// TODO_IN_THIS_COMMIT: implement this unmarshaller
func responseUnmarshallerBlockNumber(data []byte) (response, error) {
	return responseToBlockNumber{}, nil
}

// TODO_IN_THIS_COMMIT: implement this unmarshaller
func responseUnmarshallerGeneric(data []byte) (response, error) {
	return responseGeneric{}, nil
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
		return fmt.Errorf("validate chainID response: invalid ID; expected %v, got %v", id, r.ID)
	}

	return nil
}

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

// TODO_UPNEXT(@adshmh): implement the generic jsonrpc response
// (with the scope limited to an EVM-based blockchain)
type responseGeneric struct {
	ID      jsonrpc.ID      `json:"id"`
	JSONRPC jsonrpc.Version `json:"jsonrpc"`
	json.RawMessage

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
