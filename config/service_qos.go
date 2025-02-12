package config

// NOTE: Service ID list last updated 2025/01/22
//
// TODO_DOCUMENT(@commoddity): Add a README to [path docs](https://path.grove.city/) for developers.
// Consider a similar automated approach to "docs_update_gov_params_page"

import "github.com/buildwithgrove/path/protocol"

// IMPORTANT: PATH requires service IDs to be registered here for Quality of Service (QoS) endpoint checks.
// Unregistered services use NoOp QoS type with random endpoint selection and no monitoring.

// ServiceQoSType maps to a gateway.QoSService implementation that builds request QoS context
// and selects endpoints for a given service ID.
type ServiceQoSType string

const (
	// ServiceIDEVM represents the EVM service type, containing all EVM-based blockchains.
	ServiceIDEVM ServiceQoSType = "evm"

	// ServiceIDSolana represents the Solana blockchain service type.
	ServiceIDSolana ServiceQoSType = "solana"

	// ServiceIDPOKT represents the POKT blockchain service type.
	ServiceIDPOKT ServiceQoSType = "pokt"

	// ServiceIDCometBFT represents the CometBFT blockchain service type.
	ServiceIDCometBFT ServiceQoSType = "cometbft"
)

// TODO_MVP(@commoddity): figure out what these should be longer term.
const defaultEVMChainID = "0x1" // ETH Mainnet (1)
const defaultCometBFTChainID = "cosmoshub-4"

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
	for k := range shannonEVMChainIDs {
		ServiceQoSTypes[k] = ServiceIDEVM
	}
	for k := range shannonCometBFTChainIDs {
		ServiceQoSTypes[k] = ServiceIDCometBFT
	}
	for k, v := range morseQoSTypes {
		ServiceQoSTypes[k] = v
	}
	for k := range morseEVMChainIDs {
		ServiceQoSTypes[k] = ServiceIDEVM
	}
}

// GetEVMChainID returns the hexadecimal EVM chain ID for a given service ID.
// If the service ID is not found in the ShannonEVMChainIDs or MorseEVMChainIDs
// maps, the default EVM chain ID of `0x1` is returned.
func GetEVMChainID(serviceID protocol.ServiceID) string {
	if chainID, ok := shannonEVMChainIDs[serviceID]; ok {
		return chainID
	}
	if chainID, ok := morseEVMChainIDs[serviceID]; ok {
		return chainID
	}
	return defaultEVMChainID
}

func GetCometBFTChainID(serviceID protocol.ServiceID) string {
	if chainID, ok := shannonCometBFTChainIDs[serviceID]; ok {
		return chainID
	}
	return defaultCometBFTChainID
}

// IMPORTANT: To run QoS checks against a service, the service ID MUST be registered in one of the below maps.
// TODO_IMPROVE(@commoddity): consider using protocol scope for the service IDs.

// All non-EVM, non-CometBFT Shannon service IDs.
// As of the latest update, these service IDs are on Beta TestNet and intended to be moved
// over to MainNet once the network is ready.
var shannonQoSTypes = map[protocol.ServiceID]ServiceQoSType{
	// Solana Service IDs
	"solana": ServiceIDSolana,
}

// All Shannon EVM Service IDs and their corresponding EVM chain IDs.
// The map values are in hexadecimal format as this is the format returned by the
// node when making chain ID checks in the QoS hydrator.
// Reference: EVM chain IDs are sourced from https://chainlist.org
var shannonEVMChainIDs = map[protocol.ServiceID]string{
	// EVM service IDs
	"eth": "0x1", // ETH Mainnet (1)

	// Test QoS EVMservice IDs
	"anvil": "0x1", // ETH development/testing (1)
}

// All CometBFT Shannon Service IDs and their corresponding CometBFT chain IDs.
var shannonCometBFTChainIDs = map[protocol.ServiceID]string{
	"pocket-beta-rpc": "pocket-beta",
	"cometbft":        "cosmoshub-4",
}

// All non-EVM Morse Service IDs.
var morseQoSTypes = map[protocol.ServiceID]ServiceQoSType{
	// Solana Service IDs
	"F025": ServiceIDSolana, // Solana
}

// All Morse EVM Service IDs and their corresponding EVM chain IDs.
// The map values are in hexadecimal format as this is the format returned by the
// node when making chain ID checks in the QoS hydrator.
// Reference: EVM chain IDs are sourced from https://chainlist.org
var morseEVMChainIDs = map[protocol.ServiceID]string{
	"F001": "0xa4b1",     // Arbitrum One (42161)
	"F002": "0x66EEE",    // Arbitrum Sepolia Testnet (421614)
	"F003": "0xa86a",     // Avalanche (43114)
	"F004": "0xd2af",     // Avalanche-DFK (53935)
	"F005": "0x2105",     // Base (8453)
	"F006": "0x14a34",    // Base Sepolia Testnet (84660)
	"F008": "0x13e31",    // Blast (81649)
	"F009": "0x38",       // BNB Smart Chain (56)
	"F00A": "0x120",      // Boba (288)
	"F00B": "0xa4ec",     // Celo (42220)
	"F00C": "0x1",        // Ethereum (1)
	"F00D": "0x4268",     // Ethereum Holesky Testnet (17000)
	"F00E": "0xaa36a7",   // Ethereum Sepolia Testnet (11155420)
	"F00F": "0x2329",     // Evmos (9001)
	"F010": "0xfa",       // Fantom (250)
	"F011": "0xfc",       // Fraxtal (252)
	"F012": "0x7a",       // Fuse (122)
	"F013": "0x64",       // Gnosis (100)
	"F014": "0x63564c40", // Harmony-0 (1666600000)
	"F015": "0x1251",     // IoTeX (4681)
	"F016": "0x2019",     // Kaia (8217)
	"F017": "0x8ae",      // Kava (2222)
	"F018": "0x440",      // Metis (1088)
	"F019": "0x504",      // Moonbeam (1284)
	"F01A": "0x505",      // Moonriver (1285)
	"F01C": "0xf8",       // Oasys (248)
	"F01D": "0xa",        // Optimism (10)
	"F01E": "0xAA37DC",   // Optimism Sepolia Testnet (11155420)
	"F01F": "0xcc",       // opBNB (204)
	"F021": "0x89",       // Polygon (137)
	"F022": "0x13882",    // Polygon Amoy Testnet (80002)
	"F024": "0x82750",    // Scroll (534992)
	"F027": "0x28c58",    // Taiko (167000)
	"F028": "0x28c61",    // Taiko Hekla Testnet (167009)
	"F029": "0x44d",      // Polygon zkEVM (1101)
	"F02A": "0xc5cc4",    // zkLink (812564)
	"F02B": "0x144",      // zkSync (324)
	"F02C": "0x15f902",   // XRPL EVM Devnet (1440002)
	"F02D": "0x92",       // Sonic (146)
	"F02E": "0x2b6653dc", // TRON (728426128)
	"F030": "0xe708",     // Linea (59144)
	"F031": "0x138d4",    // Berachain bArtio Testnet (80084)
	"F032": "0xdef1",     // Ink (57073)
	"F033": "0x1388",     // Mantle (5000)
	"F034": "0x531",      // Sei (1329)
	"F035": "0x138de",    // Berachain (80094)
}
