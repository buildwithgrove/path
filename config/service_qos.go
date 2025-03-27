package config

// TODO_DOCUMENT(@commoddity): Add a README to [path docs](https://path.grove.city/) for developers.
// Consider a similar automated approach to "docs_update_gov_params_page"

// NOTE: Service ID list last updated 2025/03/27

// IMPORTANT: PATH requires service IDs to be registered here for Quality of Service (QoS) endpoint checks.
// Unregistered services use NoOp QoS type with random endpoint selection and no monitoring.

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

// shannonServices is the list of QoS service configs for the Shannon protocol.
var shannonServices = []ServiceConfig{
	EVMServiceConfig{
		serviceID:  "eth",             // Ethereum
		evmChainID: defaultEVMChainID, // (1)
	},
	EVMServiceConfig{
		serviceID:  "anvil",           // Anvil (Ethereum development/testing)
		evmChainID: defaultEVMChainID, // (1)
	},
	EVMServiceConfig{
		serviceID:  "anvilws",         // Anvil WebSockets (Ethereum WebSockets development/testing)
		evmChainID: defaultEVMChainID, // (1)
	},
	CometBFTServiceConfig{
		serviceID:       "pocket-beta-rpc", // Pocket Beta Testnet
		cometBFTChainID: "pocket-beta",
	},
	CometBFTServiceConfig{
		serviceID:       "cometbft",             // CometBFT (Cosmos Hub)
		cometBFTChainID: defaultCometBFTChainID, // Cosmos Hub
	},
	SolanaServiceConfig{
		serviceID: "solana", // Solana
	},
}

