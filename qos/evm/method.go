package evm

import (
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const (
	methodChainID     = jsonrpc.Method("eth_chainId")
	methodBlockNumber = jsonrpc.Method("eth_blockNumber")
)
