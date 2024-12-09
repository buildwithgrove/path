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

# Table of Contents <!-- omit in toc -->

- [1. Introduction](#1-introduction)
  - [1.1. Prerequisites](#11-prerequisites)
- [2. Path Releases](#2-path-releases)
- [3. Quickstart](#3-quickstart)
  - [3.1 Shannon Quickstart](#31-shannon-quickstart)
  - [3.2 Morse Quickstart](#32-morse-quickstart)
- [4. Configuration](#4-configuration)
  - [4.1 Configuration File](#41-configuration-file)
- [5. Running PATH](#5-running-path)
  - [5.1. Setup Config YAML](#51-setup-config-yaml)
  - [5.2. Run the PATH binary](#52-run-the-path-binary)
- [6. E2E Tests](#6-e2e-tests)
  - [6.1. Running the E2E tests against Shannon Testnet](#61-running-the-e2e-tests-against-shannon-testnet)
    - [6.1.1. Preparing the configuration](#611-preparing-the-configuration)
    - [6.1.2. Running the E2E tests](#612-running-the-e2e-tests)
  - [6.2. Running the E2E tests against Morse](#62-running-the-e2e-tests-against-morse)
    - [6.2.1. Preparing the configuration](#621-preparing-the-configuration)
    - [6.2.2. Running the E2E tests](#622-running-the-e2e-tests)
- [7. Running Localnet](#7-running-localnet)
  - [7.1. Spinning up / Tearing down Localnet](#71-spinning-up--tearing-down-localnet)
- [Special Thanks](#special-thanks)
- [License](#license)

<!--
TODO_UPNEXT(@commoddity): Convert all the tips, notes & warnings int his file into
docusaurus admonitions; https://docusaurus.io/docs/markdown-features/admonitions.
-->

## 1. Introduction

**PATH** (Path API & Toolkit Harness) is an open source framework for enabling
access to a decentralized supply network.

It provides various tools and libraries to streamline the integration and
interaction with decentralized protocols.

We use Tilt + Kind to spin up local environment for development and local testing purposes.

<!--TODO_UPNEXT(@HebertCL): Create a FAQ just like Poktroll for additional explanation on the chosen tooling -->

Kind is intentionally used instead of Docker Kubernetes cluster since we have observed that images created through Tilt are not accesible when using Docker K8s cluster.

### 1.1. Prerequisites

**Deployment:**

- [Docker](https://docs.docker.com/get-docker/)
- [Kind](https://kind.sigs.k8s.io/#installation-and-usage)
- [Tilt](https://docs.tilt.dev/install.html)
- [Helm](https://helm.sh/docs/intro/install/)

**Development only:**

- [Uber Mockgen](https://github.com/uber-go/mock)

## 2. Path Releases

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

## 3. Quickstart

### 3.1 Shannon Quickstart

1. **Stake Apps and Gateway:** Refer to the [Poktroll Docker Compose Walkthrough](https://dev.poktroll.com/operate/quickstart/docker_compose_walkthrough) for instructions on staking your Application and Gateway on Shannon.

2. **Populate Config File:** Run `make config_shannon_localnet` to copy the example configuration file to `local/path/config/.config.yaml`.

   Update the configuration file `local/path/config/.config.yaml` with your Gateway's private key & address and your Application's address.

   \*TIP: If you followed the [Debian Cheat Sheet](https://dev.poktroll.com/operate/quickstart/docker_compose_debian_cheatsheet#start-the-relayminer), you can run `path_prepare_config`
   to get you most of the way there. Make sure to review the `gateway_private_key` field.\*

3. **Start the PATH Container:** Run `make path_up` to build and start the PATH gateway in the Local development environment using Tilt.

4. **Run a curl command**: Example `eth_blockNumber` request to a PATH supporting `eth`:

   ```bash
   curl http://eth.localhost:3000/v1 \
       -X POST \
       -H "Content-Type: application/json" \
       -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
   ```

### 3.2 Morse Quickstart

1. **Retrieve Application Authentication Token & Keys**

   This is a relatively manual process in Morse that is not well documented.

   You should reach out to the team directly if you are doing this, but can refer to the following resources as references:

   - [Host a Gateway on Morse](https://docs.pokt.network/gateways/host-a-gateway)
   - [pocket-core/doc/specs/application-auth-token.md](https://github.com/pokt-network/pocket-core/blob/7f936ff7353249b161854e24435e4bc32d47aa3f/doc/specs/application-auth-token.md)
   - [pocket-core/doc/specs/cli/apps.md](https://github.com/pokt-network/pocket-core/blob/7f936ff7353249b161854e24435e4bc32d47aa3f/doc/specs/cli/apps.md)
   - [Gateway Server Kit instructions (as a reference)](https://github.com/pokt-network/gateway-server/blob/main/docs/quick-onboarding-guide.md#5-insert-app-stake-private-keys)

2. **Populate Config File:** Run `make config_morse_localnet` to copy the example configuration file to `local/path/config/.config.yaml`.

   Update the configuration file `local/path/config/.config.yaml` with your Gateway's private key, address and your delegated Application's address.

   2.1 **If you're a Grove employee**, you can use copy-paste the PROD configs from [here](https://www.notion.so/buildwithgrove/PATH-Morse-Configuration-Helpers-Instructions-111a36edfff6807c8846f78244394e74?pvs=4).

   2.2 **If you're a community member**, run the following command to get started quickly with a prefilled configuration
   for Bitcoin MainNet on Pocket Morse TestNet: `cp ./cmd/.config.morse_example_testnet.yaml ./cmd/.config.yaml`

3. **Start the PATH Container:** Run `make path_up` to build and start the PATH gateway in the Local development environment using Tilt.

4. **Run a curl command**: Example `eth_blockNumber` request to a PATH supporting `eth`:

   ```bash
   curl http://eth.localhost:3000/v1 \
       -X POST \
       -H "Content-Type: application/json" \
       -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
   ```

## 4. Configuration

### 4.1 Configuration File

The configuration for PATH is defined in a YAML file, which should be named `.config.yaml`.

This file is required for setting up a PATH instance and must be populated with the appropriate values.

The configuration is divided into several sections:

1. **Morse Configuration (`morse_config`)**:

   - **Required for Morse gateways.**
   - Must include full node URL and relay signing key.
   - Must include AAT data for all onchain staked applications assigned to the gateway operator

2. **Shannon Configuration (`shannon_config`)**:

   - **Required for Shannon gateways.**
   - Must include RPC URL, gRPC host/port, and gateway address/private key.
   - Must include the addresses of the onchain Applications that are delegated to the onchain Gateway.

3. **Services Configuration (`services`)**:

   - **Required for all gateways; at least one service must be listed.**
   - The key is the Service ID (e.g. `F00C`) and the value is the service configuration.
   - Only the Service ID is required. All other fields are optional.

4. **Router Configuration (`router_config`)**:

   - _Optional. Default values will be used if not specified._
   - Configures router settings such as port and timeouts.

- [Example Shannon Configuration File](https://github.com/buildwithgrove/path/tree/main/cmd/config/examples/config.shannon_example.yaml)
- [Example Morse Configuration File](https://github.com/buildwithgrove/path/tree/main/cmd/config/examples/config.morse_example.yaml)
- [Config YAML Schema File](https://github.com/buildwithgrove/path/tree/main/config/config.schema.yaml)

## 5. Running PATH

By default, the PATH service runs without any authorization or rate limiting. This means all requests are allowed.

To enable authorization and rate limiting, you can run the PATH service with the dependencies using the `make path_up` target.

<!-- TODO_MVP(@commoddity): Update this section to replace the docker-compose references with Local development / Tilt. -->
This will start the PATH service with all the appropriate dependencies, seen in the [docker-compose.yml](./docker-compose.yml) file, under the **Profile 2: PATH Entire Stack** section.

  > 💡 For more information about PATH's authorization and rate limiting, see the [Envoy Proxy & Auth Server README.md](./envoy/README.md).

### 5.1. Setup Config YAML

1. Run `make copy_shannon_config` or `make copy_morse_config` to prepare the `bin/config/.config.yaml` file.

   **NOTE: For a full example of the config YAML format for both Shannon and Morse protocols, see the [example config YAML files](https://github.com/buildwithgrove/path/tree/main/cmd/config/examples).**

2. You will then need to populate the `.config.yaml` file with the appropriate values for the protocol you wish to use.

   **⚠️ IMPORTANT: The data required to populate the `.config.yaml` file is sensitive and the contents of this file must never be shared outside of your organization. ⚠️**

### 5.2. Run the PATH binary

1. Once the `.config.yaml` file is populated under the `bin/config` directory, to start the PATH service for a specific protocol, use the following make target to run path:

   ```sh
   make path_run
   ```

   - All requests pass through Envoy Proxy on port `3001`
   - The PATH service runs on port `3000`

2. Once PATH is running, you may send service requests to it.

   By default, PATH will listen on port `3000`.

3. To stop the PATH instance, press Ctrl-C in the terminal from which the `make path_run` command was issued.

## 6. E2E Tests

This repository contains end-to-end (E2E) tests for the Shannon relay protocol. The tests ensure that the protocol behaves as expected under various conditions.

### 6.1. Running the E2E tests against Shannon Testnet

#### 6.1.1. Preparing the configuration

A `make` target is provided to copy the example Shannon configuration file to the `e2e/.shannon.config.yaml` needed by the E2E tests on Shannon.

```sh
make copy_shannon_e2e_config
```

Then update the `shannon_config.gateway_config` values with the appropriate values.

You can find the example Shannon configuration file [here](https://github.com/buildwithgrove/path/tree/main/e2e/shannon.example.yaml).

#### 6.1.2. Running the E2E tests

To run the tests, use the following `make` targets:

```sh
# Run E2E tests against Shannon Testnet
make test_e2e_shannon_relay

# Run all tests
make test_all
```

### 6.2. Running the E2E tests against Morse

#### 6.2.1. Preparing the configuration

A `make` target is provided to copy the example Morse configuration file to the `e2e/.morse.config.yaml` needed by the E2E tests on Morse.
To run the tests, use the following `make` targets:

```sh
make copy_morse_e2e_config
```

Then update the `morse_config.full_node_config` and `morse_config.signed_aats` values with the appropriate values.

You can find the example Morse configuration file [here](https://github.com/buildwithgrove/path/tree/main/e2e/morse.example.yaml).

#### 6.2.2. Running the E2E tests

To run the tests, use the following `make` targets:

```sh
# Run E2E tests against Morse
make test_e2e_morse_relay

# Run all tests
make test_all
```

## 7. Running Localnet

You can use path configuration under `/local` to spin up a local development environment using `Kind` + `Tilt`.

Make sure to review [Tiltfile](https://github.com/buildwithgrove/path/tree/main/Tiltfile) and [values file](https://github.com/buildwithgrove/path/tree/main/local/path/config/path-values.yaml) to make sure they have your desired configuration.

### 7.1. Spinning up / Tearing down Localnet

Localnet can be spin up/tear down using the following targets:

- `path_up` -> Spins up localnet environment using Kind + Tilt
- `path_down` -> Tears down localnet.

<!-- TODO(@adshmh): Reintroduce the Troubleshooting section once we encounter relevant issues -->

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
