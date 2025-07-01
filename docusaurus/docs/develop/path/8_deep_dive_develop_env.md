---
sidebar_position: 8
title: Deep Dive - Local Development Environment
description: Technical details and background information
---

# Developer Deep Dive

Technical details for developers who want to understand PATH's development environment.

## Development Environment Architecture

**PATH**'s development mode uses a Kubernetes-based local development environment with [Tilt](https://tilt.dev/).

We use [Kind](https://kind.sigs.k8s.io/) (Kubernetes in Docker) for running the local Kubernetes cluster, as it provides better compatibility with Tilt's image building process compared to Docker Desktop's Kubernetes cluster.

## Installed Tools

**Tools installed by `make install_tools`**:
- [**pocketd CLI**](https://dev.poktroll.com/category/pocketd-cli): CLI for interacting with Pocket Network's Shannon protocol
- [**Docker**](https://docs.docker.com/get-docker/): Container runtime
- [**Kind**](https://kind.sigs.k8s.io/#installation-and-usage): Local Kubernetes cluster
- [**kubectl**](https://kubernetes.io/docs/tasks/tools/#kubectl): CLI for interacting with Kubernetes
- [**Helm**](https://helm.sh/docs/intro/install/): Package manager for Kubernetes
- [**Tilt**](https://docs.tilt.dev/install.html): Local Kubernetes development environment

**Optional development tools** (`make install_optional_tools`):
- [**Relay Util**](https://github.com/commoddity/relay-util): Load testing tool for sending configurable batches of relays concurrently
- [**Graphviz**](https://graphviz.org): Required for generating profiling & debugging performance
- [**Uber Mockgen**](https://github.com/uber-go/mock): Mock interface generator for testing

## Protocol Support

| Protocol | Status (04/2025)       | Documentation                                     |
| -------- | ---------------------- | ------------------------------------------------- |
| Shannon  | Beta & Private MainNet | [Shannon Protocol Guide](2_cheatsheet_shannon.md) |
| Morse    | MainNet (deprecated)   | [Morse Protocol Guide](10_cheatsheet_morse.md)    |

## Architecture Overview

*This section will be expanded as more technical details are moved from other documentation files.*
