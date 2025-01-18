---
sidebar_position: 3
title: PATH Config
description: PATH configuration details
---

<div align="center">
<h1>PATH<br/>Path Configration YAML File</h1>
<img src="https://storage.googleapis.com/grove-brand-assets/Presskit/Logo%20Joined-2.png" alt="Grove logo" width="500"/>
</div>

:::info PATH Configuration

These instructions are intended for configuring the **PATH Gateway**.

**Envoy Proxy** has its own set of configuration files. See the details in the [Envoy Configuration Guide](../envoy/envoy_config.md).

:::

## Table of Contents <!-- omit in toc -->

- [Configuration YAML File](#configuration-yaml-file)
  - [Config File Location](#config-file-location)
  - [Example Config Files](#example-config-files)
  - [Config YAML Schema](#config-yaml-schema)
- [YAML Fields](#yaml-fields)
  - [Protocol Section](#protocol-section)
  - [`morse_config`](#morse_config)
    - [Morse Field Descriptions](#morse-field-descriptions)
    - [AAT Generation](#aat-generation)
  - [`shannon_config`](#shannon_config)
    - [Shannon Field Descriptions](#shannon-field-descriptions)
  - [`router_config` (optional)](#router_config-optional)
  - [`hydrator_config` (optional)](#hydrator_config-optional)
  - [`auth_server_config` (optional)](#auth_server_config-optional)
  - [`logger_config` (optional)](#logger_config-optional)
  - [`messaging_config` (TODO)](#messaging_config-todo)

## Configuration YAML File

All configuration for the PATH gateway is defined in a single YAML file named `.config.yaml`.

The following is quick example of a Gateway configured to support Shannon protocol on Beta TestNet.

```yaml
# Protocol Configuration
shannon_config:
  full_node_config:
    rpc_url: https://shannon-testnet-grove-rpc.beta.poktroll.com
    grpc_config:
      host_port: shannon-testnet-grove-grpc.beta.poktroll.com:443
  gateway_config:
    gateway_mode: "centralized"
    gateway_address: pokt1up7zlytnmvlsuxzpzvlrta95347w322adsxslw
    gateway_private_key_hex: 40af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d388
    owned_apps_private_keys_hex:
      - 40af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d388

# Quality of Service (QoS) Configuration
hydrator_config:
  service_ids:
    - "eth"
    - "solana"
    - "pokt"

# Authorization Server Configuration
auth_server_config:
  grpc_host_port: path-auth-data-server:50051
  grpc_use_insecure_credentials: true
  endpoint_id_extractor_type: url_path

# Logger Configuration
logger_config:
  level: "info"  # Valid values: debug, info, warn, error
```

### Config File Location

The default location of the configuration file is `./config/.config.yaml` relative to the location of the PATH binary.

For example, when running the compiled PATH binary from the `./bin` directory, the configuration should be located at `./bin/config/.config.yaml`.

As another example, when running PATH in Tilt, the configuration file is mounted in the container at `/app/config/.config.yaml`.

:::tip

The location of the configuration file may be overridden using the `-config` flag.

For example, you may run`./path -config ./config/.config.custom.yaml`.

:::

### Example Config Files

Example configuration files for both Shannon and Morse gateways are provided below.

- [Example Shannon Config YAML File](https://github.com/buildwithgrove/path/blob/main/config/examples/config.shannon_example.yaml)
- [Example Morse Config YAML File](https://github.com/buildwithgrove/path/blob/main/config/examples/config.morse_example.yaml)

The example files contain extensive comments and explanations for every field.

### Config YAML Schema

A YAML schema is provided for the configuration file.

This schema is used to validate the configuration file and ensure that it is populated with the appropriate values.

- [Config YAML Schema File](https://github.com/buildwithgrove/path/tree/main/config/config.schema.yaml)

:::tip

For VSCode users, the [YAML Language Support by Red Hat](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) plugin may be used to provide in-editor syntax highlighting and validation by installing the plugin and placing the following comment annotation at the top of your `.config.yaml` file:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/buildwithgrove/path/refs/heads/main/config/config.schema.yaml
```

:::

## YAML Fields

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

#### AAT Generation

Assuming you have access to a staked application, you can build your own `pocket-core` binary to generate an AAT following the steps below.

```bash
git clone git@github.com:pokt-network/pocket-core.git
cd pocket-core
go build -o pocket ./app/cmd/pocket_core/main.go
./pocket-core create-aat <ADDR_APP> <CLIENT_PUB>
```

Which will output a JSON object similar to the following:

```json
{
  "version": "0.0.1",
  "app_pub_key": <APP_PUB>,
  "client_pub_key": <CLIENT_PUB>,
  "signature": <APP_SIG>
}
```

```yaml
morse_config:
  # ...
  relay_signing_key: "CLIENT_PRIV"
  # ...
signed_aats:
  <ADDR_APP>:
    client_public_key: "<CLIENT_PUB>"
    application_public_key: "<APP_PUB>"
    application_signature: "<APP_SIG>"
```

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

| Field     | Type   | Required | Default | Description                     |
| --------- | ------ | -------- | ------- | ------------------------------- |
| `rpc_url` | string | Yes      | -       | URL of the Shannon RPC endpoint |

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

Allows specifying server parameters for how the gateway handles incoming requests.

| Field                   | Type    | Required | Default           | Description                                     |
| ----------------------- | ------- | -------- | ----------------- | ----------------------------------------------- |
| `port`                  | integer | No       | 3069              | Port number on which the gateway server listens |
| `max_request_body_size` | integer | No       | 1MB               | Maximum request size in bytes                   |
| `read_timeout`          | string  | No       | "5000ms" (5s)     | Time limit for reading request data             |
| `write_timeout`         | string  | No       | "10000ms" (10s)   | Time limit for writing response data            |
| `idle_timeout`          | string  | No       | "120000ms" (120s) | Time limit for closing idle connections         |

---

### `hydrator_config` (optional)

To enable QoS for a service, the service ID must be provided here.

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

### `auth_server_config` (optional)

**Required in order to authorize requests with Envoy Proxy.** Configures the External Authorization Server.

:::info

[For detailed information on the External Auth Server, please refer to the External Auth Server Documentation](../envoy/walkthrough.md#external-auth-server).

:::

| Field                           | Type    | Required | Default | Description                                                                                                                                                      |
| ------------------------------- | ------- | -------- | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `grpc_host_port`                | string  | Yes      | -       | Host and port for the remote gRPC connection to the `Remote gRPC Server` (eg. PADS). Pattern requires a `host:port` format.                                      |
| `grpc_use_insecure_credentials` | boolean | No       | false   | Set to true if the `Remote gRPC Server` does not use TLS                                                                                                         |
| `endpoint_id_extractor_type`    | string  | No       | -       | Either `url_path` or `header`. Specifies how endpoint IDs are extracted. [See here for more details](../envoy/walkthrough.md#specifying-the-gateway-endpoint-id) |
| `port`                          | integer | No       | 10003   | The local port for running the Auth Server                                                                                                                       |

:::caution

This IS NOT used by the PATH Gateway logic itself, but was placed in the PATH
Gateway's config file for convenience with the goal of avoiding another config file.
This may change in the future.

:::

---

### `logger_config` (optional)

Controls the logging behavior of the PATH gateway.

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
