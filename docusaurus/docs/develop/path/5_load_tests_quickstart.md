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
- [Load Testing using Grove Portal](#load-testing-using-grove-portal)

## Load Testing using Local PATH

**Prerequisites**: Complete the [Quick Start](1_quick_start.md) and [Shannon Cheat Sheet](2_cheatsheet_shannon.md) guides.

Then, run load tests for specific service IDs:

```bash
make load_test eth,xrplevm
```

Or, run load tests for all service IDs:

```bash
make load_test_all
```

## Load Testing using Grove Portal

:::danger Production Grove Portal Testing Only

**🛑 STOP HERE if you only need local PATH testing!**

The remainder of this document is only relevant if you intend to load test the Grove's Portal in production.

If you're only testing your local PATH instance, the commands above are sufficient.

:::

:::info 🌿 **Grove Employees**

You can obtain the required **Portal Application ID** and **API Key** from the [Grove Portal App Credentials for PATH Load Testing on 1Password](https://start.1password.com/open/i?a=4PU7ZENUCRCRTNSQWQ7PWCV2RM&v=kudw25ob4zcynmzmv2gv4qpkuq&i=iznzvqegxbl4y73d5lppm4y6r4&h=buildwithgrove.1password.com).

:::

1. **Setup credentials**

```bash
make copy_e2e_load_test_config
```

You will be prompted to enter your Grove Portal Application ID and API Key.

2. **Run load tests against Grove Portal**

```bash
# Shannon load tests with specified service IDs only
make load_test eth,anvil

# Shannon load tests with all service IDs
make load_test_all
```
