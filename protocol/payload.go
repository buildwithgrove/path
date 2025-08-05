package protocol

import (
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
)

// Payload represents the HTTP(s) requests proxied between clients and backend services.
// TODO_DOCUMENT(@adshmh): Add more examples (e.g. for RESTful services)
// TODO_IMPROVE(@adshmh): Use an interface here that returns the serialized form of the request.
type Payload struct {
	Data    string
	Method  string
	Path    string
	Headers map[string]string
	// RPCType indicates the type of RPC protocol for routing to appropriate backend ports:
	// - RPCType_REST: Cosmos SDK REST API (typically port 1317)
	// - RPCType_COMET_BFT: CometBFT RPC (typically port 26657)
	// - RPCType_JSON_RPC: EVM JSON-RPC (typically port 8545)
	RPCType sharedtypes.RPCType
}

// EmptyPayload returns an empty payload struct.
// It should only be used when an error is encountered and the actual request cannot be
// proxied, parsed or otherwise processed.
func EmptyErrorPayload() Payload {
	return Payload{
		// This is an intentional placeholder to distinguish errors in retrieving payloads
		// from explicit empty payloads set by PATH.
		Data:    "PATH_EmptyErrorPayload",
		Method:  "",
		Path:    "",
		Headers: map[string]string{},
		RPCType: sharedtypes.RPCType_UNKNOWN_RPC,
	}
}
