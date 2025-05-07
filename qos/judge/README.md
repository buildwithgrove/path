# JUDGE: JSON-RPC Unified Decentralized Gateway Evaluator

**TLDR:** Build a QoS system in minutes.

## Quick Start: Fully Functional QoS System in 2 Minutes

```go
// Create a QoS specification with probes for each method
spec := toolkit.QoSSpec{
    ServiceName: "EVM",
    ServiceProbes: map[jsonrpc.Method]toolkit.ServiceProbe{
        // Chain ID must be exactly 0x01 for Ethereum mainnet
        "eth_chainId": toolkit.NewServiceProbe(
            toolkit.ExactValue("0x01"),
            toolkit.WithSanctionOnFail(15 * time.Minute),
        ),
        
        // Block number should be the highest seen (most recent)
        "eth_blockNumber": toolkit.NewServiceProbe(
            toolkit.AggregatedValue(toolkit.AggregationStrategyMax),
        ),
    },
}

// That's it! Just instantiate and use
qosService := spec.NewQoSService(logger)
```

## What You Need

Building a QoS system with JUDGE takes just three steps:

1. **Identify key methods** your service needs (e.g., `eth_chainId`, `eth_blockNumber`)

2. **Choose ready-to-use ServiceProbes** for each method:
   - `ExactValue` for responses that must match exactly
   - `AggregatedValue` for values that should follow consensus
   - Add options like `WithSanctionOnFail(duration)` as needed

3. **Instantiate your QoS service** and you're done!

That's it! JUDGE handles all the complexity behind the scenes - request parsing, endpoint management, state tracking, and error handling.

## What You Get

A QoS service built with JUDGE provides:

- **Automatic Endpoint Quality Control**: Bad endpoints are automatically detected and excluded
- **Smart Request Routing**: Requests go to the best available endpoints
- **Service Consistency**: Users get consistent responses despite variable endpoint quality
- **Resilient Service Layer**: Handles edge cases like timeouts and empty responses
- **Rich Monitoring Data**: Generates structured metrics for analytics
- **Zero Maintenance**: Self-healing service that adapts to changing conditions

## Examples

See [EXAMPLES.md](EXAMPLES.md) for complete implementations.

## Behind the Scenes

JUDGE is a framework for building Quality of Service (QoS) systems for JSON-RPC services. It handles the complex infrastructure so you can focus on service-specific logic. Perfect for decentralized services where endpoint quality varies.

For more details on the framework architecture, see [ARCHITECTURE.md](ARCHITECTURE.md).

## Advanced Customization

For complex scenarios requiring fine-grained control, you can implement the ServiceProbe interface directly. 

JUDGE requires four simple implementations:

1. **Result Building**
   - Extract values from endpoint responses
   - Apply sanctions for endpoint failures

2. **State Update**
   - Update service parameters based on endpoint data
   - Example: Set consensus block height

3. **Endpoint Selection**
   - Define qualification criteria for endpoints
   - Example: Select nodes within sync threshold

4. **Quality Checks**
   - Specify verification requests for endpoints
   - Example: Check blockchain sync status

This gives you complete control over how your QoS system validates, tracks, and selects endpoints while still leveraging JUDGE's powerful infrastructure.

## License

[License Information]
