---
sidebar_position: 1
title: Introduction
---

<div align="center">
<h1>PADS<br/>PATH Auth Data Server</h1>
<img src="https://storage.googleapis.com/grove-brand-assets/Presskit/Logo%20Joined-2.png" alt="Grove logo" width="500"/>

</div>
<br/>

![Static Badge](https://img.shields.io/badge/Maintained_by-Grove-green)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/buildwithgrove/path-auth-data-server/main-build.yml)
![GitHub last commit](https://img.shields.io/github/last-commit/buildwithgrove/path-auth-data-server)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/buildwithgrove/path-auth-data-server)
![GitHub Release](https://img.shields.io/github/v/release/buildwithgrove/path-auth-data-server)
![GitHub Issues or Pull Requests](https://img.shields.io/github/issues/buildwithgrove/path-auth-data-server)
![GitHub Issues or Pull Requests](https://img.shields.io/github/issues-pr/buildwithgrove/path-auth-data-server)
![GitHub Issues or Pull Requests](https://img.shields.io/github/issues-closed/buildwithgrove/path-auth-data-server)


# Table of Contents <!-- omit in toc -->

- [Introduction](#introduction)
- [Gateway Endpoints](#gateway-endpoints)
- [Data Sources](#data-sources)
- [gRPC Proto File](#grpc-proto-file)

## Introduction

<!-- TODO_MVP(@commoddity): Move these documents over to path.grove.city -->

**PADS** (PATH Auth Data Server) is a gRPC server that provides `Gateway Endpoint` data from a data source to the `External Auth Server` in order to enable authorization for [the PATH Gateway](https://github.com/buildwithgrove/path). The nature of the data source is configurable, for example it could be a YAML file or a Postgres database.

## Gateway Endpoints

A `GatewayEndpoint` represents a single authorized endpoint of the PATH Gateway service, which may be authorized for use by any number of users.

:::info

[See the Envoy Section of the PATH documentation for more details.](../envoy/introduction.md#external-auth-server)

:::

[This package also defines the `gateway_endpoint.proto` file](https://github.com/buildwithgrove/path/blob/main/envoy/auth_server/proto/gateway_endpoint.proto), which contains the definitions for the `GatewayEndpoints` that PADS must provides to the `Go External Authorization Server`.


```go
// Simplified representation of the GatewayEndpoint proto message that
// PADS must provide to the `Go External Authorization Server`.
type GatewayEndpoint struct {
    EndpointId string
    // AuthType will be one of the following structs:
    AuthType {
        // 1. No Authorization Required
        NoAuth struct{}
        // 2. Static API Key
        StaticApiKey struct {
          ApiKey string
        }
        // 3. JSON Web Token
        Jwt struct {
            AuthorizedUsers map[string]struct{}
        }
    }
    RateLimiting struct {
        ThroughputLimit int32
        CapacityLimit int32
        CapacityLimitPeriod CapacityLimitPeriod
    }
}
```

## Data Sources

The `server` package contains the `DataSource` interface, which abstracts the data source that provides GatewayEndpoints to the `Go External Authorization Server`.

```go
// AuthDataSource is an interface that abstracts the data source.
// It can be implemented by any data provider (e.g., YAML, Postgres).
type AuthDataSource interface {
   FetchAuthDataSync() (*proto.AuthDataResponse, error)
   AuthDataUpdatesChan() (<-chan *proto.AuthDataUpdate, error)
}

```

- `FetchAuthDataSync()` returns the full set of Gateway Endpoints.
  - This is called when `PADS` starts to populate its Gateway Endpoint Data Store.
- `AuthDataUpdatesChan()` returns a channel that receives auth data updates to the Gateway Endpoints.
  - Updates are streamed as changes are made to the data source.
  
## gRPC Proto File

[The PATH `auth_server` package](https://github.com/buildwithgrove/path/tree/main/envoy/auth_server) contains the file `gateway_endpoint.proto`, which contains:

- The gRPC auto-generated Go struct definitions for the GatewayEndpoints.
- The `FetchAuthDataSync` and `StreamAuthDataUpdates` methods that the `Go External Authorization Server` uses to populate and update its Gateway Endpoint Data Store.

:::info gRPC Proto File

[See the `gateway_endpoint.proto` documentation for more details.](../envoy/introduction.md#gateway_endpointproto-file) 

:::
