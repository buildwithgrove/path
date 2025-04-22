package config

import (
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/cometbft"
	"github.com/buildwithgrove/path/qos/evm"
	"github.com/buildwithgrove/path/qos/solana"
)

// NOTE: Service ID list last updated 2025/04/10

// IMPORTANT: PATH requires service IDs to be registered here for Quality of Service (QoS) endpoint checks.
// Unregistered services use NoOp QoS type with random endpoint selection and no monitoring.

// TODO_QOS(@commoddity): Add archival check configurations for all EVM services.
// This includes hydrating the entire EVMArchivalCheckConfig struct.
// See the following archival check configurations as reference:
//   - F00C (Ethereum)
//   - F021 (Polygon)
//   - F01C (Oasys)
//   - F036 (XRPL EVM Testnet)

var _ ServiceQoSConfig = (evm.EVMServiceQoSConfig)(nil)
var _ ServiceQoSConfig = (cometbft.CometBFTServiceQoSConfig)(nil)
var _ ServiceQoSConfig = (solana.SolanaServiceQoSConfig)(nil)

type ServiceQoSConfig interface {
	GetServiceID() protocol.ServiceID
	GetServiceQoSType() string
}

// qosServiceConfigs captures the list of blockchains that PATH supports QoS for.
type qosServiceConfigs struct {
	shannonServices []ServiceQoSConfig
	morseServices   []ServiceQoSConfig
}

// GetServiceConfigs returns the service configs for the provided protocol supported by the Gateway.
func (c qosServiceConfigs) GetServiceConfigs(config GatewayConfig) []ServiceQoSConfig {
	// Shannon configurations
	if shannonConfig := config.GetShannonConfig(); shannonConfig != nil {
		return c.shannonServices
	}

	// Morse configurations
	if morseConfig := config.GetMorseConfig(); morseConfig != nil {
		return c.morseServices
	}

	// If no configuration is found, return an empty slice.
	return nil
}

// The ServiceConfigs map associates each supported service ID with a specific
// implementation of the gateway.QoSService interface.
var ServiceConfigs = qosServiceConfigs{
	shannonServices: shannonServices,
	morseServices:   morseServices,
}

const (
	defaultEVMChainID      = "0x1" // ETH Mainnet (1)
	defaultCometBFTChainID = "cosmoshub-4"
)

// shannonServices is the list of QoS service configs for the Shannon protocol.
var shannonServices = []ServiceQoSConfig{
	// *** EVM Services ***
	// Ethereum - ETH Mainnet
	evm.NewEVMServiceQoSConfig("eth", defaultEVMChainID, nil),

	// Anvil - Ethereum development/testing
	evm.NewEVMServiceQoSConfig("anvil", defaultEVMChainID, nil),

	// Anvil WebSockets - Ethereum WebSockets development/testing
	evm.NewEVMServiceQoSConfig("anvilws", defaultEVMChainID, nil),

	// *** CometBFT Services ***
	// CometBFT - Pocket Beta Testnet
	cometbft.NewCometBFTServiceQoSConfig("pocket-beta-rpc", "pocket-beta"),

	// CometBFT - Cosmos Hub
	cometbft.NewCometBFTServiceQoSConfig("cometbft", "cosmoshub-4"),

	// *** Solana Services ***
	// Solana
	solana.NewSolanaServiceQoSConfig("solana"),
}

