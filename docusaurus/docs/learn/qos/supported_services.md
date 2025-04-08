---
sidebar_position: 1
title: Supported QoS Services
description: Supported Quality of Service Implementations in PATH
---

# ⛓️ Supported QoS Services

The following table lists the Quality of Service (QoS) implementations currently supported by PATH.

:::warning 🚧 Under Construction 🚧

If a Service ID is not specified in the tables below, it means that the service does not currently have a QoS implementation in PATH. **This means no QoS checks are performed for that service and endpoints are selected at random from the available endpoints for that service.**

PATH's QoS functionality is under extremely active development and support for new services is added regularly.

:::

:::info Example Hydrator Configuration

In order to utilize automated QoS checks, the `Service ID` field must be specified in the `.config.yaml` file's `hydrator_config` section.

For example, for a Morse PATH gateway, to enable QoS checks for the Ethereum & Polygon services, the following configuration would be added to the `.config.yaml` file:

```yaml 
hydrator_config:
  service_ids:
    - "F00C"
    - "F021"
```

💡 *For more information on PATH's configuration file, please refer to the [configuration documentation](../../develop/path/configuration.md).*

:::

# 🌿 Current PATH QoS Support

**🗓️ Document Last Updated: 2025-04-07**

## Shannon Protocol Services

| Service Name | Authoritative Service ID | Service QoS Type | Chain ID (if applicable) | Archival Check Configured |
|-------------|------------|-----------------|----------|---------------------------|
| Ethereum - ETH Mainnet | eth | EVM |  | ✅ |
| Anvil WebSockets - Ethereum WebSockets development/testing | anvilws | EVM |  |  |
| CometBFT - Pocket Beta Testnet | pocket-beta-rpc | CometBFT | pocket-beta |  |
| CometBFT - Cosmos Hub | cometbft | CometBFT | cosmoshub-4 |  |
| Solana | solana | Solana |  |  |

## Morse Protocol Services

| Service Name | Authoritative Service ID | Service QoS Type | Chain ID (if applicable) | Archival Check Configured |
|-------------|------------|-----------------|----------|---------------------------|
| Arbitrum One (42161) | F001 | EVM | 42161 |  |
| Arbitrum Sepolia Testnet (421614) | F002 | EVM | 421614 |  |
| Avalanche (43114) | F003 | EVM | 43114 |  |
| Avalanche-DFK (53935) | F004 | EVM | 53935 |  |
| Base (8453) | F005 | EVM | 8453 |  |
| Base Sepolia Testnet (84660) | F006 | EVM | 84532 |  |
| Blast (81649) | F008 | EVM | 81457 |  |
| BNB Smart Chain (56) | F009 | EVM | 56 |  |
| Boba (288) | F00A | EVM | 288 |  |
| Celo (42220) | F00B | EVM | 42220 | ✅ |
| Ethereum Holesky Testnet (17000) | F00D | EVM | 17000 |  |
| Ethereum Sepolia Testnet (11155111) | F00E | EVM | 11155111 |  |
| Evmos (9001) | F00F | EVM | 9001 |  |
| Fantom (250) | F010 | EVM | 250 |  |
| Fraxtal (252) | F011 | EVM | 252 |  |
| Fuse (122) | F012 | EVM | 122 |  |
| Gnosis (100) | F013 | EVM | 100 |  |
| Harmony-0 (1666600000) | F014 | EVM | 1666600000 |  |
| IoTeX (4681) | F015 | EVM | 4689 |  |
| Kaia (8217) | F016 | EVM | 8217 |  |
| Kava (2222) | F017 | EVM | 2222 |  |
| Metis (1088) | F018 | EVM | 1088 |  |
| Moonbeam (1284) | F019 | EVM | 1284 |  |
| Moonriver (1285) | F01A | EVM | 1285 | ✅ |
| Optimism (10) | F01D | EVM | 10 |  |
| Optimism Sepolia Testnet (11155420) | F01E | EVM | 11155420 |  |
| opBNB (204) | F01F | EVM | 204 | ✅ |
| Polygon Amoy Testnet (80002) | F022 | EVM | 80002 |  |
| Scroll (534992) | F024 | EVM | 534352 |  |
| Taiko (167000) | F027 | EVM | 167000 |  |
| Taiko Hekla Testnet (167009) | F028 | EVM | 167009 |  |
| Polygon zkEVM (1101) | F029 | EVM | 1101 |  |
| zkLink (812564) | F02A | EVM | 810180 |  |
| zkSync (324) | F02B | EVM | 324 |  |
| XRPL EVM Devnet (1440002) | F02C | EVM | 1440002 | ✅ |
| Sonic (146) | F02D | EVM | 146 |  |
| TRON (728426128) | F02E | EVM | 728126428 |  |
| Linea (59144) | F030 | EVM | 59144 |  |
| Berachain Testnet (80084) | F031 | EVM | 80084 |  |
| Ink (57073) | F032 | EVM | 57073 |  |
| Mantle (5000) | F033 | EVM | 5000 |  |
| Sei (1329) | F034 | EVM | 1329 |  |
| Berachain (80094) | F035 | EVM | 80094 |  |
| Solana | solana | Solana |  |  |
