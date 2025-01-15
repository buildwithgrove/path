---
sidebar_position: 3
title: Cheat Sheet
description: Quick reference guide for setting up and running a local `PATH` instance in **Tilt**.
---

This guide provides a quick reference for setting up and running a local `PATH` instance in **Tilt**.

It is a pre-requisite for the following protocol-specific guides:

- [Shannon Protocol Guide](./shannon_cheatsheet.md) - For the new Shannon protocol (Beta TestNet as of 11/2025)
- [Morse Protocol Guide](./morse_cheatsheet.md) - For the original Morse protocol (MainNet as of 2020)

## Table of Contents <!-- omit in toc -->

- [Prerequisites](#prerequisites)
  - [1. Clone the Repository](#1-clone-the-repository)
  - [2. Install Required Tools](#2-install-required-tools)
- [Choose Your Protocol](#choose-your-protocol)
- [Additional Resources](#additional-resources)

## Prerequisites

### 1. Clone the Repository

```bash
git clone https://github.com/buildwithgrove/path.git
cd ./path
```

### 2. Install Required Tools

The following tools are required to start a local PATH instance:

- [**Docker**](https://docs.docker.com/get-docker/): Container runtime
- [**Kind**](https://kind.sigs.k8s.io/#installation-and-usage): Local Kubernetes cluster
- [**kubectl**](https://kubernetes.io/docs/tasks/tools/#kubectl): CLI for interacting with Kubernetes
- [**Helm**](https://helm.sh/docs/intro/install/): Package manager for Kubernetes
- [**Tilt**](https://docs.tilt.dev/install.html): Local Kubernetes development environment

:::tip

To install all dependencies automatically:

```bash
make install_deps
```

:::

## Choose Your Protocol

| Protocol | Version | Status   | Network | Documentation                                   |
| -------- | ------- | -------- | ------- | ----------------------------------------------- |
| Shannon  | v1      | Beta     | TestNet | [Shannon Protocol Guide](shannon_cheatsheet.md) |
| Morse    | v0      | Original | MainNet | [Morse Protocol Guide](morse_cheatsheet.md)     |

## Additional Resources

- [PATH Walkthrough](../path/walkthrough.md) - Detailed explanation of PATH architecture
- [PATH Config Docs](../path/path_config.md) - Detailed configuration guide
- [Envoy Config Docs](../envoy/envoy_config.md) - Envoy proxy configuration guide
