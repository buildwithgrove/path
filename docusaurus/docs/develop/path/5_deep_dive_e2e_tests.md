---
sidebar_position: 5
title: Deep Dive - E2E Tests
description: Deep dive into End-to-End Tests for PATH
---

# Deep Dive: E2E Tests

<!-- TODO_UPNEXT(@adshmh): 
* Use Local Development Environment to run E2E tests
* Update this doc accordingly: e.g. on accessing PATH logs.
-->

## Overview

**The E2E tests verify:**

- Correct request routing
- Service responses (data + latency)
- System reliability under load
- Success metrics for Shannon protocol

**We use the [Vegeta library](https://github.com/tsenart/vegeta) for HTTP load testing:**

- Can generate thousands of requests/sec
- Collects detailed metrics including latency percentiles (p50, p95, p99)
- Supports custom configurations and attack parameters
- Validates JSON-RPC responses and success rates

<div align="center">
![Vegeta](../../../static/img/9000.png)
</div>

## E2E Test Mode

PATH E2E tests run in a single mode:

| Mode          | Make Targets                | Purpose                                                                  | How it Works                                                                                                                                                                             | Use Cases                                                                          |
| ------------- | --------------------------- | ------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| **E2E Test**  | `make e2e_test`             | Full end-to-end testing that starts PATH in an isolated Docker container | 1. Spins up PATH in a Docker container using Dockertest <br/> 2. Uses protocol config (`.shannon.config.yaml`) <br/> 3. Runs tests <br/> 4. Tears down container | - Full system validation <br/> - Continuous integration <br/> - Regression testing |

## E2E Test Config Files

E2E Test mode requires protocol-specific configuration because it spins up a local PATH instance.

| Configuration File                                 | E2E Test (Required?) |             Default available?              |
| -------------------------------------------------- | :------------------: | :-----------------------------------------: |
| `./e2e/config/.shannon.config.yaml` (for Shannon)  |          ✅           |                      ❌                      |
| `./e2e/config/.e2e_load_test.config.yaml` (custom) |          ❌           | `e2e/config/e2e_load_test.config.tmpl.yaml` |

:::tip Populate Configs

You can use the following command to copy example configs and follow the instructions in your CLI:

For E2E tests:

- `make shannon_prepare_e2e_config`

:::

## Schema and Validation

The configuration uses a formal YAML schema with validation:

**Schema Location**: `./e2e/config/e2e_load_test.config.schema.yaml`

:::tip VSCode Validation

If you are using VSCode, we recommend using the [YAML Language Support](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) extension for in-editor validation of the `.config.yaml` file.

Enable it by ensuring the following annotation is present at the top of your config file:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/buildwithgrove/path/refs/heads/main/e2e/config/e2e_load_test.config.schema.yaml
```

:::

## Supported Services in E2E Tests

**All currently supported Grove Portal services are supported in the E2E tests.**

:::tip

To see the list of supported services for the tests, see the `test_cases` array in the [E2E Test Config](https://github.com/buildwithgrove/path/blob/main/e2e/config/e2e_load_test.config.tmpl.yaml) file.

:::

## Environment Variables

These environment variables are set by the test make targets, but if you wish to set them manually, see the table below:

<details>
<summary>Env Vars Table</summary>
| Variable         | Description                                                                                       | Values                              | Required |
| ---------------- | ------------------------------------------------------------------------------------------------- | ----------------------------------- | -------- |
| TEST_MODE        | Determines the test execution mode                                                                | `e2e`                               | Yes      |
| TEST_PROTOCOL    | Specifies which protocol to test                                                                  | `shannon`                           | Yes      |
| TEST_SERVICE_IDS | Specifies which service IDs to test. If not set, all service IDs for the protocol will be tested. | Comma-separated list of service IDs | No       |
</details>

## Extending/Updating/Adding EVM E2E Tests

To add new services or methods to the E2E tests, you will need to open a new PR to PATH's `main` branch.

1. **Add new service definitions** to the `services` array in the `e2e/config/services_shannon.yaml` configuration file
2. **Configure service parameters** including contract addresses, start blocks, and transaction hashes for archival tests

**Example new service configuration:**

_`./config/services_shannon.yaml`_

```yaml
services:
  - name: "New Chain E2E Test"
    protocol: "shannon"
    service_id: "newchain"
    archival: true
    service_params:
      contract_address: "0x..."
      contract_start_block: 1000000
      transaction_hash: "0x..."
      call_data: "0x18160ddd"
```

## Test Metrics and Validation

The E2E tests collect and validate comprehensive metrics across multiple dimensions:

| **Category**              | **Metrics Collected**                                                                                                                                        |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **HTTP Metrics**          | - Success rates (HTTP 200) <br/> - Status code distribution <br/> - HTTP error categorization                                                                |
| **Latency Metrics**       | - P50, P95, P99 latency percentiles <br/> - Average latency <br/> - Per-method latency analysis                                                              |
| **JSON-RPC Validation**   | - Response unmarshaling success <br/> - JSON-RPC error field validation <br/> - Result field validation <br/> - Protocol-specific validation                 |
| **Service-Level Metrics** | - Per-service success aggregation <br/> - Cross-method performance comparison <br/> - Service reliability scoring <br/> - Error categorization and reporting |

:::important Threshold Validation

Tests will **fail** if any configured thresholds are exceeded, ensuring consistent service quality and performance.

:::

## Reviewing PATH Logs

In E2E test mode, logs may be written to `./path_log_e2e_test_{timestamp}.txt`.

**In order to enable this, set the log_to_file field:**

```bash
yq eval '.e2e_load_test_config.e2e_config.docker_config.log_to_file = true' -i ./e2e/config/.e2e_load_test.config.yaml
```

You should see the following log line at the bottom of the test summary:

```bash
===== 👀 LOGS 👀 =====

 ✍️ PATH container output logged to /tmp/path_log_e2e_test_1745527319.txt ✍️

===== 👀 LOGS 👀 =====

```

## Debugging Anvil on Shannon Beta TestNet

🌿 Grove Employees Only

Review the [Anvil Shannon Beta TestNet Debugging Playbook](https://www.notion.so/buildwithgrove/Playbook-Debugging-Anvil-E2E-on-Beta-TestNet-177a36edfff6809c9f24e865ec5adbf8?pvs=4) if you believe the Anvil Supplier is broken.
