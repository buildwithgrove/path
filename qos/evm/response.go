package evm

import (
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseUnmarshaller is the entrypoint function for any
// new supported response types.
// E.g. to handle "eth_getBalance" requests, the following need to be fined:
//  1. A new custom responseUnmarshaller
//  2. A new custom struct  to handle the details of the particular response.
type responseUnmarshaller func([]byte) (response, error)

var (
	// All response types needs to implement the response interface.
	// Any new response struct needs to be added to the following list.
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
