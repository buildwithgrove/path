package evm

import (
	framework "github.com/buildwithgrove/path/qos/framework/jsonrpc"
)

const (
	// methodBlockNumber is the JSON-RPC method for getting the latest block number.
	// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_blocknumber
	methodBlockNumber = jsonrpc.Method("eth_blockNumber")

	// Endpoint attirbute to track its response to `eth_blockNumber`
	attrETHBlockNumber = methodBlockNumber
)

// responseToBlockNumber provides the functionality required from a response by a requestContext instance.
var _ framework.EndpointResultBuilder = responseBuilderBlockNumber

// TODO_TECHDEBT(@adshmh): validate the `eth_blockNumber` request that was sent to the endpoint.
//
// responseBuilderBlockNumber handles attribute building and sanctioning of endpoints based on the data returned to `eth_blockNumber` requests.
func responseBuilderBlockNumber(ctx *framework.EndpointQueryResultContext) *framework.EndpointQueryResult {
	// ===> single value from result: eg. eth_blockNumber
	result := ctx.BuildIntResult(ethBlockNumber)
	switch {
	case result.IsJSONRPCError():
		return result.SetError("endpoint returned a JSONRPC error response")
	case result.GetValue() <= 0:
		return result.SanctionEndpoint(5*time.Minute, "endpoint returned invalid value as block number")
	default:
		return result
	}

	// Complex values: e.g. Solana getEpochInfo:




	// TODO_MVP(@adshmh): implement the framework's RequestValidator interface to filter out invalid `eth_blockNumber` requests.
	//
	// The endpoint returned an error response: no further processing needed.
	if ctx.IsJSONRPCError() {
		return ctx.Error("endpoint returned a JSONRPC error response.")
	}

	// TODO_MVP(@adshmh): use the contents of the result field to determine the validity of the response.
	// e.g. a response that fails parsing as a number is not valid.
	//
	// The endpoint returned an error: no need to do further processing of the response.
	blockNumber, err := ctx.GetResultAsInt()
	if err != nil {
		return ctx.SanctionEndpoint(5 * time.Minute, fmt.Sprintf("endpoint returned malformed response to eth_blockNumber: %v", err))
	}

	// Sanction the endpoint if it returned an invalid block number.
	if blockNumber <= 0 {
		return ctx.SanctionEndpoint(5 * time.Minute, "endpoint returned invalid value as block number")
	}

	// Store the endpoint's reported block number as its attribute.
	// This attribute will be used in:
	// - state update: to determine the perceived block number on the blockchain.
	// - endpoint selection: to drop out-of-sync endpoints.
	return ctx.Success(ctx.BuildIntAttribute(ethBlockNumber, blockNumber)
}
