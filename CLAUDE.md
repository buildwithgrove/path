# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

PATH (Path API & Toolkit Harness) is an open-source Go framework for enabling access to a decentralized supply network. It serves as a gateway that handles service requests and relays them through different protocols (Shannon and Morse) to blockchain endpoints.

## Development Commands

### Building and Running

- `make path_build` - Build the PATH binary locally
- `make path_run` - Run PATH as a standalone binary (requires CONFIG_PATH)
- `make path_up` - Start local Tilt development environment with dependencies
- `make path_down` - Tear down local Tilt development environment

### Testing

- `make test_unit` - Run all unit tests (`go test ./... -short -count=1`)
- `make test_all` - Run unit tests plus E2E tests for key services
- `make e2e_test SERVICE_IDS` - Run E2E tests for specific Shannon service IDs (e.g., `make e2e_test eth,poly`)
- `make morse_e2e_test SERVICE_IDS` - Run E2E tests for specific Morse service IDs (e.g., `make morse_e2e_test F00C,F021`)
- `make load_test SERVICE_IDS` - Run Shannon load tests
- `make go_lint` - Run Go linters (`golangci-lint run --timeout 5m --build-tags test`)

### Configuration

- `make shannon_prepare_e2e_config` - Prepare Shannon E2E configuration
- `make morse_prepare_e2e_config` - Prepare Morse E2E configuration

## Architecture Overview

PATH operates as a multi-layered gateway system:

### Core Components

- **Gateway** (`gateway/`) - Main entry point that handles HTTP requests and coordinates request processing
- **Protocol** (`protocol/`) - Protocol implementations (Shannon and Morse) that manage endpoint communication
- **QoS** (`qos/`) - Quality of Service implementations for different blockchain services (EVM, Solana, CosmosSDK)
- **Router** (`router/`) - HTTP routing and API endpoint management
- **Config** (`config/`) - Configuration management for different protocol modes

### Protocol Implementations

- **Shannon** (`protocol/shannon/`) - Main protocol implementation with gRPC communication
- **Morse** (`protocol/morse/`) - Legacy protocol implementation (being phased out)

### QoS Services

- **EVM** (`qos/evm/`) - Ethereum-compatible blockchain QoS with archival data checks
- **Solana** (`qos/solana/`) - Solana blockchain QoS
- **CosmosSDK** (`qos/cosmos/`) - Cosmos SDK blockchain QoS with support for REST, CometBFT, and JSON-RPC
- **JSONRPC** (`qos/jsonrpc/`) - Generic JSON-RPC handling
- **NoOp** (`qos/noop/`) - Pass-through QoS for unsupported services

### Data Flow

1. HTTP requests arrive at the Gateway
2. Request Parser maps requests to appropriate QoS services
3. QoS services validate requests and select optimal endpoints
4. Protocol implementations relay requests to blockchain endpoints
5. Responses are processed through QoS validation
6. Metrics and observations are collected throughout the pipeline

### Configuration

PATH uses YAML configuration files that support both Shannon and Morse protocols. Configuration includes:

- Protocol-specific settings (gRPC endpoints, signing keys)
- Service definitions and endpoint mappings
- QoS parameters and validation rules
- Gateway routing and middleware settings

## Key Files and Directories

- `cmd/main.go` - Application entry point and initialization
- `config/config.go` - Configuration loading and management
- `gateway/gateway.go` - Main gateway implementation
- `protocol/protocol.go` - Protocol interface definitions
- `Makefile` - Build and development commands
- `makefiles/` - Modular Makefile components for different tasks
- `e2e/` - End-to-end tests and configuration
- `local/` - Local development configuration for Kubernetes/Tilt
- `proto/` - Protocol buffer definitions
- `observation/` - Generated protobuf code for metrics and observations

## Development Environment

PATH uses Tilt for local development with Kubernetes (kind). The development stack includes:

- PATH gateway
- Envoy Proxy for load balancing
- Prometheus for metrics
- Grafana for observability
- Rate limiting and authentication services

## Testing Strategy

- **Unit Tests** - Standard Go tests with `-short` flag
- **E2E Tests** - Full integration tests against live blockchain endpoints
- **Load Tests** - Performance testing using Vegeta load testing tool
- **Protocol Tests** - Separate test suites for Shannon and Morse protocols
