---
sidebar_position: 5
title: Load Tests Quickstart
description: Load Tests for PATH
---

:::warning TODO

Add a gif of load tests running locally.

:::

_tl;dr Fully featured Load Tests to verify PATH works correctly._

- [Load Testing using Local PATH](#load-testing-using-local-path)
  - [WebSocket Load Tests](#websocket-load-tests)
- [Load Testing using Grove Portal](#load-testing-using-grove-portal)
- [Load Testing Grove Fallback Endpoints](#load-testing-grove-fallback-endpoints)

## Load Testing using Local PATH

⚠️ **Prerequisites**: Complete the [Quick Start](1_quick_start.md) and [Shannon Cheat Sheet](2_cheatsheet_shannon.md) guides.

First, prepare the Shannon E2E test config file:

```bash
make config_copy_path_local_config_shannon_e2e
```

Then, run HTTP load tests for specific service IDs:

```bash
make load_test eth,xrplevm
```

Or, run HTTP load tests for all service IDs:

```bash
make load_test_all
```

### WebSocket Load Tests

For services that support WebSocket connections (like XRPLEVM), you can run WebSocket-specific load tests separately:

```bash
# Run WebSocket load tests for specific service IDs
make load_test_websocket xrplevm,xrplevm-testnet

# Run WebSocket load tests for all WebSocket-compatible services  
make load_test_websocket_all
```

**Note:** WebSocket load tests are completely separate from HTTP tests. Use the regular `make load_test` commands for HTTP-only testing, and the `make load_test_websocket` commands for WebSocket-only testing.

## Load Testing using Grove Portal

:::danger Production Grove Portal Testing Only

**🛑 STOP HERE if you only need local PATH testing!**

The remainder of this document is only relevant if you intend to load test the Grove's Portal in production.

If you're only testing your local PATH instance, the commands above are sufficient.

:::

:::info 🌿 **Grove Employees**

You can obtain the required **Portal Application ID** and **API Key** from the [Grove Portal App Credentials for PATH Load Testing on 1Password](https://start.1password.com/open/i?a=4PU7ZENUCRCRTNSQWQ7PWCV2RM&v=kudw25ob4zcynmzmv2gv4qpkuq&i=iznzvqegxbl4y73d5lppm4y6r4&h=buildwithgrove.1password.com).

:::

**First, setup your configs & credentials**:

```bash
make config_copy_e2e_load_test
```

You will be prompted to enter your Grove Portal Application ID and API Key.

**Next, run load tests against Grove Portal**:

```bash
# HTTP load tests with specified service IDs only
make load_test eth,xrplevm

# HTTP load tests with all service IDs
make load_test_all

# WebSocket load tests with specified service IDs only
make load_test_websocket xrplevm,xrplevm-testnet

# WebSocket load tests with all WebSocket-compatible service IDs
make load_test_websocket_all
```

## Load Testing Grove Fallback Endpoints

:::info 🌿 **Grove Employees Only**

This section is exclusively for Grove employees who need to test PATH's fallback endpoint functionality.

You can obtain the required PATH config from the [Grove Portal App Credentials for PATH Load Testing on 1Password](https://start.1password.com/open/i?a=4PU7ZENUCRCRTNSQWQ7PWCV2RM&v=kudw25ob4zcynmzmv2gv4qpkuq&i=iznzvqegxbl4y73d5lppm4y6r4&h=buildwithgrove.1password.com).

:::

**First, download the PATH config from 1Password** (see the existing Grove Portal section above for credentials access).

**Next, enable fallback endpoints for all services**:

```bash
make config_enable_grove_fallback
```

**Then, restart PATH to apply the config updates**:

```bash
make path_down; make path_up
```

**Finally, run load tests**:

```bash
make load_test eth,xrplevm
```