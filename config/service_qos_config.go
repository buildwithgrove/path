package config

import (
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/cosmos"
	"github.com/buildwithgrove/path/qos/evm"
	"github.com/buildwithgrove/path/qos/solana"
)

// NOTE: Service ID list last updated 2025/04/10

// IMPORTANT: PATH requires service IDs to be registered here for Quality of Service (QoS) endpoint checks.
// Unregistered services use NoOp QoS type with random endpoint selection and no monitoring.

var _ ServiceQoSConfig = (evm.EVMServiceQoSConfig)(nil)
var _ ServiceQoSConfig = (cosmos.CosmosSDKServiceQoSConfig)(nil)
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

// The QoSServiceConfigs map associates each supported service ID with a specific
// implementation of the gateway.QoSService interface.
var QoSServiceConfigs = qosServiceConfigs{
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

	// Arbitrum One
	evm.NewEVMServiceQoSConfig("arb-one", "0xa4b1", evm.NewEVMArchivalCheckConfig(
		// https://arbiscan.io/address/0xb38e8c17e38363af6ebdcb3dae12e0243582891d
		"0xb38e8c17e38363af6ebdcb3dae12e0243582891d",
		// Contract start block
		3_057_700,
	)),

	// Arbitrum Sepolia Testnet
	evm.NewEVMServiceQoSConfig("arb-sepolia-testnet", "0x66EEE", evm.NewEVMArchivalCheckConfig(
		// https://sepolia.arbiscan.io/address/0x22b65d0b9b59af4d3ed59f18b9ad53f5f4908b54
		"0x22b65d0b9b59af4d3ed59f18b9ad53f5f4908b54",
		// Contract start block
		132_000_000,
	)),

	// Avalanche
	evm.NewEVMServiceQoSConfig("avax", "0xa86a", evm.NewEVMArchivalCheckConfig(
		// https://avascan.info/blockchain/c/address/0x9f8c163cBA728e99993ABe7495F06c0A3c8Ac8b9
		"0x9f8c163cBA728e99993ABe7495F06c0A3c8Ac8b9",
		// Contract start block
		5_000_000,
	)),

	// Avalanche-DFK
	evm.NewEVMServiceQoSConfig("avax-dfk", "0xd2af", evm.NewEVMArchivalCheckConfig(
		// https://avascan.info/blockchain/dfk/address/0xCCb93dABD71c8Dad03Fc4CE5559dC3D89F67a260
		"0xCCb93dABD71c8Dad03Fc4CE5559dC3D89F67a260",
		// Contract start block
		45_000_000,
	)),

	// Base
	evm.NewEVMServiceQoSConfig("base", "0x2105", evm.NewEVMArchivalCheckConfig(
		// https://basescan.org/address/0x3304e22ddaa22bcdc5fca2269b418046ae7b566a
		"0x3304E22DDaa22bCdC5fCa2269b418046aE7b566A",
		// Contract start block
		4_504_400,
	)),

	// Base Sepolia Testnet
	evm.NewEVMServiceQoSConfig("base-sepolia-testnet", "0x14a34", evm.NewEVMArchivalCheckConfig(
		// https://sepolia.basescan.org/address/0xbab76e4365a2dff89ddb2d3fc9994103b48886c0
		"0xbab76e4365a2dff89ddb2d3fc9994103b48886c0",
		// Contract start block
		13_000_000,
	)),

	// Berachain
	evm.NewEVMServiceQoSConfig("bera", "0x138de", evm.NewEVMArchivalCheckConfig(
		// https://berascan.com/address/0x6969696969696969696969696969696969696969
		"0x6969696969696969696969696969696969696969",
		// Contract start block
		2_000_000,
	)),

	// Blast
	evm.NewEVMServiceQoSConfig("blast", "0x13e31", evm.NewEVMArchivalCheckConfig(
		// https://blastscan.io/address/0x4300000000000000000000000000000000000004
		"0x4300000000000000000000000000000000000004",
		// Contract start block
		1_000_000,
	)),

	// BNB Smart Chain
	evm.NewEVMServiceQoSConfig("bsc", "0x38", evm.NewEVMArchivalCheckConfig(
		// https://bsctrace.com/address/0xfb50526f49894b78541b776f5aaefe43e3bd8590
		"0xfb50526f49894b78541b776f5aaefe43e3bd8590",
		// Contract start block
		33_049_200,
	)),

	// Boba
	evm.NewEVMServiceQoSConfig("boba", "0x120", evm.NewEVMArchivalCheckConfig(
		// https://bobascan.com/address/0x3A92cA39476fF84Dc579C868D4D7dE125513B034
		"0x3A92cA39476fF84Dc579C868D4D7dE125513B034",
		// Contract start block
		3_060_300,
	)),

	// Celo
	evm.NewEVMServiceQoSConfig("celo", "0xa4ec", evm.NewEVMArchivalCheckConfig(
		// https://celo.blockscout.com/address/0xf89d7b9c864f589bbF53a82105107622B35EaA40
		"0xf89d7b9c864f589bbF53a82105107622B35EaA40",
		// Contract start block
		20_000_000,
	)),

	// Ethereum
	evm.NewEVMServiceQoSConfig("eth", defaultEVMChainID, evm.NewEVMArchivalCheckConfig(
		// https://etherscan.io/address/0x28C6c06298d514Db089934071355E5743bf21d60
		"0x28C6c06298d514Db089934071355E5743bf21d60",
		// Contract start block
		12_300_000,
	)),

	// Ethereum Holesky Testnet
	evm.NewEVMServiceQoSConfig("eth-holesky-testnet", "0x4268", evm.NewEVMArchivalCheckConfig(
		// https://holesky.etherscan.io/address/0xc6392ad8a14794ea57d237d12017e7295bea2363
		"0xc6392ad8a14794ea57d237d12017e7295bea2363",
		// Contract start block
		1_900_384,
	)),

	// Ethereum Sepolia Testnet
	evm.NewEVMServiceQoSConfig("eth-sepolia-testnet", "0xaa36a7", evm.NewEVMArchivalCheckConfig(
		// https://sepolia.etherscan.io/address/0xc0f3833b7e7216eecd9f6bc2c7927a7aa36ab58b
		"0xc0f3833b7e7216eecd9f6bc2c7927a7aa36ab58b",
		// Contract start block
		6_412_177,
	)),

	// Fantom
	evm.NewEVMServiceQoSConfig("fantom", "0xfa", evm.NewEVMArchivalCheckConfig(
		// https://explorer.fantom.network/address/0xaabf86ab3646a7064aa2f61e5959e39129ca46b6
		"0xaabf86ab3646a7064aa2f61e5959e39129ca46b6",
		// Contract start block
		110_633_000,
	)),

	// Fuse
	evm.NewEVMServiceQoSConfig("fuse", "0x7a", evm.NewEVMArchivalCheckConfig(
		// https://explorer.fuse.io/address/0x3014ca10b91cb3D0AD85fEf7A3Cb95BCAc9c0f79
		"0x3014ca10b91cb3D0AD85fEf7A3Cb95BCAc9c0f79",
		// Contract start block
		15_000_000,
	)),

	// Gnosis
	evm.NewEVMServiceQoSConfig("gnosis", "0x64", evm.NewEVMArchivalCheckConfig(
		// https://gnosisscan.io/address/0xe91d153e0b41518a2ce8dd3d7944fa863463a97d
		"0xe91d153e0b41518a2ce8dd3d7944fa863463a97d",
		// Contract start block
		20_000_000,
	)),

	// Harmony-0
	evm.NewEVMServiceQoSConfig("harmony", "0x63564c40", evm.NewEVMArchivalCheckConfig(
		// https://explorer.harmony.one/address/one19senwle0ezp3he6ed9xkc7zeg5rs94r0ecpp0a?shard=0
		"one19senwle0ezp3he6ed9xkc7zeg5rs94r0ecpp0a",
		// Contract start block
		60_000_000,
	)),

	// Ink
	evm.NewEVMServiceQoSConfig("ink", "0xdef1", evm.NewEVMArchivalCheckConfig(
		// https://explorer.inkonchain.com/address/0x4200000000000000000000000000000000000006
		"0x4200000000000000000000000000000000000006",
		// Contract start block
		4_500_000,
	)),

	// IoTeX
	evm.NewEVMServiceQoSConfig("iotex", "0x1251", evm.NewEVMArchivalCheckConfig(
		// https://iotexscan.io/address/0x0a7f9ea31ca689f346e1661cf73a47c69d4bd883#transactions
		"0x0a7f9ea31ca689f346e1661cf73a47c69d4bd883",
		// Contract start block
		6_440_916,
	)),

	// Kaia
	evm.NewEVMServiceQoSConfig("kaia", "0x2019", evm.NewEVMArchivalCheckConfig(
		// https://www.kaiascan.io/address/0x0051ef9259c7ec0644a80e866ab748a2f30841b3
		"0x0051ef9259c7ec0644a80e866ab748a2f30841b3",
		// Contract start block
		170_000_000,
	)),

	// Linea
	evm.NewEVMServiceQoSConfig("linea", "0xe708", evm.NewEVMArchivalCheckConfig(
		// https://lineascan.build/address/0xf89d7b9c864f589bbf53a82105107622b35eaa40
		"0xf89d7b9c864f589bbf53a82105107622b35eaa40",
		// Contract start block
		10_000_000,
	)),

	// Mantle
	evm.NewEVMServiceQoSConfig("mantle", "0x1388", evm.NewEVMArchivalCheckConfig(
		// https://explorer.mantle.xyz/address/0x588846213A30fd36244e0ae0eBB2374516dA836C
		"0x588846213A30fd36244e0ae0eBB2374516dA836C",
		// Contract start block
		60_000_000,
	)),

	// Metis
	evm.NewEVMServiceQoSConfig("metis", "0x440", evm.NewEVMArchivalCheckConfig(
		// https://explorer.metis.io/address/0xfad31cd4d45Ac7C4B5aC6A0044AA05Ca7C017e62
		"0xfad31cd4d45Ac7C4B5aC6A0044AA05Ca7C017e62",
		// Contract start block
		15_000_000,
	)),

	// Moonbeam
	evm.NewEVMServiceQoSConfig("moonbeam", "0x504", evm.NewEVMArchivalCheckConfig(
		// https://moonscan.io/address/0xf89d7b9c864f589bbf53a82105107622b35eaa40
		"0xf89d7b9c864f589bbf53a82105107622b35eaa40",
		// Contract start block
		677_000,
	)),

	// Oasys
	evm.NewEVMServiceQoSConfig("oasys", "0xf8", evm.NewEVMArchivalCheckConfig(
		// https://explorer.oasys.games/address/0xf89d7b9c864f589bbF53a82105107622B35EaA40
		"0xf89d7b9c864f589bbF53a82105107622B35EaA40",
		// Contract start block
		424_300,
	)),

	// Optimism
	evm.NewEVMServiceQoSConfig("op", "0xa", evm.NewEVMArchivalCheckConfig(
		// https://optimistic.etherscan.io/address/0xacd03d601e5bb1b275bb94076ff46ed9d753435a
		"0xacD03D601e5bB1B275Bb94076fF46ED9D753435A",
		// Contract start block
		8_121_800,
	)),

	// Optimism Sepolia Testnet
	evm.NewEVMServiceQoSConfig("op-sepolia-testnet", "0xAA37DC", evm.NewEVMArchivalCheckConfig(
		// https://sepolia-optimism.etherscan.io/address/0x734d539a7efee15714a2755caa4280e12ef3d7e4
		"0x734d539a7efee15714a2755caa4280e12ef3d7e4",
		// Contract start block
		18_241_388,
	)),

	// Polygon
	evm.NewEVMServiceQoSConfig("poly", "0x89", evm.NewEVMArchivalCheckConfig(
		// https://polygonscan.com/address/0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270
		"0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270",
		// Contract start block
		5_000_000,
	)),

	// Polygon Amoy Testnet
	evm.NewEVMServiceQoSConfig("poly-amoy-testnet", "0x13882", evm.NewEVMArchivalCheckConfig(
		// https://amoy.polygonscan.com/address/0x54d03ec0c462e9a01f77579c090cde0fc2617817
		"0x54d03ec0c462e9a01f77579c090cde0fc2617817",
		// Contract start block
		10_453_569,
	)),

	// Polygon zkEVM
	evm.NewEVMServiceQoSConfig("poly-zkevm", "0x44d", evm.NewEVMArchivalCheckConfig(
		// https://zkevm.polygonscan.com/address/0xee1727f5074e747716637e1776b7f7c7133f16b1
		"0xee1727f5074E747716637e1776B7F7C7133f16b1",
		// Contract start block
		111,
	)),

	// Scroll
	evm.NewEVMServiceQoSConfig("scroll", "0x82750", evm.NewEVMArchivalCheckConfig(
		// https://scrollscan.com/address/0x5300000000000000000000000000000000000004
		"0x5300000000000000000000000000000000000004",
		// Contract start block
		5_000_000,
	)),

	// Sonic
	evm.NewEVMServiceQoSConfig("sonic", "0x92", evm.NewEVMArchivalCheckConfig(
		// https://sonicscan.org/address/0xfc00face00000000000000000000000000000000
		"0xfc00face00000000000000000000000000000000",
		// Contract start block
		10_769_279,
	)),

	// Taiko
	evm.NewEVMServiceQoSConfig("taiko", "0x28c58", evm.NewEVMArchivalCheckConfig(
		// https://taikoscan.io/address/0x1670000000000000000000000000000000000001
		"0x1670000000000000000000000000000000000001",
		// Contract start block
		170_163,
	)),

	// Taiko Hekla Testnet
	evm.NewEVMServiceQoSConfig("taiko-hekla-testnet", "0x28c61", evm.NewEVMArchivalCheckConfig(
		// https://hekla.taikoscan.io/address/0x1670090000000000000000000000000000010001
		"0x1670090000000000000000000000000000010001",
		// Contract start block
		420_139,
	)),
	// zkLink
	evm.NewEVMServiceQoSConfig("zklink-nova", "0xc5cc4", evm.NewEVMArchivalCheckConfig(
		// https://explorer.zklink.io/address/0xa3cb8648d12bD36e713af27D92968B370D7A9546
		"0xa3cb8648d12bD36e713af27D92968B370D7A9546",
		// Contract start block
		5_004_627,
	)),

	// zkSync
	evm.NewEVMServiceQoSConfig("zksync-era", "0x144", evm.NewEVMArchivalCheckConfig(
		// https://explorer.zksync.io/address/0x03AC0b1b952C643d66A4Dc1fBc75118109cC074C
		"0x03AC0b1b952C643d66A4Dc1fBc75118109cC074C",
		// Contract start block
		55_405_668,
	)),

	// *** EVM Services (testing) ***

	// Anvil - Ethereum development/testing
	evm.NewEVMServiceQoSConfig("anvil", "0x7a69", nil),

	// Anvil WebSockets - Ethereum WebSockets development/testing
	evm.NewEVMServiceQoSConfig("anvilws", "0x7a69", nil),

	// *** EVM Services (Non-Archival) ***

	// Evmos
	evm.NewEVMServiceQoSConfig("evmos", "0x2329", nil),

	// Fraxtal
	evm.NewEVMServiceQoSConfig("fraxtal", "0xfc", nil),

	// Kava
	evm.NewEVMServiceQoSConfig("kava", "0x8ae", nil),

	// Moonriver
	evm.NewEVMServiceQoSConfig("moonriver", "0x505", nil),

	// opBNB
	evm.NewEVMServiceQoSConfig("opbnb", "0xcc", nil),

	// Radix
	evm.NewEVMServiceQoSConfig("radix", "0x1337", nil),

	// Sui
	evm.NewEVMServiceQoSConfig("sui", "0x101", nil),

	// XRPL EVM Devnet
	evm.NewEVMServiceQoSConfig("xrpl_evm_dev", "0x15f902", nil),

	// TRON
	evm.NewEVMServiceQoSConfig("tron", "0x2b6653dc", nil),

	// Sei
	evm.NewEVMServiceQoSConfig("sei", "0x531", nil),

	// *** Near EVM Services ***

	// Near
	// TODO_TECHDEBT: Add support for Near QoS
	// near.NewNearServiceQoSConfig("near", "0x18d", nil),

	// *** CometBFT Services ***

	// TODO_MVP(@commoddity): Ensure that QoS observations are being applied correctly and that
	// the correct chain ID is being used for each service in the CometBFT config.

	// Osmosis
	cosmos.NewCosmosSDKServiceQoSConfig("osmosis", "osmosis"),

	// *** Cosmos SDK Services ***

	// Pocket Mainnet and Beta Testnet
	cosmos.NewCosmosSDKServiceQoSConfig("pocket", "pocket"),

	// Pocket Mainnet
	cosmos.NewCosmosSDKServiceQoSConfig("pocket-alpha", "pocket-alpha"),
	cosmos.NewCosmosSDKServiceQoSConfig("pocket-beta", "pocket-beta"),

	// Pocket Beta Testnet
	cosmos.NewCosmosSDKServiceQoSConfig("pocket-beta1", "pocket-beta1"),
	cosmos.NewCosmosSDKServiceQoSConfig("pocket-beta2", "pocket-beta2"),
	cosmos.NewCosmosSDKServiceQoSConfig("pocket-beta3", "pocket-beta3"),
	cosmos.NewCosmosSDKServiceQoSConfig("pocket-beta4", "pocket-beta4"),
	cosmos.NewCosmosSDKServiceQoSConfig("pocket-beta5", "pocket-beta5"),

	// Cosmos Hub
	cosmos.NewCosmosSDKServiceQoSConfig("cometbft", "cosmoshub-4"),

	// XRPL EVM Testnet
	cosmos.NewCosmosSDKServiceQoSConfig("xrplevm-testnet", "xrplevm_1449000-1"),
	// TODO_IN_THIS_PR(@commoddity): Remove this once the Relay Miner changes are completed.
	cosmos.NewCosmosSDKServiceQoSConfig("xrplevm-testnet-dev", "xrplevm_1449000-1"),

	// *** Solana Services ***

	// Solana
	solana.NewSolanaServiceQoSConfig("solana", "solana"),
}

