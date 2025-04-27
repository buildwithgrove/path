package evm

import (
	framework "github.com/buildwithgrove/path/qos/framework/jsonrpc"
)

const (
	// methodChainID is the JSON-RPC method for getting the chain ID.
	// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_chainid
	methodChainID = jsonrpc.Method("eth_chainId")

	// Endpoint attirbute to track its response to `eth_chainId`
	attrETHChainID = methodChainID
)

// endpointResultBuilderChainID provides the functionality required from an endpoint result builder.
var _ framework.EndpointResultBuilder = endpointAttributeBuilderChainID

// TODO_MVP(@adshmh): handle the following scenarios:
//  1. An endpoint returned a malformed, i.e. Not in JSONRPC format, response.
//     The user-facing response should include the request's ID.
//  2. An endpoint returns a JSONRPC response indicating a user error:
//     This should be returned to the user as-is.
//  3. An endpoint returns a valid JSONRPC response to a valid user request:
//     This should be returned to the user as-is.
//
// TODO_TECHDEBT(@adshmh): validate the `eth_chainId` request that was sent to the endpoint.
//
// endpointResultBuilderChainID handles attribute building and sanctioning of endpoints based on the data returned to `eth_chainId` requests.
func endpointResultBuilderChainID(
	ctx *framework.EndpointQueryResultContext,
	config EVMQoSServiceConfig,
) *framework.ResultData {
	// TODO_MVP(@adshmh): Sanction endpoints that fail to respond to `eth_chainId` requests:
	//   1. Implement the framework's RequestValidator interface to filter out invalid `eth_chainId` requests.
	//   2. Sanction an endpoint that returns an error, as the requests are guaranteed to be valid.
	//
	// The endpoint returned an error response: no further processing needed.
	if ctx.IsJSONRPCError() {
		return ctx.Error("endpoint returned a JSONRPC error response to an eth_chainId request.")
	}

	// TODO_MVP(@adshmh): use the contents of the result field to determine the validity of the response.
	// e.g. a response that fails parsing as a number is not valid.
	//
	// The endpoint returned an error: no need to do further processing of the response.
	chainID, err := ctx.GetResultAsString()
	if err != nil {
		return ctx.SanctionEndpoint(5 * time.Minute, "endpoint returned malformed response to eth_chainID")
	}

	// Sanction the endpoint if it returned an invalid chain ID
	if chainID != config.GetChainID() {
		return ctx.SanctionEndpoint(5 * time.Minute, "endpoint returned invalid value as chain ID")
	}

	// Store the endpoint's reported chainID as its attribute.
	// This attribute will be used in:
	// - endpoint selection: to drop misconfigured endpoints.
	return ctx.Success(ctx.BuildStringAttribute(ethChainID, chainID))
}
