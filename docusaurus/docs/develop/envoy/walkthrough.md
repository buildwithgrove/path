---
sidebar_position: 1
title: Walkthrough
description: High-level architecture overview and detailed walkthrough
---

<div align="center">
<h1>PATH<br/>Authorization & Rate Limiting</h1>
<img src="https://storage.googleapis.com/grove-brand-assets/Presskit/Logo%20Joined-2.png" alt="Grove logo" width="500"/>
</div>
<br/>

## Table of Contents <!-- omit in toc -->

- [Quickstart](#quickstart)
  - [Shannon Quickstart](#shannon-quickstart)
- [Overview](#overview)
  - [Components](#components)
  - [Architecture Diagram](#architecture-diagram)
- [Envoy Proxy](#envoy-proxy)
  - [Envoy Configuration](#envoy-configuration)
  - [Envoy HTTP Filters](#envoy-http-filters)
  - [Request Lifecycle](#request-lifecycle)
- [Service ID Specification](#service-id-specification)
  - [Allowed Services File](#allowed-services-file)
  - [Target Service ID Header](#target-service-id-header)
  - [URL Subdomain](#url-subdomain)
- [Specifying the Gateway Endpoint ID](#specifying-the-gateway-endpoint-id)
  - [URL Path Endpoint ID Extractor](#url-path-endpoint-id-extractor)
  - [Header Endpoint ID Extractor](#header-endpoint-id-extractor)
- [Gateway Endpoint Authorization](#gateway-endpoint-authorization)
  - [JSON Web Token (JWT) Authorization](#json-web-token-jwt-authorization)
  - [API Key Authorization](#api-key-authorization)
  - [No Authorization](#no-authorization)
- [External Auth Server](#external-auth-server)
  - [External Auth Server Sequence Diagram](#external-auth-server-sequence-diagram)
  - [External Auth Server Configuration](#external-auth-server-configuration)
  - [`gateway_endpoint.proto` file](#gateway_endpointproto-file)
  - [Remote gRPC Auth Server](#remote-grpc-auth-server)
    - [PATH Auth Data Server](#path-auth-data-server)
    - [Implementing a Custom Remote gRPC Server](#implementing-a-custom-remote-grpc-server)
- [Rate Limiter](#rate-limiter)
  - [Rate Limit Configuration](#rate-limit-configuration)
  - [Documentation and Examples](#documentation-and-examples)

## Quickstart

### Shannon Quickstart

See the [PATH Quickstart Cheat Sheet](../path/cheatsheet.md) for instructions on how to get started with a local PATH instance on Shannon.

:::info MAKE SURE TO UPDATE THESE FILES

ℹ️ After copying config files by running `make init_envoy`, which will create the required files in the `local/path/envoy` directory, please update the following files:

- `.allowed-services.lua` with the service IDs allowed by your PATH instance
- `.gateway-endpoints.yaml` with your own data

:::

## Overview

This folder contains everything necessary for managing authorization and rate limiting in the PATH service.
Specifically, this is split into two logical parts:

1. The `Envoy Proxy configuration`
2. The `External Auth Server`

### Components

:::tip

A [Tiltfile](https://github.com/buildwithgrove/path/blob/main/Tiltfile) is provided to run all of these services locally.

:::

- **PATH Service**: The service that handles requests after they have been authorized.
- **Envoy Proxy**: A proxy server that handles incoming requests, performs auth checks, and routes authorized requests to the `PATH` service.
- **External Auth Server**: A Go/gRPC server that evaluates whether incoming requests are authorized to access the `PATH` service.
- **Rate Limiter**: A service that coordinates all rate limiting.
- **Redis**: A key-value store used by the rate limiter to share state and coordinate rate limiting across any number of PATH instances behind the same Envoy Proxy.
- **Remote gRPC Server**: A server that provides the external authorization server with data on which endpoints are authorized to use the PATH service.
  - _PADS (PATH Auth Data Server) is provided as a functional implementation of the remote gRPC server that loads data from a YAML file or simple Postgres database._
  - _See [PATH Auth Data Server](#path-auth-data-server) for more information._

### Architecture Diagram

```mermaid
graph TD
    User[/"<big>PATH<br>User</big>"\]
    Envoy[<big>Envoy Proxy</big>]

    AUTH["External Auth Server"]
    AUTH_DECISION{Did<br>Authorize<br>Request?}
    PATH[<big>PATH Service</big>]

    Error[[Error Returned to User]]
    Result[[Result Returned to User]]

    GRPCServer["Remote gRPC Server<br>(e.g. PADS)"]
    GRPCDB[("Postgres<br>Database")]
    GRPCConfig["YAML Config File"]

    subgraph AUTH["External Auth Server"]
        GRPCClient["gRPC Client"]
        Cache["Gateway Endpoint<br>Data Store"]
    end

    User -->|1.Send Request| Envoy
    Envoy -->|2.Authorization Check| AUTH
    AUTH -->|3.Authorization Result| Envoy
    Envoy --> AUTH_DECISION
    AUTH_DECISION -->|4.No <br> Forward Request| Error
    AUTH_DECISION -->|4.Yes <br> Forward Request| PATH
    PATH -->|5.Response| Result

    subgraph DataSource["Gateway Endpoint<br>Data Source<br>"]
        GRPCDB
        GRPCConfig
    end

    GRPCServer <-.-> |Fetch & Stream<br>Gateway Endpoint Data<br>Over gRPC Connection| AUTH
    GRPCServer <-.-> DataSource
```

## Envoy Proxy

<div align="center">
  <a href="https://www.envoyproxy.io/docs/envoy/latest/">
    <img src="https://www.envoyproxy.io/docs/envoy/latest/_static/envoy-logo.png" alt="Envoy logo" width="200"/>
  <p><b>Envoy Proxy Docs</b></p>
  </a>
</div>

PATH uses Envoy Proxy to handle authorization and rate limiting.

The `/envoy` directory houses the configuration files and settings for Envoy Proxy.

Envoy acts as a gateway, handling incoming requests, determining allowed services, performing auth checks, and routing authorized requests to the PATH service.

### Envoy Configuration

See the [Envoy Config docs](../envoy/envoy_config.md) for a simplified explanation of the Envoy Proxy configuration files and how they work together.

### Envoy HTTP Filters

The PATH Auth Server uses the following [Envoy HTTP filters](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/http_filters) to handle authorization:

- **[lua](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/lua_filter)**: Extracts the Service ID from the subdomain of the request's host field and attaches it to the request as the `target-service-id` header.
- **[header_mutation](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/header_mutation_filter)**: Ensures the request does not have the `jwt-user-id` header set before it is forwarded upstream.
  - `header_mutation` is used only if the PATH instance has JWT auth enabled.
- **[jwt_authn](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter)**: Performs JWT verification and sets the `jwt-user-id` header.
  - `jwt_authn` is used only if the PATH instance has JWT auth enabled.
- **[ext_authz](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_authz_filter)**: Performs authorization checks using the PATH Auth Server external authorization server.
- **[ratelimit](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/rate_limit_filter)**: Performs rate limiting checks using the Rate Limiter service.

### Request Lifecycle

```mermaid
sequenceDiagram
    participant Client
    participant Envoy as Envoy<br>Proxy
    participant JWTFilter as JWT HTTP Filter<br>(jwt_authn)
    participant AuthServer as External<br>Auth Server<br>(ext_authz)
    participant RateLimiter as PATH Rate Limiter<br>(ratelimit)
    participant Service as PATH<br>Service

    %% Add bidirectional arrow for Upstream and Downstream
    Note over Client,Service: Downstream <-------------------------------------> Upstream

    Client->>Envoy: 1. Send Request

    opt if JWT is present
      Envoy->>+JWTFilter: 2. Parse JWT (if present)
      JWTFilter->>-Envoy: 3. Return parsed jwt-user-id (if present)
    end

    Envoy->>+AuthServer: 4. Forward Request

    opt if auth is required
      AuthServer->>AuthServer: 5. Authorize (if required)
      AuthServer->>AuthServer: 6. Set Rate Limit headers (if required)
    end

    alt Auth Failed
      AuthServer->>+Envoy: 7a. Auth Failed (if rejected)
      Envoy->>-Client: 7a. Reject Request (Auth Failed)
    else Auth Success
      AuthServer->>-Envoy: 7b. Auth Success (if accepted)
      Envoy->>Envoy: Set Rate Limit descriptors from headers
    end

    Envoy->>+RateLimiter: 8. Perform Rate Limit Check
    RateLimiter->>RateLimiter: 9. Rate Limit Check

    alt Rate Limit Check Failed
      RateLimiter->>Envoy: 10a. Rate Limit Check Failed
      Envoy->>Client: 10a. Reject Request (Rate Limit Exceeded)
    else
      RateLimiter->>-Envoy: 10b. Rate Limit Check Passed
    end

    Envoy->>+Service: 11. Forward Request
    Service->>-Client: 12. Return Response
```

## Service ID Specification

The `target-service-id` header is used to specify the Service ID in the request.

There are two methods for specifying this header in the request:

1. [Target Service ID Header](#target-service-id-header) (_e.g. -H "target-service-id: anvil"_)
2. [URL Subdomain](#url-subdomain) (_e.g. http://anvil.localhost:3070/v1_)

### Allowed Services File

The file `local/path/envoy/.allowed-services.lua` defines the mapping of service IDs to the service IDs used by the PATH service.

All service IDs (and optional service aliases) used by the PATH service must be defined in this file.

:::info `.allowed-services.lua` format

```lua
return {
  -- 1. Shannon Service IDs
  ["anvil"] = "anvil", -- Anvil (Authoritative ID)

  -- 2. Morse Service IDs
  ["F000"] = "F000",   -- Pocket (Authoritative ID)
  ["pocket"] = "F000", -- Pocket (Alias)
}
```

:::

### Target Service ID Header

The service ID (or a configured alias) may be specified in the `target-service-id` header.

:::info _Example request:_

```bash
curl http://localhost:3070/v1 \
  -X POST \
  -H "Content-Type: application/json" \
  -H "target-service-id: anvil" \
  -H "endpoint-id: endpoint_3_no_auth" \
  -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```

:::

### URL Subdomain

The service ID (or a configured alias) may be specified in the URL subdomain.

eg. `host = "anvil.path.grove.city" -> Header: "target-service-id: anvil"`

:::info _Example request:_

```bash
curl http://anvil.localhost:3070/v1 \
  -X POST \
  -H "Content-Type: application/json" \
  -H "endpoint-id: endpoint_3_no_auth" \
  -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```

:::

## Specifying the Gateway Endpoint ID

The Auth Server may extract the Gateway Endpoint ID from the request in one of two ways:

1. [URL Path Endpoint ID Extractor](#url-path-endpoint-id-extractor)
2. [Header Endpoint ID Extractor](#header-endpoint-id-extractor)

This is determined by the **`auth_server.endpoint_id_extractor`** field in the PATH `.config.yaml` file.

:::tip

[For detailed information on the PATH configuration file, see the PATH Config Docs](../path/path_config.md).

:::

One of:

- `url_path` (default)
- `header`

:::warning

Requests are rejected if either of the following are true:

- The `<GATEWAY_ENDPOINT_ID>` is missing
- ID is not present in the `External Auth Server`'s `Gateway Endpoint Store`

:::

:::info

Regardless of which extractor is used, the Gateway Endpoint ID will always be set in the `endpoint-id` header if the request is forwarded to the PATH Service.

:::

### URL Path Endpoint ID Extractor

When using the `url_path` extractor, the Gateway Endpoint ID must be specified in the URL path.

```bash
https://<SERVICE_NAME>.<PATH_DOMAIN>/v1/<GATEWAY_ENDPOINT_ID>
```

For example, if the `SERVICE_NAME` is `eth` and the `GATEWAY_ENDPOINT_ID` is `endpoint_3_no_auth`:

```bash
curl http://anvil.localhost:3070/v1/endpoint_3_no_auth \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```

### Header Endpoint ID Extractor

When using the `header` extractor, the Gateway Endpoint ID must be specified in the `endpoint-id` header.

```bash
-H "endpoint-id: <GATEWAY_ENDPOINT_ID>"
```

For example, if the `endpoint-id` header is set to `endpoint_3_no_auth`:

```bash
curl http://anvil.localhost:3070/v1 \
  -X POST \
  -H "Content-Type: application/json" \
  -H "endpoint-id: endpoint_3_no_auth" \
  -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```

:::tip Example curl requests

A variety of example cURL requests to the PATH service can be found in the [`test_requests.mk` file](https://github.com/buildwithgrove/path/blob/main/makefiles/test_requests.mk).

:::

:::info

`endpoint_3_no_auth` is the endpoint from the example `.gateway-endpoints.yaml` file that requires no authorization.

See the [Gateway Endpoint YAML File](#gateway-endpoint-yaml-file) section for more information on the `GatewayEndpoint` data structure.

:::

<br/>

## Gateway Endpoint Authorization

The `External Auth Server` evaluates whether incoming requests are authorized to access the PATH service based on the `AuthType` field of the `GatewayEndpoint` proto struct.

Three authorization types are supported:

1. [JSON Web Token (JWT) Authorization](#json-web-token-jwt-authorization)
2. [API Key Authorization](#api-key-authorization)
3. [No Authorization](#no-authorization)

### JSON Web Token (JWT) Authorization

For GatewayEndpoints with the `AuthType` field set to `JWT_AUTH`, a valid JWT issued by the auth provider specified in the `envoy.yaml` file is required to access the PATH service.

:::tip

The `make init_envoy` command will prompt you about whether you wish to use JWT authorization.

If you do wish to use it, you will be asked to enter your auth provider's domain and audience.

Auth0 is an example of a JWT issuer that can be used with PATH.

Their docs page on JWTs gives a good overview of the JWT format and how to issue JWTs to your users:

- [Auth0 JWT Docs](https://auth0.com/docs/secure/tokens/json-web-tokens)
  :::

_Example Request Header:_

```bash
-H "Authorization: Bearer <JWT>"
```

The `jwt_authn` filter will verify the JWT and, if valid, set the `jwt-user-id` header from the `sub` claim of the JWT. An invalid JWT will result in an error.

The `External Auth Server` will use the `jwt-user-id` header to make an authorization decision; if the `GatewayEndpoint`'s `Auth.AuthorizedUsers` field contains the `jwt-user-id` value, the request will be authorized.

_Example auth provider user ID header:_

```bash
jwt-user-id: auth0|a12b3c4d5e6f7g8h9
```

:::info

For more information, see the [Envoy JWT Authn Docs](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter)

:::

### API Key Authorization

For GatewayEndpoints with the `AuthType` field set to `API_KEY_AUTH`, a static API key is required to access the PATH service.

_Example Request Header:_

```bash
-H "Authorization: <API_KEY>"
```

The `External Auth Server` will use the `authorization` header to make an authorization decision; if the `GatewayEndpoint`'s `Auth.ApiKey` field matches the `API_KEY` value, the request will be authorized.

### No Authorization

For GatewayEndpoints with the `AuthType` field set to `NO_AUTH`, no authorization is required to access the PATH service.

All requests for GatewayEndpoints with the `AuthType` field set to `NO_AUTH` will be authorized by the `External Auth Server`.

## External Auth Server

The `envoy/auth_server` directory contains the `External Auth Server` called by the Envoy `ext_authz` filter. It evaluates whether incoming requests are authorized to access the PATH service.

This server communicates with a `Remote gRPC Server` to populate its in-memory `Gateway Endpoint Store`, which provides data on which endpoints are authorized to use the PATH service.

### External Auth Server Sequence Diagram

```mermaid
sequenceDiagram
    participant EnvoyProxy as Envoy Proxy<br>(ext_authz filter)
    participant GoAuthServer as External<br>Auth Server
    participant RemoteGRPC as Remote gRPC Server<br>(eg. PADS)
    participant DataSource as Data Source<br>(YAML, Postgres, etc.)

    %% Grouping "Included in PATH"
    Note over EnvoyProxy, GoAuthServer: Included in PATH

    %% Grouping "Must be implemented by operator"
    Note over RemoteGRPC, DataSource: Must be implemented by operator<br>(PADS Docker image available)

    DataSource-->>RemoteGRPC: User/Authorization Data
    RemoteGRPC<<-->>GoAuthServer: Populate Gateway Endpoint Store
    EnvoyProxy->>+GoAuthServer: 1. Check Request
    GoAuthServer->>GoAuthServer: 1a. Authorize Request
    GoAuthServer->>GoAuthServer: 1b. Set Rate Limit Headers
    GoAuthServer->>-EnvoyProxy: 2. Check Response<br>(Approve/Reject)
```

### External Auth Server Configuration

The external authorization server loads its config from the `auth_server_config` section of PATH's `.config.yaml` file. For the local Tilt stack, this file is located at `local/path/config/.config.yaml`.

_`.config.yaml auth_server_config` example:_

```yaml
auth_server_config:
  grpc_host_port: path-auth-data-server:50051
  grpc_use_insecure_credentials: true
  endpoint_id_extractor_type: url_path
```

See the [PATH Configuration](../path/walkthrough.md#configuration) section for more information on the `.config.yaml` file.

For more information on external authorization using Envoy Proxy, see the Envoy documentation:

- [Envoy External Authorization Docs](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_authz_filter)
- [Envoy Go Control Plane Auth Package](https://pkg.go.dev/github.com/envoyproxy/go-control-plane@v0.13.0/envoy/service/auth/v3)

### `gateway_endpoint.proto` file

Both the `External Auth Server` and the `Remote gRPC Server` use the gRPC service and types defined in the [`gateway_endpoint.proto`](https://github.com/buildwithgrove/path/blob/main/envoy/auth_server/proto/gateway_endpoint.proto) file.

This service defines two main methods for populating the `External Auth Server`'s `Gateway Endpoint Store`:

```protobuf
service GatewayEndpoints {
  // GetInitialData requests the initial set of GatewayEndpoints from the remote gRPC server.
  rpc GetInitialData(InitialDataRequest) returns (InitialDataResponse);

  // StreamUpdates listens for updates from the remote gRPC server and streams them to the client.
  rpc StreamUpdates(UpdatesRequest) returns (stream Update);
}
```

### Remote gRPC Auth Server

The `Remote gRPC Server` is responsible for providing the `External Auth Server` with data on which endpoints are authorized to use the PATH service.

**The PATH Auth Data Server (PADS) is a pre-built implementation of the `Remote gRPC Server` that should be sufficient for most users.**

#### PATH Auth Data Server

The [PADS repo](https://github.com/buildwithgrove/path-auth-data-server) provides a pre-built implementation of the remote gRPC server.

:::info Repository

See the [PADS Documentation](../pads/introduction.md) for more information on the PATH Auth Data Server.

:::

This service is available as a Docker image and may be configured to load data from a YAML file or using an opinionated Postgres database driver.

```bash
ghcr.io/buildwithgrove/path-auth-data-server:latest
```

_This Docker image is loaded by default in the [Tiltfile](https://github.com/buildwithgrove/path/blob/main/Tiltfile) file at the root of the PATH repo._

#### Implementing a Custom Remote gRPC Server

If the Gateway operator wishes to implement a custom remote gRPC server, the implementation must import the Go `github.com/buildwithgrove/path/envoy/auth_server/proto` package, which is autogenerated from the [`gateway_endpoint.proto`](https://github.com/buildwithgrove/path/blob/main/envoy/auth_server/proto/gateway_endpoint.proto) file.

The custom implementation must use the methods defined in the `GatewayEndpoints` service:

- `FetchAuthDataSync`
- `StreamAuthDataUpdates`

:::tip

If you wish to implement your own custom database driver, forking the PADS repo is the easiest way to get started, though any gRPC server implementation that adheres to the `gateway_endpoint.proto` service definition should suffice.

:::

## Rate Limiter

### Rate Limit Configuration

1. The `External Auth Server` sets the `rl-endpoint-id` and `rl-throughput` headers if the `GatewayEndpoint` for the request should be rate limited.

2. Envoy Proxy is configured to forward the `rl-endpoint-id` and `rl-throughput` headers to the rate limiter service as descriptors.

   _.envoy.yaml_

   ```yaml
   rate_limits:
     - actions:
         - request_headers:
             header_name: "rl-endpoint-id"
             descriptor_key: "rl-endpoint-id"
         - request_headers:
             header_name: "rl-throughput"
             descriptor_key: "rl-throughput"
   ```

3. Rate limiting is configured through the `.ratelimit.yaml` file.

   _.ratelimit.yaml_

   ```yaml
   domain: rl
   descriptors:
     - key: rl-endpoint-id
       descriptors:
         - key: rl-throughput
           value: "PLAN_FREE"
           rate_limit:
             unit: second
             requests_per_unit: 30
   ```

:::info

The default throughput limit is **30 requests per second** for GatewayEndpoints with the `PLAN_FREE` plan type based on the `rl-endpoint-id` and `rl-throughput` descriptors.

_The rate limiting configuration may be configured to suit the needs of the Gateway Operator in the `.ratelimit.yaml` file._

:::

### Documentation and Examples

As Envoy's rate limiting configuration is fairly complex, this blog article provides a good overview of the configuration options:

- [Understanding Envoy Rate Limits](https://www.aboutwayfair.com/tech-innovation/understanding-envoy-rate-limits)

For more advanced configuration options, refer to the Envoy documentation:

- [Envoy Proxy Rate Limit Docs](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/rate_limit_filter)

- [Envoy Rate Limit Github](https://github.com/envoyproxy/ratelimit)
