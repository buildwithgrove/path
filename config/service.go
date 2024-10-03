package config

import (
	"github.com/buildwithgrove/path/relayer"
)

const (
	// TODO_IMPROVE: consider using protocol scope for the service IDs.
	ServiceIDEVM    = relayer.ServiceID("evm")    // ServiceIDEVM represents the EVM service type, containing all EVM-based blockchains.
	ServiceIDSolana = relayer.ServiceID("solana") // ServiceIDSolana represents the Solana blockchain service type.
	ServiceIDPOKT   = relayer.ServiceID("pokt")   // ServiceIDPOKT represents the POKT blockchain service type.
)
