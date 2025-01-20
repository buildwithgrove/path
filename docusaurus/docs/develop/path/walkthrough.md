---
sidebar_position: 2
title: Walkthrough
description: High-level architecture overview and detailed walkthrough
---

<div align="center">
<h1>PATH<br/>Path API & Toolkit Harness</h1>
<img src="https://storage.googleapis.com/grove-brand-assets/Presskit/Logo%20Joined-2.png" alt="Grove logo" width="500"/>

</div>
<br/>

![Static Badge](https://img.shields.io/badge/Maintained_by-Grove-green)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/buildwithgrove/path/main-build.yml)
![GitHub last commit](https://img.shields.io/github/last-commit/buildwithgrove/path)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/buildwithgrove/path)
![GitHub Release](https://img.shields.io/github/v/release/buildwithgrove/path)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/buildwithgrove/path/total)
![GitHub Issues or Pull Requests](https://img.shields.io/github/issues/buildwithgrove/path)
![GitHub Issues or Pull Requests](https://img.shields.io/github/issues-pr/buildwithgrove/path)
![GitHub Issues or Pull Requests](https://img.shields.io/github/issues-closed/buildwithgrove/path)
![App Status](https://argocd.tooling.buildintheshade.com/api/badge?name=path-gateway&revision=true&showAppName=true)

## Table of Contents <!-- omit in toc -->

- [Introduction](#introduction)
  - [Prerequisites](#prerequisites)
- [Path Releases](#path-releases)
- [Quickstart](#quickstart)
- [Running PATH](#running-path)
  - [Run PATH in Tilt](#run-path-in-tilt)
  - [Run the PATH binary](#run-the-path-binary)
- [E2E Tests](#e2e-tests)
  - [Running the E2E tests against Shannon Testnet](#running-the-e2e-tests-against-shannon-testnet)
    - [Preparing the configuration](#preparing-the-configuration)
    - [Running the E2E tests](#running-the-e2e-tests)
  - [Running the E2E tests against Morse](#running-the-e2e-tests-against-morse)
    - [Preparing the configuration](#preparing-the-configuration-1)
    - [Running the E2E tests](#running-the-e2e-tests-1)
- [Running Localnet](#running-localnet)
  - [Spinning up / Tearing down Localnet](#spinning-up--tearing-down-localnet)
- [Troubleshooting](#troubleshooting)
  - [Docker Permissions Issues - Need to run sudo?](#docker-permissions-issues---need-to-run-sudo)
- [Special Thanks](#special-thanks)
- [License](#license)

## Introduction

**PATH** (Path API & Toolkit Harness) is an open source framework for enabling
access to a decentralized supply network.

It provides various tools and libraries to streamline the integration and
interaction with decentralized protocols.

We use Tilt + Kind to spin up local environment for development and local testing purposes.

Kind is intentionally used instead of Docker Kubernetes cluster since we have observed that images created through Tilt are not accesible when using Docker K8s cluster.

### Prerequisites

**Local Deployment:**

- [Docker](https://docs.docker.com/get-docker/)
- [Kind](https://kind.sigs.k8s.io/#installation-and-usage)
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
- [Helm](https://helm.sh/docs/intro/install/)
- [Tilt](https://docs.tilt.dev/install.html)

**Development only:**

- [Uber Mockgen](https://github.com/uber-go/mock)

## Path Releases

Path releases provide a Docker image you can start using right away to bootstrap
your Path gateway without the need of building your own image.

You can find:

- All the releases [here](https://github.com/buildwithgrove/path/releases)
- All the package versions [here](https://github.com/buildwithgrove/path/pkgs/container/path/versions)
- The containers page [here](https://github.com/buildwithgrove/path/pkgs/container/path)

You can pull them directly using the following command:

```sh
docker pull ghcr.io/buildwithgrove/path
```

## Quickstart

:::tip

See the [PATH Cheat Sheet](../path/cheatsheet.md) for instructions on how to get started with a local PATH instance on Shannon or Morse.

:::

## Running PATH

There are two ways to run PATH locally for development purposes.

### Run PATH in Tilt

<div align="center">
  <a href="https://docs.tilt.dev/">
    <img src="https://blog.tilt.dev/assets/img/blog-default-preview.png" alt="Tilt logo" width="200"/>
  <p><b>Tilt Docs</b></p>
  </a>
</div>

To enable the full suite of PATH functionality - including authorization, rate limiting, service aliasing and more - a `Tiltfile` is provided to easily run the PATH gateway with all its dependencies in a local Kubernetes cluster.

:::tip
You can follow the instructions [in the PATH cheat sheet](../path/cheatsheet.md) to get started with PATH in Tilt.
:::

### Run the PATH binary

For an even simpler way to get started with PATH, you can run the PATH binary standalone.

In this mode, PATH will not use Envoy Proxy, which disables authorization, rate limiting & service aliasing.

To run the PATH binary, first run one of the following commands to copy an example config file to `./bin/config/.config.yaml`.

```sh
# To copy an example Shannon config file
make copy_shannon_config

# To copy an example Morse config file
make copy_morse_config
```

:::tip

[For a guide on how to get your config file ready, see the PATH Cheat Sheet's "Setup Protocol" section](../path/cheatsheet.md#13-setup-protocol).

[For detailed information on the PATH configuration file, see the PATH Config Docs](../path/path_config.md).

:::

Once you have your config file ready, run the following command to build the PATH binary and run it in standalone mode:

```sh
make run_path
```

## E2E Tests

This repository contains end-to-end (E2E) tests for the Shannon relay protocol. The tests ensure that the protocol behaves as expected under various conditions.

### Running the E2E tests against Shannon Testnet

#### Preparing the configuration

A `make` target is provided to copy the example Shannon configuration file to the `e2e/.shannon.config.yaml` needed by the E2E tests on Shannon.

```sh
make copy_shannon_e2e_config
```

Then update the `shannon_config.gateway_config` values with the appropriate values.

You can find the example Shannon configuration file [here](https://github.com/buildwithgrove/path/tree/main/e2e/shannon.example.yaml).

#### Running the E2E tests

To run the tests, use the following `make` targets:

```sh
# Run E2E tests against Shannon Testnet
make test_e2e_shannon_relay

# Run all tests
make test_all
```

### Running the E2E tests against Morse

#### Preparing the configuration

A `make` target is provided to copy the example Morse configuration file to the `e2e/.morse.config.yaml` needed by the E2E tests on Morse.
To run the tests, use the following `make` targets:

```sh
make copy_morse_e2e_config
```

Then update the `morse_config.full_node_config` and `morse_config.signed_aats` values with the appropriate values.

You can find the example Morse configuration file [here](https://github.com/buildwithgrove/path/tree/main/e2e/morse.example.yaml).

:::note Grove Employee Only

If you are a Grove employee, download [Grove's Morse configuration file for PATH E2E tests](https://start.1password.com/open/i?a=4PU7ZENUCRCRTNSQWQ7PWCV2RM&v=kudw25ob4zcynmzmv2gv4qpkuq&i=2qk5qlmrduh7irgjzih3hejfxu&h=buildwithgrove.1password.com) and COPY IT OVER the `e2e/.morse.config.yaml` file.\*\*

⚠️ The above configuration file is sensitive and the contents of this file must never be shared outside of your organization. ⚠️

:::

#### Running the E2E tests

To run the tests, use the following `make` targets:

```sh
# Run E2E tests against Morse
make test_e2e_morse_relay
```

## Running Localnet

You can use path configuration under `/local` to spin up a local development environment using `Kind` + `Tilt`.

Make sure to review [Tiltfile](https://github.com/buildwithgrove/path/tree/main/Tiltfile) and [values file](https://github.com/buildwithgrove/path/tree/main/local/path/config/path-values.yaml) to make sure they have your desired configuration.

### Spinning up / Tearing down Localnet

Localnet can be spun up/torn down using the following targets:

- `path_up` -> Spins up localnet environment using Kind + Tilt
- `path_down` -> Tears down localnet.

## Troubleshooting

### Docker Permissions Issues - Need to run sudo?

If you're hitting docker permission issues (e.g. you need to use sudo),
see the solution [here](https://github.com/jgsqware/clairctl/issues/60#issuecomment-358698788)
or just copy-paste the following command:

```bash
sudo chmod 666 /var/run/docker.sock
```

## Special Thanks

The origins of this repository were inspired by the work kicked off in [gateway-server](https://github.com/pokt-network/gateway-server) by the
[Nodies](https://nodies.app/) team. We were inspired and heavily considering forking and building off of that effort.

However, after a week-long sprint, the team deemed that starting from scratch was the better path forward for multiple reasons. These include but are not limited to:

- Enabling multi-protocol support; Morse, Shanon and beyond
- Set a foundation to migrate Grove's quality of service and data pipelineta
- Integrating with web2 standards like [Envoy](https://www.envoyproxy.io/), [gRPC](https://grpc.io/), [Stripe](https://stripe.com/), [NATS](https://nats.io/), [Auth0](https://auth0.com/), etc...
- Etc...

<!-- TODO(@olshansk): Move over the docs from [gateway-server](https://github.com/pokt-network/gateway-server) to a Morse section under [path.grove.city](https://path.grove.city) -->

---

## License

This project is licensed under the MIT License; see the [LICENSE](https://github.com/buildwithgrove/path/blob/main/LICENSE) file for details.
