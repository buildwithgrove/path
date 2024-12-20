---
sidebar_position: 2
title: YAML Data Source
---

If the `YAML_FILEPATH` environment variable is set, PADS will load the data from a YAML file at the specified path.

Hot reloading is supported, so changes to the YAML file will be reflected in the `Go External Authorization Server` without the need to restart PADS.

# Table of Contents <!-- omit in toc -->

- [Example YAML File](#example-yaml-file)
- [YAML Schema](#yaml-schema)

## Example YAML File

_`PADS` loads data from the Gateway Endpoints YAML file specified by the `YAML_FILEPATH` environment variable._\

:::info

[The example `gateway-endpoints.yaml` file may be seen in the PADS repo](https://github.com/buildwithgrove/path-auth-data-server/blob/main/yaml/testdata/gateway-endpoints.example.yaml).

:::

The yaml file below provides an example of a gateway operator's `gateway-endpoints.yaml` file where:

- `endpoint_1_static_key` is authorized with a static API Key
- `endpoint_2_jwt` is authorized using an auth-provider issued JWT for two users
- `endpoint_3_no_auth` requires no authorization and has a rate limit set

```yaml
endpoints:
  # 1. Example of a gateway endpoint using API Key Authorization
  endpoint_1_static_key:
    auth:
      api_key: "api_key_1"

  # 2. Example of a gateway endpoint using JWT Authorization
  endpoint_2_jwt:
    auth:
      jwt_authorized_users:
        - "auth0|user_1"
        - "auth0|user_2"

  # 3. Example of a gateway endpoint with no authorization and rate limiting set
  endpoint_3_no_auth:
    rate_limiting:
      throughput_limit: 30
      capacity_limit: 100000
      capacity_limit_period: "CAPACITY_LIMIT_PERIOD_MONTHLY"
```

[Full Example Gateway Endpoints YAML File](./yaml/testdata/gateway-endpoints.example.yaml)

:::tip
## YAML Schema

[The YAML Schema](./yaml/gateway-endpoints.schema.yaml) defines the expected structure of the YAML file.

You may install the RedHat YAML extension for VSCode to validate the YAML file against the schema.
https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml

:::
