package config

import "github.com/buildwithgrove/path/protocol"

// TODO_DOCUMENT(@commoddity): Add a README to [path docs](https://path.grove.city/) for developers.
// IMPORTANT: All service IDs supported by PATH must be registered in this file.

// The ServiceQoSType type corresponds to a specific implementation of the
// gateway.QoSService interface, which is used to build the request QoS context
// and select the endpoint for a request for a given service ID.
type ServiceQoSType string

const (
	// TODO_IMPROVE(@commoddity): consider using protocol scope for the service IDs.
	ServiceIDEVM    ServiceQoSType = "evm"        // ServiceIDEVM represents the EVM service type, containing all EVM-based blockchains.
	ServiceIDSolana ServiceQoSType = "solana"     // ServiceIDSolana represents the Solana blockchain service type.
	ServiceIDPOKT   ServiceQoSType = "pokt"       // ServiceIDPOKT represents the POKT blockchain service type.
	ServiceIDE2E    ServiceQoSType = "gatewaye2e" // ServiceIDE2E represents the service created for running PATH gateway's E2E tests.
)

// The ServiceQoSTypes map associates each supported service ID with a specific implementation of the
// gateway.QoSService interface.
// THis is to handle requests for a given service ID.
var ServiceQoSTypes = map[protocol.ServiceID]ServiceQoSType{
	// All Morse EVM Services as of 12/17/2024 (#103)
	// TODO_TECHDEBT(@fredteumer): Revisit and consider removing these once #105 is complete.
	"F001": ServiceIDEVM, // Arbitrum One
	"F002": ServiceIDEVM, // Arbitrum Sepolia Testnet
	"F003": ServiceIDEVM, // Avalanche
	"F004": ServiceIDEVM, // Avalanche-DFK
	"F005": ServiceIDEVM, // Base
	"F006": ServiceIDEVM, // Base Sepolia Testnet
	"F008": ServiceIDEVM, // Blast
	"F009": ServiceIDEVM, // BNB Smart Chain
	"F00A": ServiceIDEVM, // Boba
	"F00B": ServiceIDEVM, // Celo
	"F00C": ServiceIDEVM, // Ethereum
	"F00D": ServiceIDEVM, // Ethereum Holesky Testnet
	"F00E": ServiceIDEVM, // Ethereum Sepolia Testnet
	"F00F": ServiceIDEVM, // Evmos
	"F010": ServiceIDEVM, // Fantom
	"F011": ServiceIDEVM, // Fraxtal
	"F012": ServiceIDEVM, // Fuse
	"F013": ServiceIDEVM, // Gnosis
	"F014": ServiceIDEVM, // Harmony-0
	"F015": ServiceIDEVM, // IoTeX
	"F016": ServiceIDEVM, // Kaia
	"F017": ServiceIDEVM, // Kava
	"F018": ServiceIDEVM, // Metis
	"F019": ServiceIDEVM, // Moonbeam
	"F01A": ServiceIDEVM, // Moonriver
	"F01C": ServiceIDEVM, // Oasys
	"F01D": ServiceIDEVM, // Optimism
	"F01E": ServiceIDEVM, // Optimism Sepolia Testnet
	"F01F": ServiceIDEVM, // opBNB
	"F021": ServiceIDEVM, // Polygon
	"F022": ServiceIDEVM, // Polygon Amoy Testnet
	"F024": ServiceIDEVM, // Scroll
	"F027": ServiceIDEVM, // Taiko
	"F028": ServiceIDEVM, // Taiko Hekla Testnet
	"F029": ServiceIDEVM, // Polygon zkEVM
	"F02A": ServiceIDEVM, // zkLink
	"F02B": ServiceIDEVM, // zkSync

	// TODO_IMPROVE Add all non-EVM Morse Services and requisite initialization

	// Shannon Service IDs
	"anvil": ServiceIDEVM, // ETH Local (development/testing)

	// TODO_IMPROVE(@commoddity): Use actual service IDs for Solana and POKT.
	"solana": ServiceIDSolana,

	"pokt":  ServiceIDPOKT,
	"morse": ServiceIDPOKT,

	// Gateway E2E service ID is used only for running PATH's Morse and Shannon E2E tests.
	"gatewaye2e": ServiceIDE2E,
}
