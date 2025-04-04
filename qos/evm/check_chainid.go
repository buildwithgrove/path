package evm

import (
	"fmt"
	"time"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// methodChainID is the JSON-RPC method for getting the chain ID.
// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_chainid
const methodChainID = jsonrpc.Method("eth_chainId")

// TODO_IMPROVE(@commoddity): determine an appropriate interval for checking the chain ID.
const checkChainIDInterval = 60 * time.Minute

var (
	errNoChainIDObs      = fmt.Errorf("endpoint has not had an observation of its response to a %q request", methodChainID)
	errInvalidChainIDObs = fmt.Errorf("endpoint returned an invalid response to a %q request", methodChainID)
)

// endpointCheckChainID is a check that ensures the endpoint's chain ID is the same as the expected chain ID.
// It is used to ensure that the endpoint is on the correct chain.
type endpointCheckChainID struct {
	// chainIDResponse stores the result of processing the endpoint's response to an `eth_chainId` request.
	// It is nil if there has NOT been an observation of the endpoint's response to an `eth_chainId` request.
	chainID   *string
	expiresAt time.Time
}

// isValid returns an error if the endpoint's chain ID does not match the expected chain ID in the service state.
func (e *endpointCheckChainID) isValid(evmChainID string) error {
	if e.chainID == nil {
		return errNoChainIDObs
	}
	if *e.chainID != evmChainID {
		return errInvalidChainIDObs
	}
	return nil
}

// shouldRun returns true if the check is not yet initialized or has expired.
func (e *endpointCheckChainID) shouldRun() bool {
	return e.expiresAt.IsZero() || e.expiresAt.Before(time.Now())
}

// getRequest returns a JSONRPC request to check the chain ID.
// eg. '{"jsonrpc":"2.0","id":1,"method":"eth_chainId"}'
func (e *endpointCheckChainID) getRequest() jsonrpc.Request {
	return jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(idChainIDCheck),
		Method:  jsonrpc.Method(methodChainID),
	}
}
