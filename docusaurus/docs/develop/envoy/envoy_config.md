---
title: Envoy Config
sidebar_position: 2
---

<div align="center">
  <a href="https://www.envoyproxy.io/docs/envoy/latest/">
    <img src="https://www.envoyproxy.io/theme/images/envoy-logo.svg" alt="Envoy logo" width="275"/>
  <p><b>Envoy Proxy Docs</b></p>
  </a>
</div>

This document describes the configuration options for PATH's Envoy Proxy which is responsible for:

1. Defining the set of allowed services
2. Authorizing incoming requests
3. Rate limiting

### tl;dr Just show me the config files <!-- omit in toc -->

There are a total of four files used to configure Envoy Proxy in PATH:

1. `.allowed-services.lua` ([template example](https://github.com/buildwithgrove/path/blob/main/envoy/allowed-services.template.lua))
2. `.envoy.yaml` ([template example](https://github.com/buildwithgrove/path/blob/main/envoy/envoy.template.yaml))
3. `.ratelimit.yaml` ([template example](https://github.com/buildwithgrove/path/blob/main/envoy/ratelimit.yaml))
4. `.gateway-endpoints.yaml` ([template example](https://github.com/buildwithgrove/path-auth-data-server/blob/main/yaml/testdata/gateway-endpoints.example.yaml))

## Table of Contents <!-- omit in toc -->

- [Initialization of configuration files](#initialization-of-configuration-files)
- [Allowed Services - `.allowed-services.lua`](#allowed-services---allowed-serviceslua)
  - [Terminology](#terminology)
  - [Allowed Services Functionality](#allowed-services-functionality)
  - [Allowed Services File Format](#allowed-services-file-format)
- [Envoy Proxy Configuration - `.envoy.yaml`](#envoy-proxy-configuration---envoyyaml)
- [Ratelimit Configuration - `.ratelimit.yaml`](#ratelimit-configuration---ratelimityaml)
  - [Ratelimit File Format](#ratelimit-file-format)
  - [Ratelimit Customizations](#ratelimit-customizations)
- [Gateway Endpoints Data - `.gateway-endpoints.yaml`](#gateway-endpoints-data---gateway-endpointsyaml)
  - [Gateway Endpoint Functionality](#gateway-endpoint-functionality)
  - [Gateway Endpoint File Format](#gateway-endpoint-file-format)

## Initialization of configuration files

The required Envoy configuration files can be generated from their templates by running:

```bash
make init_envoy
```

This will generate the following files in the `local/path/envoy` directory:

- `.allowed-services.lua`
- `.envoy.yaml`
- `.ratelimit.yaml`
- `.gateway-endpoints.yaml`

:::note

All of these files are git ignored from the PATH repo as they are specific to
each PATH instance and may contain sensitive information.

:::

## Allowed Services - `.allowed-services.lua`

The `.allowed-services.lua` file is used to define the allowed services for the Envoy Proxy.

Once created in `local/path/envoy`, the `.allowed-services.lua` file is mounted as a
file in the Envoy Proxy container at `/etc/envoy/.allowed-services.lua`.

### Terminology

- **Authoritative ID**: The service ID that that PATH uses to identify a service.
- **Alias**: A string that resolves to a service ID, useful for creating human-readable subdomains for services.

The **key** in the config file may either be the **authoritative service ID** or an **alias**.

The **value** in the config file must be the **authoritative service ID**.

### Allowed Services Functionality

The Envoy Proxy's Lua filter will forward requests to PATH with the `authoritative ID` set in the `target-service-id` header.

For more information, see the [Service ID Specification section of the Envoy Proxy documentation](../../develop/envoy/introduction.md#service-id-specification).

:::warning

All service IDs allowed by the PATH instance must be defined in the `.allowed-services.lua` file.

**Requests for services not defined in this file will be rejected.**

:::

### Allowed Services File Format

Below is the expect file format for `.allowed-services.lua`.
You can find a template file [here](https://github.com/buildwithgrove/path/tree/main/envoy/allowed-services.template.lua).

```lua
return {
  -- 1. Shannon Service IDs
  ["anvil"] = "anvil", -- Anvil (Authoritative ID)

  -- 2. Morse Service IDs
  ["F000"] = "F000",   -- Pocket (Authoritative ID)
  ["pocket"] = "F000", -- Pocket (Alias)
}
```

## Envoy Proxy Configuration - `.envoy.yaml`

The `.envoy.yaml` file is used to configure the Envoy Proxy.

Once created in `local/path/envoy`, the `.envoy.yaml` file is mounted as a file in the Envoy Proxy container at `/etc/envoy/.envoy.yaml`.

For more information on Envoy Proxy configuration file, see the [Envoy Proxy documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/overview/examples#static).

:::warning

Once configured using the prompts in the `make init_envoy` target, the `.envoy.yaml` file does not require further modification.

**It is not recommended to modify the `.envoy.yaml` file directly unless you know what you are doing.**

:::

## Ratelimit Configuration - `.ratelimit.yaml`

The `.ratelimit.yaml` file is used to configure the Ratelimit service.

Once created in `local/path/envoy`, the `.ratelimit.yaml` file is mounted as a file in the Envoy Proxy container at `/etc/envoy/.ratelimit.yaml`.

**To make changes to the Ratelimit service, modify the `.ratelimit.yaml` file.**

### Ratelimit File Format

Below is the expect file format for `.ratelimit.yaml`.
You can find a template file [here](https://github.com/buildwithgrove/path/tree/main/envoy/ratelimit.template.lua).

For more information on Rate Limit descriptors, see the [documentation in the Envoy Rate Limit repository](https://github.com/envoyproxy/ratelimit?tab=readme-ov-file#definitions).

`.ratelimit.yaml` format:

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

### Ratelimit Customizations

To add new throughput limits, add a new descriptor array item under the `descriptors` key.

For more information on Rate Limit descriptors, see the [documentation in the Envoy Rate Limit repository](https://github.com/envoyproxy/ratelimit?tab=readme-ov-file#definitions).

## Gateway Endpoints Data - `.gateway-endpoints.yaml`

A `GatewayEndpoint` is how PATH defines a single endpoint that is authorized to use the PATH service.

It is used to define the **authorization method**, **rate limits**, and **metadata** for an endpoint.

For more information, see the [Gateway Endpoint section in the Envoy docs](../../develop/envoy/introduction.md#gateway-endpoint-authorization).

### Gateway Endpoint Functionality

The `.gateway-endpoints.yaml` file is used to define the Gateway Endpoints that are authorized to make requests to the PATH instance.

Once created in `local/path/envoy`, the `.gateway-endpoints.yaml` file is mounted as a file in the [PATH Auth Data Server (PADS)](https://github.com/buildwithgrove/path-auth-data-server) container at `/app/.gateway-endpoints.yaml`.

:::tip

This YAML file is provided as an easy default way to define Gateway Endpoints to get started with PATH.

For more complex use cases, you may wish to use a database as the data source for Gateway Endpoints.

**For more information on how to use a database as the data source for Gateway Endpoints, [see the PATH Auth Data Server (PADS) section of the Envoy docs](../../develop/envoy/introduction.md#path-auth-data-server).**

:::

### Gateway Endpoint File Format

Below is the expect file format for `.gateway-endpoints.yaml`.
You can find an example file [here](https://github.com/buildwithgrove/path-auth-data-server/blob/main/yaml/testdata/gateway-endpoints.example.yaml) which
uses [this schema](https://github.com/buildwithgrove/path-auth-data-server/blob/main/yaml/gateway-endpoints.schema.yaml).

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