// morseServices is the list of QoS service configs for the Morse protocol.
var morseServices = []ServiceConfig{
	EVMServiceConfig{
		serviceID:  "F001",   // Arbitrum One
		evmChainID: "0xa4b1", // (42161)
	},
	EVMServiceConfig{
		serviceID:  "F002",    // Arbitrum Sepolia Testnet
		evmChainID: "0x66EEE", // (421614)
	},
	EVMServiceConfig{
		serviceID:  "F003",   // Avalanche
		evmChainID: "0xa86a", // (43114)
	},
	EVMServiceConfig{
		serviceID:  "F004",   // Avalanche-DFK
		evmChainID: "0xd2af", // (53935)
	},
	EVMServiceConfig{
		serviceID:  "F005",   // Base
		evmChainID: "0x2105", // (8453)
	},
	EVMServiceConfig{
		serviceID:  "F006",    // Base Sepolia Testnet
		evmChainID: "0x14a34", // (84660)
	},
	EVMServiceConfig{
		serviceID:  "F008",    // Blast
		evmChainID: "0x13e31", // (81649)
	},
	EVMServiceConfig{
		serviceID:  "F009", // BNB Smart Chain
		evmChainID: "0x38", // (56)
	},
	EVMServiceConfig{
		serviceID:  "F00A",  // Boba
		evmChainID: "0x120", // (288)
	},
	EVMServiceConfig{
		serviceID:  "F00B",   // Celo
		evmChainID: "0xa4ec", // (42220)
	},
	EVMServiceConfig{
		serviceID:  "F00C",            // Ethereum
		evmChainID: defaultEVMChainID, // (1)
	},
	EVMServiceConfig{
		serviceID:  "F00D",   // Ethereum Holesky Testnet
		evmChainID: "0x4268", // (17000)
	},
	EVMServiceConfig{
		serviceID:  "F00E",     // Ethereum Sepolia Testnet
		evmChainID: "0xaa36a7", // (11155420)
	},
	EVMServiceConfig{
		serviceID:  "F00F",   // Evmos
		evmChainID: "0x2329", // (9001)
	},
	EVMServiceConfig{
		serviceID:  "F010", // Fantom
		evmChainID: "0xfa", // (250)
	},
	EVMServiceConfig{
		serviceID:  "F011", // Fraxtal
		evmChainID: "0xfc", // (252)
	},
	EVMServiceConfig{
		serviceID:  "F012", // Fuse
		evmChainID: "0x7a", // (122)
	},
	EVMServiceConfig{
		serviceID:  "F013", // Gnosis
		evmChainID: "0x64", // (100)
	},
	EVMServiceConfig{
		serviceID:  "F014",       // Harmony-0
		evmChainID: "0x63564c40", // (1666600000)
	},
	EVMServiceConfig{
		serviceID:  "F015",   // IoTeX
		evmChainID: "0x1251", // (4681)
	},
	EVMServiceConfig{
		serviceID:  "F016",   // Kaia
		evmChainID: "0x2019", // (8217)
	},
	EVMServiceConfig{
		serviceID:  "F017",  // Kava
		evmChainID: "0x8ae", // (2222)
	},
	EVMServiceConfig{
		serviceID:  "F018",  // Metis
		evmChainID: "0x440", // (1088)
	},
	EVMServiceConfig{
		serviceID:  "F019",  // Moonbeam
		evmChainID: "0x504", // (1284)
	},
	EVMServiceConfig{
		serviceID:  "F01A",  // Moonriver
		evmChainID: "0x505", // (1285)
	},
	EVMServiceConfig{
		serviceID:  "F01C", // Oasys
		evmChainID: "0xf8", // (248)
	},
	EVMServiceConfig{
		serviceID:  "F01D", // Optimism
		evmChainID: "0xa",  // (10)
	},
	EVMServiceConfig{
		serviceID:  "F01E",     // Optimism Sepolia Testnet
		evmChainID: "0xAA37DC", // (11155420)
	},
	EVMServiceConfig{
		serviceID:  "F01F", // opBNB
		evmChainID: "0xcc", // (204)
	},
	EVMServiceConfig{
		serviceID:  "F021", // Polygon
		evmChainID: "0x89", // (137)
	},
	EVMServiceConfig{
		serviceID:  "F022",    // Polygon Amoy Testnet
		evmChainID: "0x13882", // (80002)
	},
	EVMServiceConfig{
		serviceID:  "F024",    // Scroll
		evmChainID: "0x82750", // (534992)
	},
	EVMServiceConfig{
		serviceID:  "F027",    // Taiko
		evmChainID: "0x28c58", // (167000)
	},
	EVMServiceConfig{
		serviceID:  "F028",    // Taiko Hekla Testnet
		evmChainID: "0x28c61", // (167009)
	},
	EVMServiceConfig{
		serviceID:  "F029",  // Polygon zkEVM
		evmChainID: "0x44d", // (1101)
	},
	EVMServiceConfig{
		serviceID:  "F02A",    // zkLink
		evmChainID: "0xc5cc4", // (812564)
	},
	EVMServiceConfig{
		serviceID:  "F02B",  // zkSync
		evmChainID: "0x144", // (324)
	},
	EVMServiceConfig{
		serviceID:  "F02C",     // XRPL EVM Devnet
		evmChainID: "0x15f902", // (1440002)
	},
	EVMServiceConfig{
		serviceID:  "F036",     // XRPL EVM Testnet
		evmChainID: "0x161c28", // (1449000)
	},
	EVMServiceConfig{
		serviceID:  "F02D", // Sonic
		evmChainID: "0x92", // (146)
	},
	EVMServiceConfig{
		serviceID:  "F02E",       // TRON
		evmChainID: "0x2b6653dc", // (728426128)
	},
	EVMServiceConfig{
		serviceID:  "F030",   // Linea
		evmChainID: "0xe708", // (59144)
	},
	EVMServiceConfig{
		serviceID:  "F031",    // Berachain bArtio Testnet
		evmChainID: "0x138d4", // (80084)
	},
	EVMServiceConfig{
		serviceID:  "F032",   // Ink
		evmChainID: "0xdef1", // (57073)
	},
	EVMServiceConfig{
		serviceID:  "F033",   // Mantle
		evmChainID: "0x1388", // (5000)
	},
	EVMServiceConfig{
		serviceID:  "F034",  // Sei
		evmChainID: "0x531", // (1329)
	},
	EVMServiceConfig{
		serviceID:  "F035",    // Berachain
		evmChainID: "0x138de", // (80094)
	},
	SolanaServiceConfig{
		serviceID: "solana", // Solana
	},
}