// morseServices is the list of QoS service configs for the Morse protocol.
var morseServices = []ServiceQoSConfig{
	// *** EVM Services (Archival) ***

	// Arbitrum One
	evm.NewEVMServiceQoSConfig("F001", "0xa4b1", evm.NewEVMArchivalCheckConfig(
		// https://arbiscan.io/address/0xb38e8c17e38363af6ebdcb3dae12e0243582891d
		"0xb38e8c17e38363af6ebdcb3dae12e0243582891d",
		// Contract start block
		3_100_000,
	)),

	// Arbitrum Sepolia Testnet
	evm.NewEVMServiceQoSConfig("F002", "0x66EEE", evm.NewEVMArchivalCheckConfig(
		// https://sepolia.arbiscan.io/address/0x22b65d0b9b59af4d3ed59f18b9ad53f5f4908b54
		"0x22b65d0b9b59af4d3ed59f18b9ad53f5f4908b54",
		// Contract start block
		132_000_000,
	)),

	// Avalanche
	evm.NewEVMServiceQoSConfig("F003", "0xa86a", evm.NewEVMArchivalCheckConfig(
		// https://avascan.info/blockchain/c/address/0x9f8c163cBA728e99993ABe7495F06c0A3c8Ac8b9
		"0x9f8c163cBA728e99993ABe7495F06c0A3c8Ac8b9",
		// Contract start block
		5_000_000,
	)),

	// Avalanche-DFK
	evm.NewEVMServiceQoSConfig("F004", "0xd2af", evm.NewEVMArchivalCheckConfig(
		// https://avascan.info/blockchain/dfk/address/0xCCb93dABD71c8Dad03Fc4CE5559dC3D89F67a260
		"0xCCb93dABD71c8Dad03Fc4CE5559dC3D89F67a260",
		// Contract start block
		45_000_000,
	)),

	// Base
	evm.NewEVMServiceQoSConfig("F005", "0x2105", evm.NewEVMArchivalCheckConfig(
		// https://basescan.org/address/0x3304e22ddaa22bcdc5fca2269b418046ae7b566a
		"0x3304E22DDaa22bCdC5fCa2269b418046aE7b566A",
		// Contract start block
		4_504_400,
	)),

	// Base Sepolia Testnet
	evm.NewEVMServiceQoSConfig("F006", "0x14a34", evm.NewEVMArchivalCheckConfig(
		// https://sepolia.basescan.org/address/0xbab76e4365a2dff89ddb2d3fc9994103b48886c0
		"0xbab76e4365a2dff89ddb2d3fc9994103b48886c0",
		// Contract start block
		13_000_000,
	)),

	// Berachain
	evm.NewEVMServiceQoSConfig("F035", "0x138de", evm.NewEVMArchivalCheckConfig(
		// https://berascan.com/address/0x6969696969696969696969696969696969696969
		"0x6969696969696969696969696969696969696969",
		// Contract start block
		2_000_000,
	)),

	// Blast
	evm.NewEVMServiceQoSConfig("F008", "0x13e31", evm.NewEVMArchivalCheckConfig(
		// https://blastscan.io/address/0x4300000000000000000000000000000000000004
		"0x4300000000000000000000000000000000000004",
		// Contract start block
		1_000_000,
	)),

	// BNB Smart Chain
	evm.NewEVMServiceQoSConfig("F009", "0x38", evm.NewEVMArchivalCheckConfig(
		// https://bsctrace.com/address/0xfb50526f49894b78541b776f5aaefe43e3bd8590
		"0xfb50526f49894b78541b776f5aaefe43e3bd8590",
		// Contract start block
		33_049_200,
	)),

	// Boba
	evm.NewEVMServiceQoSConfig("F00A", "0x120", evm.NewEVMArchivalCheckConfig(
		// https://bobascan.com/address/0x3A92cA39476fF84Dc579C868D4D7dE125513B034
		"0x3A92cA39476fF84Dc579C868D4D7dE125513B034",
		// Contract start block
		3_060_300,
	)),

	// Celo
	evm.NewEVMServiceQoSConfig("F00B", "0xa4ec", evm.NewEVMArchivalCheckConfig(
		// https://celo.blockscout.com/address/0xf89d7b9c864f589bbF53a82105107622B35EaA40
		"0xf89d7b9c864f589bbF53a82105107622B35EaA40",
		// Contract start block
		20_000_000,
	)),

	// Ethereum
	evm.NewEVMServiceQoSConfig("F00C", defaultEVMChainID, evm.NewEVMArchivalCheckConfig(
		// https://etherscan.io/address/0x28C6c06298d514Db089934071355E5743bf21d60
		"0x28C6c06298d514Db089934071355E5743bf21d60",
		// Contract start block
		12_300_000,
	)),

	// Ethereum Holesky Testnet
	evm.NewEVMServiceQoSConfig("F00D", "0x4268", evm.NewEVMArchivalCheckConfig(
		// https://holesky.etherscan.io/address/0xc6392ad8a14794ea57d237d12017e7295bea2363
		"0xc6392ad8a14794ea57d237d12017e7295bea2363",
		// Contract start block
		1_900_384,
	)),

	// Ethereum Sepolia Testnet
	evm.NewEVMServiceQoSConfig("F00E", "0xaa36a7", evm.NewEVMArchivalCheckConfig(
		// https://sepolia.etherscan.io/address/0xc0f3833b7e7216eecd9f6bc2c7927a7aa36ab58b
		"0xc0f3833b7e7216eecd9f6bc2c7927a7aa36ab58b",
		// Contract start block
		6_412_177,
	)),

	// Fuse
	evm.NewEVMServiceQoSConfig("F012", "0x7a", evm.NewEVMArchivalCheckConfig(
		// https://explorer.fuse.io/address/0x3014ca10b91cb3D0AD85fEf7A3Cb95BCAc9c0f79
		"0x3014ca10b91cb3D0AD85fEf7A3Cb95BCAc9c0f79",
		// Contract start block
		15_000_000,
	)),

	// Gnosis
	evm.NewEVMServiceQoSConfig("F013", "0x64", evm.NewEVMArchivalCheckConfig(
		// https://gnosisscan.io/address/0xe91d153e0b41518a2ce8dd3d7944fa863463a97d
		"0xe91d153e0b41518a2ce8dd3d7944fa863463a97d",
		// Contract start block
		20_000_000,
	)),

	// Harmony-0
	evm.NewEVMServiceQoSConfig("F014", "0x63564c40", evm.NewEVMArchivalCheckConfig(
		// https://explorer.harmony.one/address/one19senwle0ezp3he6ed9xkc7zeg5rs94r0ecpp0a?shard=0
		"one19senwle0ezp3he6ed9xkc7zeg5rs94r0ecpp0a",
		// Contract start block
		60_000_000,
	)),

	// Ink
	evm.NewEVMServiceQoSConfig("F032", "0xdef1", evm.NewEVMArchivalCheckConfig(
		// https://explorer.inkonchain.com/address/0x4200000000000000000000000000000000000006
		"0x4200000000000000000000000000000000000006",
		// Contract start block
		4_500_000,
	)),

	// IoTeX
	evm.NewEVMServiceQoSConfig("F015", "0x1251", evm.NewEVMArchivalCheckConfig(
		// https://iotexscan.io/address/0x0a7f9ea31ca689f346e1661cf73a47c69d4bd883#transactions
		"0x0a7f9ea31ca689f346e1661cf73a47c69d4bd883",
		// Contract start block
		6_440_916,
	)),

	// Kaia
	evm.NewEVMServiceQoSConfig("F016", "0x2019", evm.NewEVMArchivalCheckConfig(
		// https://www.kaiascan.io/address/0x0051ef9259c7ec0644a80e866ab748a2f30841b3
		"0x0051ef9259c7ec0644a80e866ab748a2f30841b3",
		// Contract start block
		170_000_000,
	)),

	// Linea
	evm.NewEVMServiceQoSConfig("F030", "0xe708", evm.NewEVMArchivalCheckConfig(
		// https://lineascan.build/address/0xf89d7b9c864f589bbf53a82105107622b35eaa40
		"0xf89d7b9c864f589bbf53a82105107622b35eaa40",
		// Contract start block
		10_000_000,
	)),

	// Mantle
	evm.NewEVMServiceQoSConfig("F033", "0x1388", evm.NewEVMArchivalCheckConfig(
		// https://explorer.mantle.xyz/address/0x588846213A30fd36244e0ae0eBB2374516dA836C
		"0x588846213A30fd36244e0ae0eBB2374516dA836C",
		// Contract start block
		60_000_000,
	)),

	// Metis
	evm.NewEVMServiceQoSConfig("F018", "0x440", evm.NewEVMArchivalCheckConfig(
		// https://explorer.metis.io/address/0xfad31cd4d45Ac7C4B5aC6A0044AA05Ca7C017e62
		"0xfad31cd4d45Ac7C4B5aC6A0044AA05Ca7C017e62",
		// Contract start block
		15_000_000,
	)),

	// Moonbeam
	evm.NewEVMServiceQoSConfig("F019", "0x504", evm.NewEVMArchivalCheckConfig(
		// https://moonscan.io/address/0xf89d7b9c864f589bbf53a82105107622b35eaa40
		"0xf89d7b9c864f589bbf53a82105107622b35eaa40",
		// Contract start block
		677_000,
	)),

	// Oasys
	evm.NewEVMServiceQoSConfig("F01C", "0xf8", evm.NewEVMArchivalCheckConfig(
		// https://explorer.oasys.games/address/0xf89d7b9c864f589bbF53a82105107622B35EaA40
		"0xf89d7b9c864f589bbF53a82105107622B35EaA40",
		// Contract start block
		424_300,
	)),

	// Optimism
	evm.NewEVMServiceQoSConfig("F01D", "0xa", evm.NewEVMArchivalCheckConfig(
		// https://optimistic.etherscan.io/address/0xacd03d601e5bb1b275bb94076ff46ed9d753435a
		"0xacD03D601e5bB1B275Bb94076fF46ED9D753435A",
		// Contract start block
		8_121_800,
	)),

	// Optimism Sepolia Testnet
	evm.NewEVMServiceQoSConfig("F01E", "0xAA37DC", evm.NewEVMArchivalCheckConfig(
		// https://sepolia-optimism.etherscan.io/address/0x734d539a7efee15714a2755caa4280e12ef3d7e4
		"0x734d539a7efee15714a2755caa4280e12ef3d7e4",
		// Contract start block
		18_241_388,
	)),

	// opBNB
	evm.NewEVMServiceQoSConfig("F01F", "0xcc", evm.NewEVMArchivalCheckConfig(
		// https://opbnbscan.com/address/0x001ceb373c83ae75b9f5cf78fc2aba3e185d09e2
		"0x001ceb373c83ae75b9f5cf78fc2aba3e185d09e2",
		// Contract start block
		20_000_000,
	)),

	// Polygon
	evm.NewEVMServiceQoSConfig("F021", "0x89", evm.NewEVMArchivalCheckConfig(
		// https://polygonscan.com/address/0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270
		"0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270",
		// Contract start block
		5_000_000,
	)),

	// Polygon Amoy Testnet
	evm.NewEVMServiceQoSConfig("F022", "0x13882", evm.NewEVMArchivalCheckConfig(
		// https://amoy.polygonscan.com/address/0x54d03ec0c462e9a01f77579c090cde0fc2617817
		"0x54d03ec0c462e9a01f77579c090cde0fc2617817",
		// Contract start block
		10_453_569,
	)),

	// Polygon zkEVM
	evm.NewEVMServiceQoSConfig("F029", "0x44d", evm.NewEVMArchivalCheckConfig(
		// https://zkevm.polygonscan.com/address/0xee1727f5074e747716637e1776b7f7c7133f16b1
		"0xee1727f5074E747716637e1776B7F7C7133f16b1",
		// Contract start block
		111,
	)),

	// Scroll
	evm.NewEVMServiceQoSConfig("F024", "0x82750", evm.NewEVMArchivalCheckConfig(
		// https://scrollscan.com/address/0x5300000000000000000000000000000000000004
		"0x5300000000000000000000000000000000000004",
		// Contract start block
		5_000_000,
	)),

	// Sonic
	evm.NewEVMServiceQoSConfig("F02D", "0x92", evm.NewEVMArchivalCheckConfig(
		// https://sonicscan.org/address/0xfc00face00000000000000000000000000000000
		"0xfc00face00000000000000000000000000000000",
		// Contract start block
		10_769_279,
	)),

	// Taiko
	evm.NewEVMServiceQoSConfig("F027", "0x28c58", evm.NewEVMArchivalCheckConfig(
		// https://taikoscan.io/address/0x1670000000000000000000000000000000000001
		"0x1670000000000000000000000000000000000001",
		// Contract start block
		170_163,
	)),

	// Taiko Hekla Testnet
	evm.NewEVMServiceQoSConfig("F028", "0x28c61", evm.NewEVMArchivalCheckConfig(
		// https://hekla.taikoscan.io/address/0x1670090000000000000000000000000000010001
		"0x1670090000000000000000000000000000010001",
		// Contract start block
		420_139,
	)),

	// XRPL EVM Testnet
	evm.NewEVMServiceQoSConfig("F036", "0x161c28", evm.NewEVMArchivalCheckConfig(
		// https://explorer.testnet.xrplevm.org/address/0xc29e2583eD5C77df8792067989Baf9E4CCD4D7fc
		"0xc29e2583eD5C77df8792067989Baf9E4CCD4D7fc",
		// Contract start block
		368_266,
	)),

	// zkLink
	evm.NewEVMServiceQoSConfig("F02A", "0xc5cc4", evm.NewEVMArchivalCheckConfig(
		// https://explorer.zklink.io/address/0xa3cb8648d12bD36e713af27D92968B370D7A9546
		"0xa3cb8648d12bD36e713af27D92968B370D7A9546",
		// Contract start block
		5_004_627,
	)),

	// zkSync
	evm.NewEVMServiceQoSConfig("F02B", "0x144", evm.NewEVMArchivalCheckConfig(
		// https://explorer.zksync.io/address/0x03AC0b1b952C643d66A4Dc1fBc75118109cC074C
		"0x03AC0b1b952C643d66A4Dc1fBc75118109cC074C",
		// Contract start block
		55_405_668,
	)),

	// *** EVM Services (Non-Archival) ***

	// Evmos
	evm.NewEVMServiceQoSConfig("F00F", "0x2329", nil),

	// Fantom
	evm.NewEVMServiceQoSConfig("F010", "0xfa", nil),

	// Fraxtal
	evm.NewEVMServiceQoSConfig("F011", "0xfc", nil),

	// Kava
	evm.NewEVMServiceQoSConfig("F017", "0x8ae", nil),

	// Moonriver
	evm.NewEVMServiceQoSConfig("F01A", "0x505", nil),

	// Near
	evm.NewEVMServiceQoSConfig("F01B", "0x18d", nil),

	// Radix
	evm.NewEVMServiceQoSConfig("F023", "0x1337", nil),

	// Sui
	evm.NewEVMServiceQoSConfig("F026", "0x101", nil),

	// XRPL EVM Devnet
	evm.NewEVMServiceQoSConfig("F02C", "0x15f902", nil),

	// TRON
	evm.NewEVMServiceQoSConfig("F02E", "0x2b6653dc", nil),

	// Berachain Testnet
	evm.NewEVMServiceQoSConfig("F031", "0x138d4", nil),

	// Sei
	evm.NewEVMServiceQoSConfig("F034", "0x531", nil),

	// *** CometBFT Services ***
	// TODO_MVP(@commoddity): Ensure that QoS observations are being applied correctly and that
	// the correct chain ID is being used for each service in the CometBFT config.

	// Osmosis
	cosmos.NewCosmosSDKServiceQoSConfig("F020", "osmosis"),

	// *** Solana Services ***

	// Solana
	// TODO_MVP(@adshmh): Drop the Chain ID for Solana.
	solana.NewSolanaServiceQoSConfig("F025", "F025"),
}
