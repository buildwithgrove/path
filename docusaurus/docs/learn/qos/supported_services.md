---
sidebar_position: 2
title: Quality of Service
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

**🗓️ Document Last Updated: 2025-04-01**

## Shannon Protocol Services

| Service Name                                               | Authoritative Service ID | Service QoS Type | Chain ID (if applicable) | Archival Check Configured |
| ---------------------------------------------------------- | ------------------------ | ---------------- | ------------------------ | ------------------------- |
| Ethereum                                                   | eth                      | EVM              | 1                        |                           |
| Anvil (Ethereum development/testing)                       | anvil                    | EVM              | 1                        |                           |
| Anvil WebSockets (Ethereum WebSockets development/testing) | anvilws                  | EVM              | 1                        |                           |
| Pocket Beta Testnet                                        | pocket-beta-rpc          | CometBFT         | pocket-beta              |                           |
| CometBFT (Cosmos Hub)                                      | cometbft                 | CometBFT         | cosmoshub-4              |                           |
| Solana                                                     | solana                   | Solana           |                          |                           |

## Morse Protocol Services

| Service Name             | Authoritative Service ID | Service QoS Type | Chain ID (if applicable) | Archival Check Configured |
| ------------------------ | ------------------------ | ---------------- | ------------------------ | ------------------------- |
| Arbitrum One             | F001                     | EVM              | 42161                    |                           |
| Arbitrum Sepolia Testnet | F002                     | EVM              | 421614                   |                           |
| Avalanche                | F003                     | EVM              | 43114                    |                           |
| Avalanche-DFK            | F004                     | EVM              | 53935                    |                           |
| Base                     | F005                     | EVM              | 8453                     |                           |
| Base Sepolia Testnet     | F006                     | EVM              | 84660                    |                           |
| Blast                    | F008                     | EVM              | 81649                    |                           |
| BNB Smart Chain          | F009                     | EVM              | 56                       |                           |
| Boba                     | F00A                     | EVM              | 288                      |                           |
| Celo                     | F00B                     | EVM              | 42220                    |                           |
| Ethereum                 | F00C                     | EVM              | 1                        | ✅                         |
| Ethereum Holesky Testnet | F00D                     | EVM              | 17000                    |                           |
| Ethereum Sepolia Testnet | F00E                     | EVM              | 11155420                 |                           |
| Evmos                    | F00F                     | EVM              | 9001                     |                           |
| Fantom                   | F010                     | EVM              | 250                      |                           |
| Fraxtal                  | F011                     | EVM              | 252                      |                           |
| Fuse                     | F012                     | EVM              | 122                      |                           |
| Gnosis                   | F013                     | EVM              | 100                      |                           |
| Harmony-0                | F014                     | EVM              | 1666600000               |                           |
| IoTeX                    | F015                     | EVM              | 4681                     |                           |
| Kaia                     | F016                     | EVM              | 8217                     |                           |
| Kava                     | F017                     | EVM              | 2222                     |                           |
| Metis                    | F018                     | EVM              | 1088                     |                           |
| Moonbeam                 | F019                     | EVM              | 1284                     |                           |
| Moonriver                | F01A                     | EVM              | 1285                     |                           |
| Oasys                    | F01C                     | EVM              | 248                      | ✅                         |
| Optimism                 | F01D                     | EVM              | 10                       |                           |
| Optimism Sepolia Testnet | F01E                     | EVM              | 11155420                 |                           |
| opBNB                    | F01F                     | EVM              | 204                      |                           |
| Polygon                  | F021                     | EVM              | 137                      | ✅                         |
| Polygon Amoy Testnet     | F022                     | EVM              | 80002                    |                           |
| Scroll                   | F024                     | EVM              | 534992                   |                           |
| Taiko                    | F027                     | EVM              | 167000                   |                           |
| Taiko Hekla Testnet      | F028                     | EVM              | 167009                   |                           |
| Polygon zkEVM            | F029                     | EVM              | 1101                     |                           |
| zkLink                   | F02A                     | EVM              | 812564                   |                           |
| zkSync                   | F02B                     | EVM              | 324                      |                           |
| XRPL EVM Devnet          | F02C                     | EVM              | 1440002                  |                           |
| XRPL EVM Testnet         | F036                     | EVM              | 1449000                  | ✅                         |
| Sonic                    | F02D                     | EVM              | 146                      |                           |
| TRON                     | F02E                     | EVM              | 728426128                |                           |
| Linea                    | F030                     | EVM              | 59144                    |                           |
| Berachain bArtio Testnet | F031                     | EVM              | 80084                    |                           |
| Ink                      | F032                     | EVM              | 57073                    |                           |
| Mantle                   | F033                     | EVM              | 5000                     |                           |
| Sei                      | F034                     | EVM              | 1329                     |                           |
| Berachain                | F035                     | EVM              | 80094                    |                           |
| Solana                   | solana                   | Solana           |                          |                           |
