---
sidebar_position: 7
title: Local Development Environment Setup
description: Technical details and background information
---

Technical details for developers who want to understand PATH's development environment.

## Development Environment Architecture

**PATH**'s development mode uses a Kubernetes-based local development environment with [Tilt](https://tilt.dev/).

We use [Kind](https://kind.sigs.k8s.io/) (Kubernetes in Docker) for running the local Kubernetes cluster, as it provides better compatibility with Tilt's image building process compared to Docker Desktop's Kubernetes cluster.

## Installed Tools

**Tools installed by `make install_tools`**:

- [**pocketd CLI**](https://dev.poktroll.com/category/pocketd-cli): CLI for interacting with Pocket Network's Shannon protocol
- [**Docker**](https://docs.docker.com/get-docker/): Container runtime

**Optional development tools** (`make install_optional_tools`):

- [**Websocket Load Test**](https://github.com/commoddity/websocket-load-test): Websocket load testing tool
- [**Relay Util**](https://github.com/commoddity/relay-util): Load testing tool for sending configurable batches of relays concurrently
- [**Graphviz**](https://graphviz.org): Required for generating profiling & debugging performance
- [**Uber Mockgen**](https://github.com/uber-go/mock): Mock interface generator for testing

## Protocol Support

| Protocol | Status (04/2025)       | Documentation                                     |
| -------- | ---------------------- | ------------------------------------------------- |
| Shannon  | Beta & Private MainNet | [Shannon Protocol Guide](2_cheatsheet_shannon.md) |

## Architecture Overview

_This section will be expanded as more technical details are moved from other documentation files._
