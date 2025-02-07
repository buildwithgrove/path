---
sidebar_position: 3
title: Environment Setup
description: Quick reference to setup your environment
---

This guide provides a quick reference for setting up and running a local `PATH` instance in **Tilt**.

It is a pre-requisite for the following protocol-specific guides:

- [**Shannon Protocol Guide**](./cheat_sheet_shannon.md): For the new Shannon protocol (Beta TestNet as of 11/2025)
- [**Morse Protocol Guide**](./cheat_sheet_morse.md): For the original Morse protocol (MainNet as of 2020)

## Table of Contents <!-- omit in toc -->

- [Development Environment](#development-environment)
- [Prerequisites](#prerequisites)
  - [1. Clone the Repository](#1-clone-the-repository)
  - [2. Install Required Tools](#2-install-required-tools)
- [3. Setup Envoy Proxy](#3-setup-envoy-proxy)
- [4. Choose Your Protocol](#4-choose-your-protocol)
- [Additional Resources](#additional-resources)

## Development Environment

PATH uses a Kubernetes-based local development environment. We use Kind (Kubernetes in Docker)
for running the local Kubernetes cluster, as it provides better compatibility with Tilt's
image building process compared to Docker Desktop's Kubernetes cluster.

## Prerequisites

### 1. Clone the Repository

```bash
git clone https://github.com/buildwithgrove/path.git
cd ./path
```

### 2. Install Required Tools

The following tools are required to start a local PATH instance:

**Local Deployment Tools:**

- [**Docker**](https://docs.docker.com/get-docker/): Container runtime
- [**Kind**](https://kind.sigs.k8s.io/#installation-and-usage): Local Kubernetes cluster
- [**kubectl**](https://kubernetes.io/docs/tasks/tools/#kubectl): CLI for interacting with Kubernetes
- [**Helm**](https://helm.sh/docs/intro/install/): Package manager for Kubernetes
- [**Tilt**](https://docs.tilt.dev/install.html): Local Kubernetes development environment
- [**Graphviz**](https://graphviz.org) (Optional): Required for generating profiling & debugging performance

**Development Tools:**

- **[Uber Mockgen](https://github.com/uber-go/mock)**: Mock interface generator for testing

:::tip

To install all dependencies automatically:

```bash
make install_deps
```

:::

## 3. Setup Envoy Proxy

Setup all the configurations to enable authorization, service aliasing, and rate limiting:

```bash
make init_envoy
```

This will generate four configuration files:

- `.allowed-services.lua`
- `.envoy.yaml`
- `.ratelimit.yaml`
- `.gateway-endpoints.yaml`

:::tip

For a quick initial setup, choose **Option 2 (no authorization)** when prompted.

:::

## 4. Choose Your Protocol

| Protocol | Version | Status   | Network | Documentation                                      |
| -------- | ------- | -------- | ------- | -------------------------------------------------- |
| Shannon  | v1      | Beta     | TestNet | [Shannon Protocol Guide](./cheat_sheet_shannon.md) |
| Morse    | v0      | Original | MainNet | [Morse Protocol Guide](./cheat_sheet_morse.md)     |

## Additional Resources

- [PATH Walkthrough](introduction.md) - Detailed explanation of PATH architecture
- [PATH Config Docs](path_config.md) - Detailed configuration guide
- [Envoy Config Docs](../envoy/envoy_config.md) - Envoy proxy configuration guide
