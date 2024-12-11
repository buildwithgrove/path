---
sidebar_position: 2
title: PATH Cheat Sheet
---

# PATH Cheat Sheet <!-- omit in toc -->

This guide provides quick reference commands for setting up and running a local PATH
instance in Tilt.

:::warning

These instructions are intended to run on a Linux machine.

TODO_TECHDEBT(@commoddity): Adapt the instructions to be macOS friendly.

:::

:::info

The following instructions are specific to setting up a PATH instance on Shannon.

:::

## Table of Contents <!-- omit in toc -->

- [1. Pre-Requisites](#1-pre-requisites)
  - [1.1 Clone the `PATH` Repository](#11-clone-the-path-repository)
  - [1.2 Install Dependencies](#12-install-dependencies)
  - [1.3 Setup Shannon Account](#13-setup-shannon-account)
- [2. Populate Required Config Files](#2-populate-required-config-files)
  - [2.1 Populate the `PATH` config YAML file](#21-populate-the-path-config-yaml-file)
  - [2.2 Populate the `Envoy Proxy` config files](#22-populate-the-envoy-proxy-config-files)
- [3. Run the `PATH` Gateway](#3-run-the-path-gateway)
- [4. Send a Relay](#4-send-a-relay)

## 1. Pre-Requisites

### 1.1 Clone the `PATH` Repository

```bash
mkdir -p ~/workspace
cd ~/workspace
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


### 1.3 Setup Shannon Account

Before starting a PATH instance, you will need to set up both a [Gateway](https://docs.pokt.network/pokt-protocol/the-shannon-upgrade/shannon-actors/gateways) and [Application](https://docs.pokt.network/pokt-protocol/the-shannon-upgrade/shannon-actors/sovereign-applications) account on Shannon.

For a quick and easy way to set up your Shannon account, see [the Account Setup section of the Gateway Cheetsheet](https://dev.poktroll.com/operate/quickstart/gateway_cheatsheet).

:::tip

If you ran `make install_deps`, you will already have the `poktrolld` CLI installed.

:::

## 2. Populate Required Config Files

Assuming you have followed the instructions in [the Account Setup section of the Gateway Cheetsheet](https://dev.poktroll.com/operate/quickstart/gateway_cheatsheet), the following should be true:

1. You have created, funded and staked a `Gateway`.
2. You have created, funded and staked a `Application`.
3. You have delegated the staked `Application` to the staked `Gateway`.

Now you can populate the configuration files required to run the full `PATH Gateway` instance.

### 2.1 Populate the `PATH` config YAML file

Run the following command to generate a default Shannon config in `local/path/config/.config.yaml` using the values from your `Gateway` and `Application` accounts:

_NOTE: You'll be prompted to confirm the `Gateway` account private key export._

```bash
make populate_shannon_config
```

When you're done, run `cat local/path/config/.config.yaml` to view the updated config file.

### 2.2 Populate the `Envoy Proxy` config files

Run the following command to generate the 3 Envoy config files in `local/path/envoy`.

- `.envoy.yaml`
- `.ratelimit.yaml`
- `.gateway-endpoints.yaml`

```bash
make init_envoy
```

:::tip

If you wish to use an 0Auth provider _(for example [Auth0](https://auth0.com))_ to enable authorizing requests using an issued JWT, you will need to provide the `AUTH_DOMAIN` and `AUTH_AUDIENCE` values to substitute the sensitive variables in the `envoy.yaml` file.

If you do not wish to use an OAuth provider, simply answer `no` when prompted. This will allow authorizing requests with a static API key only.

:::

<!-- TODO_TECHDEBT(@commoddity): Update the `make init_envoy` script to remove the JWT HTTP filter if the operator selects to use a static API key only. -->

## 3. Run the `PATH` Gateway

```bash
make path_up
```

This will start the `PATH` Gateway.

You should see the following output:

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

You can visit http://localhost:10350 in your browser to view the Tilt dashboard, which allows you to view the log output for all running containers.

:::note

The `PATH Gateway` stack may take a minute or more to initialize the first time you run it as it must download all required Docker images. 

You will be able to tell it is ready when you see log output like this in the `path` Resource in the  Tilt dashboard:
```json
{"level":"info","message":"Starting PATH gateway with Shannon protocol"}
{"level":"info","message":"Starting the cache update process."}
{"level":"info","package":"router","message":"PATH gateway running on port 3000"}
{"level":"info","services count":1,"message":"Running Hydrator"}
```
:::

Once the `PATH Gateway` container is ready, you may send a relay to test. 

## 4. Send a Relay

Check that the `PATH Gateway` is serving relays by running the following command yourself:

```bash
curl http://eth.localhost:3001/v1/endpoint_1 \
    -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: api_key_1" \
    -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```

:::tip

The `GatewayEndpoint` ID `endpoint_1` and API Key `api_key_1` are defined in the `local/path/envoy/.gateway-endpoints.yaml` file.

To add or modify the `GatewayEndpoints` that are authorized to use your PATH instance, you may modify this file. 

Saving this file will trigger a hot-reloading of the PATH Auth Data Server (PADS) resource in Tilt.

For detailed information on the `GatewayEndpoint` data structure, including how to use a Postgres database for storing `GatewayEndpoints`, [see the PATH Auth Data Server section of the PATH Auth README.md](https://github.com/buildwithgrove/path/tree/main/envoy#55-remote-grpc-auth-server).

:::

:::warning

Requests MAY hit unresponsive nodes. If that happens, keep retrying the request a few times.

Once `PATH`s QoS module is mature, this will be handled automatically.

:::
