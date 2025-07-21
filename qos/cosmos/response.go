package cosmos

import (
	"encoding/json"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

var (
	// CosmosSDK response IDs for different request types:
	// - REST-like responses: -1
	// TODO_NEXT(@adshmh): Use proper JSON-RPC ID response validation that works for all CosmosSDK chains.
	restLikeResponseID = jsonrpc.IDFromInt(-1)
)
