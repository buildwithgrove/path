---
sidebar_position: 1
title: Quick Introduction (1 min)
description: Quick intro & where to get started
---

<div align="center">
<h1>PATH API & Toolkit Harness</h1>
<img src="https://storage.googleapis.com/grove-brand-assets/Presskit/Logo%20Joined-2.png" alt="Grove logo" width="500"/>

</div>
<br/>

![Static Badge](https://img.shields.io/badge/Maintained_by-Grove-green)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/buildwithgrove/path/main-build.yml)
![GitHub last commit](https://img.shields.io/github/last-commit/buildwithgrove/path)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/buildwithgrove/path)
![GitHub Release](https://img.shields.io/github/v/release/buildwithgrove/path)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/buildwithgrove/path/total)
![GitHub Issues or Pull Requests](https://img.shields.io/github/issues/buildwithgrove/path)
![GitHub Issues or Pull Requests](https://img.shields.io/github/issues-pr/buildwithgrove/path)
![GitHub Issues or Pull Requests](https://img.shields.io/github/issues-closed/buildwithgrove/path)
![App Status](https://argocd.tooling.buildintheshade.com/api/badge?name=path-gateway&revision=true&showAppName=true)

**PATH (Path API & Toolkit Harness)** is an open source framework to enable access
to a permissionless network of API providers of open source services and data
sources without sacrificing enterprise-grade SLAs.

## Table of Contents <!-- omit in toc -->

- [Where to start?](#where-to-start)
- [Special Thanks](#special-thanks)
- [License](#license)

## Where to start?

If you're unsure of where to start, we recommend the following:

1. [**Environment Setup**](2_environment.md): Prepare your environment for running PATH
2. [**Configuration Files**](6_configurations_helm.md): Explore other PATH configuration files
3. [**Shannon Cheat Sheet**](3_cheatsheet_shannon.md): Get up and running with a Gateway to Shannon
4. [**Morse Cheat Sheet**](4_cheatsheet_morse.md): Get up and running with a Gateway to Morse if you're feeling adventurous

## Special Thanks

The origins of this repository were inspired by the work kicked off in [gateway-server](https://github.com/pokt-network/gateway-server) by the
[Nodies](https://nodies.app/) team. We were inspired and heavily considering forking and building off of that effort.

However, after a week-long sprint, the team deemed that starting from scratch was the better path forward for multiple reasons. These include but are not limited to:

- Enabling multi-protocol support; Morse, Shanon and beyond
- Set a foundation to migrate Grove's quality of service and data pipelineta
- Integrating with web2 standards like [Envoy](https://www.envoyproxy.io/), [gRPC](https://grpc.io/), [Stripe](https://stripe.com/), [NATS](https://nats.io/), [Auth0](https://auth0.com/), etc...
- Etc...

<!-- TODO(@olshansk): Move over the docs from [gateway-server](https://github.com/pokt-network/gateway-server) to a Morse section under [path.grove.city](https://path.grove.city) -->

---

## License

This project is licensed under the MIT License; see the [LICENSE](https://github.com/buildwithgrove/path/blob/main/LICENSE) file for details.
