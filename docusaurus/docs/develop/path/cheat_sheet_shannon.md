---
sidebar_position: 4
title: Shannon Cheat Sheet
description: Quick reference guide for setting up PATH with Shannon protocol
---

This guide covers setting up `PATH` with the **Shannon** protocol. In Beta TestNet as of 01/2025.

## Table of Contents <!-- omit in toc -->

- [0. Prerequisites](#0-prerequisites)
- [1. Setup Shannon Protocol Accounts (Gateway \& Application)](#1-setup-shannon-protocol-accounts-gateway--application)
- [2. Configure PATH](#2-configure-path)
  - [Generate Shannon Config](#generate-shannon-config)
  - [Verify Configuration](#verify-configuration)
- [3. Setup Envoy Proxy (Optional)](#3-setup-envoy-proxy-optional)
- [4. Start PATH](#4-start-path)
  - [With Envoy Proxy (Recommended)](#with-envoy-proxy-recommended)
  - [Without Envoy Proxy](#without-envoy-proxy)
- [5. Monitor PATH](#5-monitor-path)
- [6. Test Relays](#6-test-relays)
  - [With Envoy Proxy](#with-envoy-proxy)
  - [Without Envoy Proxy](#without-envoy-proxy-1)

## 0. Prerequisites

1. Prepare your environment by following the instructions in the [**main cheat sheet**](cheat_sheet.md).
2. Install the [**Poktroll CLI**](https://dev.poktroll.com/operate/user_guide/install): CLI for interacting with Pocket's Shannon Network

## 1. Setup Shannon Protocol Accounts (Gateway & Application)

Before starting, you'll need to create and configure:

1. An onchain [**Gateway**](https://docs.pokt.network/pokt-protocol/the-shannon-upgrade/shannon-actors/gateways): An onchain actor that **facilitates** _(i.e. proxies)_ relays to ensure Quality of Service (QoS)
2. An onchain [**Application**](https://docs.pokt.network/pokt-protocol/the-shannon-upgrade/shannon-actors/sovereign-applications): An onchain actor that **pays** _(i.e. the API key holder)_ for relays

:::tip

You can follow the [Gateway cheat sheets](https://dev.poktroll.com/operate/quickstart/gateway_cheatsheet) for quick setup.

:::

## 2. Configure PATH

### Generate Shannon Config

Assuming you have keys with the names `gateway` and `application` in your keyring
after following the instructions above, you can run the following command to generate
a Shannon config at `local/path/config/.config.yaml`:

```bash
make shannon_populate_config
```

:::note TODO_IMPROVE(@olshansk)

Pre-prepare a handful of apps/gateways to get users started EVEN faster. -->

:::

This generates a config file at `local/path/config/.config.yaml`.

### Verify Configuration

Check your config file:

```bash
cat local/path/config/.config.yaml
```

It should look similar to:

```yaml
shannon_config:
  full_node_config:
    rpc_url: https://shannon-testnet-grove-rpc.beta.poktroll.com
    grpc_config:
      host_port: shannon-testnet-grove-grpc.beta.poktroll.com:443
  gateway_config:
    gateway_mode: "centralized"
    gateway_address: pokt1... # Your gateway address
    gateway_private_key_hex: "0x..." # Your gateway private key
    owned_apps_private_keys_hex:
      - "0x..." # Your application private key
hydrator_config:
  service_ids:
    - "anvil"
auth_server_config:
  grpc_host_port: path-auth-data-server:50051
  grpc_use_insecure_credentials: true
  endpoint_id_extractor_type: url_path
```

## 3. Setup Envoy Proxy (Optional)

If you want to use authorization, service aliasing, and rate limiting:

```bash
make init_envoy
```

This generates four configuration files:

- `.allowed-services.lua`
- `.envoy.yaml`
- `.ratelimit.yaml`
- `.gateway-endpoints.yaml`

For initial setup, choose Option 2 (no authorization) when prompted.

## 4. Start PATH

### With Envoy Proxy (Recommended)

```bash
make path_up
```

### Without Envoy Proxy

```bash
make path_up_standalone
```

## 5. Monitor PATH

Visit [localhost:10350](<http://localhost:10350/r/(all)/overview>) to view the Tilt dashboard.

Wait for initialization logs:

```json
{"level":"info","message":"Starting PATH gateway with Shannon protocol"}
{"level":"info","message":"Starting the cache update process."}
{"level":"info","package":"router","message":"PATH gateway running on port 3069"}
```

## 6. Test Relays

### With Envoy Proxy

Using static key authorization:

```bash
curl http://localhost:3001/v1/endpoint_1_static_key \
    -X POST \
    -H "authorization: api_key_1" \
    -H "target-service-id: anvil" \
    -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```

No authorization required:

```bash
curl http://localhost:3001/v1/endpoint_3_no_auth \
    -X POST \
    -H "target-service-id: anvil" \
    -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```

### Without Envoy Proxy

```bash
curl http://localhost:3069/v1/ \
    -X POST \
    -H "target-service-id: anvil" \
    -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```

If requests fail, retry a few times as you may hit unresponsive nodes.
