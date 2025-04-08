---
sidebar_position: 2
title: Environment Setup
description: Pre-requisite to setup your local environment
---

This guide provides instructions required to setup your environment for local development.
locally in a development environment.

It is a pre-requisite for both the:

- [**Shannon Protocol Guide**](./cheatsheet_shannon.md): Pocket Network v1 (Private MainNet as of 04/2025)
- [**Morse Protocol Guide**](./cheatsheet_morse.md): For the original Morse protocol (Public MainNet as of 2020)

## Table of Contents <!-- omit in toc -->

- [Development Environment Details](#development-environment-details)
- [Technical Pre-Requisites \& Setup](#technical-pre-requisites--setup)
  - [1. Clone the Repository](#1-clone-the-repository)
  - [2. Install Required Tools](#2-install-required-tools)
- [3. Choose Your Protocol](#3-choose-your-protocol)
- [Additional Resources](#additional-resources)

## Development Environment Details

<div align="center">
  <a href="https://docs.tilt.dev/">
    <img src="https://blog.tilt.dev/assets/img/blog-default-preview.png" alt="Tilt logo" width="200"/>
  <p><b>Tilt Documentation</b></p>
  </a>
</div>

**PATH**'s development mode uses a Kubernetes-based local development environment with [Tilt](https://tilt.dev/).

We use [Kind](https://kind.sigs.k8s.io/) (Kubernetes in Docker) for running the local Kubernetes cluster, as it provides better compatibility with Tilt's
image building process compared to Docker Desktop's Kubernetes cluster.

## Technical Pre-Requisites & Setup

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
- [**Graphviz**](https://graphviz.org) (Debug only): Required for generating profiling & debugging performance
- [**Relay Util**](https://github.com/commoddity/relay-util): (Load testing tool): Used to send configurable batches of relays concurrently

**Development Tools:**

- **[Uber Mockgen](https://github.com/uber-go/mock)**: Mock interface generator for testing

:::tip

To install all dependencies automatically:

```bash
make install_deps
```

:::warning

This script currently only works on Linux. MacOS version coming soon.

:::

## 3. Choose Your Protocol

| Protocol | Version | Status   | Network | Documentation                                                |
| -------- | ------- | -------- | ------- | ------------------------------------------------------------ |
| Shannon  | v1      | Beta     | TestNet | [Shannon Protocol Quickstart Guide](./cheatsheet_shannon.md) |
| Morse    | v0      | Original | MainNet | [Morse Protocol Quickstart Guide](./cheatsheet_morse.md)     |

## Additional Resources

- [PATH Configuration Files](./configuration.md) - Detailed configuration instructions
- [PATH Helm Chart](../helm/path.md) - Full documentation for the PATH Helm chart
- [GUARD Helm Chart](../helm/guard.md) - Full documentation for the GUARD Helm chart
- [WATCH Helm Chart](../helm/watch.md) - Full documentation for the WATCH Helm chart
