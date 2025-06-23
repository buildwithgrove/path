---
sidebar_position: 4
title: Load Tests
description: Load Tests for PATH
---
## Load Testing Local PATH (Recommended for most users)

**Prerequisites**: Complete the [Quick Start](1_quick_start.md) and [Shannon Cheat Sheet](2_cheatsheet_shannon.md) guides.

```bash
# Shannon load tests with specified service IDs only
make load_test eth,anvil

# Shannon load tests with all service IDs
make load_test_all
```

<!-- TODO_UPNEXT(@adshmh): Add screenshot/video of running load tests locally -->

---

<br/>

:::warning Production Grove Portal Testing Only

**🛑 STOP HERE if you only need local PATH testing!**

The remainder of this document is only relevant if you intend to load test the Grove Portal in production. If you're only testing your local PATH instance, the commands above are sufficient.

:::

<br/>

## Load Testing the Grove Portal
:::warning **Prerequisites**: Complete the [Quick Start](1_quick_start.md) guide.
:::

:::info 🏢 **Grove Employees**

You can obtain the required Portal Application ID and API Key from the [1Password link](1Password_link_to_grove_portal_credentials).

:::

:::info 🔑 Grove Portal Credentials

A Grove Portal Application ID and API Key are required to run load tests against the Grove Portal. If you do not have these, you can get them by visiting the [Grove Portal](https://www.portal.grove.city).

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