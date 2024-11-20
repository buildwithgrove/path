package main

import "github.com/buildwithgrove/path/relayer"

// All service IDs supported by PATH must be registered in
// this file, and associated to a single Service QoS type.

type serviceQoSType string

const (
	// TODO_IMPROVE: consider using protocol scope for the service IDs.
	serviceIDEVM    serviceQoSType = "evm"        // ServiceIDEVM represents the EVM service type, containing all EVM-based blockchains.
	serviceIDSolana serviceQoSType = "solana"     // ServiceIDSolana represents the Solana blockchain service type.
	serviceIDPOKT   serviceQoSType = "pokt"       // ServiceIDPOKT represents the POKT blockchain service type.
	serviceIDE2E    serviceQoSType = "gatewaye2e" // ServiceIDE2E represents the service created for running PATH gateway's E2E tests.
)

var serviceQoSTypes = map[relayer.ServiceID]serviceQoSType{
	// TODO_IMPROVE(@commoddity): Add all EVM service IDs here.
	"0021": serviceIDEVM,
	"F00C": serviceIDEVM,

	// TODO_IMPROVE(@commoddity): Use actual service IDs for Solana and POKT.
	"solana": serviceIDSolana,
	"pokt":   serviceIDPOKT,

	// Gateway E2E service ID is used only for running PATH's Morse and Shannon E2E tests.
	"gatewaye2e": serviceIDE2E,
}
