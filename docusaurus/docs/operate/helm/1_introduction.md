---
sidebar_position: 1
title: PATH Helm Introduction
description: PATH Helm Introduction
---

:::danger 🚧 WORK IN PROGRESS 🚧

This section is not ready for public consumption.

:::

<div align="center">
  <a href="https://helm.sh/docs/">
    <img src="https://www.redhat.com/rhdc/managed-files/helm.svg" alt="Helm logo" width="100"/>
  </a>
  <br/>
  <a href="https://helm.sh/docs/">
    <h2>Helm Docs</h2>
  </a>
</div>

## PATH Components in Helm Deployment

A full PATH deployment is packaged as a single Helm chart, with 3 main components.

| Component                                                             | Description                                                               | Repository                                                                                            |
| --------------------------------------------------------------------- | ------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| **PATH** (Path API and Tooling Harness)                               | The Gateway component that provides access to decentralized API providers | [buildwithgrove/helm-charts/charts/path](https://github.com/buildwithgrove/helm-charts/charts/path)   |
| **GUARD** (Gateway Utilities for Authentication, Routing & Defense)   | The authentication, routing and security layer built using Envoy Gateway  | [buildwithgrove/helm-charts/charts/guard](https://github.com/buildwithgrove/helm-charts/charts/guard) |
| **WATCH** (Workload Analytics and Telemetry for Comprehensive Health) | The observability layer including Prometheus, Grafana, and Alertmanager   | [buildwithgrove/helm-charts/charts/watch](https://github.com/buildwithgrove/helm-charts/charts/watch) |

These three components work together to provide a complete gateway solution for accessing decentralized services through protocols like Pocket Network.

## Resource Requirements

The following are bare minimum resource requirements for PATH and will
need to be adjusted based on the expected load.

| Resource  | Minimum | Recommended |
| --------- | ------- | ----------- |
| CPU Cores | 2       | 4           |
| RAM       | 4GB     | 8GB         |
| Storage   | 10GB    | 20GB        |
