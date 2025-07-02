<div align="center">
<h1>PATH<br/>Path API & Toolkit Harness</h1>
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

## Overview

**PATH** (Path API & Toolkit Harness) is an open source framework for enabling
access to a decentralized supply network.

It provides various tools and libraries to streamline the integration and
interaction with decentralized protocols.

## Documentation

Please visit [path.grove.city](https://path.grove.city) for documentation.

The source code for the documentation is available in the `docs` directory.

## Support

For Bug Reports and Enhancement Requests, please open an [Issue](https://github.com/buildwithgrove/path/issues).

For Technical Support please open a ticket in [Grove's Discord](https://discord.gg/build-with-grove).

---

## License

This project is licensed under the MIT License; see the [LICENSE](https://github.com/buildwithgrove/path/blob/main/LICENSE) file for details.

TODO_IN_THIS_PR: Find a good place for this documentation

```
Protocol-level: Unmarshal payload from endpoint/RelayMiner  into RelayResponse struct, and validate signature.

If the above succeeds: pass the Bytes field of RelayResponse (i.e. data from the backend service) to QoS.

QoS-level unmarshal: parse the backend service's payload into expected format (JSONRPC for most services for now).
```
