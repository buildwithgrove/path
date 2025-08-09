//go:build e2e

package e2e

import (
	"net/http"

	vegeta "github.com/tsenart/vegeta/lib"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_NEXT(@commoddity): add test cases for XRPL-EVM Testnet and Pocket Shannon once
// suppliers for both are updated to latest Relay Miner version that supports the new RPC types.
// TODO_NEXT(@commoddity): add tests for Cosmos SDK & JSON-RPC types once suppliers are updated.

/* -------------------- CometBFT REST Endpoint Definitions -------------------- */

var cometbftExpectedID = jsonrpc.IDFromInt(-1)

// Reference for all CometBFT RPC URL paths:
// - https://docs.cometbft.com/v0.38/rpc/
const (
	cometbftEndpointStatus         = "/status"          // Get node status
	cometbftEndpointHealth         = "/health"          // Get node health
	cometbftEndpointNetInfo        = "/net_info"        // Get network info
	cometbftEndpointConsensusState = "/consensus_state" // Get consensus state
	cometbftEndpointCommit         = "/commit"          // Get commit
	cometbftEndpointABCIInfo       = "/abci_info"       // Get ABCI info
	cometbftEndpointBlock          = "/block"           // Get block at height
	cometbftEndpointBlockResults   = "/block_results"   // Get block results
	cometbftEndpointValidators     = "/validators"      // Get validators
)

// getCometBFTTestURLPaths returns all CometBFT URL paths for a service load test.
func getCometBFTTestURLPaths() []string {
	return []string{
		cometbftEndpointStatus,
		cometbftEndpointHealth,
		cometbftEndpointNetInfo,
		cometbftEndpointConsensusState,
		cometbftEndpointCommit,
		cometbftEndpointABCIInfo,
		cometbftEndpointBlock,
		cometbftEndpointBlockResults,
		cometbftEndpointValidators,
	}
}

func getCometBFTVegetaTargets(
	ts *TestService,
	urlPaths []string,
	gatewayURL string,
) (map[string]vegeta.Target, error) {
	headers := getRequestHeaders(ts.ServiceID)

	targets := make(map[string]vegeta.Target)
	for _, urlPath := range urlPaths {
		// Create URL with URL path
		url := gatewayURL + string(urlPath)

		// Create vegeta target
		target := vegeta.Target{
			Method: http.MethodGet,
			URL:    url,
			Header: headers,
		}

		targets[urlPath] = target
	}

	return targets, nil
}
