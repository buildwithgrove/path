package evm

import (
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const (
	// methodChainID is the JSON-RPC method for getting the chain ID.
	// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_chainid
	methodChainID = jsonrpc.Method("eth_chainId")

	// methodBlockNumber is the JSON-RPC method for getting the latest block number.
	// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_blocknumber
	methodBlockNumber = jsonrpc.Method("eth_blockNumber")

	// TODO_MVP(@adshmh): add more examples of methods here.
)
