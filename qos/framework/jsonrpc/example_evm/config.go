package evm

// Config captures the modifiable settings of the EVM QoS service.
// This will enable the QoS service to be used as part of EVM-based blockchains which may have different desired QoS properties.
// e.g. different blockchains QoS instances could have different tolerance levels for deviation from the current block height.
type Config struct {
	// TODO_TECHDEBT(@adshmh): apply the sync allowance when validating an endpoint's block height.
	// SyncAllowance specifies the maximum number of blocks an endpoint
	// can be behind, compared to the blockchain's perceived block height,
	// before being filtered out.
	SyncAllowance uint64

	// ChainID is the ID used by the corresponding blockchain.
	// It is used to verify responses to service requests with `eth_chainId` method.
	ChainID string
}
