---
sidebar_position: 7
title: E2E Regression & Performance Tests
description: End-to-End Tests for PATH
---

## Overview <!-- omit in toc -->

PATH E2E (End-to-End) tests check if the whole system works as expected, simulating real user traffic.

These tests check:

- Correct request routing
- Service responses (data + latency)
- System reliability under load
- Success metrics

## Table of Contents <!-- omit in toc -->

- [E2E Test Configuration](#e2e-test-configuration)
- [Helper Make Targets](#helper-make-targets)
  - [Vegeta Load Testing](#vegeta-load-testing)
  - [CI Integration](#ci-integration)
- [Supported Services](#supported-services)
- [Environment Variables](#environment-variables)
- [How to Add/Update Tests](#how-to-addupdate-tests)
- [Test Metrics and Validation](#test-metrics-and-validation)
- [Extending Tests](#extending-tests)

## E2E Test Configuration

E2E tests need a valid config file in `./e2e`:

- `./e2e/morse.config.yaml` for Morse
- `./e2e/shannon.config.yaml` for Shannon

Config must match the protocol/services you want to test.

- See [PATH Configuration File docs](./5_configurations_path.md) for details.

:::tip

You can use the following commands to copy example configs and follow the instructions in your CLI:

- `make morse_prepare_e2e_config`
- `make shannon_prepare_e2e_config`

:::

## Helper Make Targets

```bash
# Run all tests (unit + E2E)
make test_all

# Only E2E tests for Shannon
make test_e2e_evm_shannon

# Only E2E tests for Morse
make test_e2e_evm_morse

# E2E for a specific service
make test_e2e_evm_morse SERVICE_ID_OVERRIDE=F021

# E2E against local PATH binary (no Docker)
make test_e2e_evm_morse GATEWAY_URL_OVERRIDE=http://localhost:3069/v1

# Force Docker rebuild
make test_e2e_evm_morse DOCKER_FORCE_REBUILD=true

# Enable Docker logs
make test_e2e_evm_morse DOCKER_LOG=true

# Wait 30s for hydrator checks
make test_e2e_evm_morse WAIT_FOR_HYDRATOR=30
```

:::tip make help

Run `make help` to see all available make targets.

:::

### Vegeta Load Testing

<div align="center">
![Vegeta](../../../static/img/9000.png)
</div>

PATH's E2E tests utilize [Vegeta](https://github.com/tsenart/vegeta) for HTTP load testing, which can:

- Generate thousands of requests/sec
- Collect detailed metrics
- Support custom configs and attack configurations
- Measure latency (p50, p95, p99)

---

### CI Integration

- E2E tests run automatically in GitHub Actions on every push to `main`.

---

## Supported Services

| Protocol | Service ID | Chain Name       | Type      |
| -------- | ---------- | ---------------- | --------- |
| Morse    | F00C       | Ethereum         | Archival  |
| Morse    | F021       | Polygon          | Archival  |
| Morse    | F01C       | Oasys            | Archival  |
| Morse    | F036       | XRPL EVM Testnet | Archival  |
| Shannon  | anvil      | Local Ethereum   | Ephemeral |

---

## Environment Variables

| Variable             | Description                                                                                           | Default                  | Required |
| -------------------- | ----------------------------------------------------------------------------------------------------- | ------------------------ | -------- |
| GATEWAY_URL_OVERRIDE | Custom PATH gateway URL (useful for local dev). If set, skips Docker and runs tests against this URL. | http://localhost:3069/v1 | No       |
| DOCKER_LOG           | Log Docker container output.                                                                          | false                    | No       |
| DOCKER_FORCE_REBUILD | Force Docker image rebuild (useful after code changes).                                               | false                    | No       |
| SERVICE_ID_OVERRIDE  | Test only a specific service ID.                                                                      | All services             | No       |
| WAIT_FOR_HYDRATOR    | Seconds to wait for hydrator checks.                                                                  | 0                        | No       |

---

## How to Add/Update Tests

1. Add new service definitions in `evm_test.go`
2. (If needed) Add new methods in `evm_methods_test.go`
3. Set up thresholds/latency expectations

- For more details, check the `e2e` directory source code.

## Test Metrics and Validation

The E2E tests collect and validate various metrics:

- HTTP success rates (percentage of successful requests)
- Response latency percentiles (p50, p95, p99)
- JSON-RPC response validation
- Error rates and types

Test results include detailed output showing performance against predefined thresholds, with colored indicators for passing or failing metrics.

## Extending Tests

To add new services or methods to the E2E tests:

1. Add new service definitions to the appropriate test case array in `evm_test.go`
2. If needed, add new method definitions in `evm_methods_test.go`
3. Configure appropriate success thresholds and latency expectations

For more details on the implementation, refer to the source code in the `e2e` directory.

TODO_IN_THIS_PR:

- [ ] Add links to relevant parts of the code in the last few sections
- [ ] Embed video as an example
- [ ] Add an act trigger to run the CI locally
