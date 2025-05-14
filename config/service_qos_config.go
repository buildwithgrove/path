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
	// *** EVM Services (Archival) ***

	// Ethereum - ETH Mainnet
	evm.NewEVMServiceQoSConfig("eth", defaultEVMChainID, evm.NewEVMArchivalCheckConfig(
		// https://etherscan.io/address/0x28C6c06298d514Db089934071355E5743bf21d60
		"0x28C6c06298d514Db089934071355E5743bf21d60",
		// Contract start block
		12_300_000,
	)),

	// Polygon
	evm.NewEVMServiceQoSConfig("poly", "0x89", evm.NewEVMArchivalCheckConfig(
		// https://polygonscan.com/address/0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270
		"0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270",
		// Contract start block
		5_000_000,
	)),

	// Oasys
	evm.NewEVMServiceQoSConfig("oasys", "0xf8", evm.NewEVMArchivalCheckConfig(
		// https://explorer.oasys.games/address/0xf89d7b9c864f589bbF53a82105107622B35EaA40
		"0xf89d7b9c864f589bbF53a82105107622B35EaA40",
		// Contract start block
		424_300,
	)),

	// XRPL EVM Testnet
	evm.NewEVMServiceQoSConfig("xrpl_evm_testnet", "0x161c28", evm.NewEVMArchivalCheckConfig(
		// https://explorer.testnet.xrplevm.org/address/0xc29e2583eD5C77df8792067989Baf9E4CCD4D7fc
		"0xc29e2583eD5C77df8792067989Baf9E4CCD4D7fc",
		// Contract start block
		368_266,
	)),

	// BNB Smart Chain
	evm.NewEVMServiceQoSConfig("bsc", "0x38", evm.NewEVMArchivalCheckConfig(
		// https://bsctrace.com/address/0xfb50526f49894b78541b776f5aaefe43e3bd8590
		"0xfb50526f49894b78541b776f5aaefe43e3bd8590",
		// Contract start block
		33_049_200,
	)),

	// Optimism
	evm.NewEVMServiceQoSConfig("op", "0xa", evm.NewEVMArchivalCheckConfig(
		// https://optimistic.etherscan.io/address/0xacd03d601e5bb1b275bb94076ff46ed9d753435a
		"0xacD03D601e5bB1B275Bb94076fF46ED9D753435A",
		// Contract start block
		8_121_800,
	)),

	// *** EVM Services (testing) ***

	// Anvil - Ethereum development/testing
	evm.NewEVMServiceQoSConfig("anvil", "0x7a69", nil),

	// Anvil WebSockets - Ethereum WebSockets development/testing
	evm.NewEVMServiceQoSConfig("anvilws", "0x7a69", nil),

	// *** EVM Services (Non-Archival) ***

	// Arbitrum One
	evm.NewEVMServiceQoSConfig("arb_one", "0xa4b1", nil),

	// Arbitrum Sepolia Testnet
	evm.NewEVMServiceQoSConfig("arb_sep_test", "0x66EEE", nil),

	// Avalanche
	evm.NewEVMServiceQoSConfig("avax", "0xa86a", nil),

	// Avalanche-DFK
	evm.NewEVMServiceQoSConfig("avax-dfk", "0xd2af", nil),

	// Base
	evm.NewEVMServiceQoSConfig("base", "0x2105", nil),

	// Base Sepolia Testnet
	evm.NewEVMServiceQoSConfig("base-test", "0x14a34", nil),

	// Blast
	evm.NewEVMServiceQoSConfig("blast", "0x13e31", nil),

	// Boba
	evm.NewEVMServiceQoSConfig("boba", "0x120", nil),

	// Celo
	evm.NewEVMServiceQoSConfig("celo", "0xa4ec", nil),

	// Ethereum Holesky Testnet
	evm.NewEVMServiceQoSConfig("eth_hol_test", "0x4268", nil),

	// Ethereum Sepolia Testnet
	evm.NewEVMServiceQoSConfig("eth_sep_test", "0xaa36a7", nil),

	// Evmos
	evm.NewEVMServiceQoSConfig("evmos", "0x2329", nil),

	// Fantom
	evm.NewEVMServiceQoSConfig("fantom", "0xfa", nil),

	// Fraxtal
	evm.NewEVMServiceQoSConfig("fraxtal", "0xfc", nil),

	// Fuse
	evm.NewEVMServiceQoSConfig("fuse", "0x7a", nil),

	// Gnosis
	evm.NewEVMServiceQoSConfig("gnosis", "0x64", nil),

	// Harmony-0
	evm.NewEVMServiceQoSConfig("harmony", "0x63564c40", nil),

	// IoTeX
	evm.NewEVMServiceQoSConfig("iotex", "0x1251", nil),

	// Kaia
	evm.NewEVMServiceQoSConfig("kaia", "0x2019", nil),

	// Kava
	evm.NewEVMServiceQoSConfig("kava", "0x8ae", nil),

	// Metis
	evm.NewEVMServiceQoSConfig("metis", "0x440", nil),

	// Moonbeam
	evm.NewEVMServiceQoSConfig("moonbeam", "0x504", nil),

	// Moonriver
	evm.NewEVMServiceQoSConfig("moonriver", "0x505", nil),

	// Near
	evm.NewEVMServiceQoSConfig("near", "0x18d", nil),

	// Optimism Sepolia Testnet
	evm.NewEVMServiceQoSConfig("op_sep_test", "0xAA37DC", nil),

	// opBNB
	evm.NewEVMServiceQoSConfig("opbnb", "0xcc", nil),

	// Polygon Amoy Testnet
	evm.NewEVMServiceQoSConfig("poly_amoy_test", "0x13882", nil),

	// Radix
	evm.NewEVMServiceQoSConfig("radix", "0x1337", nil),

	// Scroll
	evm.NewEVMServiceQoSConfig("scroll", "0x82750", nil),

	// Sui
	evm.NewEVMServiceQoSConfig("sui", "0x101", nil),

	// Taiko
	evm.NewEVMServiceQoSConfig("taiko", "0x28c58", nil),

	// Taiko Hekla Testnet
	evm.NewEVMServiceQoSConfig("taiko_hek_test", "0x28c61", nil),

	// Polygon zkEVM
	evm.NewEVMServiceQoSConfig("poly_zkevm", "0x44d", nil),

	// zkLink
	evm.NewEVMServiceQoSConfig("zklink_nova", "0xc5cc4", nil),

	// zkSync
	evm.NewEVMServiceQoSConfig("zksync_era", "0x144", nil),

	// XRPL EVM Devnet
	evm.NewEVMServiceQoSConfig("xrpl_evm_dev", "0x15f902", nil),

	// Sonic
	evm.NewEVMServiceQoSConfig("sonic", "0x92", nil),

	// TRON
	evm.NewEVMServiceQoSConfig("tron", "0x2b6653dc", nil),

	// Linea
	evm.NewEVMServiceQoSConfig("linea", "0xe708", nil),

	// Ink
	evm.NewEVMServiceQoSConfig("ink", "0xdef1", nil),

	// Mantle
	evm.NewEVMServiceQoSConfig("mantle", "0x1388", nil),

	// Sei
	evm.NewEVMServiceQoSConfig("sei", "0x531", nil),

	// Berachain
	evm.NewEVMServiceQoSConfig("bera", "0x138de", nil),

	// *** CometBFT Services ***

	// TODO_MVP(@commoddity): Ensure that QoS observations are being applied correctly and that
	// the correct chain ID is being used for each service in the CometBFT config.

	// Celestia Archival
	cometbft.NewCometBFTServiceQoSConfig("tia_da", "celestia-archival"),

	// Celestia Consensus Archival
	cometbft.NewCometBFTServiceQoSConfig("tia_cons", "celestia-consensus-archival"),

	// Celestia Testnet DA Archival
	cometbft.NewCometBFTServiceQoSConfig("tia_da_test", "celestia-testnet-da-archival"),

	// Celestia Testnet Consensus Archival
	cometbft.NewCometBFTServiceQoSConfig("tia_cons_test", "celestia-testnet-consensus-archival"),

	// Osmosis
	cometbft.NewCometBFTServiceQoSConfig("osmosis", "osmosis"),

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
	// *** EVM Services (Archival) ***

	// Ethereum
	evm.NewEVMServiceQoSConfig("F00C", defaultEVMChainID, evm.NewEVMArchivalCheckConfig(
		// https://etherscan.io/address/0x28C6c06298d514Db089934071355E5743bf21d60
		"0x28C6c06298d514Db089934071355E5743bf21d60",
		// Contract start block
		12_300_000,
	)),

	// Polygon
	evm.NewEVMServiceQoSConfig("F021", "0x89", evm.NewEVMArchivalCheckConfig(
		// https://polygonscan.com/address/0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270
		"0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270",
		// Contract start block
		5_000_000,
	)),

	// Oasys
	evm.NewEVMServiceQoSConfig("F01C", "0xf8", evm.NewEVMArchivalCheckConfig(
		// https://explorer.oasys.games/address/0xf89d7b9c864f589bbF53a82105107622B35EaA40
		"0xf89d7b9c864f589bbF53a82105107622B35EaA40",
		// Contract start block
		424_300,
	)),

	// XRPL EVM Testnet
	evm.NewEVMServiceQoSConfig("F036", "0x161c28", evm.NewEVMArchivalCheckConfig(
		// https://explorer.testnet.xrplevm.org/address/0xc29e2583eD5C77df8792067989Baf9E4CCD4D7fc
		"0xc29e2583eD5C77df8792067989Baf9E4CCD4D7fc",
		// Contract start block
		368_266,
	)),

	// BNB Smart Chain
	evm.NewEVMServiceQoSConfig("F009", "0x38", evm.NewEVMArchivalCheckConfig(
		// https://bsctrace.com/address/0xfb50526f49894b78541b776f5aaefe43e3bd8590
		"0xfb50526f49894b78541b776f5aaefe43e3bd8590",
		// Contract start block
		33_049_200,
	)),

	// Optimism
	evm.NewEVMServiceQoSConfig("F01D", "0xa", evm.NewEVMArchivalCheckConfig(
		// https://optimistic.etherscan.io/address/0xacd03d601e5bb1b275bb94076ff46ed9d753435a
		"0xacD03D601e5bB1B275Bb94076fF46ED9D753435A",
		// Contract start block
		8_121_800,
	)),

	// *** EVM Services (Non-Archival) ***

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

	// Boba
	evm.NewEVMServiceQoSConfig("F00A", "0x120", nil),

	// Celo
	evm.NewEVMServiceQoSConfig("F00B", "0xa4ec", nil),

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

	// Optimism Sepolia Testnet
	evm.NewEVMServiceQoSConfig("F01E", "0xAA37DC", nil),

	// opBNB
	evm.NewEVMServiceQoSConfig("F01F", "0xcc", nil),

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

	// *** CometBFT Services ***
	// TODO_MVP(@commoddity): Ensure that QoS observations are being applied correctly and that
	// the correct chain ID is being used for each service in the CometBFT config.

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
