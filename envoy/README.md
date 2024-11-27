<div align="center">
<h1>PATH<br/>Authorization & Rate Limiting</h1>
<img src="https://storage.googleapis.com/grove-brand-assets/Presskit/Logo%20Joined-2.png" alt="Grove logo" width="500"/>
</div>
<br/>

<!-- TODO_UPNEXT(@commoddity): Document a cheatsheet to have a envoy/cheatsheet.md for a very quick & simple copy-pasta -->

# Table of Contents <!-- omit in toc -->

- [1. Overview](#1-overview)
  - [1.1. Components](#11-components)
  - [1.2 URL Format](#12-url-format)
- [2. Quickstart](#2-quickstart)
- [3. Envoy Proxy](#3-envoy-proxy)
  - [3.1. Contents](#31-contents)
  - [3.2. Envoy HTTP Filters](#32-envoy-http-filters)
  - [3.3. Request Lifecycle](#33-request-lifecycle)
- [4. Gateway Endpoint Authorization](#4-gateway-endpoint-authorization)
  - [4.1 JSON Web Token (JWT) Authorization](#41-json-web-token-jwt-authorization)
  - [4.2 API Key Authorization](#42-api-key-authorization)
  - [4.3 No Authorization](#43-no-authorization)
- [5. External Authorization Server](#5-external-authorization-server)
  - [5.1. Gateway Endpoints gRPC Service](#51-gateway-endpoints-grpc-service)
  - [5.2. Remote gRPC Auth Server](#52-remote-grpc-auth-server)
    - [5.2.1. PATH Auth Data Server](#521-path-auth-data-server)
    - [5.2.2. Gateway Endpoint YAML File](#522-gateway-endpoint-yaml-file)
    - [5.2.3. Implementing a Custom Remote gRPC Server](#523-implementing-a-custom-remote-grpc-server)
- [6. Rate Limiter](#6-rate-limiter)
  - [6.1. Rate Limit Configuration](#61-rate-limit-configuration)
  - [6.2. Documentation and Examples](#62-documentation-and-examples)

## 1. Overview

This folder contains everything necessary for managing authorization and rate limiting in the PATH service.
Specifically, this is split into two logical parts:

1. The `Envoy Proxy configuration`
2. The `Go External Authorization Server`

### 1.1. Components

- **PATH Service**: The service that handles requests after they have been authorized.
- **Envoy Proxy**: A proxy server that handles incoming requests, performs auth checks, and routes authorized requests to the `PATH` service.
- **External Authorization Server**: A Go/gRPC server that evaluates whether incoming requests are authorized to access the `PATH` service.
- **Rate Limiter**: A service that coordinates all rate limiting.
- **Redis**: A key-value store used by the rate limiter to share state and coordinate rate limiting across any number of PATH instances behind the same Envoy Proxy.
- **Remote gRPC Server**: A server that provides the external authorization server with data on which endpoints are authorized to use the PATH service.
  - _PADS (PATH Auth Data Server) is provided as a functional implementation of the remote gRPC server that loads data from a YAML file or simple Postgres database._
  - _See [5.2.1. PATH Auth Data Server](#521-path-auth-data-server) for more information._

A [docker-compose.yaml](./docker-compose.yaml) file is provided to run all of these services locally.

```mermaid
graph TD
    User@{ shape: trapezoid, label: "<big>PATH<br>User</big>" }
    Envoy[<big>Envoy Proxy</big>]

    AUTH["Auth Server <br> "]
    AUTH_DECISION{Did<br>Authorize<br>Request?}
    PATH[<big>PATH Service</big>]

    Error[[Error Returned to User]]
    Result[[Result Returned to User]]

    GRPCServer["Remote gRPC Server<br>(eg. PADS)"]
    GRPCDB[("Postgres<br>Database")]
    GRPCConfig@{ shape: notch-rect, label: "YAML Config File" }

    subgraph AUTH["Auth Server (ext_authz)"]
        GRPCClient["gRPC Client"]
         Cache@{ shape: odd, label: "Gateway Endpoint<br>Data Store" }
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

### 1.2 URL Format

When auth is enabled, the required URL format for the PATH service is:

```
https://<SERVICE_NAME>.<PATH_DOMAIN>/v1/<GATEWAY_ENDPOINT_ID>
```

For example, if `GATEWAY_ENDPOINT_ID` is `a1b2c3d4`:

```
https://eth-mainnet.rpc.grove.city/v1/a1b2c3d4
```

Requests are rejected if:

- The `<GATEWAY_ENDPOINT_ID>` is missing
- ID is not present in `Gateway Endpoint Store`
- ID is not returned by `Go External Authorization Server`

## 2. Quickstart

1. Run `make init_envoy` to create all the required config files
   - `envoy.yaml` is created with your auth provider's domain and audience.
   - `auth_server/.env` is created with the host and port of the provided remote gRPC server.
   - `gateway-endpoints.yaml` is populated with example data; you can modify this to your needs.
2. Run `make path_up` to start the services with all auth and rate limiting dependencies.

> **Tip:** For instructions on how to run PATH without any auth or rate limiting, see the [PATH README - Quickstart Section](../README.md#quickstart).

## 3. Envoy Proxy

<div align="center">
  <a href="https://www.envoyproxy.io/docs/envoy/latest/">
    <img src="https://www.envoyproxy.io/theme/images/envoy-logo.svg" alt="Envoy logo" width="200"/>
  <p><b>Envoy Proxy Docs</b></p>
  </a>
</div>

PATH uses Envoy Proxy to handle authorization and rate limiting.

The `/envoy` directory houses the configuration files and settings for Envoy Proxy.

Envoy acts as a gateway, handling incoming requests, performing auth checks, and routing authorized requests to the PATH service.

### 3.1. Contents

- **envoy.template.yaml**: A template configuration file for Envoy Proxy.
  - To create `envoy.yaml`, run `make copy_envoy_config`.
  - This will prompt you to enter your auth provider's domain and audience and will output the result to `envoy.yaml`.
  - `envoy.yaml` is Git ignored as it contains sensitive information.
- **gateway-endpoints.example.yaml**: An example file containing data on which endpoints are authorized to use the PATH service.
  - To create `gateway-endpoints.yaml`, run `make copy_envoy_gateway_endpoints`.
  - This file is only required if loading `GatewayEndpoint` data from a YAML file and used to load data in the `external authorization server` from the `remote gRPC server`.
  - `gateway-endpoints.yaml` is Git ignored as it may contain sensitive information.
- **ratelimit.yaml**: Configuration for the rate limiting service.

### 3.2. Envoy HTTP Filters

The PATH Auth Server uses the following [Envoy HTTP filters](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/http_filters) to handle authorization:

- **[header_mutation](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/header_mutation_filter)**: Ensures the request does not have the `x-jwt-user-id` header set before it is forwarded upstream.
- **[jwt_authn](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter)**: Performs JWT verification and sets the `x-jwt-user-id` header.
- **[ext_authz](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_authz_filter)**: Performs authorization checks using the PATH Auth Server external authorization server.
- **[ratelimit](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/rate_limit_filter)**: Performs rate limiting checks using the Rate Limiter service.

### 3.3. Request Lifecycle

```mermaid
sequenceDiagram
    participant Client
    participant Envoy as Envoy<br>Proxy
    participant JWTFilter as JWT HTTP Filter<br>(jwt_authn)
    participant AuthServer as PATH Auth Server<br>(ext_authz)
    participant RateLimiter as PATH Rate Limiter<br>(ratelimit)
    participant Service as PATH<br>Service

    %% Add bidirectional arrow for Upstream and Downstream
    Note over Client,Service: Downstream <-------------------------------------> Upstream

    Client->>Envoy: 1. Send Request
    Envoy->>JWTFilter: 2. Parse JWT (if present)
    JWTFilter-->>Envoy: 3. Return parsed x-jwt-user-id (if present)
    Envoy->>AuthServer: 4. Forward Request
    AuthServer->>AuthServer: 5. Authorize (if required)
    AuthServer->>AuthServer: 6. Set Rate Limit headers (if required)
    AuthServer-->>Envoy: 7a. Auth Failed (if rejected)
    Envoy-->>Client: 7a. Reject Request (Auth Failed)
    AuthServer-->>Envoy: 7b. Auth Success (if accepted)
    Envoy->>Envoy: Set Rate Limit descriptors from headers
    Envoy->>RateLimiter: 8. Perform Rate Limit Check
    RateLimiter->>RateLimiter: 9. Rate Limit Check
    RateLimiter-->>Envoy: 10a. Rate Limit Check Failed
    Envoy-->>Client: 10a. Reject Request (Rate Limit Exceeded)
    RateLimiter-->>Envoy: 10b. Rate Limit Check Passed
    Envoy->>Service: 11. Forward Request
    Service-->>Client: 12. Return Response
```

## 4. Gateway Endpoint Authorization

The `Go External Authorization Server` evaluates whether incoming requests are authorized to access the PATH service based on the `AuthType` field of the `GatewayEndpoint` proto struct.

Three authorization types are supported:

- [JSON Web Token (JWT) Authorization](#41-json-web-token-jwt-authorization)
- [API Key Authorization](#42-api-key-authorization)
- [No Authorization](#43-no-authorization)

### 4.1 JSON Web Token (JWT) Authorization

For GatewayEndpoints with the `AuthType` field set to `JWT_AUTH`, a valid JWT issued by the auth provider specified in the `envoy.yaml` file is required to access the PATH service.

_Example Request Header:_

```bash
-H "Authorization: Bearer <JWT>"
```

The `jwt_authn` filter will verify the JWT and, if valid, set the `x-jwt-user-id` header from the `sub` claim of the JWT. An invalid JWT will result in an error.

The `Go External Authorization Server` will use the `x-jwt-user-id` header to make an authorization decision; if the `GatewayEndpoint`'s `Auth.AuthorizedUsers` field contains the `x-jwt-user-id` value, the request will be authorized.

_Example auth provider user ID header:_

```
x-jwt-user-id: auth0|a12b3c4d5e6f7g8h9
```

For more information, see:

- [Envoy JWT Authn Docs](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter)

### 4.2 API Key Authorization

For GatewayEndpoints with the `AuthType` field set to `API_KEY_AUTH`, a static API key is required to access the PATH service.

_Example Request Header (both Bearer and non-Bearer are supported):_

```bash
-H "Authorization: <API_KEY>"
-H "Authorization: Bearer <API_KEY>"
```

The `Go External Authorization Server` will use the `authorization` header to make an authorization decision; if the `GatewayEndpoint`'s `Auth.ApiKey` field matches the `API_KEY` value, the request will be authorized.

### 4.3 No Authorization

For GatewayEndpoints with the `AuthType` field set to `NO_AUTH`, no authorization is required to access the PATH service.

All requests for GatewayEndpoints with the `AuthType` field set to `NO_AUTH` will be authorized by the `Go External Authorization Server`.

## 5. External Authorization Server

The `envoy/auth_server` directory contains the `Go External Authorization Server` called by the Envoy `ext_authz` filter. It evaluates whether incoming requests are authorized to access the PATH service.

This server communicates with a `Remote gRPC Server` to populate its in-memory `Gateway Endpoint Store`, which provides data on which endpoints are authorized to use the PATH service.

```mermaid
sequenceDiagram
    participant EnvoyProxy as Envoy Proxy<br>(ext_authz filter)
    participant GoAuthServer as Go External<br>Authorization Server
    participant RemoteGRPC as Remote gRPC Server<br>(eg. PADS)
    participant DataSource as Data Source<br>(YAML, Postgres, etc.)

    %% Grouping "Included in PATH"
    Note over EnvoyProxy, GoAuthServer: Included in PATH

    %% Grouping "Must be implemented by operator"
    Note over RemoteGRPC, DataSource: Must be implemented by operator<br>(PADS Docker image available)

    DataSource-->>RemoteGRPC: Get Data
    RemoteGRPC<<-->>GoAuthServer: Populate Gateway Endpoint Store
    EnvoyProxy->>GoAuthServer: 1. Check Request
    GoAuthServer->>GoAuthServer: 1a. Authorize Request
    GoAuthServer->>GoAuthServer: 1b. Set Rate Limit Headers
    GoAuthServer->>EnvoyProxy: 2. Check Response<br>(Approve/Deny)
```

The external authorization server requires the following environment variables to be set:

- `GRPC_HOST_PORT`: The host and port of the remote gRPC server.
- `GRPC_USE_INSECURE`: Set to `true` if the remote gRPC server does not use TLS (default: `false`).

Run `make copy_envoy_env` to create the `.env` file needed to run the external authorization server locally in Docker.

For more information, see:

- [Envoy External Authorization Docs](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_authz_filter)
- [Envoy Go Control Plane Auth Package](https://pkg.go.dev/github.com/envoyproxy/go-control-plane@v0.13.0/envoy/service/auth/v3)

### 5.1. Gateway Endpoints gRPC Service

Both the `Go External Authorization Server` and the `Remote gRPC Server` use the gRPC service and types defined in the [`gateway_endpoint.proto`](./auth_server/proto/gateway_endpoint.proto) file.

This service defines two main methods for populating the `Go External Authorization Server`'s `Gateway Endpoint Store`:

```proto
service GatewayEndpoints {
  // GetInitialData requests the initial set of GatewayEndpoints from the remote gRPC server.
  rpc GetInitialData(InitialDataRequest) returns (InitialDataResponse);

  // StreamUpdates listens for updates from the remote gRPC server and streams them to the client.
  rpc StreamUpdates(UpdatesRequest) returns (stream Update);
}
```

### 5.2. Remote gRPC Auth Server

The `Remote gRPC Server` is responsible for providing the `Go External Authorization Server` with data on which endpoints are authorized to use the PATH service.

**The implementation of the remote gRPC server is up to the Gateway operator.**

#### 5.2.1. PATH Auth Data Server

[The PADS repo provides a functioning implementation of the remote gRPC server.](https://github.com/buildwithgrove/path-auth-data-server)

This service is available as a Docker image and may be configured to load data from a YAML file or using a simple Postgres database that adheres to the provided minimal schema.

**Docker Image Registry:**

```bash
ghcr.io/buildwithgrove/path-auth-data-server:latest
```

<!-- TODO_NEXT(@commoddity): Update this section to refer to Tilt instead of docker-compose.yml once Envoy Tilt PR reconciled with `main` -->

_This Docker image is loaded by default in the [docker-compose.yml](../docker-compose.yml#L90) file at the root of the PATH repo._

If the Gateway Operator wishes to implement a custom remote gRPC server, see the [Implementing a Custom Remote gRPC Server](#523-implementing-a-custom-remote-grpc-server) section.

#### 5.2.2. Gateway Endpoint YAML File

_`PADS` loads data from the Gateway Endpoints YAML file specified by the `YAML_FILEPATH` environment variable._

[An example `gateway-endpoints.yaml` file may be seen in the PADS repo](https://github.com/buildwithgrove/path-auth-data-server/blob/main/yaml/testdata/gateway-endpoints.example.yaml).

```yaml
endpoints:
  # 1. Example of a gateway endpoint using API Key Authorization
  endpoint_1:
    auth:
      auth_type: "AUTH_TYPE_API_KEY"
      api_key: "api_key_1"

  # 2. Example of a gateway endpoint using JWT Authorization
  endpoint_2:
    auth:
      auth_type: "AUTH_TYPE_JWT"
      jwt_authorized_users:
        - "auth0|user_1"
        - "auth0|user_2"

  # 3. Example of a gateway endpoint with no authorization and rate limiting set
  endpoint_3:
    rate_limiting:
      throughput_limit: 30
      capacity_limit: 100000
      capacity_limit_period: "CAPACITY_LIMIT_PERIOD_MONTHLY"
```

_In this example:_

- `endpoint_1` is authorized with a static API Key
- `endpoint_2` is authorized using an auth-provider issued JWT for two users
- `endpoint_3` requires no authorization and has a rate limit set

The contents of this file represent the gateway endpoints that are authorized to use the PATH service for a specific gateway operator.

#### 5.2.3. Implementing a Custom Remote gRPC Server

If the Gateway operator wishes to implement a custom remote gRPC server, the implementation must import the Go `github.com/buildwithgrove/path/envoy/auth_server/proto` package, which is autogenerated from the [`gateway_endpoint.proto`](./auth_server/proto/gateway_endpoint.proto) file.

The custom implementation must use the methods defined in the `GatewayEndpoints` service:

- `FetchAuthDataSync`
- `StreamAuthDataUpdates`

Forking the PADS repo is the easiest way to get started, though any gRPC server implementation that adheres to the `gateway_endpoint.proto` service definition should suffice.

## 6. Rate Limiter

Rate limiting is configured through the [`/envoy/ratelimit.yaml`](./ratelimit.yaml) file.

The default throughput limit is 30 requests per second for GatewayEndpoints with the `PLAN_FREE` plan type.

### 6.1. Rate Limit Configuration

1. The `Go External Authorization Server` sets the `x-rl-endpoint-id` and `x-rl-plan` headers if the `GatewayEndpoint` for the request should be rate limited.

2. Envoy Proxy is configured to forward the `x-rl-endpoint-id` and `x-rl-plan` headers to the rate limiter service as descriptors.

   ```yaml
   rate_limits:
     - actions:
         - request_headers:
             header_name: "x-rl-endpoint-id"
             descriptor_key: "x-rl-endpoint-id"
         - request_headers:
             header_name: "x-rl-plan"
             descriptor_key: "x-rl-plan"
   ```

   _envoy.yaml_

3. The rate limiter service is configured to limit the rate for `PLAN_FREE` GatewayEndpoints to 30 requests per second based on the `x-rl-endpoint-id` and `x-rl-plan` descriptors.

   ```yaml
   domain: rl
   descriptors:
     - key: x-rl-endpoint-id
       descriptors:
         - key: x-rl-plan
           value: "PLAN_FREE"
           rate_limit:
             unit: second
             requests_per_unit: 30
   ```

   _ratelimit.yaml_

### 6.2. Documentation and Examples

As Envoy's rate limiting configuration is fairly complex, this blog article provides a good overview of the configuration options:

- [Understanding Envoy Rate Limits](https://www.aboutwayfair.com/tech-innovation/understanding-envoy-rate-limits)

For more advanced configuration options, refer to the Envoy documentation:

- [Envoy Proxy Rate Limit Docs](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/rate_limit_filter)

- [Envoy Rate Limit Github](https://github.com/envoyproxy/ratelimit)