// morseServices is the list of QoS service configs for the Morse protocol.
var morseServices = []ServiceQoSConfig{
	// *** EVM Services ***

	// Arbitrum One (42161)
	evm.NewEVMServiceQoSConfig("F001", "0xa4b1", nil),

	// Arbitrum Sepolia Testnet (421614)
	evm.NewEVMServiceQoSConfig("F002", "0x66EEE", nil),

	// Avalanche (43114)
	evm.NewEVMServiceQoSConfig("F003", "0xa86a", nil),

	// Avalanche-DFK (53935)
	evm.NewEVMServiceQoSConfig("F004", "0xd2af", nil),

	// Base (8453)
	evm.NewEVMServiceQoSConfig("F005", "0x2105", nil),

	// Base Sepolia Testnet (84660)
	evm.NewEVMServiceQoSConfig("F006", "0x14a34", nil),

	// Blast (81649)
	evm.NewEVMServiceQoSConfig("F008", "0x13e31", nil),

	// BNB Smart Chain (56)
	evm.NewEVMServiceQoSConfig("F009", "0x38", nil),

	// Boba (288)
	evm.NewEVMServiceQoSConfig("F00A", "0x120", nil),

	// Celo (42220)
	evm.NewEVMServiceQoSConfig("F00B", "0xa4ec", nil),

	// Ethereum (1)
	evm.NewEVMServiceQoSConfig(
		"F00C",
		defaultEVMChainID,
		evm.NewEVMArchivalCheckConfig(
			// https://etherscan.io/address/0x28C6c06298d514Db089934071355E5743bf21d60
			"0x28C6c06298d514Db089934071355E5743bf21d60",
			// Contract start block
			12_300_000,
		),
	),

	// Ethereum Holesky Testnet (17000)
	evm.NewEVMServiceQoSConfig("F00D", "0x4268", nil),

	// Ethereum Sepolia Testnet (11155111)
	evm.NewEVMServiceQoSConfig("F00E", "0xaa36a7", nil),

	// Evmos (9001)
	evm.NewEVMServiceQoSConfig("F00F", "0x2329", nil),

	// Fantom (250)
	evm.NewEVMServiceQoSConfig("F010", "0xfa", nil),

	// Fraxtal (252)
	evm.NewEVMServiceQoSConfig("F011", "0xfc", nil),

	// Fuse (122)
	evm.NewEVMServiceQoSConfig("F012", "0x7a", nil),

	// Gnosis (100)
	evm.NewEVMServiceQoSConfig("F013", "0x64", nil),

	// Harmony-0 (1666600000)
	evm.NewEVMServiceQoSConfig("F014", "0x63564c40", nil),

	// IoTeX (4681)
	evm.NewEVMServiceQoSConfig("F015", "0x1251", nil),

	// Kaia (8217)
	evm.NewEVMServiceQoSConfig("F016", "0x2019", nil),

	// Kava (2222)
	evm.NewEVMServiceQoSConfig("F017", "0x8ae", nil),

	// Metis (1088)
	evm.NewEVMServiceQoSConfig("F018", "0x440", nil),

	// Moonbeam (1284)
	evm.NewEVMServiceQoSConfig("F019", "0x504", nil),

	// Moonriver (1285)
	evm.NewEVMServiceQoSConfig("F01A", "0x505", nil),

	// Near
	evm.NewEVMServiceQoSConfig("F01B", "0x18d", nil),

	// Oasys (248)
	evm.NewEVMServiceQoSConfig(
		"F01C",
		"0xf8",
		evm.NewEVMArchivalCheckConfig(
			// https://explorer.oasys.games/address/0xf89d7b9c864f589bbF53a82105107622B35EaA40
			"0xf89d7b9c864f589bbF53a82105107622B35EaA40",
			// Contract start block
			424_300,
		),
	),

	// Optimism (10)
	evm.NewEVMServiceQoSConfig("F01D", "0xa", nil),

	// Optimism Sepolia Testnet (11155420)
	evm.NewEVMServiceQoSConfig("F01E", "0xAA37DC", nil),

	// opBNB (204)
	evm.NewEVMServiceQoSConfig("F01F", "0xcc", nil),

	// Polygon (137)
	evm.NewEVMServiceQoSConfig(
		"F021",
		"0x89",
		evm.NewEVMArchivalCheckConfig(
			// https://polygonscan.com/address/0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270
			"0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270",
			// Contract start block
			5_000_000,
		),
	),

	// Polygon Amoy Testnet (80002)
	evm.NewEVMServiceQoSConfig("F022", "0x13882", nil),

	// Radix
	evm.NewEVMServiceQoSConfig("F023", "0x1337", nil),

	// Scroll (534992)
	evm.NewEVMServiceQoSConfig("F024", "0x82750", nil),

	// Sui
	evm.NewEVMServiceQoSConfig("F026", "0x101", nil),

	// Taiko (167000)
	evm.NewEVMServiceQoSConfig("F027", "0x28c58", nil),

	// Taiko Hekla Testnet (167009)
	evm.NewEVMServiceQoSConfig("F028", "0x28c61", nil),

	// Polygon zkEVM (1101)
	evm.NewEVMServiceQoSConfig("F029", "0x44d", nil),

	// zkLink (812564)
	evm.NewEVMServiceQoSConfig("F02A", "0xc5cc4", nil),

	// zkSync (324)
	evm.NewEVMServiceQoSConfig("F02B", "0x144", nil),

	// XRPL EVM Devnet (1440002)
	evm.NewEVMServiceQoSConfig("F02C", "0x15f902", nil),

	// Sonic (146)
	evm.NewEVMServiceQoSConfig("F02D", "0x92", nil),

	// TRON (728426128)
	evm.NewEVMServiceQoSConfig("F02E", "0x2b6653dc", nil),

	// Linea (59144)
	evm.NewEVMServiceQoSConfig("F030", "0xe708", nil),

	// Berachain Testnet (80084)
	evm.NewEVMServiceQoSConfig("F031", "0x138d4", nil),

	// Ink (57073)
	evm.NewEVMServiceQoSConfig("F032", "0xdef1", nil),

	// Mantle (5000)
	evm.NewEVMServiceQoSConfig("F033", "0x1388", nil),

	// Sei (1329)
	evm.NewEVMServiceQoSConfig("F034", "0x531", nil),

	// Berachain (80094)
	evm.NewEVMServiceQoSConfig("F035", "0x138de", nil),

	// XRPL EVM Testnet (1449000)
	evm.NewEVMServiceQoSConfig(
		"F036",
		"0x161c28",
		evm.NewEVMArchivalCheckConfig(
			// https://explorer.testnet.xrplevm.org/address/0xc29e2583eD5C77df8792067989Baf9E4CCD4D7fc
			"0xc29e2583eD5C77df8792067989Baf9E4CCD4D7fc",
			// Contract start block
			368_266,
		),
	),

	// *** CometBFT Services ***
	cometbft.NewCometBFTServiceQoSConfig("A0CA", "celestia-archival"),
	cometbft.NewCometBFTServiceQoSConfig("A0CB", "celestia-consensus-archival"),
	cometbft.NewCometBFTServiceQoSConfig("A0CC", "celestia-testnet-da-archival"),
	cometbft.NewCometBFTServiceQoSConfig("A0CD", "celestia-testnet-consensus-archival"),
	cometbft.NewCometBFTServiceQoSConfig("F020", "osmosis"),

	// *** Solana Services ***
	// Solana
	solana.NewSolanaServiceQoSConfig("solana"),
	solana.NewSolanaServiceQoSConfig("F025"),
}

// Configuration now aligned with the service_ids list provided
