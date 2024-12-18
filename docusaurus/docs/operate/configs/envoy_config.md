---
title: Envoy Proxy config
sidebar_position: 4
---

# Envoy Proxy config <!-- omit in toc -->

This document describes the configuration options for PATH's Envoy Proxy.

In PATH, Envoy Proxy is responsible for:

- Defining allowed services
- Request authorization
- Rate limiting

:::info

There are a total of four files used to configure Envoy Proxy in PATH:

**Envoy Config Files**

1. `.allowed-services.lua`
2. `.envoy.yaml`
3. `.ratelimit.yaml`

    The templates used to generate these Envoy config files [may be found here](https://github.com/buildwithgrove/path/tree/main/envoy).

**Gateway Endpoints File**

4. `.gateway-endpoints.yaml`

    The example `.gateway-endpoints.yaml` file is located in the [PADS repo](https://github.com/buildwithgrove/path-auth-data-service/tree/main/envoy/gateway-endpoints.yaml).

:::

- [Initialization](#initialization)
- [Allowed Services - `.allowed-services.lua`](#allowed-services---allowed-serviceslua)
  - [File Format](#file-format)
  - [Terminology](#terminology)
- [Envoy Proxy Configuration - `.envoy.yaml`](#envoy-proxy-configuration---envoyyaml)
- [Ratelimit Configuration - `.ratelimit.yaml`](#ratelimit-configuration---ratelimityaml)
  - [File Format](#file-format-1)
- [Gateway Endpoints Data - `.gateway-endpoints.yaml`](#gateway-endpoints-data---gateway-endpointsyaml)
  - [File Format](#file-format-2)


## Initialization

The required Envoy configuration files may be generated from their templates by running the make target:

```bash
make init_envoy
```

This will generate the following files in the `local/path/envoy` directory:

- `.allowed-services.lua`
- `.envoy.yaml`
- `.ratelimit.yaml`
- `.gateway-endpoints.yaml`

:::warning

All of these files are git ignored from the PATH repo as they are specific to each PATH instance and may contain sensitive information.

:::

## Allowed Services - `.allowed-services.lua`

The `.allowed-services.lua` file is used to define the allowed services for the Envoy Proxy.

Once created in `local/path/envoy`, the `.allowed-services.lua` file is mounted as a file in the Envoy Proxy container at `/etc/envoy/.allowed-services.lua`.

### File Format

_`.allowed-services.lua` format:_
```lua
return {
  -- 1. Shannon Service IDs
  ["anvil"] = "anvil", -- Anvil (Authoritative ID)

  -- 2. Morse Service IDs
  ["F000"] = "F000",   -- Pocket (Authoritative ID)
  ["pocket"] = "F000", -- Pocket (Alias)
}
```

The key may either be the **authoritative service ID** or an **alias**. The value must be the **authoritative service ID**.

- [`.allowed-services.lua` template file](https://github.com/buildwithgrove/path/tree/main/envoy/allowed-services.template.lua).

:::warning

All service IDs allowed by the PATH instance must be defined in the `.allowed-services.lua` file.

**Requests for services not defined in this file will be rejected.**

:::

### Terminology

- **Authoritative ID**: The service ID that that PATH uses to identify a service.
- **Alias**: A string that resolves to a service ID, which is useful for creating human-readable subdomains for services.

  **Regardless of the method used to pass the service ID or alias to Envoy Proxy, the Envoy Proxy's Lua filter will forward requests to PATH with the `authoritative ID` set in the `target-service-id` header.**

## Envoy Proxy Configuration - `.envoy.yaml`

The `.envoy.yaml` file is used to configure the Envoy Proxy.

Once created in `local/path/envoy`, the `.envoy.yaml` file is mounted as a file in the Envoy Proxy container at `/etc/envoy/.envoy.yaml`.

:::warning

Once configured using the prompts in the `make init_envoy` target, the `.envoy.yaml` file does not require further modification.

**It is not recommended to modify the `.envoy.yaml` file directly unless you know what you are doing.**

:::

:::tip

[For more information on Envoy Proxy configuration file, see the Envoy Proxy documentation.](https://www.envoyproxy.io/docs/envoy/latest/configuration/overview/examples#static)

:::

## Ratelimit Configuration - `.ratelimit.yaml`

The `.ratelimit.yaml` file is used to configure the Ratelimit service.

Once created in `local/path/envoy`, the `.ratelimit.yaml` file is mounted as a file in the Envoy Proxy container at `/etc/envoy/.ratelimit.yaml`.

**To make changes to the Ratelimit service, modify the `.ratelimit.yaml` file.**

### File Format

_`.ratelimit.yaml` format:_
```yaml
---
domain: rl 
descriptors:
  - key: rl-endpoint-id
    descriptors:
      - key: rl-throughput
        value: "30"
        rate_limit:
          unit: second
          requests_per_unit: 30
```

- [`.ratelimit.yaml` template file](https://github.com/buildwithgrove/path/tree/main/envoy/ratelimit.template.yaml).

To add new throughput limits, add a new descriptor array item under the `descriptors` key.

:::tip

[For more information on Rate Limit descriptors, see the documentation in the Envoy Rate Limit repository.](https://github.com/envoyproxy/ratelimit?tab=readme-ov-file#definitions)

:::


## Gateway Endpoints Data - `.gateway-endpoints.yaml`

:::info

A `GatewayEndpoint` is how PATH defines a single endpoint that is authorized to use the PATH service.

It is used to define the **authorization method**, **rate limits**, and **metadata** for an endpoint.

For more information, see the [Gateway Endpoint section in the Envoy docs](../../develop/envoy/introduction.md#gateway-endpoint-authorization).

:::

The `.gateway-endpoints.yaml` file is used to define the Gateway Endpoints that are authorized to make requests to the PATH instance.

Once created in `local/path/envoy`, the `.gateway-endpoints.yaml` file is mounted as a file in the [PATH Auth Data Server (PADS)](https://github.com/buildwithgrove/path-auth-data-server) container at `/app/.gateway-endpoints.yaml`.

:::tip

This YAML file is provided as an easy default way to define Gateway Endpoints to get started with PATH. For more complex use cases, you may wish to use a database as the data source for Gateway Endpoints.

**For more information on how to use a database as the data source for Gateway Endpoints, [see the PATH Auth Data Server (PADS) section of the Envoy docs](../../develop/envoy/introduction.md#path-auth-data-server).**

:::


### File Format

```yaml
endpoints:
  endpoint_1_static_key: 
    auth: 
      api_key: "api_key_1" 

  endpoint_2_jwt:
    auth:
      jwt_authorized_users:
        - "auth0|user_1"
        - "auth0|user_2"

  endpoint_3_no_auth:
    rate_limiting:
      throughput_limit: 30
      capacity_limit: 100000
      capacity_limit_period: "CAPACITY_LIMIT_PERIOD_MONTHLY"
```

- [`.gateway-endpoints.yaml` example file](https://github.com/buildwithgrove/path-auth-data-server/blob/main/yaml/testdata/gateway-endpoints.example.yaml).

- [Gateway Endpoint YAML File Schema](https://github.com/buildwithgrove/path-auth-data-server/blob/main/yaml/gateway-endpoints.schema.yaml).

To define the Gateway Endpoints that are authorized to use the PATH service, edit the `.gateway-endpoints.yaml` file.
