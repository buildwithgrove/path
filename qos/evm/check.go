package evm

import (
	"encoding/json"

	"github.com/pokt-network/poktroll/pkg/polylog"

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

func (es *EndpointStore) GetRequiredQualityChecks(endpointAddr protocol.EndpointAddr) []gateway.RequestQoSContext {
	// TODO_IMPROVE(@adshmh): skip any checks for which the endpoint already has
	// a valid (i.e. not expired) QoS data point.

	qualityChecks := []gateway.RequestQoSContext{
		getEndpointCheck(es.logger, es, endpointAddr, withChainIDCheck),
		getEndpointCheck(es.logger, es, endpointAddr, withBlockHeightCheck),
	}
	// If the service is configured to perform an archival check and has calculated an expected archival balance,
	// add the archival check to the list of quality checks to perform on every hydrator run.
	if es.serviceState.performArchivalCheck() {
		qualityChecks = append(qualityChecks, getEndpointCheck(es.logger, es, endpointAddr, withArchivalBlockCheck))
	}

	return qualityChecks
}

// getEndpointCheck prepares a request context for a specific endpoint check.
func getEndpointCheck(
	logger polylog.Logger,
	endpointStore *EndpointStore,
	endpointAddr protocol.EndpointAddr,
	options ...func(*requestContext),
) *requestContext {
	requestCtx := requestContext{
		logger:        logger,
		endpointStore: endpointStore,
	}

	for _, option := range options {
		option(&requestCtx)
	}

	return &requestCtx
}

// withChainIDCheck updates the request context to make an EVM JSON-RPC eth_chainId request.
// eg. '{"jsonrpc":"2.0","id":1,"method":"eth_chainId"}'
func withChainIDCheck(requestCtx *requestContext) {
	requestCtx.jsonrpcReq = buildJSONRPCReq(idChainIDCheck, methodChainID)
}

// withBlockHeightCheck updates the request context to make an EVM JSON-RPC eth_blockNumber request.
// eg. '{"jsonrpc":"2.0","id":1,"method":"eth_blockNumber"}'
func withBlockHeightCheck(requestCtx *requestContext) {
	requestCtx.jsonrpcReq = buildJSONRPCReq(idBlockNumberCheck, methodBlockNumber)
}

// withArchivalBlockCheck updates the request context to make an EVM JSON-RPC eth_getBalance request with a random archival block number.
// eg. '{"jsonrpc":"2.0","id":1,"method":"eth_getBalance","params":["0x28C6c06298d514Db089934071355E5743bf21d60", "0xe71e1d"]}'
func withArchivalBlockCheck(requestCtx *requestContext) {
	// Get the archival block number from the endpoint store.
	archivalCheckConfig := requestCtx.endpointStore.serviceState.archivalCheckConfig

	requestCtx.jsonrpcReq = buildJSONRPCReq(
		idArchivalBlockCheck,
		methodGetBalance,
		// Pass params in this order, eg. "params":["0x28C6c06298d514Db089934071355E5743bf21d60", "0xe71e1d"]
		// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_getbalance
		archivalCheckConfig.ContractAddress,
		archivalCheckConfig.archivalBlockNumber,
	)
}

func buildJSONRPCReq(id int, method jsonrpc.Method, params ...any) jsonrpc.Request {
	request := jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(id),
		Method:  method,
	}

	if len(params) > 0 {
		jsonParams, err := json.Marshal(params)
		if err == nil {
			request.Params = jsonrpc.NewParams(jsonParams)
		}
	}

	return request
}
