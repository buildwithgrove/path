//go:build e2e

package e2e

import (
	"net/http"

	vegeta "github.com/tsenart/vegeta/lib"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

/* -------------------- CometBFT REST Endpoint Definitions -------------------- */

var cometbftExpectedID = jsonrpc.IDFromInt(-1)

// Reference for all CometBFT RPC endpoints:
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

// getCometBFTTestEndpoints returns all CometBFT endpoints for a service load test.
func getCometBFTTestEndpoints() []string {
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
	endpoints []string,
	gatewayURL string,
) ([]vegeta.Target, error) {
	headers := getRequestHeaders(ts.ServiceID)

	targets := make([]vegeta.Target, 0, len(endpoints))
	for _, endpoint := range endpoints {
		// Create URL with endpoint
		url := gatewayURL + string(endpoint)

		// Create vegeta target
		target := vegeta.Target{
			Method: http.MethodGet,
			URL:    url,
			Header: headers,
		}

		targets = append(targets, target)
	}

	return targets, nil
}
