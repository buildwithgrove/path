package solana

import (
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const (
	methodGetEpochInfo = jsonrpc.Method("getEpochInfo")
	methodGetBlock     = jsonrpc.Method("getBlock")
	methodGetHealth    = jsonrpc.Method("getHealth")
)
