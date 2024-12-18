-- IMPORTANT: All Services for the PATH Service Gateway must be listed here for Envoy Proxy to forward requests to PATH.
--
-- If you wish to define aliases for existing services, you must define the alias as the key and the service ID as the value.
--
-- eg. the alias "eth" = F00C" enables the URL "http://eth.path.grove.city" to be routed to the service with the ID "F00C".
--
-- To utilize PATH's Quality of Service (QoS) features, the service ID must match the value in PATH's `qos` module.
-- TODO_IMPROVE(@commoddity): Add link to the file & line in the QoS module once 'no-op' QoS feature is completed.
return {
    -- 1. Shannon Service IDs
    ["anvil"] = "anvil", -- Anvil (Authoritative ID)

    -- 2. Morse Service IDs
    ["F000"] = "F000",                      -- Pocket (Authoritative ID)
    ["mainnet"] = "F000",                   -- Pocket (Alias)
    ["mainnet-archival"] = "F000",          -- Pocket (Alias)
    ["pocket"] = "F000",                    -- Pocket (Alias)
    ["pokt-archival"] = "F000",             -- Pocket (Alias)

    ["F001"] = "F001",                      -- Arbitrum-one (Authoritative ID)
    ["arbitrum-one"] = "F001",              -- Arbitrum-one (Alias)

    ["F002"] = "F002",                      -- Arbitrum-sepolia-testnet (Authoritative ID)
    ["arbitrum-sepolia-archival"] = "F002", -- Arbitrum-sepolia-testnet (Alias)
    ["arbitrum-sepolia-testnet"] = "F002",  -- Arbitrum-sepolia-testnet (Alias)

    ["F003"] = "F003",                      -- Avax (Authoritative ID)
    ["avax"] = "F003",                      -- Avax (Alias)
    ["avax-archival"] = "F003",             -- Avax (Alias)
    ["avax-mainnet"] = "F003",              -- Avax (Alias)

    ["F004"] = "F004",                      -- Avax-dfk (Authoritative ID)
    ["avax-dfk"] = "F004",                  -- Avax-dfk (Alias)

    ["F005"] = "F005",                      -- Base (Authoritative ID)
    ["base"] = "F005",                      -- Base (Alias)
    ["base-mainnet"] = "F005",              -- Base (Alias)
    ["base-mainnet-archival"] = "F005",     -- Base (Alias)

    ["F006"] = "F006",                      -- Base-testnet (Authoritative ID)
    ["base-testnet"] = "F006",              -- Base-testnet (Alias)
    ["base-testnet-archival"] = "F006",     -- Base-testnet (Alias)

    ["F008"] = "F008",                      -- Blast (Authoritative ID)
    ["blast"] = "F008",                     -- Blast (Alias)
    ["blast-archival"] = "F008",            -- Blast (Alias)
    ["blast-mainnet"] = "F008",             -- Blast (Alias)

    ["F009"] = "F009",                      -- Bsc (Authoritative ID)
    ["bsc"] = "F009",                       -- Bsc (Alias)
    ["bsc-archival"] = "F009",              -- Bsc (Alias)
    ["bsc-mainnet"] = "F009",               -- Bsc (Alias)

    ["F00A"] = "F00A",                      -- Boba (Authoritative ID)
    ["boba"] = "F00A",                      -- Boba (Alias)
    ["boba-mainnet"] = "F00A",              -- Boba (Alias)

    ["F00B"] = "F00B",                      -- Celo (Authoritative ID)
    ["celo"] = "F00B",                      -- Celo (Alias)
    ["celo-mainnet"] = "F00B",              -- Celo (Alias)

    ["F00C"] = "F00C",                      -- Eth (Authoritative ID)
    ["eth"] = "F00C",                       -- Eth (Alias)
    ["eth-archival"] = "F00C",              -- Eth (Alias)
    ["eth-mainnet"] = "F00C",               -- Eth (Alias)
    ["eth-trace"] = "F00C",                 -- Eth (Alias)

    ["F00D"] = "F00D",                      -- Eth-holesky-testnet (Authoritative ID)
    ["eth-holesky-testnet"] = "F00D",       -- Eth-holesky-testnet (Alias)
    ["holesky-fullnode-testnet"] = "F00D",  -- Eth-holesky-testnet (Alias)

    ["F00E"] = "F00E",                      -- Eth-sepolia-testnet (Authoritative ID)
    ["eth-sepolia-testnet"] = "F00E",       -- Eth-sepolia-testnet (Alias)
    ["sepolia"] = "F00E",                   -- Eth-sepolia-testnet (Alias)
    ["sepolia-archival"] = "F00E",          -- Eth-sepolia-testnet (Alias)

    ["F00F"] = "F00F",                      -- Evmos (Authoritative ID)
    ["evmos"] = "F00F",                     -- Evmos (Alias)
    ["evmos-mainnet"] = "F00F",             -- Evmos (Alias)

    ["F010"] = "F010",                      -- Fantom (Authoritative ID)
    ["fantom"] = "F010",                    -- Fantom (Alias)
    ["fantom-mainnet"] = "F010",            -- Fantom (Alias)

    ["F011"] = "F011",                      -- Fraxtal (Authoritative ID)
    ["fraxtal"] = "F011",                   -- Fraxtal (Alias)
    ["fraxtal-archival"] = "F011",          -- Fraxtal (Alias)

    ["F012"] = "F012",                      -- Fuse (Authoritative ID)
    ["fuse"] = "F012",                      -- Fuse (Alias)
    ["fuse-archival"] = "F012",             -- Fuse (Alias)
    ["fuse-mainnet"] = "F012",              -- Fuse (Alias)

    ["F013"] = "F013",                      -- Gnosis (Authoritative ID)
    ["gnosis"] = "F013",                    -- Gnosis (Alias)
    ["gnosischain-archival"] = "F013",      -- Gnosis (Alias)
    ["gnosischain-mainnet"] = "F013",       -- Gnosis (Alias)
    ["poa-xdai"] = "F013",                  -- Gnosis (Alias)

    ["F014"] = "F014",                      -- Harmony (Authoritative ID)
    ["harmony"] = "F014",                   -- Harmony (Alias)
    ["harmony-0"] = "F014",                 -- Harmony (Alias)

    ["F015"] = "F015",                      -- Iotex (Authoritative ID)
    ["iotex"] = "F015",                     -- Iotex (Alias)
    ["iotex-mainnet"] = "F015",             -- Iotex (Alias)

    ["F016"] = "F016",                      -- Kaia (Authoritative ID)
    ["kaia"] = "F016",                      -- Kaia (Alias)
    ["kaia-mainnet"] = "F016",              -- Kaia (Alias)
    ["klaytn-mainnet"] = "F016",            -- Kaia (Alias)

    ["F017"] = "F017",                      -- Kava (Authoritative ID)
    ["kava"] = "F017",                      -- Kava (Alias)
    ["kava-mainnet"] = "F017",              -- Kava (Alias)
    ["kava-mainnet-archival"] = "F017",     -- Kava (Alias)

    ["F018"] = "F018",                      -- Metis (Authoritative ID)
    ["metis"] = "F018",                     -- Metis (Alias)
    ["metis-mainnet"] = "F018",             -- Metis (Alias)

    ["F019"] = "F019",                      -- Moonbeam (Authoritative ID)
    ["moonbeam"] = "F019",                  -- Moonbeam (Alias)
    ["moonbeam-mainnet"] = "F019",          -- Moonbeam (Alias)

    ["F01A"] = "F01A",                      -- Moonriver (Authoritative ID)
    ["moonriver"] = "F01A",                 -- Moonriver (Alias)
    ["moonriver-mainnet"] = "F01A",         -- Moonriver (Alias)

    ["F01B"] = "F01B",                      -- Near (Authoritative ID)
    ["near"] = "F01B",                      -- Near (Alias)
    ["near-mainnet"] = "F01B",              -- Near (Alias)

    ["F01C"] = "F01C",                      -- Oasys (Authoritative ID)
    ["oasys"] = "F01C",                     -- Oasys (Alias)
    ["oasys-mainnet"] = "F01C",             -- Oasys (Alias)
    ["oasys-mainnet-archival"] = "F01C",    -- Oasys (Alias)

    ["F01D"] = "F01D",                      -- Optimism (Authoritative ID)
    ["optimism"] = "F01D",                  -- Optimism (Alias)
    ["optimism-archival"] = "F01D",         -- Optimism (Alias)
    ["optimism-mainnet"] = "F01D",          -- Optimism (Alias)

    ["F01E"] = "F01E",                      -- Optimism-sepolia-testnet (Authoritative ID)
    ["optimism-sepolia-archival"] = "F01E", -- Optimism-sepolia-testnet (Alias)
    ["optimism-sepolia-testnet"] = "F01E",  -- Optimism-sepolia-testnet (Alias)

    ["F01F"] = "F01F",                      -- Opbnb (Authoritative ID)
    ["opbnb"] = "F01F",                     -- Opbnb (Alias)
    ["opbnb-archival"] = "F01F",            -- Opbnb (Alias)

    ["F020"] = "F020",                      -- Osmosis (Authoritative ID)
    ["osmosis"] = "F020",                   -- Osmosis (Alias)
    ["osmosis-mainnet"] = "F020",           -- Osmosis (Alias)

    ["F021"] = "F021",                      -- Polygon (Authoritative ID)
    ["poly-archival"] = "F021",             -- Polygon (Alias)
    ["polygon"] = "F021",                   -- Polygon (Alias)
    ["poly-mainnet"] = "F021",              -- Polygon (Alias)

    ["F022"] = "F022",                      -- Polygon-amoy-testnet (Authoritative ID)
    ["amoy-testnet-archival"] = "F022",     -- Polygon-amoy-testnet (Alias)
    ["polygon-amoy-testnet"] = "F022",      -- Polygon-amoy-testnet (Alias)

    ["F023"] = "F023",                      -- Radix (Authoritative ID)
    ["radix"] = "F023",                     -- Radix (Alias)
    ["radix-mainnet"] = "F023",             -- Radix (Alias)

    ["F024"] = "F024",                      -- Scroll (Authoritative ID)
    ["scroll"] = "F024",                    -- Scroll (Alias)

    ["F025"] = "F025",                      -- Solana (Authoritative ID)
    ["solana"] = "F025",                    -- Solana (Alias)
    ["solana-mainnet"] = "F025",            -- Solana (Alias)
    ["solana-mainnet-custom"] = "F025",     -- Solana (Alias)

    ["F026"] = "F026",                      -- Sui (Authoritative ID)
    ["sui"] = "F026",                       -- Sui (Alias)
    ["sui-mainnet"] = "F026",               -- Sui (Alias)

    ["F027"] = "F027",                      -- Taiko (Authoritative ID)
    ["taiko"] = "F027",                     -- Taiko (Alias)

    ["F028"] = "F028",                      -- Taiko-hekla-testnet (Authoritative ID)
    ["taiko-hekla-testnet"] = "F028",       -- Taiko-hekla-testnet (Alias)

    ["F029"] = "F029",                      -- Polygon-zkevm (Authoritative ID)
    ["polygon-zkevm"] = "F029",             -- Polygon-zkevm (Alias)
    ["polygon-zkevm-mainnet"] = "F029",     -- Polygon-zkevm (Alias)
    ["zkevm-polygon-mainnet"] = "F029",     -- Polygon-zkevm (Alias)

    ["F02A"] = "F02A",                      -- Zklink-nova (Authoritative ID)
    ["zklink-nova"] = "F02A",               -- Zklink-nova (Alias)
    ["zklink-nova-archival"] = "F02A",      -- Zklink-nova (Alias)

    ["F02B"] = "F02B",                      -- Zksync-era (Authoritative ID)
    ["zksync-era"] = "F02B",                -- Zksync-era (Alias)
    ["zksync-era-mainnet"] = "F02B",        -- Zksync-era (Alias)
}
