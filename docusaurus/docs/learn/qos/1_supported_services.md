---
sidebar_position: 1
title: Supported QoS Services
description: Supported Quality of Service Implementations in PATH
---

:::danger DO NOT EDIT

This file was auto-generated via `make gen_service_qos_docs`.

:::

## Configuring PATH QoS Checks

PATH uses an **opt-out** rather than an **opt-in** approach to QoS checks.

This means that PATH **automatically** performs QoS checks for all services the applications it manages are staked for.

### Disable QoS Checks for a particular Service

In order to disable QoS checks for a specific service, the `service_id` field may be specified in the `.config.yaml` file's `qos_disabled_service_ids` field.

For example, to disable QoS checks for the Ethereum service on a Morse PATH instance, the following configuration would be added to the `.config.yaml` file:

```yaml
hydrator_config:
  qos_disabled_service_ids:
    - "F00C"
```

See [PATH Configuration File](../../develop/path/5_configurations_path.md#hydrator_config-optional) for more details.

## ⛓️ Supported QoS Services

The table below lists the Quality of Service (QoS) implementations currently supported by PATH.

:::warning **🚧 QoS Support 🚧**

If a Service ID is not specified in the tables below, it does not have a QoS implementation in PATH.

:::

## 🌿 Current PATH QoS Support

**🗓️ Document Last Updated: 2025-05-26**

## Shannon Protocol Services

| Service Name | Authoritative Service ID | Service QoS Type | Chain ID (if applicable) | Archival Check Configured |
|-------------|------------|-----------------|----------|---------------------------|
| Arbitrum One | arb_one | EVM | 42161 | ✅ |
| Arbitrum Sepolia Testnet | arb_sep_test | EVM | 421614 | ✅ |
| Avalanche | avax | EVM | 43114 | ✅ |
| Avalanche-DFK | avax-dfk | EVM | 53935 | ✅ |
| Base | base | EVM | 8453 | ✅ |
| Base Sepolia Testnet | base-test | EVM | 84532 | ✅ |
| Berachain | bera | EVM | 80094 | ✅ |
| Blast | blast | EVM | 81457 | ✅ |
| BNB Smart Chain | bsc | EVM | 56 | ✅ |
| Boba | boba | EVM | 288 | ✅ |
| Celo | celo | EVM | 42220 | ✅ |
| Ethereum - ETH Mainnet | eth | EVM | 1 | ✅ |
| Fuse | fuse | EVM | 122 | ✅ |
| Gnosis | gnosis | EVM | 100 | ✅ |
| Harmony-0 | harmony | EVM | 1666600000 | ✅ |
| Ink | ink | EVM | 57073 | ✅ |
| IoTeX | iotex | EVM | 4689 | ✅ |
| Kaia | kaia | EVM | 8217 | ✅ |
| Linea | linea | EVM | 59144 | ✅ |
| Mantle | mantle | EVM | 5000 | ✅ |
| Metis | metis | EVM | 1088 | ✅ |
| Moonbeam | moonbeam | EVM | 1284 | ✅ |
| Oasys | oasys | EVM | 248 | ✅ |
| Optimism | op | EVM | 10 | ✅ |
| Polygon | poly | EVM | 137 | ✅ |
| Polygon zkEVM | poly_zkevm | EVM | 1101 | ✅ |
| Scroll | scroll | EVM | 534352 | ✅ |
| Sonic | sonic | EVM | 146 | ✅ |
| Taiko | taiko | EVM | 167000 | ✅ |
| XRPL EVM Testnet | xrpl_evm_testnet | EVM | 1449000 | ✅ |
| Anvil - Ethereum development/testing | anvil | EVM | 31337 |  |
| Anvil WebSockets - Ethereum WebSockets development/testing | anvilws | EVM | 31337 |  |
| Ethereum Holesky Testnet | eth_hol_test | EVM | 17000 |  |
| Ethereum Sepolia Testnet | eth_sep_test | EVM | 11155111 |  |
| Evmos | evmos | EVM | 9001 |  |
| Fantom | fantom | EVM | 250 |  |
| Fraxtal | fraxtal | EVM | 252 |  |
| Kava | kava | EVM | 2222 |  |
| Moonriver | moonriver | EVM | 1285 |  |
| Near | near | EVM | 397 |  |
| Optimism Sepolia Testnet | op_sep_test | EVM | 11155420 |  |
| opBNB | opbnb | EVM | 204 |  |
| Polygon Amoy Testnet | poly_amoy_test | EVM | 80002 |  |
| Radix | radix | EVM | 4919 |  |
| Sui | sui | EVM | 257 |  |
| Taiko Hekla Testnet | taiko_hek_test | EVM | 167009 |  |
| zkLink | zklink_nova | EVM | 810180 |  |
| zkSync | zksync_era | EVM | 324 |  |
| XRPL EVM Devnet | xrpl_evm_dev | EVM | 1440002 |  |
| TRON | tron | EVM | 728126428 |  |
| Sei | sei | EVM | 1329 |  |
| Celestia Archival | tia_da | CometBFT | celestia-archival |  |
| Celestia Consensus Archival | tia_cons | CometBFT | celestia-consensus-archival |  |
| Celestia Testnet DA Archival | tia_da_test | CometBFT | celestia-testnet-da-archival |  |
| Celestia Testnet Consensus Archival | tia_cons_test | CometBFT | celestia-testnet-consensus-archival |  |
| Osmosis | osmosis | CometBFT | osmosis |  |
| Pocket Beta Testnet | pocket-beta-rpc | CometBFT | pocket-beta |  |
| Cosmos Hub | cometbft | CometBFT | cosmoshub-4 |  |
| Solana | solana | Solana |  |  |

## Morse Protocol Services

| Service Name | Authoritative Service ID | Service QoS Type | Chain ID (if applicable) | Archival Check Configured |
|-------------|------------|-----------------|----------|---------------------------|
| Arbitrum One | F001 | EVM | 42161 | ✅ |
| Arbitrum Sepolia Testnet | F002 | EVM | 421614 | ✅ |
| Avalanche | F003 | EVM | 43114 | ✅ |
| Avalanche-DFK | F004 | EVM | 53935 | ✅ |
| Base | F005 | EVM | 8453 | ✅ |
| Base Sepolia Testnet | F006 | EVM | 84532 | ✅ |
| Berachain | F035 | EVM | 80094 | ✅ |
| Blast | F008 | EVM | 81457 | ✅ |
| BNB Smart Chain | F009 | EVM | 56 | ✅ |
| Boba | F00A | EVM | 288 | ✅ |
| Celo | F00B | EVM | 42220 | ✅ |
| Ethereum | F00C | EVM | 1 | ✅ |
| Fuse | F012 | EVM | 122 | ✅ |
| Gnosis | F013 | EVM | 100 | ✅ |
| Harmony-0 | F014 | EVM | 1666600000 | ✅ |
| Ink | F032 | EVM | 57073 | ✅ |
| IoTeX | F015 | EVM | 4689 | ✅ |
| Kaia | F016 | EVM | 8217 | ✅ |
| Linea | F030 | EVM | 59144 | ✅ |
| Mantle | F033 | EVM | 5000 | ✅ |
| Metis | F018 | EVM | 1088 | ✅ |
| Moonbeam | F019 | EVM | 1284 | ✅ |
| Oasys | F01C | EVM | 248 | ✅ |
| Optimism | F01D | EVM | 10 | ✅ |
| opBNB | F01F | EVM | 204 | ✅ |
| Polygon | F021 | EVM | 137 | ✅ |
| Polygon zkEVM | F029 | EVM | 1101 | ✅ |
| Scroll | F024 | EVM | 534352 | ✅ |
| Sonic | F02D | EVM | 146 | ✅ |
| Taiko | F027 | EVM | 167000 | ✅ |
| XRPL EVM Testnet | F036 | EVM | 1449000 | ✅ |
| Ethereum Holesky Testnet | F00D | EVM | 17000 |  |
| Ethereum Sepolia Testnet | F00E | EVM | 11155111 |  |
| Evmos | F00F | EVM | 9001 |  |
| Fantom | F010 | EVM | 250 |  |
| Fraxtal | F011 | EVM | 252 |  |
| Kava | F017 | EVM | 2222 |  |
| Moonriver | F01A | EVM | 1285 |  |
| Near | F01B | EVM | 397 |  |
| Optimism Sepolia Testnet | F01E | EVM | 11155420 |  |
| Polygon Amoy Testnet | F022 | EVM | 80002 |  |
| Radix | F023 | EVM | 4919 |  |
| Sui | F026 | EVM | 257 |  |
| Taiko Hekla Testnet | F028 | EVM | 167009 |  |
| zkLink | F02A | EVM | 810180 |  |
| zkSync | F02B | EVM | 324 |  |
| XRPL EVM Devnet | F02C | EVM | 1440002 |  |
| TRON | F02E | EVM | 728126428 |  |
| Berachain Testnet | F031 | EVM | 80084 |  |
| Sei | F034 | EVM | 1329 |  |
| Celestia Archival | A0CA | CometBFT | celestia-archival |  |
| Celestia Consensus Archival | A0CB | CometBFT | celestia-consensus-archival |  |
| Celestia Testnet DA Archival | A0CC | CometBFT | celestia-testnet-da-archival |  |
| Celestia Testnet Consensus Archival | A0CD | CometBFT | celestia-testnet-consensus-archival |  |
| Osmosis | F020 | CometBFT | osmosis |  |
| TODO_MVP(@adshmh): Drop the Chain ID for Solana. | F025 | Solana |  |  |
