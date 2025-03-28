---
sidebar_position: 7
title: Quality of Service
description: High level overview of Quality of Service (QoS) in PATH
---

:::warning 🚧 Under Construction 🚧

The QoS package is under active development and does not currently reflect the final design as described in the diagram below.

:::

## Table of Contents <!-- omit in toc -->

- [Quality of Service Structure](#quality-of-service-structure)

## Quality of Service Structure

The diagram below is meant for informative purposes only. It is a high-level direction
which the QoS package is taking as of 01/2025.

```mermaid
classDiagram
    class QoSService {
        <<interface>>
    }
    class EVM {
        <<struct>>
        evm.QoS
    }
    class SolanaVM {
        <<struct>>
        solana.QoS
    }
    class Cosmos {
        <<struct>>
        cosmos.QoS
    }
    class MoveVM {
        <<struct>>
        move.QoS
    }
    class DefaultEVM {
        +implementation
        JSON-RPC
        ---
        eth_blockNumber
    }
    class DefaultCosmos {
        +implementation
        **REST**
        ---
        /v1/query/height

    }
    class ETH {
        +chainId: 0x1
        --
        eth_chainId
    }
    class FUSE {
        +chainId: 0x7a
        --
        eth_chainId
    }

    class POKT {
        +chainId: 0x7a
        --
        eth_chainId
    }

    EVM --|> QoSService : implements
    SolanaVM --|> QoSService : implements
    Cosmos --|> QoSService : implements
    MoveVM --|> QoSService : implements

    DefaultEVM ..> EVM : NewQoSService
    DefaultCosmos ..> Cosmos : NewQoSService

    ETH --> DefaultEVM : extends
    FUSE --> DefaultEVM : extends

    POKT --> DefaultCosmos: extends
```