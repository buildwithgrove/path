---
sidebar_position: 4
title: PATH Config File (`.config.yaml`)
description: PATH Configurations
---

:::info CONFIGURATION FILES

A `PATH` stack is configured via two files:

| File           | Required | Description                                   |
| -------------- | -------- | --------------------------------------------- |
| `.config.yaml` | ✅        | PATH **gateway** configurations               |
| `.values.yaml` | ❌        | PATH **Helm chart deployment** configurations |

:::

## Table of Contents <!-- omit in toc -->

- [Config File Validation](#config-file-validation)
- [Config File Location (Local Development)](#config-file-location-local-development)
- [`shannon_config` (required)](#shannon_config-required)
- [`hydrator_config` (optional)](#hydrator_config-optional)
  - [Manually Disable QoS Checks for a Service](#manually-disable-qos-checks-for-a-service)
- [`router_config` (optional)](#router_config-optional)
- [`logger_config` (optional)](#logger_config-optional)
- [`data_reporter_config` (optional)](#data_reporter_config-optional)

All configuration for the `PATH` gateway is defined in a single YAML file named `.config.yaml`.

`shannon_config` **MUST** be provided. This field configures the protocol that the gateway will use.

<details>

<summary>Example **Shannon** Config (click to expand)</summary>

```yaml
# (Required) Shannon Protocol Configuration
shannon_config:
  full_node_config:
    rpc_url: https://shannon-testnet-grove-rpc.beta.poktroll.com
    grpc_config:
      host_port: shannon-testnet-grove-grpc.beta.poktroll.com:443
    lazy_mode: false

  gateway_config:
    gateway_mode: "centralized"
    gateway_address: pokt1up7zlytnmvlsuxzpzvlrta95347w322adsxslw
    gateway_private_key_hex: 40af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d388
    owned_apps_private_keys_hex:
      - 40af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d388

# (Optional) Logger Configuration
logger_config:
  level: "info" # Valid values: debug, info, warn, error
```

</details>

## Config File Validation

:::tip VSCode Validation

If you are using VSCode, we recommend using the [YAML Language Support](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) extension for in-editor validation of the `.config.yaml` file. Enable it by ensuring the following annotation is present at the top of your config file:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/buildwithgrove/path/refs/heads/main/config/config.schema.yaml
```

:::

## Config File Location (Local Development)

In development mode, the config file must be located at:

```bash
./local/path/.config.yaml
```

## Protocol Selection <!-- omit in toc -->

The config file **MUST** contain `shannon_config`.

---

<!--

:::warning TODO_MVP(@commoddity): Auto-generate this file.

Update this file so it is auto-generated based on config.schema.yaml

:::

-->

## `shannon_config` (required)

Configuration for the Shannon protocol gateway.

```yaml
shannon_config:
  full_node_config:
    lazy_mode: false
    rpc_url: "https://shannon-testnet-grove-rpc.beta.poktroll.com"
    grpc_config: # Required
      host_port: "shannon-testnet-grove-grpc.beta.poktroll.com:443" # Required: gRPC host and port

      # Optional backoff and keepalive configs...
      insecure: false # Optional: whether to use insecure connection
      backoff_base_delay: "1s" # Optional: initial backoff delay duration
      backoff_max_delay: "120s" # Optional: maximum backoff delay duration
      min_connect_timeout: "20s" # Optional: minimum timeout for connection attempts
      keep_alive_time: "20s" # Optional: frequency of keepalive pings
      keep_alive_timeout: "20s" # Optional: timeout for keepalive pings

  gateway_config: # Required
    gateway_mode: "centralized" # Required: centralized, delegated, or permissionless
    gateway_address: "pokt1up7zlytnmvlsuxzpzvlrta95347w322adsxslw" # Required: Bech32 address
    gateway_private_key_hex: "<64-char-hex>" # Required: Gateway private key
    owned_apps_private_keys_hex: # Required for centralized mode only
      - "<64-char-hex>" # Application private key
      - "<64-char-hex>" # Additional application private keys...
```

**`full_node_config`**

| Field         | Type    | Required | Default | Description                                                     |
| ------------- | ------- | -------- | ------- | --------------------------------------------------------------- |
| `rpc_url`     | string  | Yes      | -       | URL of the Shannon RPC endpoint                                 |
| `grpc_config` | object  | Yes      | -       | gRPC connection configuration                                   |
| `lazy_mode`   | boolean | No       | false   | If true, disables caching of onchain data (e.g. apps, sessions) |

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

## `hydrator_config` (optional)

:::info

For a full list of supported QoS service implementations, refer to the [QoS Documentation](../../learn/qos/1_supported_services.md).

:::

Configures the QoS hydrator. By default, all services configured in the `shannon_config` section will have QoS checks run against them.

### Manually Disable QoS Checks for a Service

To manually disable QoS checks for a specific service, the `qos_disabled_service_ids` field may be specified in the `.config.yaml` file.

For example, to disable QoS checks for the Ethereum service on a Shannon PATH instance, the following configuration would be added to the `.config.yaml` file:

```yaml
hydrator_config:
  qos_disabled_service_ids:
    - "eth"
```

| Field                        | Type          | Required | Default   | Description                                                                                                                                                 |
| ---------------------------- | ------------- | -------- | --------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `run_interval_ms`            | string        | No       | "10000ms" | Interval at which the hydrator will run QoS checks                                                                                                          |
| `max_endpoint_check_workers` | integer       | No       | 100       | Maximum number of workers to run concurrent QoS checks against a service's endpoints                                                                        |
| `qos_disabled_service_ids`   | array[string] | No       | -         | List of service IDs to exclude from QoS checks. Will throw an error on startup if a service ID is provided that the PATH instance is not configured to use. |

---

## `router_config` (optional)

**Enables configuring how incoming requests are handled.**

In particular, allows specifying server parameters for how the gateway handles incoming requests.

| Field                   | Type    | Required | Default           | Description                                     |
| ----------------------- | ------- | -------- | ----------------- | ----------------------------------------------- |
| `port`                  | integer | No       | 3070              | Port number on which the gateway server listens |
| `max_request_body_size` | integer | No       | 1MB               | Maximum request size in bytes                   |
| `read_timeout`          | string  | No       | "5000ms" (5s)     | Time limit for reading request data             |
| `write_timeout`         | string  | No       | "10000ms" (10s)   | Time limit for writing response data            |
| `idle_timeout`          | string  | No       | "120000ms" (120s) | Time limit for closing idle connections         |

---

## `logger_config` (optional)

Controls the logging behavior of the PATH gateway.

```yaml
logger_config:
  level: "warn"
```

| Field   | Type   | Required | Default | Description                                                           |
| ------- | ------ | -------- | ------- | --------------------------------------------------------------------- |
| `level` | string | No       | "info"  | Minimum log level. Valid values are: "debug", "info", "warn", "error" |

---

## `data_reporter_config` (optional)

Configures HTTP-based data reporting to external services like BigQuery via data pipelines (e.g., Fluentd with HTTP input and BigQuery output plugins).

```yaml
data_reporter_config:
  target_url: "https://fluentd-service.example.com/http-input"
  post_timeout_ms: 5000
```

| Field             | Type    | Required | Default | Description                                                                                       |
| ----------------- | ------- | -------- | ------- | ------------------------------------------------------------------------------------------------- |
| `target_url`      | string  | Yes      | -       | HTTP endpoint URL where data will be reported (must start with http:// or https://)               |
| `post_timeout_ms` | integer | No       | 10000   | Timeout in milliseconds for HTTP POST operations. If zero or negative, default of 10000ms is used |

:::info
Currently, only JSON-accepting data pipelines are supported as of PR #215.
:::
