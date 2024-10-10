package evm

import (
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const (
	// TODO_UPNEXT(@adshmh): add more examples of methods here.
	methodChainID     = jsonrpc.Method("eth_chainId")
	methodBlockNumber = jsonrpc.Method("eth_blockNumber")
)
