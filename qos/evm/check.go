package evm

import (
	"crypto/rand"
	"fmt"
	"math/big"

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
	idGetBlockByNumberCheck
)

const MethodGetBlockByNumber = "eth_getBlockByNumber"

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
		getEndpointCheck(es.logger, es, endpointAddr, withBlockHeightCheck, withArchivalCheck),
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
		isValid:                 true,
		preSelectedEndpointAddr: endpointAddr,
	}

	for _, option := range options {
		option(&requestCtx)
	}

	return &requestCtx
}

// withChainIDCheck updates the request context to make an EVM JSON-RPC eth_chainId request.
func withChainIDCheck(requestCtx *requestContext) {
	requestCtx.jsonrpcReq = buildJSONRPCReq(idChainIDCheck, methodChainID)
}

// withBlockHeightCheck updates the request context to make an EVM JSON-RPC eth_blockNumber request.
func withBlockHeightCheck(requestCtx *requestContext) {
	requestCtx.jsonrpcReq = buildJSONRPCReq(idBlockNumberCheck, methodBlockNumber)
}

// withArchivalCheck updates the request context to make an EVM JSON-RPC eth_getBlockByNumber request
// for a random historical block.
func withArchivalCheck(requestCtx *requestContext) {
	requestCtx.nextStepHandler = handleArchivalCheckResponse
}

// handleArchivalCheckResponse processes the block number response and sets up
// the eth_getBlockByNumber request for a random historical block
func handleArchivalCheckResponse(requestCtx *requestContext, responseBytes []byte) error {
	// Parse the block number response
	blockNumberResp, err := jsonrpc.UnmarshalResponse(responseBytes)
	if err != nil {
		requestCtx.logger.Error().Err(err).Msg("Failed to unmarshal block number response for archival check")
		return err
	}

	// Extract the result containing the block number
	blockNumberBz, err := blockNumberResp.GetResultAsBytes()
	if err != nil {
		requestCtx.logger.Error().Err(err).Msg("Failed to get block number result for archival check")
		return err
	}

	// Remove quotes from the result if present (JSON string)
	var blockNumberHex string
	err = json.Unmarshal(blockNumberBz, &blockNumberHex)
	if err != nil {
		requestCtx.logger.Error().Err(err).Msg("Failed to unmarshal block number hex for archival check")
		return err
	}

	// Convert the hex string to a big.Int
	if len(blockNumberHex) < 3 || blockNumberHex[:2] != "0x" {
		return fmt.Errorf("invalid block number format: %s", blockNumberHex)
	}

	currentBlock, ok := new(big.Int).SetString(blockNumberHex[2:], 16)
	if !ok {
		return fmt.Errorf("failed to parse block number: %s", blockNumberHex)
	}

	// If the chain is too new (less than 10 blocks), we can't perform a meaningful archival check
	if currentBlock.Cmp(big.NewInt(10)) <= 0 {
		requestCtx.logger.Info().Msg("Chain has too few blocks for archival check, skipping")
		// Create a successful "dummy" response to avoid failing the QoS check
		dummyResp := jsonrpc.Response{
			JSONRPC: jsonrpc.Version2,
			ID:      jsonrpc.IDFromInt(idGetBlockByNumberCheck),
			Result:  json.RawMessage(`{"number":"0x0","hash":"0x0","parentHash":"0x0","transactions":[]}`),
		}
		dummyRespBytes, _ := json.Marshal(dummyResp)
		requestCtx.responseBytes = dummyRespBytes
		return nil
	}

	// Calculate a random block between currentBlock/10 and currentBlock/2
	minBlock := new(big.Int).Div(currentBlock, big.NewInt(10))
	maxBlock := new(big.Int).Div(currentBlock, big.NewInt(2))
	
	// Calculate the range size
	rangeSize := new(big.Int).Sub(maxBlock, minBlock)
	rangeSize.Add(rangeSize, big.NewInt(1)) // Make it inclusive
	
	// Generate random number within range and add to minBlock
	randomOffset, err := rand.Int(rand.Reader, rangeSize)
	if err != nil {
		return fmt.Errorf("failed to generate random block number: %w", err)
	}
	
	targetBlock := new(big.Int).Add(minBlock, randomOffset)
	
	// Convert to hex with 0x prefix for the JSON-RPC call
	targetBlockHex := "0x" + targetBlock.Text(16)
	
	requestCtx.logger.Info().
		Str("current_block", currentBlock.String()).
		Str("target_block", targetBlock.String()).
		Msg("Performing archival check")

	// Create params for eth_getBlockByNumber: [blockNumberHex, fullTxDetails]
	// We set fullTxDetails to false to minimize response size
	params := []interface{}{targetBlockHex, false}
	
	// Create the next JSON-RPC request to get the historical block
	requestCtx.jsonrpcReq = jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(idGetBlockByNumberCheck),
		Method:  MethodGetBlockByNumber,
		Params:  params,
	}
	
	// Clear the response bytes from the block number request
	requestCtx.responseBytes = nil
	
	return nil
}

func buildJSONRPCReq(id int, method jsonrpc.Method) jsonrpc.Request {
	return jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(id),
		Method:  method,
	}
}
