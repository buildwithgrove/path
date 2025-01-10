---
sidebar_position: 3
title: Interface Definition
description: Auth data source interface abstraction
---

## Auth Data Source Abstraction: Go interface to gRPC service

[**PADS**](https://github.com/buildwithgrove/path-auth-data-server/) defines an `AuthDataSource` interface in [`grpc/data_source.go`](https://github.com/buildwithgrove/path-auth-data-server/blob/main/grpc/data_source.go).

This interface is abstracted via the `gRPC Service` named `GatewayEndpoints` in [`gateway_endpoints.proto`](https://github.com/buildwithgrove/path-auth-data-server/blob/main/grpc/gateway_endpoints.proto).

Together, this is used to stream data from **Endpoint Auth Data Source** to the **Envoy Go External Auth Server** seen in the diagram above.

### PADS go Interface

```go
type AuthDataSource interface {
   FetchAuthDataSync() (*proto.AuthDataResponse, error)
   AuthDataUpdatesChan() (<-chan *proto.AuthDataUpdate, error)
}
```

| Function                | Returns                             | Details                                                               |
| ----------------------- | ----------------------------------- | --------------------------------------------------------------------- |
| `FetchAuthDataSync()`   | Full set of Gateway Endpoints       | Called when `PADS` starts to populate its Gateway Endpoint Data Store |
| `AuthDataUpdatesChan()` | Channel receiving auth data updates | Updates are streamed as changes are made to the data source           |

### PATH gRPC Interface

:::info gRPC Proto File Documentation

See the [`gateway_endpoint.proto` ](../envoy/walkthrough.md#gateway_endpointproto-file) documentation for complete details.

:::

```protobuf
service GatewayEndpoints {
  rpc FetchAuthDataSync(AuthDataRequest) returns (AuthDataResponse);
  rpc StreamAuthDataUpdates(AuthDataUpdatesRequest) returns (stream AuthDataUpdate);
}
```

| Method                  | Request                  | Response                | Description                                                          |
| ----------------------- | ------------------------ | ----------------------- | -------------------------------------------------------------------- |
| `FetchAuthDataSync`     | `AuthDataRequest`        | `AuthDataResponse`      | Fetches initial set of GatewayEndpoints from remote gRPC server      |
| `StreamAuthDataUpdates` | `AuthDataUpdatesRequest` | Stream `AuthDataUpdate` | Streams real-time updates of GatewayEndpoint changes from the server |
