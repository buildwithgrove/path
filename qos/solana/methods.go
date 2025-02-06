package solana

import (
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const (
	// methodGetEpochInfo is the JSON-RPC method for getting the epoch information.
	// Reference: https://docs.solana.com/developing/clients/jsonrpc-api#getepochinfo
	methodGetEpochInfo = jsonrpc.Method("getEpochInfo")

	// methodGetBlock is the JSON-RPC method for getting a block.
	// Reference: https://docs.solana.com/developing/clients/jsonrpc-api#getblock
	methodGetBlock = jsonrpc.Method("getBlock") // nolint:unused TODO_TECHDEBT(@adshmh): remove this once the `getBlock` method is used.

	// methodGetHealth is the JSON-RPC method for checking the health of the node.
	// Reference: https://docs.solana.com/developing/clients/jsonrpc-api#gethealth
	methodGetHealth = jsonrpc.Method("getHealth")
)
