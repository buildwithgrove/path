---
sidebar_position: 2
title: PATH Config
description: PATH configuration details
---

<div align="center">
<h1>PATH<br/>Path Configration YAML File</h1>
<img src="https://storage.googleapis.com/grove-brand-assets/Presskit/Logo%20Joined-2.png" alt="Grove logo" width="500"/>
</div>

:::info PATH Configuration

These instructions are intended for configuring the **PATH Gateway**.

**Envoy Proxy** has its own set of configuration files.

For details, see the the [**Envoy Configuration Guide**](../envoy/envoy_config.md).

:::

## Table of Contents <!-- omit in toc -->

- [Configuration YAML File: `.config.yaml`](#configuration-yaml-file-configyaml)
  - [Config File Location](#config-file-location)
  - [Example Config Files](#example-config-files)
  - [Config YAML Schema Validation](#config-yaml-schema-validation)
  - [Full Config Example](#full-config-example)
- [YAML Field Explanations](#yaml-field-explanations)
  - [Protocol Section](#protocol-section)
  - [`morse_config`](#morse_config)
    - [Morse Field Descriptions](#morse-field-descriptions)
  - [`shannon_config`](#shannon_config)
    - [Shannon Field Descriptions](#shannon-field-descriptions)
  - [`router_config` (optional)](#router_config-optional)
  - [`hydrator_config` (optional)](#hydrator_config-optional)
  - [`logger_config` (optional)](#logger_config-optional)
  - [`messaging_config` (TODO)](#messaging_config-todo)

## Configuration YAML File: `.config.yaml`

All configuration for the PATH gateway is defined in a single YAML file named `.config.yaml`.

### Config File Location

The config file `.config.yaml` is located in:

- **Default**: `./config/.config.yaml` (relative to PATH binary)
- **Tilt**: `/app/config/.config.yaml` (mounted in container from `./local/path/.config.yaml`)

:::tip Override Config Location
Use `-config` flag to specify a custom location:

```bash
./path -config ./config/.config.custom.yaml
```

:::

### Example Config Files

We provide example configuration files with detailed comments for each protocol type:

- [Shannon Gateway](https://github.com/buildwithgrove/path/blob/main/config/examples/config.shannon_example.yaml)
- [Morse Gateway](https://github.com/buildwithgrove/path/blob/main/config/examples/config.morse_example.yaml)

### Config YAML Schema Validation

The configuration is validated against our [YAML schema](https://github.com/buildwithgrove/path/tree/main/config/config.schema.yaml).

:::tip VSCode Validation

Use the [YAML Language Support](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) extension for real-time validation by adding:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/buildwithgrove/path/refs/heads/main/config/config.schema.yaml
```

Or the following to point to the local schema file:

```yaml
# yaml-language-server: $schema=../../../config/config.schema.yaml
```

:::

### Full Config Example

The following is quick example of a full config file.

Note **exactly one of** `morse_config` or `shannon_config` is acceptable and required.

<details>

<summary>Click to expand full config</summary>

```yaml
# (Required) Morse Protocol Configuration
morse_config:
  full_node_config:
    url: "https://pocket-rpc.liquify.com" # Required: Pocket node URL
    relay_signing_key: "<128-char-hex>" # Required: Relay signing private key
    http_config: # Optional
      retries: 3 # Default: 3
      timeout: "5000ms" # Default: "5000ms"

  signed_aats: # Required
    "<40-char-app-address>": # Application address (hex)
      client_public_key: "<64-char-hex>" # Client public key
      application_public_key: "<64-char-hex>" # Application public key
      application_signature: "<128-char-hex>" # Application signature

# (Required) Shannon Protocol Configuration
shannon_config:
  full_node_config:
    rpc_url: https://shannon-testnet-grove-rpc.beta.poktroll.com
    grpc_config:
      host_port: shannon-testnet-grove-grpc.beta.poktroll.com:443
    lazy_mode: true

  gateway_config:
    gateway_mode: "centralized"
    gateway_address: pokt1up7zlytnmvlsuxzpzvlrta95347w322adsxslw
    gateway_private_key_hex: 40af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d388
    owned_apps_private_keys_hex:
      - 40af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d388

# (Optional) Quality of Service (QoS) Configuration
hydrator_config:
  service_ids:
    - "eth"
    - "solana"
    - "pokt"

# (Optional) Logger Configuration
logger_config:
  level: "info" # Valid values: debug, info, warn, error
```

</details>

## YAML Field Explanations

This is a comprehensive outline and explanation of each YAML field in the configuration file.

### Protocol Section

The config file **MUST contain EXACTLY one** of the following top-level protocol-specific sections:

- `morse_config`
- `shannon_config`

---

### `morse_config`

Configuration for the Morse protocol gateway.

```yaml
morse_config:
  full_node_config:
    url: "https://pocket-rpc.liquify.com" # Required: Pocket node URL
    relay_signing_key: "<128-char-hex>" # Required: Relay signing private key
    http_config: # Optional
      retries: 3 # Default: 3
      timeout: "5000ms" # Default: "5000ms"

  signed_aats: # Required
    "<40-char-app-address>": # Application address (hex)
      client_public_key: "<64-char-hex>" # Client public key
      application_public_key: "<64-char-hex>" # Application public key
      application_signature: "<128-char-hex>" # Application signature
```

#### Morse Field Descriptions

**`full_node_config`**

| Field               | Type   | Required | Default | Description                                              |
| ------------------- | ------ | -------- | ------- | -------------------------------------------------------- |
| `url`               | string | Yes      | -       | URL of the full Pocket RPC node                          |
| `relay_signing_key` | string | Yes      | -       | 128-character hex-encoded private key for signing relays |

**`full_node_config.http_config`**

| Field     | Type    | Required | Default  | Description                           |
| --------- | ------- | -------- | -------- | ------------------------------------- |
| `retries` | integer | No       | 3        | Number of HTTP request retry attempts |
| `timeout` | string  | No       | "5000ms" | HTTP request timeout duration         |

**`signed_aats`**

| Field                    | Type   | Required | Default | Description                                     |
| ------------------------ | ------ | -------- | ------- | ----------------------------------------------- |
| `client_public_key`      | string | Yes      | -       | 64-character hex-encoded client public key      |
| `application_public_key` | string | Yes      | -       | 64-character hex-encoded application public key |
| `application_signature`  | string | Yes      | -       | 128-character hex-encoded signature             |

---

### `shannon_config`

Configuration for the Shannon protocol gateway.

```yaml
shannon_config:
  full_node_config:
    rpc_url: "https://shannon-testnet-grove-rpc.beta.poktroll.com" # Required: Shannon node RPC URL
    grpc_config: # Required
      host_port: "shannon-testnet-grove-grpc.beta.poktroll.com:443" # Required: gRPC host and port
      # Optional backoff and keepalive configs...
      insecure: false # Optional: whether to use insecure connection
      backoff_base_delay: "1s" # Optional: initial backoff delay duration
      backoff_max_delay: "120s" # Optional: maximum backoff delay duration
      min_connect_timeout: "20s" # Optional: minimum timeout for connection attempts
      keep_alive_time: "20s" # Optional: frequency of keepalive pings
      keep_alive_timeout: "20s" # Optional: timeout for keepalive pings
    lazy_mode: true

  gateway_config: # Required
    gateway_mode: "centralized" # Required: centralized, delegated, or permissionless
    gateway_address: "pokt1up7zlytnmvlsuxzpzvlrta95347w322adsxslw" # Required: Bech32 address
    gateway_private_key_hex: "<64-char-hex>" # Required: Gateway private key
    owned_apps_private_keys_hex: # Required for centralized mode only
      - "<64-char-hex>" # Application private key
      - "<64-char-hex>" # Additional application private keys...
```

#### Shannon Field Descriptions

**`full_node_config`**

| Field         | Type    | Required | Default | Description                                                     |
| ------------- | ------- | -------- | ------- | --------------------------------------------------------------- |
| `rpc_url`     | string  | Yes      | -       | URL of the Shannon RPC endpoint                                 |
| `grpc_config` | object  | Yes      | -       | gRPC connection configuration                                   |
| `lazy_mode`   | boolean | No       | true    | If true, disables caching of onchain data (e.g. apps, sessions) |

**`full_node_config.grpc_config`**

| Field                 | Type    | Required | Default | Description                             |
| --------------------- | ------- | -------- | ------- | --------------------------------------- |
| `host_port`           | string  | Yes      | -       | Host and port for gRPC connections      |
| `insecure`            | boolean | No       | false   | Whether to use insecure connection      |
| `backoff_base_delay`  | string  | No       | "1s"    | Initial backoff delay duration          |
| `backoff_max_delay`   | string  | No       | "120s"  | Maximum backoff delay duration          |
| `min_connect_timeout` | string  | No       | "20s"   | Minimum timeout for connection attempts |
| `keep_alive_time`     | string  | No       | "20s"   | Frequency of keepalive pings            |
| `keep_alive_timeout`  | string  | No       | "20s"   | Timeout for keepalive pings             |

**`gateway_config`**

| Field                         | Type     | Required                 | Default | Description                                                           |
| ----------------------------- | -------- | ------------------------ | ------- | --------------------------------------------------------------------- |
| `gateway_mode`                | string   | Yes                      | -       | Mode of operation: `centralized`, `delegated`, or `permissionless`    |
| `gateway_address`             | string   | Yes                      | -       | Bech32-formatted gateway address (starts with `pokt1`)                |
| `gateway_private_key_hex`     | string   | Yes                      | -       | 64-character hex-encoded `secp256k1` gateway private key              |
| `owned_apps_private_keys_hex` | string[] | Only in centralized mode | -       | List of 64-character hex-encoded `secp256k1` application private keys |

---

### `router_config` (optional)

**Enables configuring how incoming requests are handled.**

In particular, allows specifying server parameters for how the gateway handles incoming requests.

| Field                   | Type    | Required | Default           | Description                                     |
| ----------------------- | ------- | -------- | ----------------- | ----------------------------------------------- |
| `port`                  | integer | No       | 3069              | Port number on which the gateway server listens |
| `max_request_body_size` | integer | No       | 1MB               | Maximum request size in bytes                   |
| `read_timeout`          | string  | No       | "5000ms" (5s)     | Time limit for reading request data             |
| `write_timeout`         | string  | No       | "10000ms" (10s)   | Time limit for writing response data            |
| `idle_timeout`          | string  | No       | "120000ms" (120s) | Time limit for closing idle connections         |

---

### `hydrator_config` (optional)

**Required To enable QoS for a service**.

The service ID must be provided here.

| Field                        | Type          | Required | Default   | Description                                                                          |
| ---------------------------- | ------------- | -------- | --------- | ------------------------------------------------------------------------------------ |
| `service_ids`                | array[string] | No       | -         | List of service IDs for which the Quality of Service (QoS) logic will apply          |
| `run_interval_ms`            | string        | No       | "10000ms" | Interval at which the hydrator will run QoS checks                                   |
| `max_endpoint_check_workers` | integer       | No       | 100       | Maximum number of workers to run concurrent QoS checks against a service's endpoints |

:::info

In order to enable QoS for a service, the ID provided here must match a service ID in [`config/service_qos.go`](https://github.com/buildwithgrove/path/blob/main/config/service_qos.go).

:::warning

Not all services currently have a QoS implementation in PATH; new QoS implementations are actively being worked on.

If a service ID is not present in [`config/service_qos.go`](https://github.com/buildwithgrove/path/blob/main/config/service_qos.go), a No-Op QoS implementation will be used for that service. This means a random endpoint will be selected for requests to that service.

:::

---

### `logger_config` (optional)

**Controls the logging behavior** of the PATH gateway.

| Field   | Type   | Required | Default | Description                                                           |
| ------- | ------ | -------- | ------- | --------------------------------------------------------------------- |
| `level` | string | No       | "info"  | Minimum log level. Valid values are: "debug", "info", "warn", "error" |

Example configuration:

```yaml
logger_config:
  # Valid values are: debug, info, warn, error
  # Defaults to info if not specified
  level: "warn"
```

---

### `messaging_config` (TODO)

:::note TODO

TODO_MVP(@adshmh): Add messaging_config

:::
