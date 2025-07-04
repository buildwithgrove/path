# Session Rollover and Parallel Requests <!-- omit in toc -->

- [Overview](#overview)
- [Configuration](#configuration)
  - [Gateway Relay Configuration](#gateway-relay-configuration)
  - [Shannon Session Configuration](#shannon-session-configuration)
- [Parallel Request Feature](#parallel-request-feature)
  - [How It Works](#how-it-works)
  - [Benefits](#benefits)
  - [Configuration Options](#configuration-options)
- [Session Grace Period Handling](#session-grace-period-handling)
  - [Grace Period Logic](#grace-period-logic)
  - [Scale Down Factor](#scale-down-factor)
  - [Configuration Options](#configuration-options-1)
- [Monitoring and Metrics](#monitoring-and-metrics)
  - [Session Rollover Metrics](#session-rollover-metrics)
  - [Parallel Request Metrics](#parallel-request-metrics)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

PATH implements advanced session rollover handling and parallel request features to improve reliability and performance during Pocket Network session transitions. These features help maintain service availability and reduce latency during the critical moments when sessions are rolling over.

## Configuration

### Gateway Relay Configuration

Configure parallel request behavior at the gateway level:

```yaml
gateway_config:
  relay:
    # Number of parallel requests (1-10)
    max_parallel_requests: 4
    
    # Timeout for parallel requests
    parallel_request_timeout: 30s
    
    # Enable TLD diversity preference
    enable_endpoint_diversity: true
```

### Shannon Session Configuration

Configure session grace period handling in the Shannon protocol:

```yaml
shannon_config:
  full_node_config:
    session_config:
      # Grace period scale factor (0.0-1.0)
      grace_period_scale_down_factor: 0.8
```

## Parallel Request Feature

### How It Works

1. **Endpoint Selection**: PATH selects up to `max_parallel_requests` endpoints from available suppliers
2. **TLD Diversity**: When `enable_endpoint_diversity` is true, PATH prefers endpoints with different Top-Level Domains (TLDs)
3. **Parallel Execution**: Requests are sent to all selected endpoints simultaneously
4. **First Success Wins**: The first successful response is used, and other requests are cancelled
5. **Fallback Handling**: If all requests fail, the last error is returned

### Benefits

- **Reduced Latency**: Uses the fastest responding endpoint
- **Improved Reliability**: If one endpoint fails, others can still succeed
- **Better Resource Utilization**: Distributes load across multiple suppliers
- **Resilience**: Reduces impact of individual endpoint failures

### Configuration Options

| Setting | Default | Range | Description |
|---------|---------|-------|-------------|
| `max_parallel_requests` | 4 | 1-10 | Maximum parallel endpoints to query |
| `parallel_request_timeout` | 30s | 1s-300s | Timeout for parallel operations |
| `enable_endpoint_diversity` | true | boolean | Prefer diverse TLDs when selecting endpoints |

## Session Grace Period Handling

### Grace Period Logic

PATH implements sophisticated grace period handling to manage session transitions:

1. **Current Session Check**: Gets the current session for the service+app combination
2. **Grace Period Calculation**: Determines if we're within the grace period of the previous session
3. **Scale Factor Application**: Applies the configurable scale factor to aggressively use new sessions
4. **Previous Session Retrieval**: If within grace period, attempts to get the previous session
5. **Fallback Strategy**: Falls back to current session if previous session retrieval fails

### Scale Down Factor

The `grace_period_scale_down_factor` configuration (default: 0.8) reduces the effective grace period:

- **Original Grace Period**: 100 blocks (from onchain parameters)
- **Scaled Grace Period**: 80 blocks (100 × 0.8)
- **Purpose**: Start using new sessions sooner to reduce rollover window

### Configuration Options

| Setting | Default | Range | Description |
|---------|---------|-------|-------------|
| `grace_period_scale_down_factor` | 0.8 | 0.0-1.0 | Scale factor for grace period |

## Monitoring and Metrics

### Session Rollover Metrics

PATH exports comprehensive metrics for monitoring session behavior:

- `shannon_session_transitions_total`: Session transition events
- `shannon_session_grace_period_usage_total`: Grace period usage patterns  
- `shannon_session_operation_duration_seconds`: Session operation latencies

### Parallel Request Metrics

Monitor parallel request performance:

- `shannon_relay_latency_seconds`: End-to-end request latency
- `shannon_backend_service_latency_seconds`: Backend service response times
- `shannon_request_setup_latency_seconds`: Request setup overhead

## Best Practices

### Parallel Requests

- **Conservative Settings**: Start with default `max_parallel_requests: 4`
- **Monitor Backend Load**: Ensure suppliers can handle increased traffic
- **Enable Diversity**: Keep `enable_endpoint_diversity: true` for better resilience
- **Appropriate Timeouts**: Use 30s timeout for most use cases

### Session Configuration

- **Default Scale Factor**: Use `grace_period_scale_down_factor: 0.8` for most deployments
- **High-Frequency Rollover**: Reduce scale factor (e.g., 0.6) for services with frequent session changes
- **Conservative Rollover**: Increase scale factor (e.g., 0.9) for critical services requiring maximum stability

### Monitoring

- **Set up alerting** on session transition failures
- **Monitor grace period usage** patterns to optimize scale factor
- **Track parallel request success rates** to validate configuration

## Troubleshooting

### High Session Rollover Failures

**Symptoms**: Increased `shannon_session_transitions_total` with `transition_type="grace_period_fallback"`

**Solutions**:
- Check full node connectivity and performance
- Verify session cache configuration
- Consider increasing `grace_period_scale_down_factor`

### Parallel Request Timeouts

**Symptoms**: Requests timing out despite available endpoints

**Solutions**:
- Increase `parallel_request_timeout`
- Reduce `max_parallel_requests`
- Check backend service performance
- Verify network connectivity to suppliers

### TLD Diversity Issues

**Symptoms**: All selected endpoints using same TLD

**Solutions**:
- Verify endpoint addressing format
- Check supplier pool diversity
- Review endpoint URL extraction logic

### Performance Issues

**Symptoms**: High latency despite parallel requests

**Solutions**:
- Monitor backend service latency metrics
- Check session cache hit rates
- Verify proper configuration of both cache and session settings
- Consider adjusting parallel request count based on supplier pool size