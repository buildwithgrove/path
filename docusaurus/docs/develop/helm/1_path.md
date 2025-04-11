---
sidebar_position: 1
title: PATH Helm Chart
description: PATH Helm Chart
---

<div align="center">
  <a href="https://helm.sh/docs/">
    <img src="https://www.redhat.com/rhdc/managed-files/helm.svg" alt="Helm logo" width="100"/>
  </a>
  <br/>
  <a href="https://helm.sh/docs/">
    <h2>Helm Docs</h2>
  </a>
</div>

A full PATH deployment is packaged as a Helm chart, with 3 main components:
- [PATH (PATH API and Tooling Harness)](1_path.md)
  - The Gateway component of PATH.
  - [GitHub Repository](https://github.com/buildwithgrove/helm-charts/charts/path)
- [GUARD (Gateway Utilities for Authentication, Routing & Defense)](2_guard.md)
  - The authentication, routing and security layer for the gateway built using Envoy Gateway.
  - [GitHub Repository](https://github.com/buildwithgrove/helm-charts/charts/guard)
- [WATCH (Workload Analytics and Telemetry for Comprehensive Health)](4_watch.md)
  - The observability layer for the gateway, including Prometheus, Grafana, and Alertmanager.
  - [GitHub Repository](https://github.com/buildwithgrove/helm-charts/charts/watch)

## Table of Contents <!-- omit in toc -->

- [PATH Helm Chart](#path-helm-chart)
- [Namespace \& RBAC Considerations](#namespace--rbac-considerations)
- [Integrating with WATCH](#integrating-with-watch)
- [Accessing Grafana](#accessing-grafana)

import RemoteMarkdown from '@site/src/components/RemoteMarkdown';

## PATH Helm Chart

<RemoteMarkdown src="https://raw.githubusercontent.com/buildwithgrove/helm-charts/refs/heads/main/charts/path/README.md" />

## Namespace & RBAC Considerations

<RemoteMarkdown src="https://raw.githubusercontent.com/buildwithgrove/helm-charts/refs/heads/main/charts/path/docs/namespace-rbac-considerations.md" />

## Integrating with WATCH

<RemoteMarkdown src="https://raw.githubusercontent.com/buildwithgrove/helm-charts/refs/heads/main/charts/path/docs/path-watch-integration-guide.md" />

## Accessing Grafana

<RemoteMarkdown src="https://raw.githubusercontent.com/buildwithgrove/helm-charts/refs/heads/main/charts/path/docs/path-accessing-grafana.md" />