---
sidebar_position: 6
title: Quality of Service
description: High level overview of Quality of Service (QoS) in PATH
---

## Table of Contents <!-- omit in toc -->

- [Quality of Service Structure](#quality-of-service-structure)

## Quality of Service Structure

The diagram below shows the intended high level structure of the QoS system in PATH.

:::note

The QoS package is under active development and does not currently reflect the final design as described in the diagram below.

:::


```mermaid
classDiagram
    class gateway.QoSService {
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
    
    
    EVM ..> gateway.QoSService : implements
    SolanaVM ..> gateway.QoSService : implements
    Cosmos ..> gateway.QoSService : implements
    MoveVM ..> gateway.QoSService : implements
    
    DefaultEVM ..> EVM : NewQoSService
    DefaultCosmos ..> Cosmos : NewQoSService
    
    ETH --> DefaultEVM : extends
    FUSE --> DefaultEVM : extends

    POKT --> DefaultCosmos: extends
```