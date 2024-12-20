package config

import "github.com/buildwithgrove/path/protocol"

/* IMPORTANT: In order for PATH to run Quality of Service (QoS) checks against the endpoints for a service,
the authoritative service ID MUST be registered in this file, which is used to build the ServiceQoSTypes map.

Services that are not registered in this file will be supported but will use the NoOp service QoS type,
which selects a random endpoint for the given service and does not perform any observations or QoS checks. */

// TODO_DOCUMENT(@commoddity): Add a README to [path docs](https://path.grove.city/) for developers.

// The ServiceQoSType type corresponds to a specific implementation of the
// gateway.QoSService interface, which is used to build the request QoS context
// and select the endpoint for a request for a given service ID.
type ServiceQoSType string

const (
	ServiceIDEVM    ServiceQoSType = "evm"    // ServiceIDEVM represents the EVM service type, containing all EVM-based blockchains.
	ServiceIDSolana ServiceQoSType = "solana" // ServiceIDSolana represents the Solana blockchain service type.
	ServiceIDPOKT   ServiceQoSType = "pokt"   // ServiceIDPOKT represents the POKT blockchain service type.
)

// The ServiceQoSTypes map associates each supported service ID with a specific
// implementation of the gateway.QoSService interface.
// This is to handle requests for a given service ID.
//
// IMPORTANT: Only service IDs that are part of this map will have QoS checks performed.
// All other service IDS will be supported but will have the NoOp service QoS type,
// which does not perform any observations or QoS checks, meaning a random endpoint
// for the given service ID will be selected for the request.
//
// DEV_NOTE: The ServiceQoSTypes map is initialized in the init() function.
var ServiceQoSTypes = map[protocol.ServiceID]ServiceQoSType{}

func init() {
	for k, v := range shannonQoSTypes {
		ServiceQoSTypes[k] = v
	}
	for k, v := range morseQoSTypes {
		ServiceQoSTypes[k] = v
	}
	for k, v := range shannonTestQoSTypes {
		ServiceQoSTypes[k] = v
	}
}

// IMPORTANT: To run QoS checks against a service, the service ID MUST be registered in one of the below maps.
// TODO_IMPROVE(@commoddity): consider using protocol scope for the service IDs.

// Shannon service IDs.
// As of 12/2024, these service IDs are on Beta TestNet and intended to be moved
// over to MainNet once the network is ready.
var shannonQoSTypes = map[protocol.ServiceID]ServiceQoSType{
	// Solana Service IDs
	"solana": ServiceIDSolana,

	// EVM Service IDs
	"eth": ServiceIDEVM,

	// POKT Service IDs
	"pokt":  ServiceIDPOKT,
	"morse": ServiceIDPOKT,
}

// Shannon test service IDs.
// As of 12/2024, these service IDs are on Beta TestNet and primarily used
// for E2E testing. They may or may not be moved over to MainNet once the network.
var shannonTestQoSTypes = map[protocol.ServiceID]ServiceQoSType{
	// Shannon Service IDs
	"anvil": ServiceIDEVM, // ETH Local (development/testing)
}

// Service IDs transferred from Morse to Shannon for backwards compatibility.
var morseQoSTypes = map[protocol.ServiceID]ServiceQoSType{
	// All Morse EVM F-Chain Services as of 12/17/2024 (#103)
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

	// Solana F-Chain Service IDs as of 12/2024 (#103)
	"F025": ServiceIDSolana, // Solana
}
