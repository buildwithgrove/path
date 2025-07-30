package evm

import (
	"fmt"
	"time"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// ID for the archival check.
// This number may be any arbitrary ID and is selected
// to maintain a convention in the QoS packages of
// consistent ID for a given check type.
//
// EVM checks begin with 1.
const idArchivalCheck = 1003

// methodGetBalance is the JSON-RPC method for getting the balance of an account at a specific block number.
// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_getbalance
const methodGetBalance = jsonrpc.Method("eth_getBalance")

// TODO_IMPROVE(@commoddity): determine an appropriate interval for checking archival status.
const checkArchivalInterval = 20 * time.Minute

var (
	errNoArchivalBalanceObs      = fmt.Errorf("endpoint has not returned an archival balance response to a %q request", methodGetBalance)
	errInvalidArchivalBalanceObs = fmt.Errorf("endpoint has incorrect archival balance")
)

// endpointCheckBlockNumber is a check that ensures the endpoint's block height is greater than the perceived block height.
// It is used to ensure that the endpoint is not behind the chain.
type endpointCheckArchival struct {
	// observedArchivalBalance stores the result of processing the endpoint's response
	// to an `eth_getBalance` request for a specific contract address at a specific block number.
	observedArchivalBalance string
	expiresAt               time.Time
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
		archivalState.archivalCheckConfig.contractAddress,
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

// getArchivalBalance returns the observed archival balance for the endpoint at the archival block height.
// Returns an error if the endpoint hasn't yet returned an archival balance observation.
func (e *endpointCheckArchival) getArchivalBalance() (string, error) {
	if e.observedArchivalBalance == "" {
		return "", errNoArchivalBalanceObs
	}
	return e.observedArchivalBalance, nil
}
