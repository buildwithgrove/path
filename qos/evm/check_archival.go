package evm

import (
	"fmt"
	"time"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

var _ evmQualityCheck = &endpointCheckArchival{}

// methodGetBalance is the JSON-RPC method for getting the balance of an account at a specific block number.
// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_getbalance
const methodGetBalance = jsonrpc.Method("eth_getBalance")

// TODO_IMPROVE(@commoddity): determine an appropriate interval for checking archival status.
const checkArchivalInterval = 20 * time.Minute

var (
	errNoArchivalBalanceObs = fmt.Errorf("endpoint has not returned an archival balance response to a %q request", methodGetBalance)
	errInvalidArchivalObs   = fmt.Errorf("endpoint returned an invalid response to a %q request", methodGetBalance)
)

// endpointCheckBlockNumber is a check that ensures the endpoint's block height is greater than the perceived block height.
// It is used to ensure that the endpoint is not behind the chain.
type endpointCheckArchival struct {
	// archivalBalance stores the result of processing the endpoint's response to an `eth_getBalance` request.
	// It is empty if there has NOT been an observation of the endpoint's response to an `eth_getBalance` request.
	archivalBalance string
	expiresAt       time.Time
}

// isValid returns an error if the endpoint's block height is less than the perceived block height minus the sync allowance.
func (e *endpointCheckArchival) isValid(serviceState *ServiceState) error {
	if e.archivalBalance == "" {
		return errNoArchivalBalanceObs
	}
	if e.archivalBalance != serviceState.archivalState.balance {
		return fmt.Errorf(errInvalidArchivalObs.Error(), e.archivalBalance, serviceState.archivalState.balance)
	}
	return nil
}

// shouldRun returns true if the check is not yet initialized or has expired.
func (e *endpointCheckArchival) shouldRun() bool {
	return e.expiresAt.IsZero() || e.expiresAt.Before(time.Now())
}

// setRequestContext updates the request context to make an EVM JSON-RPC eth_getBalance request with a random archival block number.
// eg. '{"jsonrpc":"2.0","id":1,"method":"eth_getBalance","params":["0x28C6c06298d514Db089934071355E5743bf21d60", "0xe71e1d"]}'
func (e *endpointCheckArchival) setRequestContext(requestCtx *requestContext) {
	// Get the archival block number from the endpoint store.
	archivalCheckConfig := requestCtx.endpointStore.serviceState.serviceConfig.getArchivalCheckConfig()
	// Get the current state of the archival check.
	serviceArchivalState := requestCtx.endpointStore.serviceState.archivalState

	requestCtx.jsonrpcReq = buildJSONRPCReq(
		idArchivalCheck,
		methodGetBalance,
		// Pass params in this order, eg. "params":["0x28C6c06298d514Db089934071355E5743bf21d60", "0xe71e1d"]
		// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_getbalance
		archivalCheckConfig.ContractAddress,
		serviceArchivalState.blockNumberHex,
	)

	// Set the archival balance check flag to true.
	// This is used to ensure that only hydrator requests for the archival block number are used
	// to update QoS data on whether endpoints are able to service archival requests.
	requestCtx.archivalBalanceCheck = true
}
