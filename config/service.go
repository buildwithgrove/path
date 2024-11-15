package config

import (
	"github.com/buildwithgrove/path/relayer"
)

const (
	// TODO_IMPROVE: consider using protocol scope for the service IDs.
	ServiceIDEVM    = relayer.ServiceID("F00C")       // ServiceIDEVM represents the EVM service type, containing all EVM-based blockchains.
	ServiceIDSolana = relayer.ServiceID("solana")     // ServiceIDSolana represents the Solana blockchain service type.
	ServiceIDPOKT   = relayer.ServiceID("pokt")       // ServiceIDPOKT represents the POKT blockchain service type.
	ServiceIDE2E    = relayer.ServiceID("gatewaye2e") //ServiceIDE2E represents the service created for running PATH gateway's E2E tests.
)
