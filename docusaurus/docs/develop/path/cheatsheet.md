---
sidebar_position: 1
title: PATH Cheat Sheet
---

This guide provides quick reference (i.e. a cheat sheet leveraging lots of helpers)
for setting up and running a local PATH instance in Tilt. 

If you'd like to understand all the underlying details, please refer to the [PATH Introduction](../path/introduction.md).

:::warning TODOs

1. These instructions are intended to run on a Linux machine.
   - TODO_TECHDEBT(@olshansk): Adapt the instructions to be macOS friendly.

:::

## Table of Contents <!-- omit in toc -->

- [1. Prerequisites](#1-prerequisites)
  - [1.1 Clone the `PATH` Repository](#11-clone-the-path-repository)
  - [1.2 Install Dependencies](#12-install-dependencies)
  - [1.3 Setup Protocol](#13-setup-protocol)
    - [1.3a Shannon - Setup Account](#13a-shannon---setup-account)
    - [1.3b Morse - Retrieve AATs](#13b-morse---retrieve-aats)
- [2. Populate Required Config Files](#2-populate-required-config-files)
  - [2.1 Populate the `PATH` config YAML file](#21-populate-the-path-config-yaml-file)
    - [2.1a Populate the Shannon config YAML file](#21a-populate-the-shannon-config-yaml-file)
    - [2.1b Populate the Morse config YAML file](#21b-populate-the-morse-config-yaml-file)
  - [2.2 Populate the `Envoy Proxy` config](#22-populate-the-envoy-proxy-config)
- [3. Run the `PATH` Gateway](#3-run-the-path-gateway)
  - [3a. Run `PATH` with Envoy Proxy](#3a-run-path-with-envoy-proxy)
  - [3b. Run `PATH` standalone](#3b-run-path-standalone)
- [4. View PATH Resources in Tilt](#4-view-path-resources-in-tilt)
- [5. Send a Relay](#5-send-a-relay)
  - [5.1 With Envoy Proxy](#51-with-envoy-proxy)
  - [5.2 Standalone](#52-standalone)

## 1. Prerequisites

### 1.1 Clone the `PATH` Repository

```bash
git clone https://github.com/buildwithgrove/path.git
cd ./path
```

### 1.2 Install Dependencies

The following tools are required to start a local PATH instance in Tilt:

- [Poktroll CLI](https://dev.poktroll.com/operate/user_guide/install)
- [Docker](https://docs.docker.com/get-docker/)
- [Kind](https://kind.sigs.k8s.io/#installation-and-usage)
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
- [Helm](https://helm.sh/docs/intro/install/)
- [Tilt](https://docs.tilt.dev/install.html)

:::tip

A script is provided to install the dependencies to start a PATH instance in Tilt.

```bash
make install_deps
```

This will check if the required tools are installed and install them if they are not.

:::

### 1.3 Setup Protocol

#### 1.3a Shannon - Setup Account

Before starting a PATH instance, you will need to set up both a [Gateway](https://docs.pokt.network/pokt-protocol/the-shannon-upgrade/shannon-actors/gateways) and [Application](https://docs.pokt.network/pokt-protocol/the-shannon-upgrade/shannon-actors/sovereign-applications) account on Shannon.

[For a quick and easy way to set up your Shannon accounts, see the Account Setup section of the Gateway cheat sheets](https://dev.poktroll.com/operate/quickstart/gateway_cheatsheet#account-setup).

#### 1.3b Morse - Retrieve AATs

`Application Authentication Tokens` (AATs) are auth tokens that allow application clients to access the network without the need to expose their private keys.

This is a relatively manual process in Morse that is not well documented.

You should reach out to the team directly if you are doing this, but can refer to the following resources as references:

- [What are AATs?](https://docs.pokt.network/gateways/host-a-gateway/relay-process#what-are-aats)
- [Host a Gateway on Morse](https://docs.pokt.network/gateways/host-a-gateway)
- [pocket-core/doc/specs/application-auth-token.md](https://github.com/pokt-network/pocket-core/blob/7f936ff7353249b161854e24435e4bc32d47aa3f/doc/specs/application-auth-token.md)
- [pocket-core/doc/specs/cli/apps.md](https://github.com/pokt-network/pocket-core/blob/7f936ff7353249b161854e24435e4bc32d47aa3f/doc/specs/cli/apps.md)
- [Gateway Server Kit instructions (as a reference)](https://github.com/pokt-network/gateway-server/blob/main/docs/quick-onboarding-guide.md#5-insert-app-stake-private-keys)

Once you have one or more valid AATs, you can populate the configuration files required to run the full `PATH Gateway` instance.

## 2. Populate Required Config Files

### 2.1 Populate the `PATH` config YAML file

:::tip

[For detailed information on the PATH configuration file, see the PATH Config Docs](../path/path_config.md).

:::

#### 2.1a Populate the Shannon config YAML file

Assuming you have followed the instructions above, the following should be true:

1. You have created, funded and staked a `Gateway`.
2. You have created, funded and staked a `Application`.
3. You have delegated the staked `Application` to the staked `Gateway`.

Now you can populate the configuration files required to run the full `PATH Gateway` instance.

Run the following command to generate a default Shannon config in `local/path/config/.config.yaml` using the values from your `Gateway` and `Application` accounts:

```bash
make shannon_populate_config
```

:::note Exporting private keys

You'll be prompted to confirm the `Gateway` account private key export. **Say Yes**.

:::

:::note TODO(@olshansk)

Pre-prepare a handful of apps/gateways to get users started EVEN faster.

:::

#### 2.1b Populate the Morse config YAML file

Run the following command to generate a default Morse config in `local/path/config/.config.yaml` using the values from your `Gateway` and `Application` accounts:

```bash
make config_morse_localnet
```

:::warning

You will need to manually update the `url`, `relay_signing_key`, & `signed_aats` values in the `local/path/config/.config.yaml` file for your Morse Gateway configuration.

Pay close attention to the field names and note that this file contains sensitive information.

:::

#### View the config file <!-- omit in toc -->

When you're done, run the following command to view your updated config file:

```bash
cat local/path/config/.config.yaml
```

It should look something like this:

```yaml
shannon_config:
  full_node_config:
    rpc_url: https://shannon-testnet-grove-rpc.beta.poktroll.com
    grpc_config:
      host_port: shannon-testnet-grove-grpc.beta.poktroll.com:443
    lazy_mode: false
  gateway_config:
    gateway_mode: centralized
    gateway_address: pokt1...
    gateway_private_key_hex: { REDACTED }
    owned_apps_private_keys_hex:
      - { REDACTED }
services:
  "anvil":
    alias: "eth"
```

### 2.2 Populate the `Envoy Proxy` config

:::note

You may run PATH locally in Tilt without Envoy Proxy, which disables authorization, service aliasing and rate limiting.

**If you wish to run PATH locally in Tilt without Envoy Proxy, you can skip this step.**

[Instead, proceed to Section 3b. Run `PATH` without Envoy Proxy](#3b-run-path-without-envoy-proxy).

:::


Run the following command to generate the 4 Envoy config files:

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

If you wish to use an 0Auth provider _([for example Auth0](https://auth0.com))_ to enable authorizing requests using an issued JWT, you will need to provide the `AUTH_DOMAIN` and `AUTH_AUDIENCE` values to substitute the sensitive variables in the `envoy.yaml` file.

If you do not wish to use an OAuth provider, simply answer `no` when prompted. This will allow authorizing requests with a static API key only.

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

## 4. View PATH Resources in Tilt

Regardless of which mode you choose, you should see the following output:

```bash
❯ make path_up
#########################################################################
### ./local/path/config/.config.yaml already exists, not overwriting. ###
#########################################################################
No kind clusters found.
Cluster 'path-localnet' not found. Creating it...
Creating cluster "path-localnet" ...
 ✓ Ensuring node image (kindest/node:v1.31.2) 🖼
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-path-localnet"
You can now use your cluster with:

kubectl cluster-info --context kind-path-localnet

Thanks for using kind! 😊
Switched to context "kind-path-localnet".
Checking if secret 'path-config-local' exists...
Secret 'path-config-local' not found. Creating it...
secret/path-config-local created
Tilt started on http://localhost:10350/
v0.33.21, built 2024-11-08

(space) to open the browser
(s) to stream logs (--stream=true)
(t) to open legacy terminal mode (--legacy=true)
(ctrl-c) to exit
```

You can visit [localhost:10350](http://localhost:10350) in your browser to view the Tilt dashboard, which allows you to view the log output for all running containers.

:::info

The `PATH Gateway` stack may take a minute or more to initialize the first time you run it as it must download all required Docker images.

You will be able to tell it is ready when you see log output like this in the `path` Resource in the Tilt dashboard:

```json
{"level":"info","message":"Starting PATH using config file: /app/config/.config.yaml"}
{"level":"info","message":"Starting PATH gateway with Shannon protocol"}
{"level":"info","message":"Starting the cache update process."}
{"level":"info","package":"router","message":"PATH gateway running on port 3069"}
{"level":"info","services count":1,"message":"Running Hydrator"}
```

:::

**Once the `PATH Gateway` container is ready, you can send a relay to test.**

## 5. Send a Relay

Check that the `PATH Gateway` is serving relays by running the following command yourself:

### 5.1 With Envoy Proxy

When running PATH with authorization, requests are routed through Envoy Proxy, running on `port 3001`.

**1. Endpoint with Static Key Authorization**

    This endpoint requires an API key in the `authorization` header.

    ```bash
    curl http://localhost:3001/v1/endpoint_1_static_key \
        -X POST \
        -H "authorization: api_key_1" \
        -H "target-service-id: anvil" \
        -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
    ```

**2. Endpoint with No Auth Required**

This endpoint does not require an API key in the `authorization` header.

```bash
curl http://localhost:3001/v1/endpoint_3_no_auth \
    -X POST \
    -H "target-service-id: anvil" \
    -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```

:::tip

The `GatewayEndpoint` IDs `endpoint_1_static_key` and `endpoint_3_no_auth`, as well as the API key `api_key_1` are defined in the `local/path/envoy/.gateway-endpoints.yaml` file.

To add or modify the `GatewayEndpoints` that are authorized to use your PATH instance, you may modify this file.

Saving this file will trigger a hot-reloading of the PATH Auth Data Server (PADS) resource in Tilt.

[For detailed information on the `GatewayEndpoint` data structure, including how to use a Postgres database for storing `GatewayEndpoints`, see the PATH Auth Data Server section of the PATH Auth README.md](../).

:::

### 5.2 Standalone

When running PATH without authorization, requests go directly to the `PATH` Service, running on `port 3069`.

```bash
curl http://localhost:3069/v1/ \
    -X POST \
    -H "target-service-id: anvil" \
    -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```

:::warning

Requests MAY hit unresponsive nodes. If that happens, keep retrying the request a few times.

Once `PATH`s QoS module is mature, this will be handled automatically.

:::
