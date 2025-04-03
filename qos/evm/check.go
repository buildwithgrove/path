package evm

import (
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const (
	// Each endpoint check should use its own ID to avoid potential conflicts.
	// ID of JSON-RPC requests for any new checks should be added to the list below.
	_              = iota
	idChainIDCheck = 1000 + iota
	idBlockNumberCheck
	idArchivalBlockCheck
)

// EndpointStore provides the endpoint check generator required by
// the gateway package to augment endpoints' quality data,
// using synthetic service requests.
var _ gateway.QoSEndpointCheckGenerator = &EndpointStore{}

// TODO_IMPROVE(@commoddity): implement QoS check expiry functionality and use protocol.EndpointAddr
// to filter out checks for any endpoint which has acurrently valid QoS data point.
func (es *EndpointStore) GetRequiredQualityChecks(_ protocol.EndpointAddr) []gateway.RequestQoSContext {
	qualityChecks := []gateway.RequestQoSContext{
		// eg. '{"jsonrpc":"2.0","id":1,"method":"eth_chainId"}'
		getEndpointCheck(es, jsonrpc.NewRequest(idChainIDCheck, methodChainID)),
		// eg. '{"jsonrpc":"2.0","id":1,"method":"eth_blockNumber"}'
		getEndpointCheck(es, jsonrpc.NewRequest(idBlockNumberCheck, methodBlockNumber)),
	}

	// If all of the following are true:
	//  - Service is configured to perform an archival check
	//  - Has calculated an expected archival balance
	// Then:
	// - Add the archival check to the list of qos checks to perform on every hydrator run
	if archivalCheckReq, ok := es.serviceState.archivalState.getArchivalCheckRequest(); ok {
		// '{"jsonrpc":"2.0","id":1,"method":"eth_getBalance","params":["<address>", "<block_number>"]}'
		qualityChecks = append(qualityChecks, getEndpointCheck(es, archivalCheckReq))
	}

	return qualityChecks
}

// getEndpointCheck prepares a request context for a specific endpoint check.
func getEndpointCheck(endpointStore *EndpointStore, jsonrpcReq jsonrpc.Request) *requestContext {
	return &requestContext{
		logger:        endpointStore.logger,
		endpointStore: endpointStore,
		jsonrpcReq:    jsonrpcReq,
	}
}
