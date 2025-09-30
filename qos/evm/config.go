package evm

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
)

// Config represents EVM-specific service configuration
type Config struct {
	// Chain ID in hex format (e.g., "0x1" for Ethereum mainnet)
	ChainID string `yaml:"chain_id"`

	// Sync allowance (required, must be greater than 0)
	SyncAllowance uint64 `yaml:"sync_allowance"`

	// Optional: archival check configuration
	ArchivalCheck *ArchivalCheckConfig `yaml:"archival_check,omitempty"`

	// Supported RPC types for this service
	SupportedAPIs map[string]struct{} `yaml:"supported_apis"`
}

// Validate validates the Config struct
func (c *Config) Validate(logger polylog.Logger, serviceID protocol.ServiceID) error {
	if c.ChainID == "" {
		return fmt.Errorf("service %s: chain_id is required", serviceID)
	}

	if c.SyncAllowance == 0 {
		return fmt.Errorf("service %s: sync_allowance is required and must be greater than 0", serviceID)
	}

	if len(c.SupportedAPIs) == 0 {
		return fmt.Errorf("service %s: supported_apis must contain at least one API", serviceID)
	}

	// Validate ArchivalCheck if present
	if c.ArchivalCheck != nil {
		if err := c.ArchivalCheck.Validate(logger, serviceID); err != nil {
			return err
		}
	}

	return nil
}

// ArchivalCheckConfig represents the archival check configuration
type ArchivalCheckConfig struct {
	// Contract address to check for archival balance (in hex format)
	ContractAddress string `yaml:"contract_address"`

	// The block number when the contract first had a balance
	ContractStartBlock uint64 `yaml:"contract_start_block"`

	// Threshold
	Threshold uint64 `yaml:"threshold,omitempty"`
}

// Validate validates the ArchivalCheckConfig struct
func (a *ArchivalCheckConfig) Validate(logger polylog.Logger, serviceID protocol.ServiceID) error {
	if a.ContractAddress == "" {
		return fmt.Errorf("service %s: archival_check.contract_address is required when archival_check is set", serviceID)
	}

	if a.ContractStartBlock == 0 {
		return fmt.Errorf("service %s: archival_check.contract_start_block is required when archival_check is set", serviceID)
	}

	if a.Threshold == 0 {
		return fmt.Errorf("service %s: archival_check.threshold is required when archival_check is set", serviceID)
	}

	return nil
}
