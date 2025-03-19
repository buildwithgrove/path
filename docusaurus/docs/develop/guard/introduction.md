---
sidebar_position: 1
title: Introduction
description: High-level architecture overview and detailed walkthrough
---

<div align="center">
<h1>GUARD<br/>Gateway Utilities for Authentication, Routing & Defense</h1>
<img src="https://storage.googleapis.com/grove-brand-assets/Presskit/Logo%20Joined-2.png" alt="Grove logo" width="500"/>
</div>
<br/>

## Table of Contents <!-- omit in toc -->

- [Introduction](#introduction)
  - [Envoy Gateway](#envoy-gateway)
  - [Request Flow](#request-flow)

# Introduction

GUARD is the web2 gateway layer for PATH. It is responsible for authentication, rate limiting, and routing of incoming web traffic to the backend PATH service.

It uses Envoy Gateway as the underlying proxy and is configured through a Helm chart.

## Envoy Gateway

<div align="center">
  <a href="https://gateway.envoyproxy.io/docs/">
    <img src="https://raw.githubusercontent.com/cncf/artwork/refs/heads/main/projects/envoy/envoy-gateway/horizontal/color/envoy-gateway-horizontal-color.svg" alt="Envoy logo" width="200"/>
  </a>
  <br/>
  <a href="https://gateway.envoyproxy.io/docs/">
    <h2>Envoy Gateway Docs</h2>
  </a>
</div>

:::info From Envoy Gateway's Documentation

_Envoy Gateway is an open source project for managing Envoy Proxy as a standalone or Kubernetes-based application gateway. Gateway API resources are used to dynamically provision and configure the managed Envoy Proxies._

:::

<div align="center">
  <img src="https://gateway.envoyproxy.io/img/traffic.png" alt="Envoy Gateway" />
  <h2>Envoy Gateway</h2>
</div>

- [Envoy Gateway Resources](https://gateway.envoyproxy.io/docs/concepts/concepts_overview/)

## Request Flow

```mermaid
sequenceDiagram
    participant U as User
    participant G as GUARD<br/>(Envoy Gateway)
    participant A as API Key<br/>SecurityPolicy
    participant P as PATH

    U->>G: Send request with API key
    G->>A: Validate API key
    alt API key valid
        A-->>G: Valid response
        G->>P: Forward request to PATH
        P-->>G: Processed response
        G-->>U: Return response to user
    else API key invalid
        A-->>G: Invalid response
        G-->>U: Return error (Unauthorized)
    end
```
