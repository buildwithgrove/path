---
sidebar_position: 3
title: Auth Config (`.values.yaml`)
description: PATH Auth, Helm & Deployment Configurations
---

_tl;dr Configurations for request authorization and deployment._

- [Example Configuration](#example-configuration)
- [Helm Values File Location (Local Development)](#helm-values-file-location-local-development)
- [Default Values](#default-values)
- [Customizing Default Values](#customizing-default-values)
- [GUARD Configuration](#guard-configuration)
  - [`auth.apiKey` Section](#authapikey-section)
    - [`services` Section](#services-section)
  - [Example Requests](#example-requests)

## Example Configuration

<details>

<summary>Example **Values YAML** Config (click to expand)</summary>

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
    - serviceId: poly
      aliases:
        - polygon
    - serviceId: eth
      aliases:
        - ethereum
    - serviceId: pocket
      aliases:
        - pokt
```

</details>

## Helm Values File Location (Local Development)

In development mode, the config file must be located at:

```bash
./local/path/.values.yaml
```

Tilt's hot reload feature is enabled by default in the Helm chart. This means that when the `.values.yaml` file is updated, Tilt will automatically redeploy the PATH Gateway stack with the new values.

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
| `shannon` | `eth`      | `eth`, `eth-mainnet`         |
| `shannon` | `polygon`  | `polygon`, `polygon-mainnet` |

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

### Example Requests

The above `.values.yaml` files will allow the following requests to PATH.

Request to the `polygon` service using the service ID using API key `test_api_key_1`:

```bash
curl http://localhost:3070/v1 \
  -H "Target-Service-Id: polygon" \
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
