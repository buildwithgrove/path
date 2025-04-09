---
sidebar_position: 3
title: Shannon Cheat Sheet (30-60 min)
description: Introductory guide for setting up PATH w/ Shannon
---

This guide covers setting up `PATH` with Pocket Network's **Shannon** protocol.

Shannon is in Beta TestNet as of 01/2025 and private MainNet as of 04/2025.

## Table of Contents <!-- omit in toc -->

- [0. Prerequisites](#0-prerequisites)
- [1. Setup Shannon Protocol Accounts (Gateway \& Application)](#1-setup-shannon-protocol-accounts-gateway--application)
  - [1.1 Gateway and Application Account Creation](#11-gateway-and-application-account-creation)
  - [1.2 `Application` and `Gateway` Account Validation](#12-application-and-gateway-account-validation)
- [2. Configure PATH for Shannon](#2-configure-path-for-shannon)
  - [2.1 Generate Shannon Config](#21-generate-shannon-config)
  - [2.2 Verify Configuration](#22-verify-configuration)
- [3. Run PATH in development mode](#3-run-path-in-development-mode)
- [3.1 Start PATH](#31-start-path)
  - [3.1 Monitor PATH](#31-monitor-path)
- [4. Test Relays](#4-test-relays)
  - [Test Relay with `curl`](#test-relay-with-curl)
  - [Test Relay with `make`](#test-relay-with-make)

## 0. Prerequisites

1. Prepare your environment by following the instructions in the [**environment setup**](2_environment.md) guide.
2. Install the [**`pocketd` CLI**](https://dev.poktroll.com/tools/user_guide/pocketd_cli) to interact with [Pocket Network's Shannon Network](https://dev.poktroll.com).

:::tip

You can use the `make install_deps` command to install the dependencies for the PATH stack, including the `pocketd` CLI.

:::

## 1. Setup Shannon Protocol Accounts (Gateway & Application)

Before starting, you'll need to create and configure:

1. **An onchain [Gateway](https://docs.pokt.network/pokt-protocol/the-shannon-upgrade/shannon-actors/gateways)**: An onchain actor that **facilitates** _(i.e. proxies)_ relays to ensure Quality of Service (QoS)
2. **An onchain [Application](https://docs.pokt.network/pokt-protocol/the-shannon-upgrade/shannon-actors/sovereign-applications)**: An onchain actor that **pays** _(i.e. the API key holder)_ for relays

### 1.1 Gateway and Application Account Creation

We **strongly** recommend following the [**App & PATH Gateway Cheat Sheet**](https://dev.poktroll.com/operate/cheat_sheets/gateway_cheatsheet) for setting up your accounts and getting a thorough understanding of all the elements.

But, if this is your first time going through the docs, you can follow the copy-pasta instructions
below to get a feel for the end-to-end process.

<details>

<summary>tl;dr Copy-pasta to stake onchain Application & Gateway</summary>

**Prepare a gateway stake config:**

```bash
cat <<EOF >>/tmp/stake_gateway_config.yaml
stake_amount: 1000000upokt
EOF
```

**Prepare an application stake config:**

```bash
cat <<EOF > /tmp/stake_app_config.yaml
stake_amount: 100000000upokt
service_ids:
- "anvil"
EOF
```

**Create gateway and application accounts in your keyring:**

```bash
pocketd keys add gateway
pocketd keys add application
```

Fund the accounts by finding a link to the faucet [here](https://dev.poktroll.com/category/explorers-faucets-wallets-and-more).

:::tip Grove employees only

You can manually fund the accounts using the `pkd_beta_tx` helper by following the instructions [in this notion doc](https://www.notion.so/buildwithgrove/Shannon-Alpha-Beta-Environment-rc-helpers-152a36edfff680019314d468fad88864?pvs=4):

```bash
pkd_beta_tx bank send faucet_beta $(pkd keys show -a application) 6900000000042upokt
pkd_beta_tx bank send faucet_beta $(pkd keys show -a gateway) 6900000000042upokt
```

:::

**Stake the gateway:**

```bash
pocketd tx gateway stake-gateway \
 --config=/tmp/stake_gateway_config.yaml \
 --from=gateway --gas=auto --gas-prices=200upokt --gas-adjustment=1.5 --chain-id=pocket-beta \
 --node=https://shannon-testnet-grove-rpc.beta.poktroll.com \
 --yes
```

**Stake the application:**

```bash
pocketd tx application stake-application \
 --config=/tmp/stake_app_config.yaml \
 --from=application --gas=auto --gas-prices=200upokt --gas-adjustment=1.5 --chain-id=pocket-beta \
 --node=https://shannon-testnet-grove-rpc.beta.poktroll.com \
 --yes
```

**Delegate from the application to the gateway:**

```bash
pocketd tx application delegate-to-gateway $(pocketd keys show -a gateway) \
 --from=application --gas=auto --gas-prices=200upokt --gas-adjustment=1.5 --chain-id=pocket-beta \
 --node=https://shannon-testnet-grove-rpc.beta.poktroll.com \
 --yes
```

</details>

### 1.2 `Application` and `Gateway` Account Validation

After following the instructions above, you should have keys with the names `gateway` and `application` in your keyring.

You can validate it like so:

```bash
# All accounts
pocketd keys list

# Gateway only
pocketd keys show -a gateway

# Application only
pocketd keys show -a application
```

## 2. Configure PATH for Shannon

### 2.1 Generate Shannon Config

Run the following command to generate a Shannon config at `local/path/.config.yaml`:

```bash
make shannon_populate_config
```

:::important Command configuration

This command relies on `pocketd` command line interface to export the **Gateway** and **Application** address from your keyring backend.

You can set the following environment variables to override the default values:

- `POCKETD_HOME`: Path to the `pocketd` home directory (default `$HOME/.pocketd`)
- `POCKETD_KEYRING_BACKEND`: Keyring backend to use (default `test`)

:::

### 2.2 Verify Configuration

Check your config file:

```bash
cat local/path/.config.yaml
```

It should look similar to the following with the `gateway_config` filled out.

```yaml
shannon_config:
  full_node_config:
    rpc_url: https://shannon-testnet-grove-rpc.beta.poktroll.com
    grpc_config:
      host_port: shannon-testnet-grove-grpc.beta.poktroll.com:443
    lazy_mode: true
  gateway_config:
    gateway_mode: "centralized"
    gateway_address: pokt1... # Your gateway address
    gateway_private_key_hex: "0x..." # Your gateway private key
    owned_apps_private_keys_hex:
      - "0x..." # Your application private key
hydrator_config:
  service_ids:
    - "anvil"
```

:::important Gateway Configuration Validation

1. Do a manual sanity check of the addresses to ensure everything looks correct before continuing.
2. The configuration about is set up for `anvil`, so ensure your application is staked for the same service id. You can verify this by running:

   ```bash
   pocketd query application show-application \
     $(pkd keys show -a application) \
     --node=https://shannon-testnet-grove-rpc.beta.poktroll.com
   ```

:::

## 3. Run PATH in development mode

## 3.1 Start PATH

Run PATH in local development mode in Tilt by running:

```bash
make path_up
```

You can stop the PATH stack by running:

```bash
make path_down
```

### 3.1 Monitor PATH

:::warning Grab a ☕
It could take a few minutes for `path`, `guard` and `watch` to start up the first time.
:::

You should see an output similar to the following relatively quickly (~30 seconds):

![Tilt Output in Console](../../../static/img/path-in-tilt-console.png)

Once you see the above log, visit [localhost:10350](<http://localhost:10350/r/(all)/overview>) to view the Tilt dashboard.

![Tilt Dashboard in Browser](../../../static/img/path-in-tilt.png)

## 4. Test Relays

:::warning Anvil Node & Request Retries

_tl;dr Retry the requests below if the first one fails._

The instructions above were written to get you to access an [**anvil**](https://book.getfoundry.sh/anvil/) node accessible on Pocket Network.

Since `anvil` is an Ethereum node used for testing, there is a chance it may not be available.

We recommend you try the instructions below a few times to ensure you can get a successful relay. Otherwise, reach out to the community on Discord.

:::

### Test Relay with `curl`

Send a test relay to check the height of

```bash
curl http://localhost:3070/v1 \
 -H "Target-Service-Id: anvil" \
 -H "Authorization: test_api_key" \
 -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```

And you should expect to see a response similar to the following:

```json
{"id":1,"jsonrpc":"2.0","result":"0x2f01a"}%
```

### Test Relay with `make`

For your convenience, we have provided a few makefile helpers to test relays with `curl` and `jq`.
You can find that in the `makefiles/test_requests.mk` file.

For example, to mimic the `curl` command above, you can simply run:

```bash
make test_request__service_id_header_shannon
```

To see all available helpers, simply run:

```bash
make help
```
