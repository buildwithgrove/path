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

For example, to disable QoS checks for the Ethereum service on a Shannon PATH instance, the following configuration would be added to the `.config.yaml` file:

```yaml
hydrator_config:
  qos_disabled_service_ids:
    - "eth"
```

See [PATH Configuration File](../../develop/path/5_configurations_path.md#hydrator_config-optional) for more details.

## ⛓️ Supported QoS Services

The table below lists the Quality of Service (QoS) implementations currently supported by PATH.

:::warning **🚧 QoS Support 🚧**

If a Service ID is not specified in the tables below, it does not have a QoS implementation in PATH.

:::

## 🌿 Current PATH QoS Support

**🗓️ Document Last Updated: 2025-06-03**

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
| Ethereum | eth | EVM | 1 | ✅ |
| Ethereum Holesky Testnet | eth_hol_test | EVM | 17000 | ✅ |
| Ethereum Sepolia Testnet | eth_sep_test | EVM | 11155111 | ✅ |
| Fantom | fantom | EVM | 250 | ✅ |
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
| Optimism Sepolia Testnet | op_sep_test | EVM | 11155420 | ✅ |
| Polygon | poly | EVM | 137 | ✅ |
| Polygon Amoy Testnet | poly_amoy_test | EVM | 80002 | ✅ |
| Polygon zkEVM | poly_zkevm | EVM | 1101 | ✅ |
| Scroll | scroll | EVM | 534352 | ✅ |
| Sonic | sonic | EVM | 146 | ✅ |
| Taiko | taiko | EVM | 167000 | ✅ |
| Taiko Hekla Testnet | taiko_hek_test | EVM | 167009 | ✅ |
| XRPL EVM Testnet | xrpl_evm_testnet | EVM | 1449000 | ✅ |
| zkLink | zklink_nova | EVM | 810180 | ✅ |
| zkSync | zksync_era | EVM | 324 | ✅ |
| Anvil - Ethereum development/testing | anvil | EVM | 31337 |  |
| Anvil WebSockets - Ethereum WebSockets development/testing | anvilws | EVM | 31337 |  |
| Evmos | evmos | EVM | 9001 |  |
| Fraxtal | fraxtal | EVM | 252 |  |
| Kava | kava | EVM | 2222 |  |
| Moonriver | moonriver | EVM | 1285 |  |
| Near | near | EVM | 397 |  |
| opBNB | opbnb | EVM | 204 |  |
| Radix | radix | EVM | 4919 |  |
| Sui | sui | EVM | 257 |  |
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
