package cometbft

const (
	// Get node health. Returns empty result (200 OK) on success, no response - in case of an error.
	// Reference: https://docs.cometbft.com/v0.38/rpc/#/Info/health
	apiPathHealthCheck = "/health"
	// Get CometBFT status including node info, pubkey, latest block hash, app hash, block height and time.
	// Reference: https://docs.cometbft.com/v0.38/rpc/#/Info/status
	apiPathBlockHeight = "/status"
)
