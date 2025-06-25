---
sidebar_position: 6
title: PATH Helm Config (`.values.yaml`)
description: PATH Helm Configurations
---

:::danger 🚧 WORK IN PROGRESS 🚧

This document is not ready for public consumption.

:::

:::info CONFIGURATION FILES

A `PATH` stack is configured via two files:

| File           | Required | Description                                   |
| -------------- | -------- | --------------------------------------------- |
| `.config.yaml` | ✅        | PATH **gateway** configurations               |
| `.values.yaml` | ❌        | PATH **Helm chart deployment** configurations |

:::

## Table of Contents <!-- omit in toc -->

- [Default Values](#default-values)
- [Customizing Default Values](#customizing-default-values)
- [Helm Values File Location (Local Development)](#helm-values-file-location-local-development)
- [GUARD Configuration](#guard-configuration)
  - [`auth.apiKey` Section](#authapikey-section)
    - [`services` Section](#services-section)
    - [Example `.values.yaml` File](#example-valuesyaml-file)
  - [Example Requests](#example-requests)

## Default Values

:::note Optional

Using the `.values.yaml` file is optional; PATH will run with default values if the file is not present.

:::

The `.values.yaml` file is used to configure a PATH deployment by overriding the default values in the Helm chart.

By default PATH is configured as follows:

**Services:**

| Protocol  | Service ID | Aliases                      |
| --------- | ---------- | ---------------------------- |
| `shannon` | `anvil`    | -                            |
| `morse`   | `F00C`     | `eth`, `eth-mainnet`         |
| `morse`   | `F021`     | `polygon`, `polygon-mainnet` |

**API Keys:**

- `test_api_key`

## Customizing Default Values

If you wish to customize the default values, you can copy the template file to the local directory and modify it.

```bash
make configs_copy_values_yaml
```

:::info

For the full list of configurable values in the PATH Helm Chart, see the [Helm Values Documentation](../../operate/helm/5_values.md).

:::

## Helm Values File Location (Local Development)

In development mode, the config file must be located at:

```bash
./local/path/.values.yaml
```

Tilt's hot reload feature is enabled by default in the Helm chart. This means that when the `.values.yaml` file is updated, Tilt will automatically redeploy the PATH Gateway stack with the new values.

## GUARD Configuration

### `auth.apiKey` Section

:::important Default `test_api_key`

By default a single default API key value of `test_api_key` is provided. This should be overridden in the `.values.yaml` file.

:::

This section configures the list of allowed API keys for the PATH gateway. Any request without a valid API key will be rejected.

The API key is specified per-request as the `Authorization` header.

| Field     | Type          | Required | Default        | Description                                       |
| --------- | ------------- | -------- | -------------- | ------------------------------------------------- |
| `enabled` | boolean       | Yes      | true           | Whether to enforce API key authentication         |
| `apiKeys` | array[string] | Yes      | [test_api_key] | List of API keys authorized to access the gateway |

💡 _See Envoy Gateway's [API Key Authentication documentation](https://gateway.envoyproxy.io/latest/tasks/security/apikey-auth/) for more information._

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

<!--TODO_MIGRATION(@commoddity): once GUARD is updated, remove `shannonServiceId` and use `serviceId` instead. -->
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
      shannonServiceId: poly
      aliases:
        - polygon
    - serviceId: F00C
      shannonServiceId: eth
      aliases:
        - eth
    - serviceId: F000
      shannonServiceId: pocket
      aliases:
        - pocket
```

### Example Requests

The above `.values.yaml` files will allow the following requests to PATH.

Request to the `polygon` service using the service ID using API key `test_api_key_1`:

```bash
curl http://localhost:3070/v1 \
  -H "Target-Service-Id: F021" \
  -H "Authorization: test_api_key_1" \
  -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```

Request to the `polygon` service using the alias using API key `test_api_key_2`:

```bash
curl http://localhost:3070/v1 \
  -H "Target-Service-Id: polygon" \
  -H "Authorization: test_api_key_2" \
  -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```

Request to the "eth" service using an alias using API key `test_api_key_3`:

```bash
curl http://localhost:3070/v1 \
  -H "Target-Service-Id: eth" \
  -H "Authorization: test_api_key_3" \
  -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```
