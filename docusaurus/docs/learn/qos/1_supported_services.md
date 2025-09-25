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

See [PATH Configuration File](../../develop/configs/2_gateway_config.md#hydrator_config-optional) for more details.

## ⛓️ Supported QoS Services

The table below lists the Quality of Service (QoS) implementations currently supported by PATH.

:::warning **🚧 QoS Support 🚧**

If a Service ID is not specified in the tables below, it does not have a QoS implementation in PATH.

:::

## 🌿 Current PATH QoS Support

**🗓️ Document Last Updated: 2025-09-23**

## Shannon Protocol Services

| Service Name | Authoritative Service ID | Service QoS Type | Chain ID (if applicable) | Archival Check Configured |
|-------------|------------|-----------------|----------|---------------------------|
| TODO_NEXT(@commoddity): Add archival check config for Giwa once `giwa` mainnet is live. | giwa | EVM | 1 |  |
| Anvil - Ethereum development/testing | anvil | EVM | 31337 |  |
| Anvil WebSockets - Ethereum WebSockets development/testing | anvilws | EVM | 31337 |  |
| Fraxtal | fraxtal | EVM | 252 |  |
| Kava | kava | EVM | 2222 |  |
| Moonriver | moonriver | EVM | 1285 |  |
| opBNB | opbnb | EVM | 204 |  |
| Sui | sui | EVM | 257 |  |
| TRON | tron | EVM | 728126428 |  |
| Sei | sei | EVM | 1329 |  |
| TODO_TECHDEBT(@olshansk): Either remove or format this correctly | hey | EVM | 1 | ✅ |
| Hyperliquid | hyperliquid | EVM | 999 |  |
| Unichain | unichain | EVM | 130 |  |
| Akash | akash | Cosmos SDK | akashnet-2 |  |
| Arkeo | arkeo | Cosmos SDK | arkeo-main-v1 |  |
| AtomOne | atomone | Cosmos SDK | atomone-1 |  |
| Babylon | babylon | Cosmos SDK | bbn-1 |  |
| Celestia | celestia | Cosmos SDK | celestia |  |
| Cheqd | cheqd | Cosmos SDK | cheqd-mainnet-1 |  |
| Chihuahua | chihuahua | Cosmos SDK | chihuahua-1 |  |
| Cosmos Hub | cosmoshub | Cosmos SDK | cosmoshub-4 |  |
| Dungeon Chain | dungeon-chain | Cosmos SDK | dungeon-1 |  |
| Elys Network | elys-network | Cosmos SDK | elys-1 |  |
| Fetch | fetch | Cosmos SDK | fetchhub-4 |  |
| Jackal | jackal | Cosmos SDK | jackal-1 |  |
| Juno | juno | Cosmos SDK | juno-1 |  |
| KYVE | kyve | Cosmos SDK | kyve-1 |  |
| Neutron | neutron | Cosmos SDK | neutron-1 |  |
| Nillion | nillion | Cosmos SDK | nillion-1 |  |
| Osmosis | osmosis | Cosmos SDK | osmosis-1 |  |
| Passage | passage | Cosmos SDK | passage-2 |  |
| Persistence | persistence | Cosmos SDK | core-1 |  |
| Provenance | provenance | Cosmos SDK | pio-mainnet-1 |  |
| Pocket Mainnet | pocket | Cosmos SDK | pocket |  |
| Pocket Alpha Testnet - (Not in the chain registry - present here for onchain load testing) | pocket-alpha | Cosmos SDK | pocket-alpha |  |
| Pocket Beta Testnet | pocket-beta | Cosmos SDK | pocket-beta |  |
| Pocket Beta Testnet 1 - (Not in the chain registry - present here for onchain load testing) | pocket-beta1 | Cosmos SDK | pocket-beta |  |
| Pocket Beta Testnet 2 - (Not in the chain registry - present here for onchain load testing) | pocket-beta2 | Cosmos SDK | pocket-beta |  |
| Pocket Beta Testnet 3 - (Not in the chain registry - present here for onchain load testing) | pocket-beta3 | Cosmos SDK | pocket-beta |  |
| Pocket Beta Testnet 4 - (Not in the chain registry - present here for onchain load testing) | pocket-beta4 | Cosmos SDK | pocket-beta |  |
| Quicksilver | quicksilver | Cosmos SDK | quicksilver-2 |  |
| Router | router | Cosmos SDK | router_9600-1 |  |
| Seda | seda | Cosmos SDK | seda-1 |  |
| Shentu | shentu | Cosmos SDK | shentu-2.2 |  |
| Side Protocol | side-protocol | Cosmos SDK | sidechain-1 |  |
| Stargaze | stargaze | Cosmos SDK | stargaze-1 |  |
| Stride | stride | Cosmos SDK | stride-1 |  |
| XRPLEVM | xrplevm | Cosmos SDK | xrplevm_1440000-1 |  |
| XRPLEVM Testnet | xrplevm-testnet | Cosmos SDK | xrplevm_1449000-1 |  |
| Solana | solana | Solana |  |  |
