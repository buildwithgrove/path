---
sidebar_position: 1
title: Introduction
description: High-level architecture overview
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

## Introduction

**PADS** (PATH Auth Data Server) is an opinionated implementation of an authorization data server for PATH users. It provides authorization data on a per endpoint basis from an external data source for [the PATH Gateway](https://github.com/buildwithgrove/path).

Gateway Operators who want to enable authorization for their services are encouraged to use it
as a starting point, but can implement their own as well.

The nature of the data source is configurable. For example it could be a **static YAML file** or a **Postgres database**.

```mermaid
graph TD
    PATH["PATH Gateway"]
    ENVOY["Envoy Proxy"]
    GRPCServer["PADS <br> (i.e. Remote gRPC Server)"]

    subgraph AUTH["Envoy Go External Auth Server"]
        GRPCClient["gRPC Client"]
        Cache["Gateway Endpoint Data Store"]
    end

    subgraph DataSource["Endpoint Auth<br>Data Source<br>"]
        GRPCConfig["YAML Config File"]
        GRPCDB[("Postgres<br>Database")]
    end

    GRPCServer <-. "Proxy per endpoint auth data <br> (i.e. Gateway Endpoint Data)" .-> DataSource
    GRPCServer <-. "Fetch & Stream<br>Gateway Endpoint Data<br> (over gRPC)" .-> AUTH

    AUTH <-. Authorize Requests .-> ENVOY
    ENVOY <-. Proxy Authorized Requests .-> PATH
```
