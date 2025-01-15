---
sidebar_position: 5
title: Morse Cheat Sheet
description: Quick reference guide for setting up PATH with Morse protocol
---

This guide covers setting up `PATH` with the **Morse** protocol. In MainNet as of 01/2025.

## Table of Contents <!-- omit in toc -->

- [0. Prerequisites](#0-prerequisites)
- [1. Setup Morse Protocol Accounts](#1-setup-morse-protocol-accounts)
- [2. Configure PATH](#2-configure-path)
  - [Generate Morse Config](#generate-morse-config)
  - [Update Configuration](#update-configuration)
  - [Verify Configuration](#verify-configuration)
- [3. Setup Envoy Proxy (Optional)](#3-setup-envoy-proxy-optional)
- [4. Start PATH](#4-start-path)
  - [With Envoy Proxy (Recommended)](#with-envoy-proxy-recommended)
  - [Without Envoy Proxy](#without-envoy-proxy)
- [5. Monitor PATH](#5-monitor-path)
- [6. Test Relays](#6-test-relays)
  - [With Envoy Proxy](#with-envoy-proxy)
  - [Without Envoy Proxy](#without-envoy-proxy-1)
- [Additional Notes](#additional-notes)

## 0. Prerequisites

1. Prepare your environment by following the instructions in the [main cheat sheet](cheat_sheet.md).
2. Install the [**Pocket CLI**](https://github.com/pokt-network/homebrew-pocket-core): CLI for interacting with Pocket's Morse Network

## 1. Setup Morse Protocol Accounts

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

## 2. Configure PATH

### Generate Morse Config

Generate a default Morse configuration:

```bash
make copy_morse_e2e_config
```

This creates a config file at `local/path/config/.config.yaml`.

### Update Configuration

You'll need to manually update these fields in the config file:

- `url`
- `relay_signing_key`
- `signed_aats`

### Verify Configuration

Check your updated config:

```bash
cat local/path/config/.config.yaml
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
{"level":"info","message":"Starting PATH using config file: /app/config/.config.yaml"}
{"level":"info","message":"Starting PATH gateway with Morse protocol"}
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

## Additional Notes

- The Morse protocol setup process is more manual than Shannon
- Pay close attention to field names in configuration files
- Keep your configuration secure as it contains sensitive information
- For troubleshooting, consult the [Gateway Server Kit instructions](https://github.com/pokt-network/gateway-server/blob/main/docs/quick-onboarding-guide.md#5-insert-app-stake-private-keys)
