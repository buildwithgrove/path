---
sidebar_position: 1
title: E2E Tests Quickstart (< 10min)
description: End-to-End Tests for PATH
---

**_tl;dr Fully featured E2E Tests to verify PATH works correctly._**

:::warning TODO(@olshansky): Add gif

Add a gif of load tests running locally to show what and how this works

:::

- [Quickstart](#quickstart)
- [E2E Test Config Files](#e2e-test-config-files)
- [Supported Services in E2E Tests](#supported-services-in-e2e-tests)

## Quickstart

⚠️ **Prerequisites**: Complete the [Getting Started](../path/1_getting_started.md) and [Shannon Cheat Sheet](../path/2_cheatsheet_shannon.md) guides ⚠️

First, prepare the Shannon E2E test config file at `./e2e/config/.shannon.config.yaml` by running:

```bash
make config_copy_path_local_config_shannon_e2e
```

Then, run E2E tests for specific service IDs. For example:

```bash
make e2e_test eth,xrplevm
```

Alternatively, run E2E tests for all service IDs:

```bash
make e2e_test_all
```

## E2E Test Config Files

| Configuration File                        | Custom Config Required? |             Default available?              | Description                            | Command to create or customize                                                     |
| ----------------------------------------- | :---------------------: | :-----------------------------------------: | :------------------------------------- | :--------------------------------------------------------------------------------- |
| `./e2e/config/.shannon.config.yaml`       |           ✅            |                     ❌                      | Gateway service configuration for PATH | `make config_copy_path_local_config_shannon_e2e` OR `make config_shannon_populate` |
| `./e2e/config/.e2e_load_test.config.yaml` |           ❌            | `e2e/config/e2e_load_test.config.tmpl.yaml` | Custom configuration for E2E tests     | `make config_prepare_shannon_e2e`                                                  |

## Supported Services in E2E Tests

To see the list of supported services for the tests, see the `test_cases` array in the [E2E Test Config](https://github.com/buildwithgrove/path/blob/main/e2e/config/e2e_load_test.config.tmpl.yaml) file.
