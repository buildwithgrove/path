---
sidebar_position: 5
title: Debugging and Profiling Tools
description: Advanced debugging tools for PATH performance and health analysis
---

## Table of Contents

- [Health Check](#health-check)
- [Disqualified Endpoints](#disqualified-endpoints)
- [Performance Profiling (pprof)](#performance-profiling-pprof)

**Prerequisites**: Complete the [Quick Start](1_quick_start.md) and [Shannon Cheat Sheet](2_cheatsheet_shannon.md) guides.

## Health Check

```bash
curl http://localhost:3070/healthz
```

<details>
<summary>Example response</summary>

```json
{
  "status": "ready",
  "imageTag": "development",
  "readyStates": {
    "endpoint-hydrator": true,
    "pokt-shannon": true
  },
  "configuredServiceIDs": [
    "arb-one", "arb-sepolia-testnet", "avax", "avax-dfk", "base", 
    "base-sepolia-testnet", "bera", "bitcoin", "blast", "boba", "bsc", 
    "celo", "eth", "eth-holesky-testnet", "eth-sepolia-testnet", "evmos", 
    "fantom", "fraxtal", "fuse", "gnosis", "harmony", "ink", "iotex", 
    "kaia", "kava", "linea", "mantle", "metis", "moonbeam", "moonriver", 
    "near", "oasys", "op", "op-sepolia-testnet", "opbnb", "osmosis", 
    "pocket", "poly", "poly-amoy-testnet", "poly-zkevm", "radix", 
    "scroll", "sei", "solana", "sonic", "sui", "taiko", 
    "taiko-hekla-testnet", "tron", "xrplevm", "xrplevm-testnet", 
    "zklink-nova", "zksync-era"
  ]
}
```

</details>

## Disqualified Endpoints

Check which RPC endpoints have been sanctioned at protocol level (e.g. malformed relay responses) or QoS level (e.g. empty responses):

```bash
# Check disqualified endpoints for specific service
make get_disqualified_endpoints SERVICE_ID=eth
make get_disqualified_endpoints SERVICE_ID=base
```

<details>
<summary>Manual curl</summary>

```bash
curl http://localhost:3070/disqualified_endpoints \
  -H "Authorization: test_debug_api_key" \
  -H "Target-Service-Id: eth" | jq
```

</details>

## Performance Profiling (pprof)

Access Go pprof endpoints for performance analysis:

```bash
make profile_cpu       # Capture CPU profile
make profile_memory    # Capture memory profile  
make profile_goroutines # Capture goroutine profile
make pprof_index       # View available endpoints
```

<details>
<summary>Manual curls</summary>

```bash
# CPU profiling (30 seconds)
curl http://localhost:3070/debug/pprof/profile?seconds=30 \
  -H "Authorization: test_debug_api_key" -o cpu.prof

# Memory profiling
curl http://localhost:3070/debug/pprof/heap \
  -H "Authorization: test_debug_api_key" -o mem.prof

# Goroutine profiling
curl http://localhost:3070/debug/pprof/goroutine \
  -H "Authorization: test_debug_api_key" -o goroutine.prof

# View all available profiles
curl http://localhost:3070/debug/pprof/ \
  -H "Authorization: test_debug_api_key"
```

</details>

:::info 🔑 **Debug API Key**

The debug endpoints use a separate API key (`test_debug_api_key`) defined in `local/secrets.yaml` for security isolation. This is automatically configured in local development.

:::