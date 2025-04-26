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
	evm.NewEVMServiceQoSConfig("anvil", "0x7a69", nil),

	// Anvil WebSockets - Ethereum WebSockets development/testing
	evm.NewEVMServiceQoSConfig("anvilws", "0x7a69", nil),

	// *** CometBFT Services ***

	// Pocket Beta Testnet
	cometbft.NewCometBFTServiceQoSConfig("pocket-beta-rpc", "pocket-beta"),

	// Cosmos Hub
	cometbft.NewCometBFTServiceQoSConfig("cometbft", "cosmoshub-4"),

	// *** Solana Services ***

	// Solana
	solana.NewSolanaServiceQoSConfig("solana"),
}

// morseServices is the list of QoS service configs for the Morse protocol.
var morseServices = []ServiceQoSConfig{
	// *** EVM Services ***

	// Arbitrum One
	evm.NewEVMServiceQoSConfig("F001", "0xa4b1", nil),

	// Arbitrum Sepolia Testnet
	evm.NewEVMServiceQoSConfig("F002", "0x66EEE", nil),

	// Avalanche
	evm.NewEVMServiceQoSConfig("F003", "0xa86a", nil),

	// Avalanche-DFK
	evm.NewEVMServiceQoSConfig("F004", "0xd2af", nil),

	// Base
	evm.NewEVMServiceQoSConfig("F005", "0x2105", nil),

	// Base Sepolia Testnet
	evm.NewEVMServiceQoSConfig("F006", "0x14a34", nil),

	// Blast
	evm.NewEVMServiceQoSConfig("F008", "0x13e31", nil),

	// BNB Smart Chain
	evm.NewEVMServiceQoSConfig("F009", "0x38", nil),

	// Boba
	evm.NewEVMServiceQoSConfig("F00A", "0x120", nil),

	// Celo
	evm.NewEVMServiceQoSConfig("F00B", "0xa4ec", nil),

	// Ethereum
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

	// Ethereum Holesky Testnet
	evm.NewEVMServiceQoSConfig("F00D", "0x4268", nil),

	// Ethereum Sepolia Testnet
	evm.NewEVMServiceQoSConfig("F00E", "0xaa36a7", nil),

	// Evmos
	evm.NewEVMServiceQoSConfig("F00F", "0x2329", nil),

	// Fantom
	evm.NewEVMServiceQoSConfig("F010", "0xfa", nil),

	// Fraxtal
	evm.NewEVMServiceQoSConfig("F011", "0xfc", nil),

	// Fuse
	evm.NewEVMServiceQoSConfig("F012", "0x7a", nil),

	// Gnosis
	evm.NewEVMServiceQoSConfig("F013", "0x64", nil),

	// Harmony-0
	evm.NewEVMServiceQoSConfig("F014", "0x63564c40", nil),

	// IoTeX
	evm.NewEVMServiceQoSConfig("F015", "0x1251", nil),

	// Kaia
	evm.NewEVMServiceQoSConfig("F016", "0x2019", nil),

	// Kava
	evm.NewEVMServiceQoSConfig("F017", "0x8ae", nil),

	// Metis
	evm.NewEVMServiceQoSConfig("F018", "0x440", nil),

	// Moonbeam
	evm.NewEVMServiceQoSConfig("F019", "0x504", nil),

	// Moonriver
	evm.NewEVMServiceQoSConfig("F01A", "0x505", nil),

	// Near
	evm.NewEVMServiceQoSConfig("F01B", "0x18d", nil),

	// Oasys
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

	// Optimism
	evm.NewEVMServiceQoSConfig("F01D", "0xa", nil),

	// Optimism Sepolia Testnet
	evm.NewEVMServiceQoSConfig("F01E", "0xAA37DC", nil),

	// opBNB
	evm.NewEVMServiceQoSConfig("F01F", "0xcc", nil),

	// Polygon
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

	// Polygon Amoy Testnet
	evm.NewEVMServiceQoSConfig("F022", "0x13882", nil),

	// Radix
	evm.NewEVMServiceQoSConfig("F023", "0x1337", nil),

	// Scroll
	evm.NewEVMServiceQoSConfig("F024", "0x82750", nil),

	// Sui
	evm.NewEVMServiceQoSConfig("F026", "0x101", nil),

	// Taiko
	evm.NewEVMServiceQoSConfig("F027", "0x28c58", nil),

	// Taiko Hekla Testnet
	evm.NewEVMServiceQoSConfig("F028", "0x28c61", nil),

	// Polygon zkEVM
	evm.NewEVMServiceQoSConfig("F029", "0x44d", nil),

	// zkLink
	evm.NewEVMServiceQoSConfig("F02A", "0xc5cc4", nil),

	// zkSync
	evm.NewEVMServiceQoSConfig("F02B", "0x144", nil),

	// XRPL EVM Devnet
	evm.NewEVMServiceQoSConfig("F02C", "0x15f902", nil),

	// Sonic
	evm.NewEVMServiceQoSConfig("F02D", "0x92", nil),

	// TRON
	evm.NewEVMServiceQoSConfig("F02E", "0x2b6653dc", nil),

	// Linea
	evm.NewEVMServiceQoSConfig("F030", "0xe708", nil),

	// Berachain Testnet
	evm.NewEVMServiceQoSConfig("F031", "0x138d4", nil),

	// Ink
	evm.NewEVMServiceQoSConfig("F032", "0xdef1", nil),

	// Mantle
	evm.NewEVMServiceQoSConfig("F033", "0x1388", nil),

	// Sei
	evm.NewEVMServiceQoSConfig("F034", "0x531", nil),

	// Berachain
	evm.NewEVMServiceQoSConfig("F035", "0x138de", nil),

	// XRPL EVM Testnet
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

	// Celestia Archival
	cometbft.NewCometBFTServiceQoSConfig("A0CA", "celestia-archival"),

	// Celestia Consensus Archival
	cometbft.NewCometBFTServiceQoSConfig("A0CB", "celestia-consensus-archival"),

	// Celestia Testnet DA Archival
	cometbft.NewCometBFTServiceQoSConfig("A0CC", "celestia-testnet-da-archival"),

	// Celestia Testnet Consensus Archival
	cometbft.NewCometBFTServiceQoSConfig("A0CD", "celestia-testnet-consensus-archival"),

	// Osmosis
	cometbft.NewCometBFTServiceQoSConfig("F020", "osmosis"),

	// *** Solana Services ***

	// Solana
	solana.NewSolanaServiceQoSConfig("F025"),
}
