package evm

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

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

	return []gateway.RequestQoSContext{
		getEndpointCheck(es.logger, es, endpointAddr, withChainIDCheck),
		getEndpointCheck(es.logger, es, endpointAddr, withBlockHeightCheck),
		// TODO_IN_THIS_PR@(commoddity): make adding this check configurable.
		getEndpointCheck(es.logger, es, endpointAddr, withArchivalBlockCheck),
	}
}

// getEndpointCheck prepares a request context for a specific endpoint check.
func getEndpointCheck(
	logger polylog.Logger,
	endpointStore *EndpointStore,
	endpointAddr protocol.EndpointAddr,
	options ...func(*requestContext),
) *requestContext {
	requestCtx := requestContext{
		logger:                  logger,
		endpointStore:           endpointStore,
		preSelectedEndpointAddr: endpointAddr,
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

// withArchivalBlockCheck updates the request context to make an EVM JSON-RPC eth_getBlockByNumber request with a random archival block number.
// eg. '{"jsonrpc":"2.0","id":1,"method":"eth_getBlockByNumber","params":["0x1b4", false]}'
func withArchivalBlockCheck(requestCtx *requestContext) {
	// Get the current perceived block number.
	perceivedBlockNumber := requestCtx.endpointStore.getPerceivedBlockNumber()

	// Get a random archival block number.
	archivalBlockNumber := getArchivalBlockNumber(perceivedBlockNumber)

	requestCtx.jsonrpcReq = buildJSONRPCReq(
		idArchivalBlockCheck,
		methodGetBlockByNumber,
		// Pass params in this order, eg. "params":["0x1b4", false]
		// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_getblockbynumber
		archivalBlockNumber,
		false, // Return only hashes of the transactions in the block
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

// getArchivalBlockNumber returns a random archival block number as a hex string.
// The block number is a fraction of the current block height with given threshold.
//
// eg. "0x1b4" (436 in decimal)
func getArchivalBlockNumber(currentBlockHeight uint64) string {
	// If the current block height is not yet determined, use the earliest block number.
	// This is to avoid sending a request with an invalid block number until the
	// service state has calculated the current block height from other checks.
	if currentBlockHeight == 0 {
		return "0x0" // 0x0 is equivalent to "earliest"
	}

	const (
		minBlockNumber uint64  = 0
		maxFraction    float64 = 0.5
	)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomFraction := r.Float64() * maxFraction
	archivalBlockNumber := uint64(float64(currentBlockHeight) * randomFraction)

	archivalBlockNumber = uint64(math.Max(float64(archivalBlockNumber), float64(minBlockNumber)))

	return fmt.Sprintf("0x%x", archivalBlockNumber)
}
