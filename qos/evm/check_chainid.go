package evm

import (
	"fmt"
	"time"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

var _ check = &endpointCheckChainID{}

const (
	checkNameChainID endpointCheckName = "chain_id"
	// methodChainID is the JSON-RPC method for getting the chain ID.
	// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_chainid
	methodChainID = jsonrpc.Method("eth_chainId")
	// TODO_IMPROVE: determine an appropriate interval for checking the chain ID.
	checkChainIDInterval = 60 * time.Minute
)

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
func (e *endpointCheckChainID) isValid(serviceState *ServiceState) error {
	if e.chainID == nil {
		return errNoChainIDObs
	}
	if *e.chainID != serviceState.config.chainID {
		return errInvalidChainIDObs
	}
	return nil
}

func (e *endpointCheckChainID) name() endpointCheckName {
	return checkNameChainID
}

// shouldRun returns true if the check is not yet initialized or has expired.
func (e *endpointCheckChainID) shouldRun() bool {
	return e.expiresAt.IsZero() || e.expiresAt.Before(time.Now())
}

// withChainIDCheck updates the request context to make an EVM JSON-RPC eth_chainId request.
func withChainIDCheck(requestCtx *requestContext) {
	requestCtx.jsonrpcReq = buildJSONRPCReq(idChainIDCheck, methodChainID)
}
