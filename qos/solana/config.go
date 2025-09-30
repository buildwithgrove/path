package solana

// Config represents Solana-specific service configuration
type Config struct {
	// Chain ID (e.g., "solana", "mainnet-beta")
	ChainID string `yaml:"chain_id"`
}
