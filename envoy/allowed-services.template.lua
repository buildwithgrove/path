-- IMPORTANT: All Services for the PATH Service Gateway must be listed here for Envoy Proxy to forward requests to PATH.
-- The service IDs configured here are used in the `envoy.filters.http.lua` HTTP filter defined in `.envoy.yaml` config file.
-- The `.allowed-services.lua` file must be mounted as a file in the Envoy Proxy container at `/etc/envoy/.allowed-services.lua`.
--
-- If you wish to define aliases for existing services, you must define the alias as the key and the service ID as the value.
--
-- eg 1. the service ID ["F000"] = "F000" enables Envoy to forward requests with the subdomain "F000.path.grove.city" to PATH with the service ID "F000".
-- eg 2. the alias ["pocket"] = "F000" enables Envoy to forward requests with the subdomain "pocket.path.grove.city" to PATH with the service ID "F000".
--
-- To utilize PATH's Quality of Service (QoS) features, the service ID must match the value in PATH's `qos` module.
-- TODO_IMPROVE(@commoddity): Add link to the file & line in the QoS module once 'no-op' QoS feature is completed.
return {
  -- 1. Shannon Service IDs
  ["anvil"] = "anvil", -- Anvil (Authoritative ID)

  -- 2. Morse Service IDs
  ["F000"] = "F000",                     -- Pocket (Authoritative ID)
  ["pocket"] = "F000",                   -- Pocket (Alias)

  ["F001"] = "F001",                     -- Arbitrum One (Authoritative ID)
  ["arbitrum-one"] = "F001",             -- Arbitrum One (Alias)

  ["F002"] = "F002",                     -- Arbitrum Sepolia Testnet (Authoritative ID)
  ["arbitrum-sepolia-testnet"] = "F002", -- Arbitrum Sepolia Testnet (Alias)

  ["F003"] = "F003",                     -- AVAX (Authoritative ID)
  ["avax"] = "F003",                     -- AVAX (Alias)

  ["F004"] = "F004",                     -- AVAX-DFK (Authoritative ID)
  ["avax-dfk"] = "F004",                 -- AVAX-DFK (Alias)

  ["F005"] = "F005",                     -- Base (Authoritative ID)
  ["base"] = "F005",                     -- Base (Alias)

  ["F006"] = "F006",                     -- Base Testnet (Authoritative ID)
  ["base-testnet"] = "F006",             -- Base Testnet (Alias)

  ["F008"] = "F008",                     -- Blast (Authoritative ID)
  ["blast"] = "F008",                    -- Blast (Alias)

  ["F009"] = "F009",                     -- Binance Smart Chain (Authoritative ID)
  ["bsc"] = "F009",                      -- Binance Smart Chain (Alias)

  ["F00A"] = "F00A",                     -- Boba (Authoritative ID)
  ["boba"] = "F00A",                     -- Boba (Alias)

  ["F00B"] = "F00B",                     -- Celo (Authoritative ID)
  ["celo"] = "F00B",                     -- Celo (Alias)

  ["F00C"] = "F00C",                     -- Ethereum (Authoritative ID)
  ["eth"] = "F00C",                      -- Ethereum (Alias)

  ["F00D"] = "F00D",                     -- Ethereum Holesky Testnet (Authoritative ID)
  ["eth-holesky-testnet"] = "F00D",      -- Ethereum Holesky Testnet (Alias)

  ["F00E"] = "F00E",                     -- Ethereum Sepolia Testnet (Authoritative ID)
  ["sepolia"] = "F00E",                  -- Ethereum Sepolia Testnet (Alias)

  ["F00F"] = "F00F",                     -- Evmos (Authoritative ID)
  ["evmos"] = "F00F",                    -- Evmos (Alias)

  ["F010"] = "F010",                     -- Fantom (Authoritative ID)
  ["fantom"] = "F010",                   -- Fantom (Alias)

  ["F011"] = "F011",                     -- Fraxtal (Authoritative ID)
  ["fraxtal"] = "F011",                  -- Fraxtal (Alias)

  ["F012"] = "F012",                     -- Fuse (Authoritative ID)
  ["fuse"] = "F012",                     -- Fuse (Alias)

  ["F013"] = "F013",                     -- Gnosis (Authoritative ID)
  ["gnosis"] = "F013",                   -- Gnosis (Alias)

  ["F014"] = "F014",                     -- Harmony (Authoritative ID)
  ["harmony"] = "F014",                  -- Harmony (Alias)

  ["F015"] = "F015",                     -- Iotex (Authoritative ID)
  ["iotex"] = "F015",                    -- Iotex (Alias)

  ["F016"] = "F016",                     -- Kaia (Authoritative ID)
  ["kaia"] = "F016",                     -- Kaia (Alias)

  ["F017"] = "F017",                     -- Kava (Authoritative ID)
  ["kava"] = "F017",                     -- Kava (Alias)

  ["F018"] = "F018",                     -- Metis (Authoritative ID)
  ["metis"] = "F018",                    -- Metis (Alias)

  ["F019"] = "F019",                     -- Moonbeam (Authoritative ID)
  ["moonbeam"] = "F019",                 -- Moonbeam (Alias)

  ["F01A"] = "F01A",                     -- Moonriver (Authoritative ID)
  ["moonriver"] = "F01A",                -- Moonriver (Alias)

  ["F01B"] = "F01B",                     -- Near (Authoritative ID)
  ["near"] = "F01B",                     -- Near (Alias)

  ["F01C"] = "F01C",                     -- Oasys (Authoritative ID)
  ["oasys"] = "F01C",                    -- Oasys (Alias)

  ["F01D"] = "F01D",                     -- Optimism (Authoritative ID)
  ["optimism"] = "F01D",                 -- Optimism (Alias)

  ["F01E"] = "F01E",                     -- Optimism Sepolia Testnet (Authoritative ID)
  ["optimism-sepolia-testnet"] = "F01E", -- Optimism Sepolia Testnet (Alias)

  ["F01F"] = "F01F",                     -- Opbnb (Authoritative ID)
  ["opbnb"] = "F01F",                    -- Opbnb (Alias)

  ["F020"] = "F020",                     -- Osmosis (Authoritative ID)
  ["osmosis"] = "F020",                  -- Osmosis (Alias)

  ["F021"] = "F021",                     -- Polygon (Authoritative ID)
  ["polygon"] = "F021",                  -- Polygon (Alias)

  ["F022"] = "F022",                     -- Polygon Amoy Testnet (Authoritative ID)
  ["polygon-amoy-testnet"] = "F022",     -- Polygon Amoy Testnet (Alias)

  ["F023"] = "F023",                     -- Radix (Authoritative ID)
  ["radix"] = "F023",                    -- Radix (Alias)

  ["F024"] = "F024",                     -- Scroll (Authoritative ID)
  ["scroll"] = "F024",                   -- Scroll (Alias)

  ["F025"] = "F025",                     -- Solana (Authoritative ID)
  ["solana"] = "F025",                   -- Solana (Alias)

  ["F026"] = "F026",                     -- Sui (Authoritative ID)
  ["sui"] = "F026",                      -- Sui (Alias)

  ["F027"] = "F027",                     -- Taiko (Authoritative ID)
  ["taiko"] = "F027",                    -- Taiko (Alias)

  ["F028"] = "F028",                     -- Taiko Hekla Testnet (Authoritative ID)
  ["taiko-hekla-testnet"] = "F028",      -- Taiko Hekla Testnet (Alias)

  ["F029"] = "F029",                     -- Polygon Zkevm (Authoritative ID)
  ["polygon-zkevm"] = "F029",            -- Polygon Zkevm (Alias)

  ["F02A"] = "F02A",                     -- Zklink Nova (Authoritative ID)
  ["zklink-nova"] = "F02A",              -- Zklink Nova (Alias)

  ["F02B"] = "F02B",                     -- Zksync Era (Authoritative ID)
  ["zksync-era"] = "F02B",               -- Zksync Era (Alias)
}
