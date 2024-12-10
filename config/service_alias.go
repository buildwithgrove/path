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
	// TODO_IMPROVE(@commoddity): Add all EVM service IDs here.
	"0021":        ServiceIDEVM, // ETH Mainnet
	"F00C":        ServiceIDEVM, // ETH Mainnet (Full/Archival)
	"anvil":       ServiceIDEVM, // ETH Local (development/testing)
	"eth":         ServiceIDEVM, // ETH general-purpose catch-all
	"eth-mainnet": ServiceIDEVM, // ETH MainNet general-purpose catch-all

	// TODO_IMPROVE(@commoddity): Use actual service IDs for Solana and POKT.
	"solana": ServiceIDSolana,

	"pokt":  ServiceIDPOKT,
	"morse": ServiceIDPOKT,

	// Gateway E2E service ID is used only for running PATH's Morse and Shannon E2E tests.
	"gatewaye2e": ServiceIDE2E,
}
