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

- [Getting Started](#getting-started)
  - [1. Clone the Repository](#1-clone-the-repository)
  - [2. Install `pocketd` CLI](#2-install-pocketd-cli)
  - [3. Install Docker](#3-install-docker)
  - [4. Choose Your Protocol](#4-choose-your-protocol)
- [Development Environment Details](#development-environment-details)
- [Remote vs Local Helm Charts](#remote-vs-local-helm-charts)

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/buildwithgrove/path.git
cd ./path
```

### 2. Install `pocketd` CLI

To install the [`pocketd` CLI](https://dev.poktroll.com/category/pocketd-cli) on Linux or MacOS, run:

```bash
curl -sSL https://raw.githubusercontent.com/pokt-network/poktroll/main/scripts/install.sh | bash
```

### 3. Install Docker

<br/>
<div align="center">
<a href="https://docs.docker.com/get-started/get-docker">
![Docker](../../../static/img/docker.png)
<p><b>Docker Installation Instructions</b></p>
</a>
</div>

If not already installed, follow the directions at [Get Docker](https://docs.docker.com/get-started/get-docker) to install Docker in your environment.

### 4. Choose Your Protocol

Pick one of these protocols and follow the respective guide.

| Protocol | Version (aka) | Status (04/2025)       | Network | Documentation                                     |
| -------- | ------------- | ---------------------- | ------- | ------------------------------------------------- |
| Shannon  | v1            | Beta & Private MainNet | TestNet | [Shannon Protocol Guide](3_cheatsheet_shannon.md) |
| Morse    | v0            | MainNet (2020)         | MainNet | [Morse Protocol Guide](4_cheatsheet_morse.md)     |

## Development Environment Details

:::tip

The PATH local development environment runs inside a single Docker container using the `docker/dind` image as a base.

**Therefore the only dependency is a running Docker daemon.**

However, if you wish to know more about what's happening under the hood, read on.

:::

<br/>
<div align="center">
  <a href="https://docs.tilt.dev/">
    <img src="https://blog.tilt.dev/assets/img/blog-default-preview.png" alt="Tilt logo" width="200"/>
  <p><b>Tilt Documentation</b></p>
  </a>
</div>

**PATH**'s development mode uses a Kubernetes-based local development environment with [Tilt](https://tilt.dev/).

We use [Kind](https://kind.sigs.k8s.io/) (Kubernetes in Docker) for running the local Kubernetes cluster, as it provides better compatibility with Tilt's
image building process compared to Docker Desktop's Kubernetes cluster.

The full PATH stack uses [Helm Charts](https://helm.sh/) to deploy the necessary services to the `Kind` Kubernetes cluster.

For more information, see the [PATH Helm Introduction](../../operate/helm/1_introduction.md).

You may view the [PATH Helm Charts](https://github.com/buildwithgrove/helm-charts) repository if you're interested in the services deployed to the local Kubernetes cluster.

## Remote vs Local Helm Charts

By default, the PATH local development environment uses the remote PATH Helm Charts.

For local development, you may optionally choose to use the local Helm Charts by pulling the [PATH Helm Charts](https://github.com/buildwithgrove/helm-charts) repository.

To use the local Helm Charts, run:

```bash
make path_up_local_helm
```

This will prompt you for the local path to the PATH Helm Charts repository. By default this is `../helm-charts` relative to the PATH repository root.

You may also specify a custom relative or absolute path to the PATH Helm Charts repository.
