package config

import (
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/cometbft"
	"github.com/buildwithgrove/path/qos/evm"
	"github.com/buildwithgrove/path/qos/solana"
)

// TODO_DOCUMENT(@commoddity): Add a README to [path docs](https://path.grove.city/) for developers.
// Consider a similar automated approach to "docs_update_gov_params_page"

// NOTE: Service ID list last updated 2025/03/27

// IMPORTANT: PATH requires service IDs to be registered here for Quality of Service (QoS) endpoint checks.
// Unregistered services use NoOp QoS type with random endpoint selection and no monitoring.

type ServiceConfig interface {
	GetServiceID() protocol.ServiceID
	GetServiceQoSType() string
}

type qosServiceConfigs struct {
	shannonServices []ServiceConfig
	morseServices   []ServiceConfig
}

// GetServiceConfigs returns the service configs for the provided protocol.
func (c qosServiceConfigs) GetServiceConfigs(config GatewayConfig) []ServiceConfig {
	if shannonConfig := config.GetShannonConfig(); shannonConfig != nil {
		return c.shannonServices
	}
	if morseConfig := config.GetMorseConfig(); morseConfig != nil {
		return c.morseServices
	}
	return nil
}

// The ServiceConfigs map associates each supported service ID with a specific
// implementation of the gateway.QoSService interface.
// This is to handle requests for a given service ID.
var ServiceConfigs = qosServiceConfigs{
	shannonServices: shannonServices,
	morseServices:   morseServices,
}

const (
	defaultEVMChainID      = "0x1" // ETH Mainnet (1)
	defaultCometBFTChainID = "cosmoshub-4"
)

// shannonServices is the list of QoS service configs for the Shannon protocol.
var shannonServices = []ServiceConfig{
	evm.ServiceConfig{
		ServiceID:  "eth",             // Ethereum
		EVMChainID: defaultEVMChainID, // (1)
	},
	evm.ServiceConfig{
		ServiceID:  "anvil",           // Anvil (Ethereum development/testing)
		EVMChainID: defaultEVMChainID, // (1)
	},
	evm.ServiceConfig{
		ServiceID:  "anvilws",         // Anvil WebSockets (Ethereum WebSockets development/testing)
		EVMChainID: defaultEVMChainID, // (1)
	},
	cometbft.ServiceConfig{
		ServiceID:       "pocket-beta-rpc", // Pocket Beta Testnet
		CometBFTChainID: "pocket-beta",
	},
	cometbft.ServiceConfig{
		ServiceID:       "cometbft",             // CometBFT (Cosmos Hub)
		CometBFTChainID: defaultCometBFTChainID, // Cosmos Hub
	},
	solana.ServiceConfig{
		ServiceID: "solana", // Solana
	},
}

// TODO_IN_THIS_PR(@commoddity): Add archival check configurations for all EVM services.
// This means setting the following fields:
//   - Enabled
//   - ContractAddress
//   - ContractStartBlock

