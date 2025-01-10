---
sidebar_position: 4
title: YAML Data Source
description: YAML data source example configuration
---

If the `YAML_FILEPATH` environment variable is set, **PADS** will load the data from a YAML file at the specified path.

**Hot reloading is supported**, so changes to the YAML file will be reflected in the `Go External Authorization Server` without the need to restart PADS.

## Example YAML Data Source File

:::tip

The PADS repo contains a template [`gateway-endpoints.yaml`](https://github.com/buildwithgrove/path-auth-data-server/blob/main/yaml/testdata/gateway-endpoints.example.yaml) file for reference.

:::

Below are partial sections of that yaml file for explanatory purposes:

### 1. Static API Key Authorization

`endpoint_1_static_key` is authorized with a static API Key.

```yaml
endpoints:
  endpoint_1_static_key:
    auth:
      api_key: "api_key_1"
```

### 2. JWT Authorization

`endpoint_2_jwt` is authorized using an auth-provider issued JWT for two users.

```yaml
endpoints:
  endpoint_2_jwt:
    auth:
      jwt_authorized_users:
        - "auth0|user_1"
        - "auth0|user_2"
```

### 3. No Authorization

`endpoint_3_no_auth` requires no authorization and has a rate limit set

```yaml
endpoints:
  endpoint_3_no_auth:
    rate_limiting:
      throughput_limit: 30
      capacity_limit: 100000
      capacity_limit_period: "CAPACITY_LIMIT_PERIOD_MONTHLY"
```

## YAML Schema

The [YAML Schema](https://github.com/buildwithgrove/path-auth-data-server/blob/main/yaml/gateway-endpoints.schema.yaml) defines the expected structure of the YAML file.

:::tip

You can install the [RedHat YAML extension](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) for VSCode to validate the YAML file against the schema.

:::
