---
sidebar_position: 1
title: Supported QoS Services
description: Supported Quality of Service Implementations in PATH
---

:::danger DO NOT EDIT

This file was auto-generated via `make gen_service_qos_docs`.

:::

## ⛓️ Supported QoS Services

The following table lists the Quality of Service (QoS) implementations currently supported by PATH.

:::important 🚧 QoS Support 🚧

If a Service ID is not specified in the tables below, it does not have a QoS implementation in PATH.

**This means no QoS checks are performed for that service and endpoints are selected at random from the network.**

:::

### Example Hydrator Configuration

In order to utilize automated QoS checks, the `Service ID` field must be specified in the `.config.yaml` file's `hydrator_config` section.

For example, for a Morse PATH gateway supporting Ethereum & Polygon QoS, the following configuration would be added to the `.config.yaml` file:

```yaml
hydrator_config:
  service_ids:
    - "F00C"
    - "F021"
```

💡 _For more information on PATH's configuration file, please refer to the [configuration documentation](../../develop/path/6_configurations_helm.md)._

# 🌿 Current PATH QoS Support

**🗓️ Document Last Updated: 2025-04-23**

## Shannon Protocol Services

| Service Name                                               | Authoritative Service ID | Service QoS Type | Chain ID (if applicable) | Archival Check Configured |
| ---------------------------------------------------------- | ------------------------ | ---------------- | ------------------------ | ------------------------- |
| Ethereum - ETH Mainnet                                     | eth                      | EVM              |                          |                           |
| Anvil - Ethereum development/testing                       | anvil                    | EVM              | 31337                    |                           |
| Anvil WebSockets - Ethereum WebSockets development/testing | anvilws                  | EVM              | 31337                    |                           |
| CometBFT - Pocket Beta Testnet                             | pocket-beta-rpc          | CometBFT         | pocket-beta              |                           |
| CometBFT - Cosmos Hub                                      | cometbft                 | CometBFT         | cosmoshub-4              |                           |
| Solana                                                     | solana                   | Solana           |                          |                           |

## Morse Protocol Services

| Service Name                        | Authoritative Service ID | Service QoS Type | Chain ID (if applicable)            | Archival Check Configured |
| ----------------------------------- | ------------------------ | ---------------- | ----------------------------------- | ------------------------- |
| Arbitrum One                        | F001                     | EVM              | 42161                               |                           |
| Arbitrum Sepolia Testnet            | F002                     | EVM              | 421614                              |                           |
| Avalanche                           | F003                     | EVM              | 43114                               |                           |
| Avalanche-DFK                       | F004                     | EVM              | 53935                               |                           |
| Base                                | F005                     | EVM              | 8453                                |                           |
| Base Sepolia Testnet                | F006                     | EVM              | 84532                               |                           |
| Blast                               | F008                     | EVM              | 81457                               |                           |
| BNB Smart Chain                     | F009                     | EVM              | 56                                  |                           |
| Boba                                | F00A                     | EVM              | 288                                 |                           |
| Celo                                | F00B                     | EVM              | 42220                               |                           |
| Ethereum                            | F00C                     | EVM              | 1                                   | ✅                         |
| Ethereum Holesky Testnet            | F00D                     | EVM              | 17000                               |                           |
| Ethereum Sepolia Testnet            | F00E                     | EVM              | 11155111                            |                           |
| Evmos                               | F00F                     | EVM              | 9001                                |                           |
| Fantom                              | F010                     | EVM              | 250                                 |                           |
| Fraxtal                             | F011                     | EVM              | 252                                 |                           |
| Fuse                                | F012                     | EVM              | 122                                 |                           |
| Gnosis                              | F013                     | EVM              | 100                                 |                           |
| Harmony-0                           | F014                     | EVM              | 1666600000                          |                           |
| IoTeX                               | F015                     | EVM              | 4689                                |                           |
| Kaia                                | F016                     | EVM              | 8217                                |                           |
| Kava                                | F017                     | EVM              | 2222                                |                           |
| Metis                               | F018                     | EVM              | 1088                                |                           |
| Moonbeam                            | F019                     | EVM              | 1284                                |                           |
| Moonriver                           | F01A                     | EVM              | 1285                                |                           |
| Near                                | F01B                     | EVM              | 397                                 |                           |
| Oasys                               | F01C                     | EVM              | 61468                               | ✅                         |
| Optimism                            | F01D                     | EVM              | 10                                  |                           |
| Optimism Sepolia Testnet            | F01E                     | EVM              | 11155420                            |                           |
| opBNB                               | F01F                     | EVM              | 204                                 |                           |
| Polygon                             | F021                     | EVM              | 61473                               | ✅                         |
| Polygon Amoy Testnet                | F022                     | EVM              | 80002                               |                           |
| Radix                               | F023                     | EVM              | 4919                                |                           |
| Scroll                              | F024                     | EVM              | 534352                              |                           |
| Sui                                 | F026                     | EVM              | 257                                 |                           |
| Taiko                               | F027                     | EVM              | 167000                              |                           |
| Taiko Hekla Testnet                 | F028                     | EVM              | 167009                              |                           |
| Polygon zkEVM                       | F029                     | EVM              | 1101                                |                           |
| zkLink                              | F02A                     | EVM              | 810180                              |                           |
| zkSync                              | F02B                     | EVM              | 324                                 |                           |
| XRPL EVM Devnet                     | F02C                     | EVM              | 1440002                             |                           |
| Sonic                               | F02D                     | EVM              | 146                                 |                           |
| TRON                                | F02E                     | EVM              | 728126428                           |                           |
| Linea                               | F030                     | EVM              | 59144                               |                           |
| Berachain Testnet                   | F031                     | EVM              | 80084                               |                           |
| Ink                                 | F032                     | EVM              | 57073                               |                           |
| Mantle                              | F033                     | EVM              | 5000                                |                           |
| Sei                                 | F034                     | EVM              | 1329                                |                           |
| Berachain                           | F035                     | EVM              | 80094                               |                           |
| XRPL EVM Testnet                    | F036                     | EVM              | 61494                               | ✅                         |
| Celestia Archival                   | A0CA                     | CometBFT         | celestia-archival                   |                           |
| Celestia Consensus Archival         | A0CB                     | CometBFT         | celestia-consensus-archival         |                           |
| Celestia Testnet DA Archival        | A0CC                     | CometBFT         | celestia-testnet-da-archival        |                           |
| Celestia Testnet Consensus Archival | A0CD                     | CometBFT         | celestia-testnet-consensus-archival |                           |
| Osmosis                             | F020                     | CometBFT         | osmosis                             |                           |
| Solana                              | F025                     | Solana           |                                     |                           |
