---
sidebar_position: 3
title: E2E Tests - Quickstart
description: End-to-End Tests for PATH
---

:::warning TODO

Add a gif of load tests running locally.

:::

_tl;dr Fully featured E2E Tests to verify PATH works correctly._

- [Quickstart](#quickstart)
- [E2E Test Config Files](#e2e-test-config-files)
- [Supported Services in E2E Tests](#supported-services-in-e2e-tests)

## Quickstart

**Prerequisites**: Complete the [Quick Start](1_quick_start.md) and [Shannon Cheat Sheet](2_cheatsheet_shannon.md) guides.

Then, run E2E tests for specific service IDs:

```bash
make e2e_test eth,xrplevm
```

Or, run E2E tests for all service IDs:

```bash
make e2e_test_all
```

## E2E Test Config Files

| Configuration File                        | Required? |             Default available?              | Description                            | Helper Creation Command           |
| ----------------------------------------- | :-------: | :-----------------------------------------: | :------------------------------------- | :-------------------------------- |
| `./e2e/config/.shannon.config.yaml`       |    ✅     |                     ❌                      | Gateway service configuration for PATH | `make config_shannon_populate`    |
| `./e2e/config/.e2e_load_test.config.yaml` |    ❌     | `e2e/config/e2e_load_test.config.tmpl.yaml` | Custom configuration for E2E tests     | `make config_prepare_shannon_e2e` |

## Supported Services in E2E Tests

To see the list of supported services for the tests, see the `test_cases` array in the [E2E Test Config](https://github.com/buildwithgrove/path/blob/main/e2e/config/e2e_load_test.config.tmpl.yaml) file.
