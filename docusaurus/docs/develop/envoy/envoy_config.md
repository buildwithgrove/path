---
sidebar_position: 2
title: Envoy Config
description: Envoy configuration details
---

<div align="center">
  <a href="https://www.envoyproxy.io/docs/envoy/latest/">
    <img src="https://www.envoyproxy.io/docs/envoy/latest/_static/envoy-logo.png" alt="Envoy logo" width="275"/>
  <p><b>Envoy Proxy Docs</b></p>
  </a>
</div>

This document describes the configuration options for PATH's Envoy Proxy which is responsible for:

1. Authorizing incoming requests
2. Rate limiting
3. [Optional] Aliases for service IDs

### tl;dr Just show me the config files <!-- omit in toc -->

There are a total of three files used to configure Envoy Proxy in PATH:

1. `.envoy.yaml` ([template example](https://github.com/buildwithgrove/path/blob/main/envoy/envoy.template.yaml))
2. `.ratelimit.yaml` ([template example](https://github.com/buildwithgrove/path/blob/main/envoy/ratelimit.yaml))
3. `.gateway-endpoints.yaml` ([template example](https://github.com/buildwithgrove/path-auth-data-server/blob/main/yaml/testdata/gateway-endpoints.example.yaml))

:::tip

While **three config** files may seem like a lot, the default templates for **files 1 through 3** should be sufficient for most use cases.

For most use cases, only the `.gateway-endpoints.yaml` file will need to be modified.

:::

:::info

The PATH gateway is configured with its own set of configuration files.

[For detailed information on the PATH gateway configuration, please refer to the PATH Configuration Guide](../path/path_config.md).

:::

## Table of Contents <!-- omit in toc -->

- [Initialization of configuration files](#initialization-of-configuration-files)
- [1. Envoy Proxy Configuration - `.envoy.yaml`](#1-envoy-proxy-configuration---envoyyaml)
- [2. Ratelimit Configuration - `.ratelimit.yaml`](#2-ratelimit-configuration---ratelimityaml)
  - [Ratelimit File Format](#ratelimit-file-format)
  - [Ratelimit Customizations](#ratelimit-customizations)
- [3. Gateway Endpoints Data - `.gateway-endpoints.yaml`](#3-gateway-endpoints-data---gateway-endpointsyaml)
  - [Gateway Endpoint Functionality](#gateway-endpoint-functionality)
  - [Gateway Endpoint File Format](#gateway-endpoint-file-format)

## Initialization of configuration files

The required Envoy configuration files can be generated from their templates by running:

```bash
make init_envoy
```

This will generate the following files in the `local/path/envoy` directory:

1. `.envoy.yaml`
2. `.ratelimit.yaml`
3. `.gateway-endpoints.yaml`

:::note

All of these files are git ignored from the PATH repo as they are specific to
each PATH instance and may contain sensitive information.

:::

## 1. Envoy Proxy Configuration - `.envoy.yaml`

The `.envoy.yaml` file is used to configure the Envoy Proxy.

Once created in `local/path/envoy`, the `.envoy.yaml` file is mounted as a file in the Envoy Proxy container at `/etc/envoy/.envoy.yaml`.

For more information on Envoy Proxy configuration file, see the [Envoy Proxy documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/overview/examples#static).

:::warning

Once configured using the prompts in the `make init_envoy` target, the `.envoy.yaml` file does not require further modification.

**It is not recommended to modify the `.envoy.yaml` file directly unless you know what you are doing.**

:::

## 2. Ratelimit Configuration - `.ratelimit.yaml`

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

## 3. Gateway Endpoints Data - `.gateway-endpoints.yaml`

A `GatewayEndpoint` is how PATH defines a single endpoint that is authorized to use the PATH service.

It is used to define the **authorization method**, **rate limits**, and **metadata** for an endpoint.

For more information, see the [Gateway Endpoint section in the Envoy docs](../envoy/walkthrough.md#gateway-endpoint-authorization).

### Gateway Endpoint Functionality

The `.gateway-endpoints.yaml` file is used to define the Gateway Endpoints that are authorized to make requests to the PATH instance.

Once created in `local/path/envoy`, the `.gateway-endpoints.yaml` file is mounted as a file in the [PATH Auth Data Server (PADS)](https://github.com/buildwithgrove/path-auth-data-server) container at `/app/.gateway-endpoints.yaml`.

:::tip

This YAML file is provided as an easy default way to define Gateway Endpoints to get started with PATH.

For more complex use cases, you may wish to use a database as the data source for Gateway Endpoints.

**For more information on how to use a database as the data source for Gateway Endpoints, [see the PATH Auth Data Server (PADS) section of the Envoy docs](../envoy/walkthrough.md#path-auth-data-server).**

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
