---
sidebar_position: 2
title: Endpoint Configuration
description: Gateway Endpoint Protobuf Definition
---

## Gateway Endpoint Protobuf Definition

:::tip

See the [Envoy Section](../envoy/walkthrough.md#external-auth-server) of the PATH documentation for complete details.

:::

A `GatewayEndpoint` represents a single endpoint managed by the Gateway. It can be configured to be public or authorized to support one or more user accounts.

The complete protobuf definitions can be found in the [`gateway_endpoint.proto`](https://github.com/buildwithgrove/path/blob/main/envoy/auth_server/proto/gateway_endpoint.proto) file.

### Core Fields

| Field           | Type         | Required | Default | Description                                                               |
| --------------- | ------------ | -------- | ------- | ------------------------------------------------------------------------- |
| `endpoint_id`   | string       | Yes      | -       | Unique identifier used in the request URL path (e.g. `/v1/{endpoint_id}`) |
| `auth`          | Auth         | Yes      | -       | Authorization configuration for the endpoint                              |
| `rate_limiting` | RateLimiting | No       | -       | Rate limit settings for request throughput and capacity                   |
| `metadata`      | Metadata     | No       | -       | Optional fields for billing, metrics and observability                    |

### Auth Types

| Field            | Type         | Required | Default | Description                                |
| ---------------- | ------------ | -------- | ------- | ------------------------------------------ |
| `no_auth`        | NoAuth       | No       | -       | Endpoint requires no authorization         |
| `static_api_key` | StaticAPIKey | No       | -       | Uses a `Static API` key for auth           |
| `jwt`            | JWT          | No       | -       | Uses `JWT` with map of authorized user IDs |

### Rate Limiting Configuration

| Field                   | Type  | Required | Default     | Description                                      |
| ----------------------- | ----- | -------- | ----------- | ------------------------------------------------ |
| `throughput_limit`      | int32 | No       | 0           | Requests per second (TPS) limit                  |
| `capacity_limit`        | int32 | No       | 0           | Total request capacity limit                     |
| `capacity_limit_period` | enum  | No       | UNSPECIFIED | Period for capacity limit (DAILY/WEEKLY/MONTHLY) |

### Metadata Fields

| Field         | Type   | Required | Default | Description                                                          |
| ------------- | ------ | -------- | ------- | -------------------------------------------------------------------- |
| `name`        | string | No       | ""      | Name of the endpoint                                                 |
| `account_id`  | string | No       | ""      | User account identifier                                              |
| `user_id`     | string | No       | ""      | Specific user identifier                                             |
| `plan_type`   | string | No       | ""      | Subscription plan (e.g. `Free`, `Pro`, `Enterprise`)                 |
| `email`       | string | No       | ""      | Associated email address of `account_id` owner                       |
| `environment` | string | No       | ""      | Deployment environment (e.g. `development`, `staging`, `production`) |
