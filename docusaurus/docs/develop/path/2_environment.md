---
sidebar_position: 2
title: Environment Setup (< 10 min)
description: Pre-requisite to setup your local environment
---

This guide provides instructions required to setup your environment for local development.
locally in a development environment.

It is a pre-requisite for both the:

- [**Shannon Protocol Guide**](3_cheatsheet_shannon.md): Pocket Network v1 (Private MainNet as of 04/2025)
- [**Morse Protocol Guide**](4_cheatsheet_morse.md): For the original Morse protocol (Public MainNet as of 2020)

## Table of Contents <!-- omit in toc -->

- [Development Environment Details](#development-environment-details)
- [Technical Pre-Requisites \& Setup](#technical-pre-requisites--setup)
  - [1. Clone the Repository](#1-clone-the-repository)
  - [2. Install Required Tools (Linux \& MacOS)](#2-install-required-tools-linux--macos)
    - [Install `pocketd` CLI](#install-pocketd-cli)
    - [Install Required Local Tools](#install-required-local-tools)
  - [3. Choose Your Protocol](#3-choose-your-protocol)

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

### 2. Install Required Tools (Linux & MacOS)

#### Install `pocketd` CLI

To install the [`pocketd` CLI](https://dev.poktroll.com/category/pocketd-cli) on Linux or MacOS, run:

```bash
curl -sSL https://raw.githubusercontent.com/pokt-network/poktroll/main/scripts/install.sh | bash
```

#### Install Required Local Tools

To install all required local development tools on Linux or MacOS, run:

```bash
make install_tools
```

The above script will install the following tools which are required to start a PATH instance locally in development mode:

- [**Docker**](https://docs.docker.com/get-docker/): Container runtime
- [**Kind**](https://kind.sigs.k8s.io/#installation-and-usage): Local Kubernetes cluster
- [**kubectl**](https://kubernetes.io/docs/tasks/tools/#kubectl): CLI for interacting with Kubernetes
- [**Helm**](https://helm.sh/docs/intro/install/): Package manager for Kubernetes
- [**Tilt**](https://docs.tilt.dev/install.html): Local Kubernetes development environment

:::tip

The following optional tools may be helpful for your development workflow but are not required to start a PATH instance locally in development mode:

- [**Relay Util**](https://github.com/commoddity/relay-util): An easy to use load testing tool for sending configurable batches of relays concurrently
- [**Graphviz**](https://graphviz.org): Required for generating profiling & debugging performance
- [**Uber Mockgen**](https://github.com/uber-go/mock): Mock interface generator for testing

They may be installed with:

```bash
make install_optional_tools
```

:::

### 3. Choose Your Protocol

Pick one of these protocols and follow the respective guide.

| Protocol | Version (aka) | Status (04/2025)       | Network | Documentation                                     |
| -------- | ------------- | ---------------------- | ------- | ------------------------------------------------- |
| Shannon  | v1            | Beta & Private MainNet | TestNet | [Shannon Protocol Guide](3_cheatsheet_shannon.md) |
| Morse    | v0            | MainNet (2020)         | MainNet | [Morse Protocol Guide](4_cheatsheet_morse.md)     |
