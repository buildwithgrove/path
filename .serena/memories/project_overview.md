# PATH Project Overview

## Purpose
PATH (Path API & Toolkit Harness) is an open-source Go framework for enabling access to a decentralized supply network. It serves as a gateway that handles service requests and relays them through the Shannon protocol to blockchain endpoints.

## Tech Stack
- **Language**: Go
- **Protocol**: Shannon (gRPC-based communication protocol)
- **Metrics**: Prometheus with custom metrics collection
- **Development**: Kubernetes/Tilt for local development
- **Configuration**: YAML-based configuration files
- **Logging**: polylog with structured logging
- **Testing**: Standard Go testing with E2E test suites

## Architecture
- **Gateway** (`gateway/`) - Main entry point for HTTP requests
- **Protocol** (`protocol/`) - Shannon protocol implementation
- **QoS** (`qos/`) - Quality of Service for different blockchain services
- **Router** (`router/`) - HTTP routing and API endpoint management
- **Config** (`config/`) - Configuration management
- **Metrics** (`metrics/`) - Prometheus metrics collection and reporting

## Key Features
- Shannon protocol for decentralized network communication
- QoS implementations for EVM, Solana, CometBFT
- Comprehensive metrics and observability
- Load balancing and endpoint management
- Rate limiting and authentication

## Note on Morse
While the codebase contains references to Morse protocol for backward compatibility, it is no longer actively used. Shannon is the primary protocol for all operations.