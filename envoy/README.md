<div align="center">
<h1>PATH<br/>Authorization & Rate Limiting</h1>
<img src="https://storage.googleapis.com/grove-brand-assets/Presskit/Logo%20Joined-2.png" alt="Grove logo" width="500"/>

</div>
<br/>

# Table of Contents <!-- omit in toc -->
****
- [1. Overview](#1-overview)
  - [1.1. Components](#11-components)
  - [1.2 URL Format](#12-url-format)
- [2. Quickstart](#2-quickstart)
- [3. Envoy Proxy](#3-envoy-proxy)
  - [3.1. Contents](#31-contents)
  - [3.2. Envoy HTTP Filters](#32-envoy-http-filters)
  - [3.3. Request Lifecycle](#33-request-lifecycle)
- [4. JSON Web Token (JWT) Verification](#4-json-web-token-jwt-verification)
- [5. External Authorization Server](#5-external-authorization-server)
  - [5.1. Remote gRPC Server](#51-remote-grpc-server)
    - [5.1.1. Example Gateway Endpoint Data File](#511-example-gateway-endpoint-data-file)
  - [5.2. Environment Variables](#52-environment-variables)
- [6. Rate Limiter](#6-rate-limiter)
- [7. Architecture](#7-architecture)


## 1. Overview

This folder contains the `Envoy Proxy configuration` and the `Go External Authorization Server` required for managing authorization and rate limiting in the PATH service.

### 1.1. Components

- **Envoy Proxy**: A proxy server that handles incoming requests, performs auth checks, and routes authorized requests to the PATH service.
- **External Authorization Server**: A Go/gRPC server that evaluates whether incoming requests are authorized to access the PATH service.
- **Rate Limiter**: A service that coordinates rate limiting among all services.
- **Redis**: A database used by the rate limiter to coordinate rate limiting among all services.
- **PATH Service**: The service that handles requests after they have been authorized.
- **Remote gRPC Server** *(Implemented by Gateway Operator)*: A server that provides the external authorization server with data on which endpoints are authorized to use the PATH service.
  - _A basic implementation of the remote gRPC server that loads data from a YAML file is available as a Docker image._

A [docker-compose.yaml](./docker-compose.yaml) file is provided to run all of these services locally.

_[See the Architecture Diagram for more information on how these components interact.](#7-architecture)_

### 1.2 URL Format

When auth is enabled, the required URL format for the PATH service is:

```
https://<SERVICE_NAME>.<PATH_DOMAIN>/v1/<GATEWAY_ENDPOINT_ID>
```
eg.
```
https://eth-mainnet.rpc.grove.city/v1/a1b2c3d4
```
_In this example the `GATEWAY_ENDPOINT_ID` is `a1b2c3d4`._

Requests that are either missing the `<GATEWAY_ENDPOINT_ID>` or which contain an ID that is not present in the `Go External Authorization Server`'s `Gateway Endpoint Store` will be rejected.

## 2. Quickstart
1. Create all required config files by running `make init_envoy`.
   - `envoy.yaml` is created with your auth provider's domain and audience.
   - `auth_server/.env` is created with the host and port of the provided remote gRPC server.
   - `gateway-endpoints.yaml` is populated with example data; you can modify this to your needs.
2. Run `make path_up` to start the services with all auth and rate limiting dependencies.

*For instructions on how to run PATH without any auth or rate limiting, see the [PATH README - Quickstart Section](../README.md#quickstart).*

## 3. Envoy Proxy

<div align="center">
  <a href="https://www.envoyproxy.io/docs/envoy/latest/">
    <img src="https://www.envoyproxy.io/theme/images/envoy-logo.svg" alt="Envoy logo" width="200"/>
  <p><b>Envoy Proxy Docs</b></p>
  </a>
</div>

PATH uses Envoy Proxy to handle authorization and rate limiting. 

The `/envoy` directory houses the configuration files and settings for Envoy Proxy. Envoy acts as a gateway, handling incoming requests, performing auth checks, and routing authorized requests to the PATH service.

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
    participant Envoy
    participant JWTFilter as JWT HTTP Filter
    participant AuthServer as PATH Auth Server
    participant RateLimiter as PATH Rate Limiter
    participant Service as PATH Service

    Client->>Envoy: 1. Send Request
    Envoy->>JWTFilter: 2. Parse JWT (if present)
    JWTFilter-->>Envoy: Return parsed x-jwt-user-id (if present)
    Envoy->>AuthServer: 3. Forward Request
    AuthServer->>AuthServer: 4. Authorize (if required)
    AuthServer-->>Envoy: 5. Auth Failed (if rejected)
    Envoy-->>Client: Reject Request (Auth Failed)
    AuthServer-->>Envoy: 6. Auth Success (if accepted)
    Envoy->>RateLimiter: 7. Set Rate Limit Labels (if required)
    RateLimiter->>RateLimiter: 8. Rate Limit Check
    RateLimiter-->>Envoy: Rate Check Failed
    Envoy-->>Client: Reject Request (Rate Limit Exceeded)
    RateLimiter-->>Envoy: Rate Check Passed
    Envoy->>Service: 9. Forward Request
    Service-->>Client: 10. Return Response
```

## 4. JSON Web Token (JWT) Verification

For GatewayEndpoints with the `RequireAuth` field set to `true`, a valid JWT issued by the auth provider specified in the `envoy.yaml` file is required to access the PATH service.

_Example Request Header:_
```bash
-H "Authorization: Bearer <JWT>"
```

The `jwt_authn` filter will verify the JWT and, if valid, set the `x-jwt-user-id` header from the `sub` claim of the JWT. An invalid JWT will result in an error. 

The `ext_authz` filter will use the `x-jwt-user-id` header to make an authorization decision; if the `GatewayEndpoint`'s `Auth.AuthorizedUsers` field contains the `x-jwt-user-id` value, the request will be authorized.

_Example auth provider user ID header:_
```
x-jwt-user-id: auth0|a12b3c4d5e6f7g8h9
```

For GatewayEndpoints with the `RequireAuth` field set to `false`, no JWT is required to access the PATH service. The request may be sent to the PATH Envoy Proxy without the `Authorization` header set. The `jwt_authn` filter will forward the request without setting the `x-jwt-user-id` header.

For more information, see:
- [Envoy JWT Authn Docs](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter)

## 5. External Authorization Server

The `envoy/auth_server` directory contains the Go/gRPC server responsible for authorizing requests forwarded by Envoy Proxy. It evaluates whether incoming requests are authorized to access the PATH service.

This server communicates with a remote gRPC server to populate its in-memory `Gateway Endpoint Store`, which provides data on which endpoints are authorized to use the PATH service.

For more information, see:
- [Envoy External Authorization Docs](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_authz_filter)

### 5.1. Remote gRPC Server

**The implementation of the remote gRPC server is up to the Gateway operator. **

A default Docker image is provided to handle live-loading of data from a `gateway-endpoints.yaml` file for simple use cases or quick startup of PATH.

```bash
docker pull buildwithgrove/path-auth-grpc-server:latest
```

#### 5.1.1. Example Gateway Endpoint Data File

An example `gateway-endpoints.yaml` file is provided at [envoy/gateway-endpoints.example.yaml](./gateway-endpoints.example.yaml).

```yaml
endpoints:
  endpoint_1:
    endpoint_id: "endpoint_1"
    auth:
      require_auth: true
      authorized_users:
        "auth0|user_1": {}
    user_account:
      account_id: "account_1"
      plan_type: "PLAN_FREE"
    rate_limiting:
      throughput_limit: 30
      capacity_limit: 100
      capacity_limit_period: "CAPACITY_LIMIT_PERIOD_DAILY"
  endpoint_2:
    endpoint_id: "endpoint_2"
    auth:
      require_auth: false
    user_account:
      account_id: "account_2"
      plan_type: "PLAN_UNLIMITED"
    rate_limiting:
      throughput_limit: 50
      capacity_limit: 200
      capacity_limit_period: "CAPACITY_LIMIT_PERIOD_MONTHLY"
```

_In this example, `endpoint_1` is authorized for `user_1` and `endpoint_2` is authorized for all users._

Run `make copy_envoy_gateway_endpoints` to create an example `gateway-endpoints.yaml` file used by the remote gRPC server.

The contents of this file represent the gateway endpoints that are authorized to use the PATH service for a specific gateway operator.

### 5.2. Environment Variables

The external authorization server requires the following environment variables to be set:

- `GRPC_HOST_PORT`: The host and port of the remote gRPC server.
- `GRPC_USE_INSECURE`: Set to `true` if the remote gRPC server does not use TLS (default: `false`).

Run `make copy_envoy_env` to create the `.env` file needed to run the external authorization server locally in Docker.

## 6. Rate Limiter

Rate limiting is configured through the [`/envoy/ratelimit.yaml`](./ratelimit.yaml) file. 

The default throughput limit is 30 requests per second for GatewayEndpoints with the `PLAN_FREE` plan type.

For more advanced configuration options, refer to the Envoy documentation:

- [Envoy Proxy Rate Limit Docs](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/rate_limit_filter)

- [Envoy Rate Limit Github](https://github.com/envoyproxy/ratelimit)

## 7. Architecture

```mermaid
graph TD
    User([User])
    Envoy[Envoy Proxy]
    
    AUTH["Auth Server <br> "]
    AUTH_DECISION{Did Authorize Request?}
    PATH[PATH Service]
    
    Error[[Error Returned to User]]
    Result[[Result Returned to User]]

    GRPCServer["gRPC Remote Server  <br> NOT part of PATH <br> (Impl. up to Operator)"]
    GRPCDB[("Optional Database <br> (Stores User Metadata)")]
    GRPCConfig@{ shape: notch-rect, label: "Optional Config File <br> (Stores User Metadata)" }

    subgraph AUTH["Auth Server (ext_authz)"]
        GRPCClient["gRPC Client"]
         Cache@{ shape: odd, label: "Stores gRPC Server Data" }
    end

    User -->|1.Send Request| Envoy
    Envoy -->|2.Authenticate  Request| AUTH
    AUTH -->|3.Authentication Result| Envoy
    Envoy --> AUTH_DECISION
    AUTH_DECISION -->|4.No <br> Forward Request| Error
    AUTH_DECISION -->|4.Yes <br> Forward Request| PATH
    PATH -->|5.Response| Result
    
    GRPCServer <-.-> |Retrieve User <> Endpoint <br> Data over gRPC| AUTH    
    GRPCServer <-.->GRPCDB
    GRPCServer <-.->GRPCConfig
```
