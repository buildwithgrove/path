---
sidebar_position: 2
title: Shannon Cheat Sheet (30-60 min)
description: Introductory guide for setting up PATH w/ Shannon
---

_tl;dr Set up `PATH` with Pocket Network's **Shannon** protocol_

:::tip TIP: Coming from `dev.poktroll.com`?

Coming from the [App & PATH Gateway Cheat Sheet](https://dev.poktroll.com/operate/cheat_sheets/gateway_cheatsheet)?

Skip to [2.1 Generate Shannon Config](#21-generate-shannon-config).

:::

:::note NOTE: Are you a Grove 🌿 employee?

<details>

<summary>Download your configs here</summary>

1. Download the preferred config file from 1Password:
   - **[Shannon MainNet](https://start.1password.com/open/i?a=4PU7ZENUCRCRTNSQWQ7PWCV2RM&v=kudw25ob4zcynmzmv2gv4qpkuq&i=4ifsnkuifvaggwgptns6xyglsa&h=buildwithgrove.1password.com)**
   - **[Shannon Beta TestNet](https://start.1password.com/open/i?a=4PU7ZENUCRCRTNSQWQ7PWCV2RM&v=kudw25ob4zcynmzmv2gv4qpkuq&i=3treknedz5q47rgwdbreluwffu&h=buildwithgrove.1password.com)**
2. Copy to `local/path/.config.yaml` **in your PATH repository**
3. Comment out the `data_reporter_config` section
4. Comment out the `owned_apps_private_keys_hex` you're not using for testing
5. Skip to [section 3: Run PATH](#3-run-path-in-development-mode)

</details>

:::

## Table of Contents <!-- omit in toc -->

- [0. Prerequisites](#0-prerequisites)
- [1. Setup Shannon Protocol Accounts (Gateway \& Application)](#1-setup-shannon-protocol-accounts-gateway--application)
  - [1.1 Gateway and Application Account Creation](#11-gateway-and-application-account-creation)
  - [1.2 Account Validation](#12-account-validation)
- [2. Configure PATH for Shannon](#2-configure-path-for-shannon)
  - [2.1 Generate Shannon Config](#21-generate-shannon-config)
  - [2.2 Manual Configuration Verification](#22-manual-configuration-verification)
  - [2.3 Ensure onchain configuration matches](#23-ensure-onchain-configuration-matches)
- [3. Run the full PATH stack locally](#3-run-the-full-path-stack-locally)
  - [3.1 Run \& Monitor PATH](#31-run--monitor-path)
  - [3.2 Check configured services](#32-check-configured-services)
- [4. Test Relays](#4-test-relays)
  - [Test Relay with `curl`](#test-relay-with-curl)
  - [Load Testing with `relay-util`](#load-testing-with-relay-util)
- [5. Stop PATH](#5-stop-path)

## 0. Prerequisites

Complete the [**Quick Start**](1_quick_start.md) guide.

## 1. Setup Shannon Protocol Accounts (Gateway & Application)

You will need:

1. **[Gateway](https://docs.pokt.network/pokt-protocol/the-shannon-upgrade/shannon-actors/gateways)**: Facilitates relays and ensures QoS
2. **[Application](https://docs.pokt.network/pokt-protocol/the-shannon-upgrade/shannon-actors/sovereign-applications)**: Sends and pays for relays

### 1.1 Gateway and Application Account Creation

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

## 2. Configure PATH for Shannon

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

### 2.2 Manual Configuration Verification

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

### 2.3 Ensure onchain configuration matches

Verify service configuration for each application:

```bash
pocketd query application show-application \
     $(pocketd keys show -a application --keyring-backend=test) \
     --node=https://shannon-grove-rpc.mainnet.poktroll.com
```

:::

## 3. Run the full PATH stack locally

### 3.1 Run & Monitor PATH

Run the following command and wait 1-2 minutes:

```bash
make path_up
```

Your **terminal** should display the following:

![Terminal](../../../static/img/path-in-tilt-console.png)

Visit the **Tilt dashboard** at [localhost:10350](<http://localhost:10350/r/(all)/overview>) and make sure everything is 🟢.

![Tilt Dashboard](../../../static/img/path-in-tilt.png)

### 3.2 Check configured services

```bash
curl http://localhost:3070/healthz | jq
```

## 4. Test Relays

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

### Load Testing with `relay-util`

Send 100 requests with performance metrics:

```bash
SERVICE_ID=eth make test_request__shannon_relay_util_100
```

## 5. Stop PATH

When you're finished testing:

```bash
make path_down
```
