package evm

import (
	"fmt"
	"time"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// methodBlockNumber is the JSON-RPC method for getting the latest block number.
// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_blocknumber
const methodBlockNumber = jsonrpc.Method("eth_blockNumber")

// TODO_IMPROVE(@commoddity): determine an appropriate interval for checking the block height.
const checkBlockNumberInterval = 60 * time.Second

var (
	errNoBlockNumberObs      = fmt.Errorf("endpoint has not had an observation of its response to a %q request", methodBlockNumber)
	errInvalidBlockNumberObs = fmt.Errorf("endpoint returned an invalid response to a %q request", methodBlockNumber)
)

// endpointCheckBlockNumber is a check that ensures the endpoint's block height is greater than the perceived block height.
// It is used to ensure that the endpoint is not behind the chain.
type endpointCheckBlockNumber struct {
	// parsedBlockNumberResponse stores the result of processing the endpoint's response to an `eth_blockNumber` request.
	// It is nil if there has NOT been an observation of the endpoint's response to an `eth_blockNumber` request.
	parsedBlockNumberResponse *uint64
	expiresAt                 time.Time
}

// isValid returns an error if the endpoint's block height is less than the perceived block height minus the sync allowance.
func (e *endpointCheckBlockNumber) isValid(perceivedBlockNumber uint64, syncAllowance uint64) error {
	if e.parsedBlockNumberResponse == nil {
		return errNoBlockNumberObs
	}

	// If the endpoint's block height is less than the perceived block height minus the sync allowance,
	// then the endpoint is behind the chain and should be filtered out.
	minAllowedBlockNumber := perceivedBlockNumber - syncAllowance

	if *e.parsedBlockNumberResponse < minAllowedBlockNumber {
		return errInvalidBlockNumberObs
	}

	return nil
}

// shouldRun returns true if the check is not yet initialized or has expired.
func (e *endpointCheckBlockNumber) shouldRun() bool {
	return e.expiresAt.IsZero() || e.expiresAt.Before(time.Now())
}

// getRequest returns a JSONRPC request to check the block number.
// eg. '{"jsonrpc":"2.0","id":1,"method":"eth_blockNumber"}'
func (e *endpointCheckBlockNumber) getRequest() jsonrpc.Request {
	return jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(idBlockNumberCheck),
		Method:  jsonrpc.Method(methodBlockNumber),
	}
}