// morseServices is the list of QoS service configs for the Morse protocol.
var morseServices = []ServiceConfig{
	evm.ServiceConfig{
		ServiceID:  "F001",   // Arbitrum One
		EVMChainID: "0xa4b1", // (42161)
	},
	evm.ServiceConfig{
		ServiceID:  "F002",    // Arbitrum Sepolia Testnet
		EVMChainID: "0x66EEE", // (421614)
	},
	evm.ServiceConfig{
		ServiceID:  "F003",   // Avalanche
		EVMChainID: "0xa86a", // (43114)
	},
	evm.ServiceConfig{
		ServiceID:  "F004",   // Avalanche-DFK
		EVMChainID: "0xd2af", // (53935)
	},
	evm.ServiceConfig{
		ServiceID:  "F005",   // Base
		EVMChainID: "0x2105", // (8453)
	},
	evm.ServiceConfig{
		ServiceID:  "F006",    // Base Sepolia Testnet
		EVMChainID: "0x14a34", // (84660)
	},
	evm.ServiceConfig{
		ServiceID:  "F008",    // Blast
		EVMChainID: "0x13e31", // (81649)
	},
	evm.ServiceConfig{
		ServiceID:  "F009", // BNB Smart Chain
		EVMChainID: "0x38", // (56)
	},
	evm.ServiceConfig{
		ServiceID:  "F00A",  // Boba
		EVMChainID: "0x120", // (288)
	},
	evm.ServiceConfig{
		ServiceID:  "F00B",   // Celo
		EVMChainID: "0xa4ec", // (42220)
	},
	evm.ServiceConfig{
		ServiceID:  "F00C",            // Ethereum
		EVMChainID: defaultEVMChainID, // (1)
		ArchivalCheckConfig: evm.EVMArchivalCheckConfig{
			Enabled:            true,
			ContractAddress:    "0x28C6c06298d514Db089934071355E5743bf21d60",
			ContractStartBlock: 12_300_000,
		},
	},
	evm.ServiceConfig{
		ServiceID:  "F00D",   // Ethereum Holesky Testnet
		EVMChainID: "0x4268", // (17000)
	},
	evm.ServiceConfig{
		ServiceID:  "F00E",     // Ethereum Sepolia Testnet
		EVMChainID: "0xaa36a7", // (11155420)
	},
	evm.ServiceConfig{
		ServiceID:  "F00F",   // Evmos
		EVMChainID: "0x2329", // (9001)
	},
	evm.ServiceConfig{
		ServiceID:  "F010", // Fantom
		EVMChainID: "0xfa", // (250)
	},
	evm.ServiceConfig{
		ServiceID:  "F011", // Fraxtal
		EVMChainID: "0xfc", // (252)
	},
	evm.ServiceConfig{
		ServiceID:  "F012", // Fuse
		EVMChainID: "0x7a", // (122)
	},
	evm.ServiceConfig{
		ServiceID:  "F013", // Gnosis
		EVMChainID: "0x64", // (100)
	},
	evm.ServiceConfig{
		ServiceID:  "F014",       // Harmony-0
		EVMChainID: "0x63564c40", // (1666600000)
	},
	evm.ServiceConfig{
		ServiceID:  "F015",   // IoTeX
		EVMChainID: "0x1251", // (4681)
	},
	evm.ServiceConfig{
		ServiceID:  "F016",   // Kaia
		EVMChainID: "0x2019", // (8217)
	},
	evm.ServiceConfig{
		ServiceID:  "F017",  // Kava
		EVMChainID: "0x8ae", // (2222)
	},
	evm.ServiceConfig{
		ServiceID:  "F018",  // Metis
		EVMChainID: "0x440", // (1088)
	},
	evm.ServiceConfig{
		ServiceID:  "F019",  // Moonbeam
		EVMChainID: "0x504", // (1284)
	},
	evm.ServiceConfig{
		ServiceID:  "F01A",  // Moonriver
		EVMChainID: "0x505", // (1285)
	},
	evm.ServiceConfig{
		ServiceID:  "F01C", // Oasys
		EVMChainID: "0xf8", // (248)
	},
	evm.ServiceConfig{
		ServiceID:  "F01D", // Optimism
		EVMChainID: "0xa",  // (10)
	},
	evm.ServiceConfig{
		ServiceID:  "F01E",     // Optimism Sepolia Testnet
		EVMChainID: "0xAA37DC", // (11155420)
	},
	evm.ServiceConfig{
		ServiceID:  "F01F", // opBNB
		EVMChainID: "0xcc", // (204)
	},
	evm.ServiceConfig{
		ServiceID:  "F021", // Polygon
		EVMChainID: "0x89", // (137)
		ArchivalCheckConfig: evm.EVMArchivalCheckConfig{
			Enabled:            true,
			ContractAddress:    "0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270",
			ContractStartBlock: 5_000_000,
		},
	},
	evm.ServiceConfig{
		ServiceID:  "F022",    // Polygon Amoy Testnet
		EVMChainID: "0x13882", // (80002)
	},
	evm.ServiceConfig{
		ServiceID:  "F024",    // Scroll
		EVMChainID: "0x82750", // (534992)
	},
	evm.ServiceConfig{
		ServiceID:  "F027",    // Taiko
		EVMChainID: "0x28c58", // (167000)
	},
	evm.ServiceConfig{
		ServiceID:  "F028",    // Taiko Hekla Testnet
		EVMChainID: "0x28c61", // (167009)
	},
	evm.ServiceConfig{
		ServiceID:  "F029",  // Polygon zkEVM
		EVMChainID: "0x44d", // (1101)
	},
	evm.ServiceConfig{
		ServiceID:  "F02A",    // zkLink
		EVMChainID: "0xc5cc4", // (812564)
	},
	evm.ServiceConfig{
		ServiceID:  "F02B",  // zkSync
		EVMChainID: "0x144", // (324)
	},
	evm.ServiceConfig{
		ServiceID:  "F02C",     // XRPL EVM Devnet
		EVMChainID: "0x15f902", // (1440002)
	},
	evm.ServiceConfig{
		ServiceID:  "F036",     // XRPL EVM Testnet
		EVMChainID: "0x161c28", // (1449000)
	},
	evm.ServiceConfig{
		ServiceID:  "F02D", // Sonic
		EVMChainID: "0x92", // (146)
	},
	evm.ServiceConfig{
		ServiceID:  "F02E",       // TRON
		EVMChainID: "0x2b6653dc", // (728426128)
	},
	evm.ServiceConfig{
		ServiceID:  "F030",   // Linea
		EVMChainID: "0xe708", // (59144)
	},
	evm.ServiceConfig{
		ServiceID:  "F031",    // Berachain bArtio Testnet
		EVMChainID: "0x138d4", // (80084)
	},
	evm.ServiceConfig{
		ServiceID:  "F032",   // Ink
		EVMChainID: "0xdef1", // (57073)
	},
	evm.ServiceConfig{
		ServiceID:  "F033",   // Mantle
		EVMChainID: "0x1388", // (5000)
	},
	evm.ServiceConfig{
		ServiceID:  "F034",  // Sei
		EVMChainID: "0x531", // (1329)
	},
	evm.ServiceConfig{
		ServiceID:  "F035",    // Berachain
		EVMChainID: "0x138de", // (80094)
	},
	solana.ServiceConfig{
		ServiceID: "solana", // Solana
	},
}
