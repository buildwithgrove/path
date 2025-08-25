---
sidebar_position: 2
title: Shannon Cheat Sheet (30-60 min)
description: Introductory guide for setting up PATH w/ Shannon
---

<!-- TODO_TECHDEBT(@olshansk): Remove all instances of Shannon -->

**_tl;dr Set up `PATH` with Pocket Network's Shannon protocol_**

:::note 🌿 Are you a Grove employee 🌿?

<details>

<summary>Download your configs here</summary>

### 1. Download the shannon `.config.yaml`

For **MainNet**:

```bash
op item get 4ifsnkuifvaggwgptns6xyglsa --fields notesPlain --format json | jq -r '.value' > ./local/path/.config.yaml
```

For **Beta TestNet**:

```bash
op item get 3treknedz5q47rgwdbreluwffu --fields notesPlain --format json | jq -r '.value' > ./local/path/.config.yaml
```

### 2. Comment out unused config sections

In `./local/path/.config.yaml`:

1. Comment out the `owned_apps_private_keys_hex` you're not using for testing.
2. Comment out the `data_reporter_config` section:

   ```bash
   sed -i '' \
     -e 's/^[[:space:]]*data_reporter_config:/# data_reporter_config:/' \
     -e 's/^[[:space:]]*"target_url":/#   "target_url":/' \
     local/path/.config.yaml
   ```

### 3. Download the guard `.values.yaml`

op item get fkltz2wb7fegpumntqyo3w5qau --fields notesPlain --format json | jq -r '.value' > ./local/path/.values.yaml

### 4. Skip to Section 4

