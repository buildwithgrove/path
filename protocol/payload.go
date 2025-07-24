package protocol

import (
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
)

// TODO_TECHDEBT(@adshmh): use an interface here that returns the serialized form of the request, with the following requirements:
// 1. Payload should return the serialized form of the request to be delivered to the backend service,
// i.e. the onchain service to which the protocol endpoint proxies relay requests.
// 2. Use an enum to represent the underlying spec/standard, e.g. REST/JSONRPC/gRPC/etc.
//
// Payload currently supports HTTP(s) requests to blockchain services
// TODO_DOCUMENT(@adshmh): add more examples, e.g. for RESTful services, as support for more types of services
// is added.
type Payload struct {
	Data            string
	Method          string
	Path            string
	Headers         map[string]string
	TimeoutMillisec int
	// RPCType indicates the type of RPC protocol for routing to appropriate backend ports:
	// - RPCType_REST: Cosmos SDK REST API (typically port 1317)
	// - RPCType_COMET_BFT: CometBFT RPC (typically port 26657)
	// - RPCType_JSON_RPC: EVM JSON-RPC (typically port 8545)
	RPCType sharedtypes.RPCType
}