package config

import (
	"github.com/buildwithgrove/path/protocol"
)

const (
	// TODO_IMPROVE: consider using protocol scope for the service IDs.
	ServiceIDEVM    = protocol.ServiceID("0021")       // ServiceIDEVM represents the EVM service type, containing all EVM-based blockchains.
	ServiceIDSolana = protocol.ServiceID("solana")     // ServiceIDSolana represents the Solana blockchain service type.
	ServiceIDPOKT   = protocol.ServiceID("pokt")       // ServiceIDPOKT represents the POKT blockchain service type.
	ServiceIDE2E    = protocol.ServiceID("gatewaye2e") //ServiceIDE2E represents the service created for running PATH gateway's E2E tests.
)