Skip to [Section 4: Run PATH](#4-run-path-stack-locally)

</details>

:::

:::tip Are you coming from `dev.poktroll.com`?

Coming from the [App & PATH Gateway Cheat Sheet](https://dev.poktroll.com/operate/cheat_sheets/gateway_cheatsheet)?

Skip to [2.1 Generate Shannon Config](#21-generate-shannon-config).

:::

## Table of Contents <!-- omit in toc -->

- [0. Prerequisites](#0-prerequisites)
- [1. Protocol Account Setup (Applications \& Gateway)](#1-protocol-account-setup-applications--gateway)
  - [1.1 Account Creation](#11-account-creation)
  - [1.2 Account Validation](#12-account-validation)
- [2. PATH Protocol Configuration (`.config.yaml`)](#2-path-protocol-configuration-configyaml)
  - [2.1 Generate Shannon Config](#21-generate-shannon-config)
  - [2.2 Verify Shannon Configuration](#22-verify-shannon-configuration)
  - [2.3 Verify onchain configuration matches](#23-verify-onchain-configuration-matches)
- [3. PATH Envoy Configuration (`.values.yaml`)](#3-path-envoy-configuration-valuesyaml)
  - [3.1 Copy the template config](#31-copy-the-template-config)
  - [3.2 Update the services](#32-update-the-services)
- [4. Run PATH Stack Locally](#4-run-path-stack-locally)
  - [4.1 Run \& Monitor PATH](#41-run--monitor-path)
  - [4.2 Check configured services](#42-check-configured-services)
- [5. Test Relays](#5-test-relays)
  - [Test Relay with `curl`](#test-relay-with-curl)
  - [Test WebSockets with `wscat`](#test-websockets-with-wscat)
  - [Load Testing Relays with `relay-util`](#load-testing-relays-with-relay-util)
  - [Load Testing WebSockets with `websocket-load-test`](#load-testing-websockets-with-websocket-load-test)
- [6. Stop PATH](#6-stop-path)

## 0. Prerequisites

⚠️ Complete the [**Getting Started**](1_getting_started.md) guide.

## 1. Protocol Account Setup (Applications & Gateway)

You will need:

1. **[Gateway](https://docs.pokt.network/pokt-protocol/the-shannon-upgrade/shannon-actors/gateways)**: Facilitates relays and ensures QoS
2. **[Application](https://docs.pokt.network/pokt-protocol/the-shannon-upgrade/shannon-actors/sovereign-applications)**: Sends and pays for relays

### 1.1 Account Creation

Choose one of the following flows to setup your accounts:

1. **Recommended Setup**: Follow the [**App & PATH Gateway Cheat Sheet**](https://dev.poktroll.com/operate/cheat_sheets/gateway_cheatsheet) for complete setup.
2. **Quick Setup** for first-time users:

     <details>

     <summary>Copy-pasta to stake onchain Application & Gateway</summary>

   **Create gateway stake config:**

   ```bash
   cat <<EOF > /tmp/stake_gateway_config.yaml
   stake_amount: 1000000upokt
   EOF
   ```

   **Create application stake config:**

   ```bash
   cat <<EOF > /tmp/stake_app_config.yaml
   stake_amount: 100000000upokt
   service_ids:
     - "anvil"
   EOF
   ```

   **Create accounts:**

   ```bash
   pocketd keys add gateway
   pocketd keys add application
   ```

   **Fund accounts**: Use faucet links [here](https://dev.poktroll.com/category/explorers-faucets-wallets-and-more).

   :::tip Grove employees only

   Fund using `pkd_beta_tx` helper ([instructions](https://www.notion.so/buildwithgrove/Shannon-Alpha-Beta-Environment-rc-helpers-152a36edfff680019314d468fad88864?pvs=4)):

   ```bash
   pkd_beta_tx bank send faucet_beta $(pocketd keys show -a application --keyring-backend=test) 6900000000042upokt
   pkd_beta_tx bank send faucet_beta $(pocketd keys show -a gateway --keyring-backend=test) 6900000000042upokt
   ```

   :::

   **Stake gateway:**

   ```bash
   pocketd tx gateway stake-gateway \
   --config=/tmp/stake_gateway_config.yaml \
   --from=gateway --gas=auto --gas-prices=200upokt --gas-adjustment=1.5 --chain-id=pocket-beta \
   --node=https://shannon-grove-rpc.mainnet.poktroll.com \
   --keyring-backend=test \
   --yes
   ```

   **Stake application:**

   ```bash
   pocketd tx application stake-application \
   --config=/tmp/stake_app_config.yaml \
   --from=application --gas=auto --gas-prices=200upokt --gas-adjustment=1.5 --chain-id=pocket-beta \
   --node=https://shannon-grove-rpc.mainnet.poktroll.com \
   --keyring-backend=test \
   --yes
   ```

   **Delegate application to gateway:**

   ```bash
   pocketd tx application delegate-to-gateway $(pocketd keys show -a gateway --keyring-backend=test) \
   --from=application --gas=auto --gas-prices=200upokt --gas-adjustment=1.5 --chain-id=pocket-beta \
   --node=https://shannon-grove-rpc.mainnet.poktroll.com \
   --keyring-backend=test \
   --yes
   ```

     </details>

### 1.2 Account Validation

Verify the keys are in your keyring:

```bash
# All accounts
pocketd keys list --keyring-backend=test

# Gateway only
pocketd keys show -a gateway --keyring-backend=test

# Application only
pocketd keys show -a application --keyring-backend=test
```

## 2. PATH Protocol Configuration (`.config.yaml`)

### 2.1 Generate Shannon Config

Generate config at `local/path/.config.yaml`:

```bash
make config_shannon_populate
```

:::important Environment variables

Override defaults if needed:

- `POCKETD_HOME`: Path to pocketd home (default `$HOME/.pocketd`)
- `POCKETD_KEYRING_BACKEND`: Keyring backend (default `test`)

:::

### 2.2 Verify Shannon Configuration

Check config:

```bash
cat local/path/.config.yaml
```

Expected format:

```yaml
shannon_config:
  full_node_config:
    rpc_url: https://shannon-grove-rpc.mainnet.poktroll.com
    grpc_config:
      host_port: shannon-grove-grpc.mainnet.poktroll.com:443
    lazy_mode: false
  gateway_config:
    gateway_mode: "centralized"
    gateway_address: pokt1... # Your gateway address
    gateway_private_key_hex: "0x..." # Your gateway private key
    owned_apps_private_keys_hex:
      - "0x..." # Your application private key
```

### 2.3 Verify onchain configuration matches

Verify service configuration for each application:

```bash
pocketd query application show-application \
     $(pocketd keys show -a application --keyring-backend=test) \
     --node=https://shannon-grove-rpc.mainnet.poktroll.com
```

:::

## 3. PATH Envoy Configuration (`.values.yaml`)

:::tip More details about configs

You can learn more about various configurations at [Auth Configs](../configs/3_auth_config.md) or [Helm Docs](./../../operate/helm/). It covers auth, rate limiting, etc...

:::

### 3.1 Copy the template config

Run the following command to create `local/path/.values.yaml`:

```bash
make configs_copy_values_yaml
```

And check the contents of `local/path/.values.yaml`:

```bash
cat local/path/.values.yaml
```

### 3.2 Update the services

Update this section with the services you want to support. For example:

```yaml
guard:
  services:
    - serviceId: eth
    - serviceId: svc1
    - serviceId: sv2
```

Make sure this reflects both of the following:

1. What your onchain application is configured to support.
2. What your gateway `.config.yaml` is configured for

## 4. Run PATH Stack Locally

### 4.1 Run & Monitor PATH

Run the following command and wait 1-2 minutes:

```bash
make path_up
```

Your **terminal** should display the following:

![Terminal](../../../static/img/path-in-tilt-console.png)

Visit the **Tilt dashboard** at [localhost:10350](<http://localhost:10350/r/(all)/overview>) and make sure everything is 🟢.

![Tilt Dashboard](../../../static/img/path-in-tilt.png)

### 4.2 Check configured services

```bash
curl http://localhost:3070/healthz | jq
```

## 5. Test Relays

### Test Relay with `curl`

Assuming you have an app staked for `eth`, you can query `eth_blockNumber`.

By specifying the `Target-Service-Id` header:

```bash
curl http://localhost:3070/v1 \
 -H "Target-Service-Id: eth" \
 -H "Authorization: test_api_key" \
 -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```

Or by using the `eth` subdomain:

```bash
curl http://eth.localhost:3070/v1 \
 -H "Authorization: test_api_key" \
 -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```

Expected response:

```json
{ "id": 1, "jsonrpc": "2.0", "result": "0x2f01a" }
```

### Test WebSockets with `wscat`

:::tip

For `wscat` installation instructions, see [here](https://github.com/ArtiomL/wscat?tab=readme-ov-file#installation).

:::

```bash
wscat -c ws://localhost:3070/v1 \
 -H "Authorization: test_api_key" \
 -H "Target-Service-Id: xrplevm"
```

Expected terminal prompt:

```bash
Connected (press CTRL+C to quit)
>
```

And subscribe to events:

```bash
> {"jsonrpc":"2.0", "id": 1, "method": "eth_subscribe", "params": ["newHeads"]}
< {"jsonrpc":"2.0","result":"0x2dc4edb4ba815232ef2d144b5818c540","id":1}
```

Which will start sending events like so:

```bash
< {"jsonrpc":"2.0","method":"eth_subscription","params":{"subscription":"0x2dc4edb4ba815232ef2d144b5818c540","result":{"parentHash":"0xaf1ebef9181d53a61a05b328646e747b5100eaa7ea301e21f2b5b1772beda053", ...
```

:::info

This is a simple terminal-based WebSocket example and does not contain reconnection logic.

Connections will drop on session rollover, which is expected behavior.

In production environments, you should implement reconnection logic and handle errors gracefully.

:::

### Load Testing Relays with `relay-util`

Make sure you install optional tools first:

```bash
make install_tools_optional
```

Send 100 requests with performance metrics:

```bash
SERVICE_ID=eth make test_request__shannon_relay_util_100
```

### Load Testing WebSockets with `websocket-load-test`

TODO_IN_THIS_PR: Update to have this point at LocalNet.

1. Make sure you install optional tools first:

   ```bash
   make install_tools_optional
   ```

2. Get your `GROVE_PORTAL_APP_ID` and `GROVE_PORTAL_API_KEY` from the [Grove's Portal](https://portal.grove.city).

3. Subscribe to events:

   ```bash
   websocket-load-test \
   --service "xrplevm" \
   --app-id $GROVE_PORTAL_APP_ID \
   --api-key $GROVE_PORTAL_API_KEY \
   --subs "newHeads,newPendingTransactions" \
   --count 10 \
   --log
   ```

## 6. Stop PATH

When you're finished testing:

```bash
make path_down
```
