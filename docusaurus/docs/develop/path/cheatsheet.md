---
sidebar_position: 1
title: Cheat Sheet
description: Quick reference guide for setting up and running a local PATH instance in Tilt
---

This guide provides quick reference (i.e. a cheat sheet leveraging lots of helpers)
for setting up and running a local PATH instance in Tilt. If you'd like to understand
all the underlying details, please refer to the [PATH Walkthrough](../path/walkthrough.md).

## Table of Contents <!-- omit in toc -->

- [1. Prerequisites](#1-prerequisites)
  - [1.1 Clone the `PATH` Repository](#11-clone-the-path-repository)
  - [1.2 Install Dependencies](#12-install-dependencies)
  - [1.3 Setup Protocol Accounts, Secrets and Configurations](#13-setup-protocol-accounts-secrets-and-configurations)
    - [1.3a `Shannon` - Setup Gateway \& Application Accounts](#13a-shannon---setup-gateway--application-accounts)
    - [1.3b `Morse` - Setup \& Retrieve AATs](#13b-morse---setup--retrieve-aats)
- [2. Prepare Required Config Files](#2-prepare-required-config-files)
  - [2.1 Preparing `PATH` config YAML file](#21-preparing-path-config-yaml-file)
    - [2.1a `Shannon` PATH Configs](#21a-shannon-path-configs)
    - [2.1b `Morse` PATH Configs](#21b-morse-path-configs)
    - [2.2 Inspect `local/path/config/.config.yaml`](#22-inspect-localpathconfigconfigyaml)
  - [2.2 Populate the `Envoy Proxy` config](#22-populate-the-envoy-proxy-config)
- [3. Run the `PATH` Gateway](#3-run-the-path-gateway)
  - [3a. Run `PATH` with Envoy Proxy](#3a-run-path-with-envoy-proxy)
  - [3b. Run `PATH` standalone](#3b-run-path-standalone)
- [4. View `PATH` Resources in Tilt](#4-view-path-resources-in-tilt)
  - [4.1 Wait for the `PATH` stack to initialize](#41-wait-for-the-path-stack-to-initialize)
- [5. Send a Relay](#5-send-a-relay)
  - [5.1a `PATH` with Envoy Proxy](#51a-path-with-envoy-proxy)
    - [5.1a.1 Endpoint with Static Key Authorization](#51a1-endpoint-with-static-key-authorization)
    - [5.1a.2 Endpoint with No Auth Required](#51a2-endpoint-with-no-auth-required)
    - [5.1a.3 Configuring Relay Authorization](#51a3-configuring-relay-authorization)
  - [5.1b `PATH` standalone](#51b-path-standalone)

## 1. Prerequisites

### 1.1 Clone the `PATH` Repository

```bash
git clone https://github.com/buildwithgrove/path.git
cd ./path
```

### 1.2 Install Dependencies

The following tools are required to start a local PATH instance in Tilt:

- [**Poktroll CLI**](https://dev.poktroll.com/operate/user_guide/install): CLI for interacting with Poktroll (Pocket Network Shannon Upgrade)
- [**Docker**](https://docs.docker.com/get-docker/): Container runtime
- [**Kind**](https://kind.sigs.k8s.io/#installation-and-usage): Local Kubernetes cluster
- [**kubectl**](https://kubernetes.io/docs/tasks/tools/#kubectl): CLI for interacting with Kubernetes
- [**Helm**](https://helm.sh/docs/intro/install/): Package manager for Kubernetes
- [**Tilt**](https://docs.tilt.dev/install.html): Local Kubernetes development environment

:::tip

A script is provided to install the dependencies to start a PATH instance in Tilt.

It will check if the required tools are installed and install them if they are not.

```bash
make install_deps
```

:::

### 1.3 Setup Protocol Accounts, Secrets and Configurations

You can choose to use either one or both of the protocols PATH supports:

1. **Shannon** (v1): The upgrade to Pocket Network protocol; in Beta TestNet as of 01/2025.
2. **Morse** (v0): The original Pocket Network protocol; in MainNet as of 2020.

#### 1.3a `Shannon` - Setup Gateway & Application Accounts

:::tip

You can reference the [Gateway cheat sheets](https://dev.poktroll.com/operate/quickstart/gateway_cheatsheet)
for a quick and easy way to set up your Shannon accounts.

:::

Before starting a PATH instance, you will need to create an configure:

1. An onchain [Gateway](https://docs.pokt.network/pokt-protocol/the-shannon-upgrade/shannon-actors/gateways)
2. An onchain [Application](https://docs.pokt.network/pokt-protocol/the-shannon-upgrade/shannon-actors/sovereign-applications) account on Shannon.

#### 1.3b `Morse` - Setup & Retrieve AATs

:::caution

This is a manual and poorly documented process in Morse.

:::

`Application Authentication Tokens` (**AATs**) are auth tokens that allow application
clients to access the network without the need to expose their private keys.
Once you have one or more valid AATs, you can populate the configuration files required to run the full `PATH Gateway` instance.

The following resources are also good references and starting points:

- [What are AATs?](https://docs.pokt.network/gateways/host-a-gateway/relay-process#what-are-aats)
- [Host a Gateway on Morse](https://docs.pokt.network/gateways/host-a-gateway)
- [pocket-core/doc/specs/application-auth-token.md](https://github.com/pokt-network/pocket-core/blob/7f936ff7353249b161854e24435e4bc32d47aa3f/doc/specs/application-auth-token.md)
- [pocket-core/doc/specs/cli/apps.md](https://github.com/pokt-network/pocket-core/blob/7f936ff7353249b161854e24435e4bc32d47aa3f/doc/specs/cli/apps.md)
- [Gateway Server Kit instructions (as a reference)](https://github.com/pokt-network/gateway-server/blob/main/docs/quick-onboarding-guide.md#5-insert-app-stake-private-keys)

_If you are unsure of where to start, you should reach out to the team directly._

## 2. Prepare Required Config Files

There are two config files that must be prepared for `PATH` operation:

1. The `PATH` config YAML file
2. The `Envoy Proxy` config YAML file

### 2.1 Preparing `PATH` config YAML file

:::tip In Depth Config Docs

A detailed explanation of the `PATH` config YAML file is provided in the [PATH Config Docs](../path/path_config.md).

:::

#### 2.1a `Shannon` PATH Configs

<!-- TODO_IMPROVE(@olshansk): Pre-prepare a handful of apps/gateways to get users started EVEN faster. -->

Assuming you have keys with the names `gateway` and `application` in your keyring
after following the instructions above, you can run the following command to generate
a Shannon config at `local/path/config/.config.yaml`:

```bash
make shannon_populate_config
```

#### 2.1b `Morse` PATH Configs

Run the following command to generate a default Morse config in `local/path/config/.config.yaml`
using the values from your `Gateway` and `Application` accounts:

```bash
make config_morse_localnet
```

:::warning

You will need to manually update the `url`, `relay_signing_key`, & `signed_aats`
values in the `local/path/config/.config.yaml` file for your Morse Gateway configuration.

Pay close attention to the field names and note that this file contains sensitive information.

:::

#### 2.2 Inspect `local/path/config/.config.yaml`

When you're done configuring either `Shannon` or `Morse`, run the following command to view your updated config file:

```bash
cat local/path/config/.config.yaml
```

For `Shannon`, it should look something like this:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/buildwithgrove/path/refs/heads/main/config/config.schema.yaml
shannon_config:
  full_node_config:
    rpc_url: https://shannon-testnet-grove-rpc.beta.poktroll.com
    grpc_config:
      host_port: shannon-testnet-grove-grpc.beta.poktroll.com:443
  gateway_config:
    gateway_mode: "centralized"
    gateway_address: pokt1... # bech32 address of your gateway account
    gateway_private_key_hex: "0x" + 64_hex_chars # 32-byte hex-encoded private key
    owned_apps_private_keys_hex: # Array of 32-byte hex-encoded private keys
      - "0x" + 64_hex_chars # 32-byte hex-encoded private key
hydrator_config:
  service_ids:
    - "anvil"
auth_server_config:
  grpc_host_port: path-auth-data-server:50051
  grpc_use_insecure_credentials: true
  endpoint_id_extractor_type: url_path
```

### 2.2 Populate the `Envoy Proxy` config

:::note

**If you want to run PATH locally in Tilt WITHOUT Envoy Proxy, you can SKIP this step. Proceed to [Section 3b](#3b-run-path-standalone).**

You can run PATH locally in Tilt without Envoy Proxy, which disables authorization, service aliasing and rate limiting.

:::

Run the following command and choose to generate the 4 Envoy config files:

```bash
make init_envoy
```

You can view them by running the following command:

```bash
cat local/path/envoy/.allowed-services.lua
cat local/path/envoy/.envoy.yaml
cat local/path/envoy/.ratelimit.yaml
cat local/path/envoy/.gateway-endpoints.yaml
```

:::tip

For detailed information on the Envoy configuration files, see the [Envoy Config Docs](../envoy/envoy_config.md).

:::

:::note

When running `make init_envoy`, we recommend choosing Option 2 (no authorization) for now.

If you wish to use an `0Auth` provider _([for example Auth0](https://auth0.com))_ to enable
authorizing requests using an issued JWT, you will need to provide the `AUTH_DOMAIN` and
`AUTH_AUDIENCE` values to substitute the sensitive variables in the `envoy.yaml` file.

If you do not wish to use an OAuth provider, simply answer `no` when prompted.
This will allow authorizing requests with a static API key only.

:::

## 3. Run the `PATH` Gateway

PATH can be run in Tilt in two different modes:

1. **With Envoy Proxy**
   - **This is the default mode; we recommend running PATH in this mode.**
   - Requests to PATH require a Gateway Endpoint in order to be authorized.
   - Requests to PATH are routed through Envoy Proxy, running on `port 3001`.
2. **Standalone**
   - Requests to PATH do not require a Gateway Endpoint to be authorized.
   - Requests to PATH go directly to the `PATH` Service, running on `port 3069`.

### 3a. Run `PATH` with Envoy Proxy

:::warning

In order to run PATH with authorization, you must have completed [Section 2.2 of this guide](#22-populate-the-envoy-proxy-config-files).

Without the configuration files created in this section, Envoy Proxy and all other required resources will not be able to start.

:::

Start the `PATH` Gateway with Envoy Proxy and all other required resources by running the following command:

```bash
make path_up
```

### 3b. Run `PATH` standalone

To run PATH without Envoy Proxy, you can use the following command:

```bash
make path_up_standalone
```

## 4. View `PATH` Resources in Tilt

Regardless of which mode you choose, you should see the output below and can
visit [localhost:10350](<http://localhost:10350/r/(all)/overview>) in your browser
to view the Tilt dashboard.

```bash
❯ make path_up
#########################################################################
### ./local/path/config/.config.yaml already exists, not overwriting. ###
#########################################################################
No kind clusters found.
Cluster 'path-localnet' not found. Creating it...
Creating cluster "path-localnet" ...
# ...
Set kubectl context to "kind-path-localnet"
You can now use your cluster with:
# ...
kubectl cluster-info --context kind-path-localnet
# ...
(space) to open the browser
(s) to stream logs (--stream=true)
(t) to open legacy terminal mode (--legacy=true)
(ctrl-c) to exit
```

### 4.1 Wait for the `PATH` stack to initialize

The `PATH Gateway` stack may take a minute or more to initialize the first time
you run it as it must download all required Docker images.

You will be able to tell it is ready when you see log output like this in the
[`path`](http://localhost:10350/r/path/overview) resource in the Tilt dashboard:

```json
{"level":"info","message":"Starting PATH using config file: /app/config/.config.yaml"}
{"level":"info","message":"Starting PATH gateway with Shannon protocol"}
{"level":"info","message":"Starting the cache update process."}
{"level":"info","package":"router","message":"PATH gateway running on port 3069"}
{"level":"info","services count":1,"message":"Running Hydrator"}
```

## 5. Send a Relay

You can verify that servicing relays works by sending one yourself!command yourself:

:::warning

Requests MAY hit unresponsive nodes. If that happens, keep retrying the request a few times.

Once `PATH`s QoS module is mature, this will be handled automatically.

:::

### 5.1a `PATH` with Envoy Proxy

Authorized relays are routed through Envoy Proxy running on port `3001`.

#### 5.1a.1 Endpoint with Static Key Authorization

This endpoint requires an API key in the `authorization` header.

```bash
curl http://localhost:3001/v1/endpoint_1_static_key \
    -X POST \
    -H "authorization: api_key_1" \
    -H "target-service-id: anvil" \
    -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```

#### 5.1a.2 Endpoint with No Auth Required

This endpoint does not require an API key in the `authorization` header.

```bash
curl http://localhost:3001/v1/endpoint_3_no_auth \
    -X POST \
    -H "target-service-id: anvil" \
    -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```

#### 5.1a.3 Configuring Relay Authorization

:::tip

Saving `.gateway-endpoints.yaml` triggers a hot-reload of PADS (PATH Auth Data Server) in Tilt.

:::

You can view the `GatewayEndpoint`s and update `local/path/envoy/.gateway-endpoints.yaml` to configure authorization for your relays.

- `endpoint_1_static_key` requires an API key in the `authorization` header set to `api_key_1` by default.
- `endpoint_3_no_auth` does not require an API key in the `authorization` header.

For detailed information on the `GatewayEndpoint` data structure, including how to use a Postgres database for storing `GatewayEndpoints`, see the PATH Auth Data Server section of the [PATH Config Docs](../path/path_config.md).

### 5.1b `PATH` standalone

Unauthorized relays are routed directly to the `PATH` Service, running on `port 3069`.

```bash
curl http://localhost:3069/v1/ \
    -X POST \
    -H "target-service-id: anvil" \
    -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```
