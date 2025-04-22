---
sidebar_position: 7
title: E2E Tests
description: End-to-End Tests for PATH
---

## Overview

PATH's End-to-End (E2E) tests are designed to validate the entire system's functionality in a real-world-like environment. These tests simulate actual user traffic, testing not just individual components, but the full request flow from client to services and back.

The E2E test suite primarily focuses on ensuring:
- PATH correctly routes requests to the appropriate services
- Services respond with expected data within acceptable latency thresholds
- The system maintains high availability and reliability under load
- Response metrics meet predefined success criteria

## Running Tests

:::important E2E 

As the E2E tests send actual relay requests, they require a valid configuration file that:
- is configured for the protocol being tested
- contains valid configuration for the services being tested

These files must be located in the `./e2e` directory and be named after the protocol:
- `.morse.config.yaml` for Morse
- `.shannon.config.yaml` for Shannon

For instructions on how to create these configuration files, see the [PATH Configuration File documentation](./5_configurations_path.md).

:::tip

Make targets are available for each protocol to copy example configuration files to the `./e2e` directory:

- `make morse_prepare_e2e_config`
- `make shannon_prepare_e2e_config`

Once copied, you will need to update the configuration with valid values before running the tests.

:::

### Make Targets

Tests can be executed using Make targets defined in the project:

```bash
# Run all tests (unit and E2E)
make test_all

# Run only E2E tests for Shannon
make test_e2e_evm_shannon

# Run only E2E tests for Morse
make test_e2e_evm_morse

# Run E2E tests for a specific service
make test_e2e_evm_morse SERVICE_ID_OVERRIDE=F021

# Run E2E tests against a PATH binary running locally without Docker
make test_e2e_evm_morse GATEWAY_URL_OVERRIDE=http://localhost:3069/v1

# Force Docker rebuild
make test_e2e_evm_morse DOCKER_FORCE_REBUILD=true

# Enable Docker logs
make test_e2e_evm_morse DOCKER_LOG=true

# Wait 30 seconds for hydrator checks
make test_e2e_evm_morse WAIT_FOR_HYDRATOR=30
```

### Vegeta Load Testing

<div align="center">
![Vegeta](../../../static/img/9000.png)
</div>

PATH's E2E tests utilize [Vegeta](https://github.com/tsenart/vegeta), a powerful HTTP load testing tool and library. Vegeta was chosen for its:

- High-performance capabilities (can generate thousands of requests per second)
- Detailed metrics collection and reporting
- Support for custom targeting and attack configurations
- Ability to measure precise latency distributions (p50, p95, p99)

### CI Workflows

The tests are configured to run in the Github Actions workflows for the `path` repository on push to the `main` branch.

## Available Services

The E2E tests cover multiple blockchain services. The exact services available for testing depend on the protocol configuration being used.

| Protocol | Service ID | Chain Name       | Type      |
| -------- | ---------- | ---------------- | --------- |
| Morse    | F00C       | Ethereum         | Archival  |
| Morse    | F021       | Polygon          | Archival  |
| Morse    | F01C       | Oasys            | Archival  |
| Morse    | F036       | XRPL EVM Testnet | Archival  |
| Shannon  | anvil      | Local Ethereum   | Ephemeral |

## Environment Variables

The E2E tests can be configured through several environment variables:

| Variable             | Description                                                                                                                                                                                                | Default                  | Required |
| -------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------ | -------- |
| GATEWAY_URL_OVERRIDE | Custom PATH gateway URL - useful for testing a locally running PATH instance on during development. If set the test Docker image will not be used and the test will run against the provided URL directly. | http://localhost:3069/v1 | No       |
| DOCKER_LOG           | Whether to log Docker container output                                                                                                                                                                     | false                    | No       |
| DOCKER_FORCE_REBUILD | Force rebuild of Docker image. By default the Docker image is built only if not found locally. However, a rebuild may be forced with this flag. Useful if testing local development changes                | false                    | No       |
| SERVICE_ID_OVERRIDE  | Test only a specific service ID                                                                                                                                                                            | All services             | No       |
| WAIT_FOR_HYDRATOR    | Seconds to wait for hydrator checks                                                                                                                                                                        | 0                        | No       |

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

