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
  - [4.2 Example Shannon Configuration Format](#42-example-shannon-configuration-format)
  - [4.3 Example Morse Configuration Format](#43-example-morse-configuration-format)
  - [4.4 Other Examples](#44-other-examples)
- [5. Running PATH](#5-running-path)
  - [5.1. Setup Config YAML](#51-setup-config-yaml)
  - [5.2. Start the Container](#52-start-the-container)
- [6. E2E Tests](#6-e2e-tests)
  - [6.1. Running Tests](#61-running-tests)
- [8. Running Localnet](#8-running-localnet)
  - [8.1 Spinning up / Tearing down Localnet](#81-spinning-up--tearing-down-localnet)
- [9. Troubleshooting](#9-troubleshooting)
  - [9.1. Docker Permissions Issues - Need to run sudo?](#91-docker-permissions-issues---need-to-run-sudo)
- [Special Thanks](#special-thanks)
- [License](#license)

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

**Development only:**

- [SQLC](https://docs.sqlc.dev/)
- [Mockgen](https://github.com/uber-go/mock)

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

2. **Populate Config File:** Run `make copy_shannon_config` to copy the example configuration file to `cmd/.config.yaml`.

   Update the configuration file `cmd/.config.yaml` with your Gateway's private key & address and your delegated Application's address.

   \*TIP: If you followed the [Debian Cheat Sheet](https://dev.poktroll.com/operate/quickstart/docker_compose_debian_cheatsheet#start-the-relayminer), you can run `path_prepare_config`
   to get you most of the way there. Make sure to review the `gateway_private_key` field.\*

3. **Start the PATH Container:** Run `make path_up_build_gateway` or `make path_up_gateway` to start & build the PATH gateway.

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

2. **Populate Config File:** Run `make copy_morse_config` to copy the example configuration file to `cmd/.config.yaml`.

   Update the configuration file `cmd/.config.yaml` with your Gateway's private key, address and your delegated Application's address.

   2.1 **If you're a Grove employee**, you can use copy-paste the PROD configs from [here](https://www.notion.so/buildwithgrove/PATH-Morse-Configuration-Helpers-Instructions-111a36edfff6807c8846f78244394e74?pvs=4).

   2.2 **If you're a community member**, run the following command to get started quickly with a prefilled configuration
   for Bitcoin MainNet on Pocket Morse TestNet: `cp ./cmd/.config.morse_example_testnet.yaml ./cmd/.config.yaml`

3. **Start the PATH Container:** Run `make path_up_build_gateway` or `make path_up_gateway` to start & build PATH gateway

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

### 4.2 Example Shannon Configuration Format

```yaml
shannon_config:
  full_node_config:
    rpc_url: "https://rpc-url.io"
    grpc_config:
      host_port: "grpc-url.io:443"
    gateway_address: "pokt1710ed9a8d0986d808e607c5815cc5a13f15dba"
    gateway_private_key: "d5fcbfb894059a21e914a2d6bf1508319ce2b1b8878f15aa0c1cdf883feb018d"
    delegated_app_addresses:
      - "pokt1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p7q8r9s0"
      - "pokt1u2v3w4x5y6z7a8b9c0d1e2f3g4h5i6j7k8l9m0"

services:
  "F00C":
    alias: "eth"
```

### 4.3 Example Morse Configuration Format

```yaml
# For a morse gateway, the following config is required:
morse_config:
  full_node_config:
    url: "https://pocket-network-full-full-node-url.io"
    relay_signing_key: "example_relay_signing_key"
    http_config:
      retries: 3
      timeout: "5000ms"
    request_config:
      retries: 3

  signed_aats:
    "example_application_address":
      client_public_key: "example_application_client_public_key"
      application_public_key: "example_application_public_key"
      application_signature: "example_application_signature"

# services is required. At least one service must be configured with a valid id.
# All fields are optional but the id is required.
services:
  "F00C":
    alias: "eth"
    request_timeout: "3000ms"
```

### 4.4 Other Examples

- Full example config YAML files:
  - [Morse](https://github.com/buildwithgrove/path/tree/main/cmd/config/testdata/morse.example.yaml)
  - [Shannon](https://github.com/buildwithgrove/path/tree/main/cmd/config/testdata/shannon.example.yaml)
- [Config YAML Schema](https://github.com/buildwithgrove/path/tree/main/config/config.schema.yaml)

## 5. Running PATH

### 5.1. Setup Config YAML

1. Run `make copy_shannon_config` or `make copy_morse_config` to prepare the `.config.yaml` file.

   **NOTE: For a full example of the config YAML format for both Shannon and Morse protocols, see the [example config YAML files](https://github.com/buildwithgrove/path/tree/main/cmd/config/testdata).**

2. You will then need to populate the `.config.yaml` file with the appropriate values for the protocol you wish to use.

   **⚠️ IMPORTANT: The data required to populate the `.config.yaml` file is sensitive and the contents of this file must never be shared outside of your organization. ⚠️**

### 5.2. Start the Container

1. Once the `.config.yaml` file is populated, to start the PATH service for a specific protocol, use the `make` target:

   ```sh
   make path_up
   ```

   **NOTE: The protocol version (`morse` or `shannon`) depends on whether `morse_config` or `shannon_config` is populated in the `.config.yaml` file.**

2. Once the Docker container is running, you may send service requests to the PATH service.

   By default, the PATH service will run on port `3000`.

3. To stop the PATH service, use the following `make` target:

   ```sh
   make path_down
   ```

## 6. E2E Tests

This repository contains end-to-end (E2E) tests for the Shannon relay protocol. The tests ensure that the protocol behaves as expected under various conditions.

To use E2E tests, a `make` target is provided to copy the example configuration file to the `.config.test.yaml` needed by the E2E tests:

```sh
make copy_test_config
```

Then update the `protocol.shannon_config.full_node_config` values with the appropriate values.

You can find the example configuration file [here](https://github.com/buildwithgrove/path/tree/main/e2e/.example.test.yaml).

Currently, the E2E tests are configured to run against the Shannon testnet.

Future work will include adding support for other protocols.

### 6.1. Running Tests

To run the tests, use the following `make` targets:

```sh
# Run all tests
make test_all

# Unit tests only
make test_unit

# Shannon E2E test only
make test_e2e_shannon_relay
```

## 8. Running Localnet

You can use path configuration under `/local` to spin up a local development environment using `Kind` + `Tilt`.

Make sure to review [Tiltfile](https://github.com/buildwithgrove/path/tree/main/Tiltfile) and [values file](https://github.com/buildwithgrove/path/tree/main/local/path/config/path-values.yaml) to make sure they have your desired configuration.

### 8.1 Spinning up / Tearing down Localnet

Localnet can be spin up/tear down using the following targets:

- `localnet_up` -> Spins up localnet environment using Kind + Tilt
- `localnet_down` -> Tears down localnet.

## 9. Troubleshooting

### 9.1. Docker Permissions Issues - Need to run sudo?

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
