package evm

import (
	"fmt"
	"time"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// methodGetBalance is the JSON-RPC method for getting the balance of an account at a specific block number.
// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_getbalance
const methodGetBalance = jsonrpc.Method("eth_getBalance")

// TODO_IN_THIS_PR(@commoddity): determine an appropriate interval for checking archival status.
const checkArchivalInterval = 20 * time.Minute

var (
	errNoArchivalBalanceObs      = fmt.Errorf("endpoint has not returned an archival balance response to a %q request", methodGetBalance)
	errInvalidArchivalBalanceObs = "endpoint has archival balance %s, expected archival balance %s"
)

// endpointCheckBlockNumber is a check that ensures the endpoint's block height is greater than the perceived block height.
// It is used to ensure that the endpoint is not behind the chain.
type endpointCheckArchival struct {
	// observedArchivalBalance stores the result of processing the endpoint's response
	// to an `eth_getBalance` request for a specific contract address at a specific block number.
	observedArchivalBalance string
	expiresAt               time.Time
}

// isValid returns an error if the endpoint's block height is less than the perceived block height minus the sync allowance.
func (e *endpointCheckArchival) isValid(archivalState archivalState) error {
	if !archivalState.archivalCheckConfig.Enabled {
		return nil
	}

	if e.observedArchivalBalance == "" {
		return errNoArchivalBalanceObs
	}
	if e.observedArchivalBalance != archivalState.expectedBalance {
		return fmt.Errorf(errInvalidArchivalBalanceObs, e.observedArchivalBalance, archivalState.expectedBalance)
	}

	return nil
}

// shouldRun returns true if the check is not yet initialized or has expired.
func (e *endpointCheckArchival) shouldRun(archivalState archivalState) bool {
	// Do not perform an archival check if:
	// 	- The archival check is not enabled for the service.
	// 	- The archival block number has not yet been set in the archival state.
	if !archivalState.archivalCheckConfig.Enabled || archivalState.blockNumberHex == "" {
		return false
	}

	return e.expiresAt.IsZero() || e.expiresAt.Before(time.Now())
}

// getRequest returns a JSONRPC request to check the balance of:
//   - the contract specified in `a.archivalCheckConfig.ContractAddress`
//   - at the block number specified in `a.blockNumberHex`
//
// For example:
// '{"jsonrpc":"2.0","id":1,"method":"eth_getBalance","params":["0x28C6c06298d514Db089934071355E5743bf21d60", "0xe71e1d"]}'
func (e *endpointCheckArchival) getRequest(archivalState archivalState) jsonrpc.Request {
	// Pass params in this order: [<contract_address>, <block_number>]
	// eg. "params":["0x28C6c06298d514Db089934071355E5743bf21d60", "0xe71e1d"]
	// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_getbalance
	params, err := jsonrpc.BuildParamsFromStringArray([2]string{
		archivalState.archivalCheckConfig.ContractAddress,
		archivalState.blockNumberHex,
	})
	if err != nil {
		archivalState.logger.Error().Msgf("failed to build archival check request params: %v", err)
		return jsonrpc.Request{}
	}

	return jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(idArchivalCheck),
		Method:  jsonrpc.Method(methodGetBalance),
		Params:  params,
	}
}
