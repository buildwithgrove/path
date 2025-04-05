package evm

import (
	"encoding/json"

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
var _ gateway.QoSEndpointCheckGenerator = &endpointStore{}

// TODO_IMPROVE(@commoddity): implement QoS check expiry functionality and use protocol.EndpointAddr
// to filter out checks for any endpoint which has acurrently valid QoS data point.
func (es *endpointStore) GetRequiredQualityChecks(_ protocol.EndpointAddr) []gateway.RequestQoSContext {
	qualityChecks := []gateway.RequestQoSContext{
		getEndpointCheck(es, getChainIDCheckRequest()),
		getEndpointCheck(es, getBlockNumberCheckRequest()),
	}

	// If the service is expected to be archival, perform an archival check.
	if archivalCheckReq, ok := getArchivalCheckRequest(es.serviceState.archivalState); ok {
		qualityChecks = append(qualityChecks, getEndpointCheck(es, archivalCheckReq))
	}

	return qualityChecks
}

// getEndpointCheck prepares a request context for a specific endpoint check.
func getEndpointCheck(endpointStore *endpointStore, jsonrpcReq jsonrpc.Request) *requestContext {
	return &requestContext{
		logger:        endpointStore.logger,
		endpointStore: endpointStore,
		jsonrpcReq:    jsonrpcReq,
	}
}

// getChainIDCheckRequest returns a JSONRPC request to check the chain ID.
// eg. '{"jsonrpc":"2.0","id":1,"method":"eth_chainId"}'
func getChainIDCheckRequest() jsonrpc.Request {
	return jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(idChainIDCheck),
		Method:  jsonrpc.Method(methodChainID),
	}

	if len(params) > 0 {
		jsonParams, err := json.Marshal(params)
		if err == nil {
			request.Params = jsonrpc.NewParams(jsonParams)
		}
	}

	return request
}

// getBlockNumberCheckRequest returns a JSONRPC request to check the block number.
// eg. '{"jsonrpc":"2.0","id":1,"method":"eth_blockNumber"}'
func getBlockNumberCheckRequest() jsonrpc.Request {
	return jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(idBlockNumberCheck),
		Method:  jsonrpc.Method(methodBlockNumber),
	}
}

// getArchivalCheckRequest returns a JSONRPC request to check the balance of:
//   - the contract specified in `a.archivalCheckConfig.ContractAddress`
//   - at the block number specified in `a.blockNumberHex`
//
// eg.
// '{"jsonrpc":"2.0","id":1,"method":"eth_getBalance","params":["0x28C6c06298d514Db089934071355E5743bf21d60", "0xe71e1d"]}'
func getArchivalCheckRequest(archivalState archivalState) (jsonrpc.Request, bool) {
	// Do not perform an archival check if:
	// 	- The archival block number has not yet been set in the archival state.
	// 	- The archival check is not enabled for the service.
	if archivalState.blockNumberHex == "" || archivalState.archivalCheckConfig.IsEmpty() {
		return jsonrpc.Request{}, false
	}

	// Pass params in this order: [<contract_address>, <block_number>]
	// eg. "params":["0x28C6c06298d514Db089934071355E5743bf21d60", "0xe71e1d"]
	// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_getbalance
	params, err := jsonrpc.BuildParamsFromStringArray([2]string{
		archivalState.archivalCheckConfig.contractAddress,
		archivalState.blockNumberHex,
	})
	if err != nil {
		archivalState.logger.Error().Msgf("failed to build archival check request params: %v", err)
		return jsonrpc.Request{}, false
	}

	return jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(idArchivalBlockCheck),
		Method:  jsonrpc.Method(methodGetBalance),
		Params:  params,
	}, true
}
