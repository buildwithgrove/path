---
sidebar_position: 5
title: Configuration
description: PATH configuration details
---

The following documentation describes how to configure a local PATH deployment running in development mode in Tilt.


:::warning IN PROGRESS

For production deployments of PATH, the Operate documentation is currently under construction.

:::

<!-- TODO_IMPROVE(@commoddity): add a link to Operate docs for production configuration. -->

:::info CONFIGURATION FILES

A PATH deployment is configured via two files:

| File           | Required | Description                                   |
| -------------- | -------- | --------------------------------------------- |
| `.config.yaml` | ✅        | configures the PATH **gateway**               |
| `.values.yaml` | ❌        | configures the PATH **Helm chart deployment** |

:::

## Table of Contents <!-- omit in toc -->

- [PATH Config File (`.config.yaml`)](#path-config-file-configyaml)
  - [Config File Location](#config-file-location)
  - [Config File Fields](#config-file-fields)
    - [`morse_config`](#morse_config)
    - [`shannon_config`](#shannon_config)
    - [`hydrator_config` (optional)](#hydrator_config-optional)
    - [`router_config` (optional)](#router_config-optional)
    - [`logger_config` (optional)](#logger_config-optional)
- [Helm Values Config File (`.values.yaml`)](#helm-values-config-file-valuesyaml)
  - [Config File Location](#config-file-location-1)
  - [GUARD Configuration](#guard-configuration)
    - [`auth.apiKey` Section](#authapikey-section)
    - [`services` Section](#services-section)
    - [Example `.values.yaml` File](#example-valuesyaml-file)
  - [Example Requests](#example-requests)

## PATH Config File (`.config.yaml`)

All configuration for the PATH gateway is defined in a single YAML file named `.config.yaml`.

Exactly one of `shannon_config` or `morse_config` **must** be provided. This field determines the protocol that the gateway will use.

<details>

<summary>Example **Shannon** Config (click to expand)</summary>

```yaml
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
    - "anvil"

# (Optional) Logger Configuration
logger_config:
  level: "info" # Valid values: debug, info, warn, error
```

</details>

<details>

<summary>Example **Morse** Config (click to expand)</summary>

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

# (Optional) Quality of Service (QoS) Configuration
hydrator_config:
  service_ids:
    - "F00C"
    - "F021"

# (Optional) Logger Configuration
logger_config:
  level: "info" # Valid values: debug, info, warn, error
```

</details>

### Config File Location

In development mode, the config file must be located at:

```bash
./local/path/.config.yaml
```

:::tip VSCode Validation

If you are using VSCode, we recommend using the [YAML Language Support](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) extension for in-editor validation of the `.config.yaml` file. Enable it by ensuring the following annotation is present at the top of your config file:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/buildwithgrove/path/refs/heads/main/config/config.schema.yaml
```

:::

### Config File Fields

This is a comprehensive outline and explanation of each YAML field in the configuration file.

#### Protocol Section <!-- omit in toc -->

The config file **MUST contain EXACTLY one** of the following top-level protocol-specific sections:

- `morse_config`
- `shannon_config`

---

#### `morse_config`

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

#### Morse Field Descriptions <!-- omit in toc -->

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

#### `shannon_config`

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

#### Shannon Field Descriptions <!-- omit in toc -->

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

#### `hydrator_config` (optional)


Configures the QoS hydrator to run synthetic Quality of Service (QoS) checks against endpoints of the provided service IDs.

For example, to enable QoS checks for the Ethereum & Polygon services, the following configuration must be added to the `.config.yaml` file:


```yaml
hydrator_config:
  service_ids:
    - "F00C"
    - "F021"
```

:::info

For a full list of currently supported QoS service implementations, please refer to the [QoS Documentation](../../learn/qos/supported_services.md).

:::warning

⚠️ Any ID provided here must match a `Service ID` from the [QoS Documentation](../../learn/qos/supported_services.md); if an invalid ID is provided, the gateway will error.

:::

#### Hydrator Field Descriptions <!-- omit in toc -->

| Field                        | Type          | Required | Default   | Description                                                                          |
| ---------------------------- | ------------- | -------- | --------- | ------------------------------------------------------------------------------------ |
| `service_ids`                | array[string] | No       | -         | List of service IDs for which the Quality of Service (QoS) logic will apply          |
| `run_interval_ms`            | string        | No       | "10000ms" | Interval at which the hydrator will run QoS checks                                   |
| `max_endpoint_check_workers` | integer       | No       | 100       | Maximum number of workers to run concurrent QoS checks against a service's endpoints |


---

#### `router_config` (optional)

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

#### `logger_config` (optional)

Controls the logging behavior of the PATH gateway.

```yaml
logger_config:
  level: "warn"
```

| Field   | Type   | Required | Default | Description                                                           |
| ------- | ------ | -------- | ------- | --------------------------------------------------------------------- |
| `level` | string | No       | "info"  | Minimum log level. Valid values are: "debug", "info", "warn", "error" |

<br/>

## Helm Values Config File (`.values.yaml`)

The `.values.yaml` file is used to configure a PATH deployment by overriding the default values in the Helm chart.

:::info DEFAULT VALUES

**Using the `.values.yaml` file is optional; PATH will run with default values if the file is not present.**

However, it is is highly recommended to override the default values in the `.values.yaml` file to customize the local PATH deployment to your needs.

By default PATH is configured as follows:

**1. Services**
   | Protocol  | Service ID | Aliases                      |
   | --------- | ---------- | ---------------------------- |
   | `shannon` | `anvil`    | -                            |
   | `morse`   | `F00C`     | `eth`, `eth-mainnet`         |
   | `morse`   | `F021`     | `polygon`, `polygon-mainnet` |

**2. API Keys:**
   - `test_api_key`


:::tip

If you wish to customize the default values, you can copy the template file to the local directory and modify it.

```bash
make copy_values_yaml
```

:::

:::info

For the full list of configurable values in the PATH Helm Chart, see the [Helm Values Documentation](../helm/values.md).

:::

### Config File Location

In development mode, the config file must be located at:

```bash
./local/path/.values.yaml
```

Tilt's hot reload feature is enabled by default in the Helm chart. This means that when the `.values.yaml` file is updated, Tilt will automatically redeploy the PATH gateway with the new values.

### GUARD Configuration

#### `auth.apiKey` Section

This section configures the list of allowed API keys for the PATH gateway. Any request without a valid API key will be rejected.

The API key is specified per-request as the `Authorization` header.

**By default a single default API key value of `test_api_key` is provided. This should be overridden in the `.values.yaml` file.**

_See Envoy Gateway's [API Key Authentication documentation](https://gateway.envoyproxy.io/latest/tasks/security/apikey-auth/) for more information._

| Field     | Type          | Required | Default        | Description                                       |
| --------- | ------------- | -------- | -------------- | ------------------------------------------------- |
| `enabled` | boolean       | Yes      | true           | Whether to enforce API key authentication         |
| `apiKeys` | array[string] | Yes      | [test_api_key] | List of API keys authorized to access the gateway |

#### `services` Section

This section configures the list of services that are allowed to access the PATH gateway.

Each service must be assigned a unique `serviceId` and may have multiple `aliases` which map to the same service.

The service ID is specified per-request as the `Target-Service-Id` header; either the `serviceId` or any of the `aliases` will be accepted. **See examples below for more details.**

| Field                  | Type          | Required | Default | Description                           |
| ---------------------- | ------------- | -------- | ------- | ------------------------------------- |
| `services`             | array[object] | Yes      | -       | List of services                      |
| `services[].serviceId` | string        | Yes      | -       | The unique identifier for the service |
| `services[].aliases`   | array[string] | Yes      | -       | List of aliases for the service       |

#### Example `.values.yaml` File

```yaml
guard:
  auth:
    apiKey:
      enabled: true
      apiKeys:
        - test_api_key_1
        - test_api_key_2
        - test_api_key_3
  services:  
    - serviceId: F021
      aliases:
        - polygon
    - serviceId: F00C
      aliases:
        - eth
    - serviceId: F000
      aliases:
        - pocket
```

### Example Requests

The above `.values.yaml` files will allow the following requests to PATH:

```bash
# Request to the "polygon" service using the service ID
# API key: test_api_key_1
curl http://localhost:3070/v1 \
  -H "Target-Service-Id: F021" \
  -H "Authorization: test_api_key_1" \ 
  -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'

# Request to the "polygon" service using an alias
# API key: test_api_key_2
curl http://localhost:3070/v1 \
  -H "Target-Service-Id: polygon" \
  -H "Authorization: test_api_key_2" \
  -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'

# Request to the "eth" service using an alias
# API key: test_api_key_3
curl http://localhost:3070/v1 \
  -H "Target-Service-Id: eth" \
  -H "Authorization: test_api_key_3" \
  -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```
