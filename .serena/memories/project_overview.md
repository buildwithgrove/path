# PATH Project Overview

## Purpose
PATH (Path API & Toolkit Harness) is an open-source Go framework for enabling access to a decentralized supply network. It serves as a gateway that handles service requests and relays them through different protocols (Shannon and Morse) to blockchain endpoints.

## Tech Stack
- **Language**: Go
- **Protocols**: Shannon (main), Morse (legacy, being phased out)
- **Metrics**: Prometheus with custom metrics collection
- **Development**: Kubernetes/Tilt for local development
- **Configuration**: YAML-based configuration files
- **Logging**: polylog with structured logging
- **Testing**: Standard Go testing with E2E test suites

## Architecture
- **Gateway** (`gateway/`) - Main entry point for HTTP requests
- **Protocol** (`protocol/`) - Protocol implementations (Shannon/Morse)
- **QoS** (`qos/`) - Quality of Service for different blockchain services
- **Router** (`router/`) - HTTP routing and API endpoint management
- **Config** (`config/`) - Configuration management
- **Metrics** (`metrics/`) - Prometheus metrics collection and reporting

## Key Features
- Multi-protocol support (Shannon and Morse)
- QoS implementations for EVM, Solana, CometBFT
- Comprehensive metrics and observability
- Load balancing and endpoint management
- Rate limiting and authentication